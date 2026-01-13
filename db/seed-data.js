// Seed Data Script for SDLC AI Agents
// This script adds test data for development and testing

db = db.getSiblingDB('sdlc_agents');

print('Adding seed data to SDLC AI Agents database...');

// ============================================
// Seed Project Data
// ============================================
print('\nAdding test projects...');

const testProject = {
  name: "E-Commerce Platform",
  description: "Microservices-based e-commerce system for online retail",
  scope: "Backend services written in Go using Gin framework. Follow clean architecture pattern: handlers → services → repositories. Use structured logging with logrus. Follow RESTful API design principles. Write unit tests for all services.",
  jira_project_key: "ECOM",
  jira_project_name: "E-Commerce",
  jira_project_url: "https://company.atlassian.net/projects/ECOM",
  repositories: [
    {
      repository_id: "repo-001",
      url: "https://github.com/storos/sdlc-agent-test-project",
      description: "Main API backend service handling products, orders, and payments",
      git_access_token: "ghp_test_token_replace_with_real_token"
    }
  ],
  created_at: new Date(),
  updated_at: new Date()
};

// Remove existing test project if it exists
db.projects.deleteOne({ jira_project_key: "ECOM" });

// Insert test project
const projectResult = db.projects.insertOne(testProject);
print('✓ Test project created with ID: ' + projectResult.insertedId);

// Store project ID for later use
const projectId = projectResult.insertedId;

// ============================================
// Seed Webhook Event Data
// ============================================
print('\nAdding test webhook events...');

const testWebhookEvent = {
  jira_issue_id: "10001",
  jira_issue_key: "ECOM-123",
  jira_project_key: "ECOM",
  summary: "Add user authentication endpoint",
  description: "Implement JWT-based authentication for /api/auth/login endpoint with email and password validation",
  issue_status: "In Development",
  issue_type: "Story",
  repository: "https://github.com/storos/sdlc-agent-test-project",
  webhook_payload: {
    webhookEvent: "jira:issue_updated",
    issue: {
      id: "10001",
      key: "ECOM-123"
    },
    changelog: {
      items: [{
        field: "status",
        fromString: "To Do",
        toString: "In Development"
      }]
    }
  },
  created_at: new Date()
};

// Remove existing test webhook if it exists
db.webhook_events.deleteOne({ jira_issue_key: "ECOM-123" });

// Insert test webhook event
const webhookResult = db.webhook_events.insertOne(testWebhookEvent);
print('✓ Test webhook event created with ID: ' + webhookResult.insertedId);

// ============================================
// Seed Development Data
// ============================================
print('\nAdding test development records...');

// Example 1: Completed development
const completedDevelopment = {
  project_id: projectId,
  jira_issue_id: "10001",
  jira_issue_key: "ECOM-123",
  jira_project_key: "ECOM",
  summary: "Add user authentication endpoint",
  description: "Implement JWT-based authentication for /api/auth/login endpoint",
  repository_url: "https://github.com/storos/sdlc-agent-test-project",
  git_access_token: "ghp_test_token_replace_with_real_token",
  branch_name: "feature/ECOM-123",
  status: "completed",
  pr_mr_url: "https://github.com/storos/sdlc-agent-test-project/pull/1",
  development_details: "Created /api/auth/login endpoint with JWT token generation. Added user validation middleware. Updated authentication documentation.",
  error_message: null,
  created_at: new Date(Date.now() - 3600000), // 1 hour ago
  updated_at: new Date()
};

// Example 2: Failed development
const failedDevelopment = {
  project_id: projectId,
  jira_issue_id: "10002",
  jira_issue_key: "ECOM-124",
  jira_project_key: "ECOM",
  summary: "Fix payment processing bug",
  description: "Debug and fix payment gateway timeout issue",
  repository_url: "https://github.com/company/unknown-repo",
  git_access_token: null,
  branch_name: null,
  status: "failed",
  pr_mr_url: null,
  development_details: null,
  error_message: "Repository 'https://github.com/company/unknown-repo' from JIRA components not found in project configuration",
  created_at: new Date(Date.now() - 7200000), // 2 hours ago
  updated_at: new Date(Date.now() - 7200000)
};

// Example 3: Ready development (pending processing)
const readyDevelopment = {
  project_id: projectId,
  jira_issue_id: "10003",
  jira_issue_key: "ECOM-125",
  jira_project_key: "ECOM",
  summary: "Add product search functionality",
  description: "Implement full-text search for products with filters",
  repository_url: "https://github.com/storos/sdlc-agent-test-project",
  git_access_token: "ghp_test_token_replace_with_real_token",
  branch_name: "feature/ECOM-125",
  status: "ready",
  pr_mr_url: null,
  development_details: null,
  error_message: null,
  created_at: new Date(),
  updated_at: new Date()
};

// Remove existing test developments if they exist
db.developments.deleteMany({ jira_project_key: "ECOM" });

// Insert test developments
db.developments.insertMany([completedDevelopment, failedDevelopment, readyDevelopment]);
print('✓ Test development records created (3 records)');

// ============================================
// Verify Seed Data
// ============================================
print('\nVerifying seed data...');

const projectCount = db.projects.countDocuments();
const webhookCount = db.webhook_events.countDocuments();
const developmentCount = db.developments.countDocuments();

print('Projects: ' + projectCount);
print('Webhook events: ' + webhookCount);
print('Developments: ' + developmentCount);

print('\n✓ Seed data added successfully!');
print('\nNote: Replace "ghp_test_token_replace_with_real_token" with a real GitHub token for actual testing.');
