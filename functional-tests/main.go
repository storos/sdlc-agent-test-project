package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger *logrus.Logger

type TestResult struct {
	TestName string
	Passed   bool
	Message  string
	Duration time.Duration
}

type Project struct {
	ID              string       `json:"id,omitempty"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Scope           string       `json:"scope"`
	JiraProjectKey  string       `json:"jira_project_key"`
	JiraProjectName string       `json:"jira_project_name"`
	JiraProjectURL  string       `json:"jira_project_url"`
	Repositories    []Repository `json:"repositories"`
}

type Repository struct {
	RepositoryID   string `json:"repository_id"`
	URL            string `json:"url"`
	Description    string `json:"description"`
	GitAccessToken string `json:"git_access_token"`
}

type WebhookPayload struct {
	WebhookEvent     string `json:"webhookEvent"`
	IssueEventType   string `json:"issue_event_type_name"`
	Issue            Issue  `json:"issue"`
	Changelog        *Changelog `json:"changelog,omitempty"`
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Status      Status   `json:"status"`
	Project     Project2 `json:"project"`
}

type Status struct {
	Name string `json:"name"`
}

type Project2 struct {
	Key string `json:"key"`
}

type Changelog struct {
	Items []ChangelogItem `json:"items"`
}

type ChangelogItem struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}

func main() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("========================================")
	logger.Info("SDLC AI Agents - Functional Tests")
	logger.Info("========================================")

	// Get configuration from environment
	configAPIURL := getEnv("CONFIG_API_URL", "http://localhost:8081")
	webhookAPIURL := getEnv("WEBHOOK_API_URL", "http://localhost:8080")
	mongoURL := getEnv("MONGO_URL", "mongodb://localhost:27017")
	mongoDatabase := getEnv("MONGO_DATABASE", "sdlc_agent")

	// Connect to MongoDB for validation
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, nil); err != nil {
		logger.Fatalf("Failed to ping MongoDB: %v", err)
	}
	logger.Info("✓ Connected to MongoDB")

	db := mongoClient.Database(mongoDatabase)

	// Run tests
	results := []TestResult{}

	// Test 1: Create test project via Configuration API
	result := testCreateProject(configAPIURL)
	results = append(results, result)
	if !result.Passed {
		printResults(results)
		os.Exit(1)
	}

	// Test 2: Send JIRA webhook
	result = testSendWebhook(webhookAPIURL)
	results = append(results, result)
	if !result.Passed {
		printResults(results)
		os.Exit(1)
	}

	// Test 3: Validate webhook event in MongoDB
	result = testValidateWebhookEvent(db)
	results = append(results, result)
	if !result.Passed {
		printResults(results)
		os.Exit(1)
	}

	// Test 4: Wait for development completion (with timeout)
	result = testWaitForCompletion(db)
	results = append(results, result)
	if !result.Passed {
		printResults(results)
		os.Exit(1)
	}

	// Test 5: Validate development record
	result = testValidateDevelopment(db)
	results = append(results, result)
	results = append(results, result)

	// Test 6: Cleanup test data
	result = testCleanup(configAPIURL, db)
	results = append(results, result)

	// Print final results
	printResults(results)

	// Exit with appropriate code
	allPassed := true
	for _, r := range results {
		if !r.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		logger.Info("========================================")
		logger.Info("✓ ALL TESTS PASSED")
		logger.Info("========================================")
		os.Exit(0)
	} else {
		logger.Error("========================================")
		logger.Error("✗ SOME TESTS FAILED")
		logger.Error("========================================")
		os.Exit(1)
	}
}

var testProjectID string
var testIssueKey string

func testCreateProject(apiURL string) TestResult {
	start := time.Now()
	logger.Info("\n[Test 1] Creating test project via Configuration API...")

	testIssueKey = fmt.Sprintf("TEST-%s", uuid.New().String()[:8])

	project := Project{
		Name:            "Test Project",
		Description:     "Automated test project",
		Scope:           "Test scope for functional testing",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
		Repositories: []Repository{
			{
				RepositoryID:   "test-repo",
				URL:            "https://github.com/test/repo",
				Description:    "Test repository",
				GitAccessToken: "test-token",
			},
		},
	}

	jsonData, err := json.Marshal(project)
	if err != nil {
		return TestResult{
			TestName: "Create Project",
			Passed:   false,
			Message:  fmt.Sprintf("Failed to marshal project: %v", err),
			Duration: time.Since(start),
		}
	}

	resp, err := http.Post(apiURL+"/api/projects", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return TestResult{
			TestName: "Create Project",
			Passed:   false,
			Message:  fmt.Sprintf("Failed to create project: %v", err),
			Duration: time.Since(start),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return TestResult{
			TestName: "Create Project",
			Passed:   false,
			Message:  fmt.Sprintf("Expected status 201, got %d: %s", resp.StatusCode, string(body)),
			Duration: time.Since(start),
		}
	}

	var createdProject Project
	if err := json.NewDecoder(resp.Body).Decode(&createdProject); err != nil {
		return TestResult{
			TestName: "Create Project",
			Passed:   false,
			Message:  fmt.Sprintf("Failed to decode response: %v", err),
			Duration: time.Since(start),
		}
	}

	testProjectID = createdProject.ID
	logger.Infof("✓ Created test project with ID: %s", testProjectID)

	return TestResult{
		TestName: "Create Project",
		Passed:   true,
		Message:  fmt.Sprintf("Project created successfully (ID: %s)", testProjectID),
		Duration: time.Since(start),
	}
}

func testSendWebhook(apiURL string) TestResult {
	start := time.Now()
	logger.Info("\n[Test 2] Sending test JIRA webhook...")

	payload := WebhookPayload{
		WebhookEvent:   "jira:issue_updated",
		IssueEventType: "issue_generic",
		Issue: Issue{
			ID:  "10001",
			Key: testIssueKey,
			Fields: Fields{
				Summary:     "Test issue for functional testing",
				Description: "This is an automated test issue",
				Status: Status{
					Name: "In Development",
				},
				Project: Project2{
					Key: "TEST",
				},
			},
		},
		Changelog: &Changelog{
			Items: []ChangelogItem{
				{
					Field:      "status",
					FromString: "To Do",
					ToString:   "In Development",
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return TestResult{
			TestName: "Send Webhook",
			Passed:   false,
			Message:  fmt.Sprintf("Failed to marshal webhook: %v", err),
			Duration: time.Since(start),
		}
	}

	resp, err := http.Post(apiURL+"/webhook", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return TestResult{
			TestName: "Send Webhook",
			Passed:   false,
			Message:  fmt.Sprintf("Failed to send webhook: %v", err),
			Duration: time.Since(start),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TestResult{
			TestName: "Send Webhook",
			Passed:   false,
			Message:  fmt.Sprintf("Expected status 200, got %d: %s", resp.StatusCode, string(body)),
			Duration: time.Since(start),
		}
	}

	logger.Infof("✓ Webhook sent successfully for issue: %s", testIssueKey)

	return TestResult{
		TestName: "Send Webhook",
		Passed:   true,
		Message:  "Webhook processed successfully",
		Duration: time.Since(start),
	}
}

func testValidateWebhookEvent(db *mongo.Database) TestResult {
	start := time.Now()
	logger.Info("\n[Test 3] Validating webhook event in MongoDB...")

	// Wait a bit for webhook to be stored
	time.Sleep(2 * time.Second)

	collection := db.Collection("webhook_events")
	filter := bson.M{"jira_issue_key": testIssueKey}

	var event bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&event)
	if err != nil {
		return TestResult{
			TestName: "Validate Webhook Event",
			Passed:   false,
			Message:  fmt.Sprintf("Webhook event not found: %v", err),
			Duration: time.Since(start),
		}
	}

	logger.Infof("✓ Webhook event found in database")

	return TestResult{
		TestName: "Validate Webhook Event",
		Passed:   true,
		Message:  "Webhook event stored correctly",
		Duration: time.Since(start),
	}
}

func testWaitForCompletion(db *mongo.Database) TestResult {
	start := time.Now()
	logger.Info("\n[Test 4] Waiting for development completion...")

	collection := db.Collection("developments")
	filter := bson.M{"jira_issue_key": testIssueKey}

	timeout := 5 * time.Minute
	checkInterval := 5 * time.Second
	elapsed := time.Duration(0)

	for elapsed < timeout {
		var development bson.M
		err := collection.FindOne(context.Background(), filter).Decode(&development)
		if err == nil {
			status := development["status"].(string)
			logger.Infof("Development status: %s", status)

			if status == "completed" {
				logger.Infof("✓ Development completed successfully")
				return TestResult{
					TestName: "Wait for Completion",
					Passed:   true,
					Message:  "Development completed",
					Duration: time.Since(start),
				}
			} else if status == "failed" {
				errorMsg := ""
				if errField, ok := development["error_message"]; ok {
					errorMsg = errField.(string)
				}
				return TestResult{
					TestName: "Wait for Completion",
					Passed:   false,
					Message:  fmt.Sprintf("Development failed: %s", errorMsg),
					Duration: time.Since(start),
				}
			}
		}

		time.Sleep(checkInterval)
		elapsed += checkInterval
		logger.Infof("Waiting... (%s elapsed)", elapsed)
	}

	return TestResult{
		TestName: "Wait for Completion",
		Passed:   false,
		Message:  fmt.Sprintf("Timeout after %s", timeout),
		Duration: time.Since(start),
	}
}

func testValidateDevelopment(db *mongo.Database) TestResult {
	start := time.Now()
	logger.Info("\n[Test 5] Validating development record...")

	collection := db.Collection("developments")
	filter := bson.M{"jira_issue_key": testIssueKey}

	var development bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&development)
	if err != nil {
		return TestResult{
			TestName: "Validate Development",
			Passed:   false,
			Message:  fmt.Sprintf("Development record not found: %v", err),
			Duration: time.Since(start),
		}
	}

	// Validate required fields
	requiredFields := []string{"jira_issue_key", "jira_project_key", "status", "created_at"}
	for _, field := range requiredFields {
		if _, ok := development[field]; !ok {
			return TestResult{
				TestName: "Validate Development",
				Passed:   false,
				Message:  fmt.Sprintf("Missing required field: %s", field),
				Duration: time.Since(start),
			}
		}
	}

	logger.Infof("✓ Development record validated")

	return TestResult{
		TestName: "Validate Development",
		Passed:   true,
		Message:  "All required fields present",
		Duration: time.Since(start),
	}
}

func testCleanup(apiURL string, db *mongo.Database) TestResult {
	start := time.Now()
	logger.Info("\n[Test 6] Cleaning up test data...")

	// Delete test project
	if testProjectID != "" {
		req, _ := http.NewRequest("DELETE", apiURL+"/api/projects/"+testProjectID, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Warnf("Failed to delete test project: %v", err)
		} else {
			resp.Body.Close()
			logger.Infof("✓ Deleted test project")
		}
	}

	// Delete webhook event
	webhookCollection := db.Collection("webhook_events")
	webhookCollection.DeleteMany(context.Background(), bson.M{"jira_issue_key": testIssueKey})
	logger.Infof("✓ Deleted webhook events")

	// Delete development record
	devCollection := db.Collection("developments")
	devCollection.DeleteMany(context.Background(), bson.M{"jira_issue_key": testIssueKey})
	logger.Infof("✓ Deleted development records")

	return TestResult{
		TestName: "Cleanup",
		Passed:   true,
		Message:  "Test data cleaned up successfully",
		Duration: time.Since(start),
	}
}

func printResults(results []TestResult) {
	logger.Info("\n========================================")
	logger.Info("TEST RESULTS")
	logger.Info("========================================")

	for i, result := range results {
		status := "✓ PASS"
		if !result.Passed {
			status = "✗ FAIL"
		}
		logger.Infof("[%d] %s - %s (%.2fs)", i+1, status, result.TestName, result.Duration.Seconds())
		logger.Infof("    %s", result.Message)
	}

	passed := 0
	failed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	logger.Info("========================================")
	logger.Infof("Total: %d | Passed: %d | Failed: %d", len(results), passed, failed)
	logger.Info("========================================")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
