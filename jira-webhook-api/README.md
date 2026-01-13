# JIRA Webhook API

Receives JIRA webhooks and triggers automated development workflow.

## Features

- Receives JIRA webhook events
- Detects "In Development" status changes
- Stores webhook events in MongoDB
- Publishes development requests to RabbitMQ
- Structured logging with logrus
- Health check endpoint

## Endpoints

### POST /webhook
Receives JIRA webhook payloads.

**Request Body**: JIRA webhook JSON payload

**Response**:
```json
{
  "message": "Webhook processed successfully",
  "issue_key": "PROJ-123"
}
```

### GET /health
Health check endpoint.

**Response**:
```json
{
  "status": "healthy",
  "service": "jira-webhook-api"
}
```

## Environment Variables

- `PORT`: Server port (default: 8080)
- `MONGODB_URL`: MongoDB connection string (default: mongodb://localhost:27017/sdlc_agents)
- `RABBITMQ_URL`: RabbitMQ connection string (default: amqp://guest:guest@localhost:5672/)
- `LOG_LEVEL`: Logging level (default: info)

## Development

### Run Locally

```bash
# Install dependencies
go mod tidy

# Set environment variables
export MONGODB_URL=mongodb://localhost:27017/sdlc_agents
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Run the application
go run main.go
```

### Build

```bash
go build -o jira-webhook-api .
```

## Docker

### Build Image

```bash
docker build -t jira-webhook-api .
```

### Run Container

```bash
docker run -p 8080:8080 \
  -e MONGODB_URL=mongodb://mongodb:27017/sdlc_agents \
  -e RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/ \
  jira-webhook-api
```

## JIRA Webhook Configuration

Configure JIRA to send webhooks to this API:

1. Go to JIRA Settings → System → WebHooks
2. Create a new webhook
3. Set URL: `http://your-server:8080/webhook`
4. Select events: Issue Updated
5. Save

## Message Flow

1. JIRA sends webhook when issue status changes
2. API validates payload and checks for "In Development" status
3. Webhook event is stored in MongoDB `webhook_events` collection
4. Development request message is published to RabbitMQ exchange `webhook.development.request`
5. Routing key: `webhook.development.{jira_project_key}`
6. Consumer processes the message and triggers development

## RabbitMQ Message Format

```json
{
  "jira_issue_id": "10001",
  "jira_issue_key": "PROJ-123",
  "jira_project_key": "PROJ",
  "summary": "Issue summary",
  "description": "Issue description",
  "repository": "https://github.com/org/repo"
}
```
