# Functional Tests

End-to-end functional tests for the SDLC AI Agents system.

## Overview

This test suite validates the complete workflow:
1. Create a test project via Configuration API
2. Send a JIRA webhook payload
3. Validate webhook event is stored in MongoDB
4. Wait for Developer Agent Consumer to process the request
5. Validate development record is created with correct status
6. Clean up test data

## Prerequisites

- All services must be running (Configuration API, JIRA Webhook API, MongoDB, RabbitMQ, Developer Agent Consumer)
- Go 1.20+ installed (for local execution)

## Running Tests

### Local Execution

```bash
cd functional-tests
./run-tests.sh
```

### With Docker

```bash
# Build the test image
docker build -t sdlc-functional-tests .

# Run tests
docker run --rm \
  --network sdlc-network \
  -e CONFIG_API_URL=http://configuration-api:8081 \
  -e WEBHOOK_API_URL=http://jira-webhook-api:8080 \
  -e MONGO_URL=mongodb://mongodb:27017 \
  -e MONGO_DATABASE=sdlc_agent \
  sdlc-functional-tests
```

### Using Docker Compose

Add to docker-compose.yml:

```yaml
functional-tests:
  build:
    context: ./functional-tests
    dockerfile: Dockerfile
  container_name: sdlc-functional-tests
  environment:
    CONFIG_API_URL: http://configuration-api:8081
    WEBHOOK_API_URL: http://jira-webhook-api:8080
    MONGO_URL: mongodb://mongodb:27017
    MONGO_DATABASE: sdlc_agent
  depends_on:
    configuration-api:
      condition: service_healthy
    jira-webhook-api:
      condition: service_healthy
    developer-agent-consumer:
      condition: service_started
  networks:
    - sdlc-network
```

Then run:
```bash
docker-compose run --rm functional-tests
```

## Test Cases

### Test 1: Create Project
Creates a test project via Configuration API with:
- Project name: "Test Project"
- JIRA project key: "TEST"
- One test repository

**Expected**: HTTP 201 Created, project ID returned

### Test 2: Send Webhook
Sends a JIRA webhook payload with:
- Event: jira:issue_updated
- Status change: To Do → In Development
- Issue key: TEST-{random}

**Expected**: HTTP 200 OK, webhook processed

### Test 3: Validate Webhook Event
Queries MongoDB webhook_events collection for the test issue.

**Expected**: Webhook event exists with correct issue key

### Test 4: Wait for Completion
Polls the developments collection for up to 5 minutes.

**Expected**: Development status becomes "completed" or "failed"

**Note**: In a real environment with Claude Code API, this would complete. In test environment without Claude Code, it may timeout or fail, which is expected.

### Test 5: Validate Development
Checks the development record has required fields:
- jira_issue_key
- jira_project_key
- status
- created_at

**Expected**: All required fields present

### Test 6: Cleanup
Removes test data:
- Deletes test project via API
- Deletes webhook events from MongoDB
- Deletes development records from MongoDB

**Expected**: All test data removed successfully

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_API_URL` | Configuration API base URL | `http://localhost:8081` |
| `WEBHOOK_API_URL` | JIRA Webhook API base URL | `http://localhost:8080` |
| `MONGO_URL` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | MongoDB database name | `sdlc_agent` |

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed

## Limitations

**Note**: These tests validate the workflow up to the point where the Developer Agent Consumer would call Claude Code API. In a full integration environment, you would need:

1. A valid Claude Code API endpoint
2. A real Git repository with write access
3. Valid GitHub/GitLab credentials

For unit testing the individual components, see the respective test files in each service directory.

## Troubleshooting

### Services Not Running

If you see errors about services not being available:

```bash
# Start all services
cd ..
docker-compose up -d

# Wait for services to be healthy
docker-compose ps
```

### MongoDB Connection Errors

```bash
# Check MongoDB is running
docker ps | grep mongodb

# Check MongoDB logs
docker-compose logs mongodb
```

### Tests Timeout

If Test 4 (Wait for Completion) times out:
- This is expected if Claude Code API is not configured
- Check Developer Agent Consumer logs: `docker-compose logs developer-agent-consumer`
- Verify RabbitMQ has messages: http://localhost:15672

## Next Steps

For comprehensive testing, implement:
- Unit tests for each service (see User Story 5.2)
- Integration tests with mocked Claude Code API
- Performance tests with multiple concurrent requests
- CI/CD pipeline integration
