# Testing Guide

## Overview

This document covers the functional testing approach for SDLC AI Agents. A separate testing application validates the entire workflow from JIRA webhook to Git repository pull/merge request creation.

---

## Test Application

### Technology Stack

- **Language**: Go
- **HTTP Client**: `net/http` for API calls
- **MongoDB Client**: `go.mongodb.org/mongo-driver` for database validation
- **GitHub API Client**: `github.com/google/go-github` for GitHub validation
- **Logging**: `github.com/sirupsen/logrus` for structured console output
- **Containerization**: Docker

### Test Structure

```
/tests
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── run-tests.sh
└── test_data/
    ├── test_project.json
    └── test_webhook.json
```

---

## Running Tests

### Prerequisites

Main services must be running before executing tests.

```bash
# Start main services first
docker-compose up -d

# Run functional tests
cd tests
./run-tests.sh
```

The `run-tests.sh` script will:
1. Build the Docker image for the test application
2. Run the Docker container
3. Execute functional tests
4. Output test results and logs to console

### Test Execution Script

**File**: `tests/run-tests.sh`

```bash
#!/bin/bash
# run-tests.sh - Build and run functional tests

set -e  # Exit on error

echo "Building test application Docker image..."
docker build -t sdlc-agent-tests:latest .

echo "Running functional tests..."
docker run --rm \
  --network sdlc-network \
  -e CONFIGURATION_API_URL=http://configuration-api:8081 \
  -e JIRA_WEBHOOK_API_URL=http://jira-webhook-api:8080 \
  -e MONGODB_URL=mongodb://mongodb:27017 \
  -e TEST_API_KEY=${TEST_API_KEY} \
  -e GITHUB_TOKEN=${GITHUB_TOKEN} \
  sdlc-agent-tests:latest

echo "Tests completed!"
```

**Usage**:
```bash
# Make script executable
chmod +x tests/run-tests.sh

# Run tests
cd tests
./run-tests.sh
```

---

## Test Workflow

The functional tests execute the following steps:

1. **Connect to MongoDB** for validation
2. **Create test project configuration** via Configuration API
3. **Send JIRA webhook** to trigger development workflow
4. **Wait for processing** (poll developments collection)
5. **Validate database records** (webhook_events, developments)
6. **Validate GitHub** (branch, pull request)
7. **Output test results** with detailed logs
8. **Clean up test data**
9. **Exit with appropriate code** (0 = success, 1 = failure)

---

## Test Success Criteria

All of the following must be true for tests to pass:

1. ✅ Project configuration created successfully in database
2. ✅ Webhook received and processed (HTTP 200)
3. ✅ `webhook_events` collection has entry for TEST-123
4. ✅ `developments` collection has entry with status "completed"
5. ✅ Branch `feature/TEST-123` exists in GitHub repository
6. ✅ Pull request exists targeting main branch
7. ✅ Commit message follows pattern `[TEST-123] ...`
8. ✅ `development_details` field contains Claude Code output
9. ✅ No errors in application logs
10. ✅ Message acknowledged and removed from RabbitMQ queue

---

## Console Output Examples

### Success Output

```
INFO: Starting SDLC AI Agents Functional Tests
INFO: Connecting to MongoDB at mongodb://mongodb:27017
INFO: ✓ MongoDB connection successful
INFO:
INFO: [Test 1/10] Creating test project configuration
INFO: POST http://configuration-api:8081/api/projects
INFO: ✓ Test project created successfully (ID: 507f1f77bcf86cd799439011)
INFO:
INFO: [Test 2/10] Sending JIRA webhook
INFO: POST http://jira-webhook-api:8080/webhook
INFO: ✓ Webhook received and processed (HTTP 200)
INFO:
INFO: [Test 3/10] Validating webhook_events collection
INFO: Querying MongoDB for jira_issue_key: TEST-123
INFO: ✓ webhook_events entry found
INFO:
INFO: [Test 4/10] Waiting for development processing (polling every 5s, max 60s)
INFO: Polling attempt 1/12...
INFO: Development status: ready
INFO: Polling attempt 2/12...
INFO: Development status: ready
INFO: Polling attempt 3/12...
INFO: ✓ Development status: completed
INFO:
INFO: [Test 5/10] Validating developments collection
INFO: ✓ Development record found with status 'completed'
INFO: ✓ Pull/Merge request URL populated: https://github.com/storos/sdlc-agent-test-project/pull/1
INFO:
INFO: [Test 6/10] Validating GitHub branch exists
INFO: GET https://api.github.com/repos/storos/sdlc-agent-test-project/branches/feature/TEST-123
INFO: ✓ Branch feature/TEST-123 exists
INFO:
INFO: [Test 7/10] Validating GitHub pull request exists
INFO: GET https://api.github.com/repos/storos/sdlc-agent-test-project/pulls?head=feature/TEST-123
INFO: ✓ Pull request #1 found and open
INFO:
INFO: [Test 8/10] Validating commit message format
INFO: ✓ Commit message follows pattern [TEST-123]
INFO:
INFO: [Test 9/10] Validating development_details populated
INFO: ✓ development_details contains Claude Code output
INFO:
INFO: [Test 10/10] Validating RabbitMQ message acknowledged
INFO: Checking develop queue is empty
INFO: ✓ Message acknowledged and removed from queue
INFO:
INFO: Cleaning up test data...
INFO: ✓ Test cleanup completed
INFO:
INFO: ========================================
INFO: TEST RESULTS: 10/10 PASSED ✓
INFO: ========================================
INFO: All functional tests passed successfully!
```

### Error Output

```
INFO: [Test 4/10] Waiting for development processing (polling every 5s, max 60s)
INFO: Polling attempt 1/12...
INFO: Development status: ready
...
INFO: Polling attempt 12/12...
INFO: Development status: failed
ERROR: ✗ Development failed with error: Failed to push branch to GitHub
ERROR: Error message: authentication required
ERROR:
ERROR: ========================================
ERROR: TEST RESULTS: 3/10 PASSED, 1 FAILED ✗
ERROR: ========================================
ERROR: Tests failed. Check logs above for details.
```

---

## Test Cleanup

After test completion, the following cleanup actions are performed:

1. Delete test pull request from GitHub
2. Delete test branch `feature/TEST-123` from GitHub
3. Delete test project from `projects` collection
4. Delete test entries from `webhook_events` collection
5. Delete test entries from `developments` collection
6. Purge RabbitMQ queues if needed

---

## Troubleshooting Tests

### Issue: Tests failing

**Possible causes**:
- Services not running
- Test network not connected
- GitHub token invalid

**Solution**:
```bash
# Verify all services are running
docker-compose ps

# Check test network connectivity
docker network inspect sdlc-network

# Verify test environment variables
echo $GITHUB_TOKEN
```

### Issue: Test timeout during development processing

**Possible causes**:
- Developer Agent Consumer not processing messages
- Claude Code API token invalid
- Git authentication failure

**Solution**:
```bash
# Check Developer Agent Consumer logs
docker-compose logs developer-agent-consumer

# Check RabbitMQ queue
# Access http://localhost:15672 and check develop queue

# Verify Configuration API
curl http://localhost:8081/api/projects
```

### Issue: GitHub validation fails

**Possible causes**:
- Invalid GitHub token
- Repository not accessible
- Network connectivity issues

**Solution**:
```bash
# Verify GitHub token has correct permissions
# Required: repo, workflow

# Test GitHub API access manually
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/repos/storos/sdlc-agent-test-project

# Check repository exists and token has access
```

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md) - System design
- [Integration Details](INTEGRATION.md) - External system integration
- [Deployment Guide](DEPLOYMENT.md) - Docker Compose and environment variables
- [Database Schema](DATABASE.md) - MongoDB collections
