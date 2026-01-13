package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/jira-webhook-api/handlers"
	"github.com/storos/sdlc-agent/jira-webhook-api/repositories"
	"github.com/storos/sdlc-agent/jira-webhook-api/services"
	"github.com/streadway/amqp"
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
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Get configuration from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017/sdlc_agents"
	}

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	// Connect to MongoDB
	logger.Info("Connecting to MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			logger.WithError(err).Error("Failed to disconnect from MongoDB")
		}
	}()

	// Ping MongoDB
	if err = mongoClient.Ping(ctx, nil); err != nil {
		logger.WithError(err).Fatal("Failed to ping MongoDB")
	}
	logger.Info("Connected to MongoDB")

	// Get database
	db := mongoClient.Database("sdlc_agents")

	// Connect to RabbitMQ with retry
	logger.Info("Connecting to RabbitMQ...")
	var rabbitConn *amqp.Connection
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		rabbitConn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			break
		}
		logger.WithError(err).Warnf("Failed to connect to RabbitMQ (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to RabbitMQ after retries")
	}
	defer rabbitConn.Close()
	logger.Info("Connected to RabbitMQ")

	// Initialize repository
	webhookRepo := repositories.NewWebhookRepository(db)

	// Initialize service
	webhookService, err := services.NewWebhookService(webhookRepo, rabbitConn, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize webhook service")
	}
	defer webhookService.Close()

	// Initialize handler
	webhookHandler := handlers.NewWebhookHandler(webhookService, logger)

	// Setup HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", webhookHandler.HealthCheck)

	// Webhook endpoint
	router.POST("/webhook", webhookHandler.HandleWebhook)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("port", port).Info("Starting JIRA Webhook API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
