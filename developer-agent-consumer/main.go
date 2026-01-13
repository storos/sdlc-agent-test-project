package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/storos/sdlc-agent/developer-agent-consumer/clients"
	"github.com/storos/sdlc-agent/developer-agent-consumer/consumer"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
	"github.com/storos/sdlc-agent/developer-agent-consumer/repositories"
	"github.com/storos/sdlc-agent/developer-agent-consumer/services"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Developer Agent Consumer")

	// Load configuration from environment
	mongoURL := getEnv("MONGODB_URL", "mongodb://localhost:27017/sdlc_agents")
	mongoDatabase := "sdlc_agents"
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	configAPIURL := getEnv("CONFIGURATION_API_URL", "http://localhost:8081")
	claudeAPIURL := getEnv("CLAUDE_API_URL", "http://localhost:8000/generate")
	claudeSessionToken := getEnv("CLAUDE_SESSION_TOKEN", "")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Ping MongoDB
	if err := mongoClient.Ping(ctx, nil); err != nil {
		logger.Fatalf("Failed to ping MongoDB: %v", err)
	}
	logger.Info("Connected to MongoDB")

	db := mongoClient.Database(mongoDatabase)

	// Initialize repositories
	devRepo := repositories.NewDevelopmentRepository(db)

	// Initialize services
	configClient := clients.NewConfigAPIClient(configAPIURL, logger)
	gitService := services.NewGitService(logger)
	analyzerService := services.NewAnalyzerService(logger)
	claudeService := services.NewClaudeService(claudeAPIURL, logger)
	prService := services.NewPRService(logger)

	// Create application context
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Create message handler
	handler := createMessageHandler(
		devRepo,
		configClient,
		gitService,
		analyzerService,
		claudeService,
		prService,
		claudeSessionToken,
		logger,
	)

	// Initialize RabbitMQ consumer
	rabbitConsumer, err := consumer.NewRabbitMQConsumer(rabbitMQURL, handler, logger)
	if err != nil {
		logger.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}
	defer rabbitConsumer.Close()

	// Start consumer
	if err := rabbitConsumer.Start(appCtx); err != nil {
		logger.Fatalf("Failed to start consumer: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")
	appCancel()

	// Wait for consumer to finish processing
	rabbitConsumer.Wait()

	logger.Info("Developer Agent Consumer stopped")
}

func createMessageHandler(
	devRepo *repositories.DevelopmentRepository,
	configClient *clients.ConfigAPIClient,
	gitService *services.GitService,
	analyzerService *services.AnalyzerService,
	claudeService *services.ClaudeService,
	prService *services.PRService,
	claudeSessionToken string,
	logger *logrus.Logger,
) consumer.MessageHandler {
	return func(ctx context.Context, request *models.DevelopmentRequest) error {
		logger.WithFields(logrus.Fields{
			"jira_issue_key":    request.JiraIssueKey,
			"jira_project_key":  request.JiraProjectKey,
		}).Info("Processing development request")

		// Create development record
		dev := &models.Development{
			JiraIssueID:    request.JiraIssueID,
			JiraIssueKey:   request.JiraIssueKey,
			JiraProjectKey: request.JiraProjectKey,
		}

		if err := devRepo.Create(ctx, dev); err != nil {
			logger.Errorf("Failed to create development record: %v", err)
			return err
		}

		logger.WithFields(logrus.Fields{
			"development_id": dev.ID.Hex(),
		}).Info("Development record created")

		// Step 1: Get project configuration
		logger.Info("Fetching project configuration")
		project, err := configClient.GetProjectByJiraKey(request.JiraProjectKey)
		if err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		// Step 2: Find repository in project
		var repository *models.Repository
		if request.Repository != "" {
			repository, err = configClient.FindRepositoryInProject(project, request.Repository)
			if err != nil {
				devRepo.MarkFailed(ctx, dev.ID, err.Error())
				return err
			}
		} else {
			// Use first repository if not specified
			if len(project.Repositories) == 0 {
				err := "no repositories configured for project"
				devRepo.MarkFailed(ctx, dev.ID, err)
				return &ErrNoRepositories{}
			}
			repository = &project.Repositories[0]
		}

		dev.RepositoryURL = repository.URL

		logger.WithFields(logrus.Fields{
			"repository_url": repository.URL,
		}).Info("Repository selected")

		// Step 3: Clone repository
		logger.Info("Cloning repository")
		workspace, err := gitService.CloneRepository(repository.URL, repository.GitAccessToken, request.JiraIssueKey)
		if err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}
		defer gitService.Cleanup(workspace)

		// Step 4: Analyze repository
		logger.Info("Analyzing repository structure")
		analysis, err := analyzerService.AnalyzeRepository(workspace.Path)
		if err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		// Step 5: Generate code with Claude
		logger.Info("Generating code with Claude Code")
		claudeResponse, err := claudeService.GenerateCode(request, project, analysis, workspace.Path, claudeSessionToken)
		if err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		// Step 6: Create branch and commit changes
		logger.Info("Creating feature branch")
		if err := gitService.CreateAndCheckoutBranch(workspace, request.JiraIssueKey); err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		dev.BranchName = workspace.BranchName

		logger.Info("Committing changes")
		if err := gitService.CommitChanges(workspace, request.JiraIssueKey, request.Summary); err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		logger.Info("Pushing branch")
		if err := gitService.PushBranch(workspace, repository.GitAccessToken); err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		// Step 7: Create PR/MR
		logger.Info("Creating pull/merge request")
		prURL, err := prService.CreatePullRequest(
			repository.URL,
			workspace.BranchName,
			request.JiraIssueKey,
			request.Summary,
			request.Description,
			repository.GitAccessToken,
		)
		if err != nil {
			devRepo.MarkFailed(ctx, dev.ID, err.Error())
			return err
		}

		// Step 8: Update development record
		logger.Info("Marking development as completed")
		if err := devRepo.MarkCompleted(ctx, dev.ID, prURL, claudeResponse.DevelopmentDetails); err != nil {
			logger.Errorf("Failed to mark as completed: %v", err)
			return err
		}

		logger.WithFields(logrus.Fields{
			"jira_issue_key": request.JiraIssueKey,
			"pr_url":         prURL,
			"files_changed":  claudeResponse.FilesChanged,
		}).Info("Development request processed successfully")

		return nil
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type ErrNoRepositories struct{}

func (e *ErrNoRepositories) Error() string {
	return "no repositories configured for project"
}
