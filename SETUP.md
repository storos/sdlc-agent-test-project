# SDLC AI Agents - Setup Guide

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Claude Code session token
- (Optional) GitHub/GitLab Personal Access Token for testing

### Step 1: Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` and set your values:
- `CLAUDE_CODE_SESSION_TOKEN`: Your Claude Code API token
- `TEST_GITHUB_TOKEN`: (Optional) GitHub token for testing

### Step 2: Start All Services

```bash
# Build and start all services
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

This will start:
- **MongoDB** on port 27017
- **RabbitMQ** on ports 5672 (AMQP) and 15672 (Management UI)
- **Configuration API** on port 8081
- **JIRA Webhook API** on port 8080
- **Backoffice UI** on port 3000
- **Developer Agent Consumer** (background service)

### Step 3: Initialize Database

The database will be automatically initialized on first startup using `db/init-db.js`.

To add seed data for testing:

```bash
docker exec -i sdlc-mongodb mongosh sdlc_agents < db/seed-data.js
```

### Step 4: Access Services

- **Backoffice UI**: http://localhost:3000
- **Configuration API**: http://localhost:8081/api/projects
- **JIRA Webhook Endpoint**: http://localhost:8080/webhook
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **MongoDB**: mongodb://localhost:27017/sdlc_agents

### Step 5: Configure a Project

1. Open Backoffice UI at http://localhost:3000
2. Create a new project:
   - Set project name and description
   - Add JIRA project key (e.g., "ECOM")
   - Add repository URL and Git access token
   - Define project scope/guidelines for AI

### Step 6: Configure JIRA Webhook

In your JIRA instance:

1. Go to **Settings** → **System** → **WebHooks**
2. Create a new webhook:
   - **URL**: `http://your-server:8080/webhook`
   - **Events**: Issue → updated
   - **JQL Filter**: `status = "In Development"`
3. Save the webhook

### Step 7: Test the Workflow

1. Create a JIRA issue in your configured project
2. Add a component with name "repository" and value matching your configured repository URL
3. Move the issue to "In Development" status
4. Monitor the logs to see the workflow execute:

```bash
# Watch all logs
docker-compose logs -f

# Watch specific service
docker-compose logs -f developer-agent-consumer
```

## Service Management

### Stop All Services

```bash
docker-compose down
```

### Stop and Remove Volumes (Clean Start)

```bash
docker-compose down -v
```

### Restart a Specific Service

```bash
docker-compose restart jira-webhook-api
docker-compose restart developer-agent-consumer
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f jira-webhook-api
docker-compose logs -f developer-agent-consumer
docker-compose logs -f configuration-api
```

### Rebuild After Code Changes

```bash
docker-compose up --build
```

## Database Management

### Access MongoDB Shell

```bash
docker exec -it sdlc-mongodb mongosh sdlc_agents
```

### View Collections

```javascript
show collections
db.projects.find().pretty()
db.webhook_events.find().pretty()
db.developments.find().pretty()
```

### Reset Database

```bash
# Drop and reinitialize
docker exec -it sdlc-mongodb mongosh sdlc_agents --eval "db.dropDatabase()"
docker exec -i sdlc-mongodb mongosh sdlc_agents < db/init-db.js
docker exec -i sdlc-mongodb mongosh sdlc_agents < db/seed-data.js
```

## RabbitMQ Management

### Access Management UI

Open http://localhost:15672 (default credentials: guest/guest)

### Monitor Queues

- **develop**: Main processing queue
- **develop_error**: Failed message queue

### Clear Error Queue

```bash
docker exec sdlc-rabbitmq rabbitmqctl purge_queue develop_error
```

## Troubleshooting

### Service Won't Start

Check logs:
```bash
docker-compose logs <service-name>
```

### MongoDB Connection Issues

```bash
# Check MongoDB is running
docker ps | grep mongodb

# Check health
docker exec sdlc-mongodb mongosh --eval "db.runCommand('ping')"
```

### RabbitMQ Connection Issues

```bash
# Check RabbitMQ is running
docker ps | grep rabbitmq

# Check health
docker exec sdlc-rabbitmq rabbitmq-diagnostics ping
```

### Consumer Not Processing Messages

1. Check consumer logs: `docker-compose logs -f developer-agent-consumer`
2. Verify RabbitMQ queue has messages: http://localhost:15672
3. Check Configuration API is accessible: `curl http://localhost:8081/api/projects`
4. Verify Claude Code token is set correctly in `.env`

## Next Steps

- Read [docs/README.md](docs/README.md) for complete documentation
- Review [PROJECT-PLAN.md](PROJECT-PLAN.md) for development roadmap
- Run functional tests (see [docs/TESTING.md](docs/TESTING.md))
