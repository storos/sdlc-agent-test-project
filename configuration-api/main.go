package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/configuration-api/handlers"
	"github.com/storos/sdlc-agent/configuration-api/repositories"
	"github.com/storos/sdlc-agent/configuration-api/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.Info("Starting Configuration API...")

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		logger.Fatal("MONGODB_URL environment variable is required")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		logger.WithError(err).Fatal("Failed to ping MongoDB")
	}

	logger.Info("Connected to MongoDB successfully")

	// Initialize database and repositories
	db := client.Database("sdlc_agents")
	projectRepo := repositories.NewProjectRepository(db)
	developmentRepo := repositories.NewDevelopmentRepository(db)

	// Initialize services
	projectService := services.NewProjectService(projectRepo)
	developmentService := services.NewDevelopmentService(developmentRepo)

	// Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectService, logger)
	developmentHandler := handlers.NewDevelopmentHandler(developmentService, logger)

	// Setup Gin router
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Custom endpoint
		api.GET("/selimboyuk", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Project routes
		api.GET("/projects", projectHandler.GetProjects)
		api.GET("/projects/:id", projectHandler.GetProject)
		api.POST("/projects", projectHandler.CreateProject)
		api.PUT("/projects/:id", projectHandler.UpdateProject)
		api.DELETE("/projects/:id", projectHandler.DeleteProject)

		// Repository routes
		api.GET("/projects/:id/repositories", projectHandler.GetRepositories)
		api.POST("/projects/:id/repositories", projectHandler.AddRepository)
		api.PUT("/repositories/:id", projectHandler.UpdateRepository)
		api.DELETE("/repositories/:id", projectHandler.DeleteRepository)

		// Development routes
		api.GET("/developments", developmentHandler.GetDevelopments)
		api.GET("/developments/:id", developmentHandler.GetDevelopment)
	}

	// Start server in a goroutine
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		logger.Infof("Configuration API listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	// Disconnect from MongoDB
	if err := client.Disconnect(ctx); err != nil {
		logger.WithError(err).Error("Failed to disconnect from MongoDB")
	}

	logger.Info("Server exited")
}
