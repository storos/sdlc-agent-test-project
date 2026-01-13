# Deployment Guide

## Overview

This document covers:
- Docker Compose configuration
- Container networking
- Environment variables for all services
- Running and monitoring the system

---

## Docker Compose Configuration

### Container Networking

All services communicate via Docker internal network: `sdlc-network`

**Note**: The test application (`tests/`) is NOT included in the main `docker-compose.yml`. It runs separately via the `run-tests.sh` script but connects to the same Docker network to test the running services.

### Service Ports

#### External Ports (Host → Container)

| Service | Host Port | Container Port | Description |
|---------|-----------|----------------|-------------|
| JIRA Webhook API | 8080 | 8080 | Receives webhooks from external JIRA |
| Configuration API | 8081 | 8081 | Backend for Backoffice UI |
| Backoffice UI | 3000 | 3000 | Web interface for project management |
| MongoDB | 27017 | 27017 | Database access (for development/debugging) |
| RabbitMQ | 5672 | 5672 | Message broker port |
| RabbitMQ Management | 15672 | 15672 | RabbitMQ management console |

#### Internal Communication (Container ↔ Container)

Services use Docker service names for internal communication:
- **MongoDB**: `mongodb:27017`
- **RabbitMQ**: `rabbitmq:5672`
- **Configuration API**: `configuration-api:8081` (called by Backoffice UI and Developer Agent Consumer)

### Container Details

| Service | Container Name | Type | Dependencies |
|---------|---------------|------|--------------|
| JIRA Webhook API | jira-webhook-api | REST API | mongodb, rabbitmq |
| Developer Agent Consumer | developer-agent-consumer | Worker | mongodb, rabbitmq, configuration-api |
| Configuration API | configuration-api | REST API | mongodb |
| Backoffice UI | backoffice-ui | Web UI | configuration-api |
| MongoDB | mongodb | Database | None |
| RabbitMQ | rabbitmq | Message Broker | None |

---

## Environment Variables

### JIRA Webhook API

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | HTTP server port |
| `MONGODB_URL` | Yes | - | MongoDB connection string |
| `RABBITMQ_URL` | Yes | - | RabbitMQ connection string (AMQP) |
| `LOG_LEVEL` | No | info | Logging level (debug, info, warn, error) |

**Example**:
```bash
PORT=8080
MONGODB_URL=mongodb://mongodb:27017/sdlc_agents
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
LOG_LEVEL=info
```

### Developer Agent Consumer

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MONGODB_URL` | Yes | - | MongoDB connection string |
| `RABBITMQ_URL` | Yes | - | RabbitMQ connection string |
| `CONFIGURATION_API_URL` | Yes | - | Configuration API base URL |
| `CLAUDE_CODE_SESSION_TOKEN` | Yes | - | Claude Code API token |
| `LOG_LEVEL` | No | info | Logging level |
| `TEMP_DIR` | No | /tmp | Temporary directory for cloning |

**Example**:
```bash
MONGODB_URL=mongodb://mongodb:27017/sdlc_agents
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CONFIGURATION_API_URL=http://configuration-api:8081
CLAUDE_CODE_SESSION_TOKEN=sk-xxxxxxxxxxxxx
LOG_LEVEL=info
TEMP_DIR=/tmp
```

### Configuration API

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8081 | HTTP server port |
| `MONGODB_URL` | Yes | - | MongoDB connection string |
| `LOG_LEVEL` | No | info | Logging level |

**Example**:
```bash
PORT=8081
MONGODB_URL=mongodb://mongodb:27017/sdlc_agents
LOG_LEVEL=info
```

### Backoffice UI

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_URL` | Yes | - | Configuration API URL |
| `PORT` | No | 3000 | Web server port |

**Example**:
```bash
VITE_API_URL=http://localhost:8081
PORT=3000
```

### MongoDB

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MONGO_INITDB_DATABASE` | No | sdlc_agents | Initial database name |

**Example**:
```bash
MONGO_INITDB_DATABASE=sdlc_agents
```

### RabbitMQ

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RABBITMQ_DEFAULT_USER` | No | guest | Default username |
| `RABBITMQ_DEFAULT_PASS` | No | guest | Default password |

**Example**:
```bash
RABBITMQ_DEFAULT_USER=guest
RABBITMQ_DEFAULT_PASS=guest
```

---

## Running the System

### Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for local development)
- Node.js 18+ (for UI development)
- Access to JIRA instance
- GitHub or GitLab account with Personal Access Token
- Claude Code session token

### Start All Services

```bash
# From project root
docker-compose up --build
```

This will start all services:
- **JIRA Webhook API** (http://localhost:8080)
- **Developer Agent Consumer** (background service)
- **Backoffice UI** (http://localhost:3000)
- **Configuration API** (http://localhost:8081)
- **MongoDB** (mongodb://localhost:27017)
- **RabbitMQ** (amqp://localhost:5672, Management UI: http://localhost:15672)

### Stop All Services

```bash
docker-compose down
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f jira-webhook-api
docker-compose logs -f developer-agent-consumer
docker-compose logs -f configuration-api
docker-compose logs -f backoffice-ui
```

### Restart a Service

```bash
docker-compose restart jira-webhook-api
docker-compose restart developer-agent-consumer
```

---

## Monitoring

### RabbitMQ Management UI

Access: http://localhost:15672
Default credentials: `guest` / `guest`

**Monitor**:
- Queue depth (`develop` queue)
- Failed messages (`develop_error` queue)
- Message rates
- Consumer status

### MongoDB

Access: `mongodb://localhost:27017`

**Monitor**:
- `webhook_events` collection size
- `developments` collection status distribution
- Failed developments (`status: "failed"`)

### Application Logs

All services output structured JSON logs via logrus.

**Log Levels**:
- `INFO`: Major workflow steps
- `WARN`: Non-critical issues
- `ERROR`: Failures requiring attention
- `DEBUG`: Detailed processing information

**View logs**:
```bash
docker-compose logs -f developer-agent-consumer
```

### Key Metrics

- Messages consumed per hour
- Success vs failure rate
- Average processing time
- Failed messages in `develop_error` queue

### Health Indicators

- RabbitMQ connection status
- MongoDB connection status
- Configuration API availability
- Claude Code API response time

---

## Troubleshooting

### Issue: JIRA Webhook API returns 500 error

**Possible causes**:
- MongoDB connection failure
- RabbitMQ connection failure
- Invalid webhook payload

**Solution**:
```bash
# Check logs
docker-compose logs jira-webhook-api

# Verify MongoDB connection
docker-compose ps mongodb

# Verify RabbitMQ connection
docker-compose ps rabbitmq
```

### Issue: Developer Agent Consumer not processing messages

**Possible causes**:
- RabbitMQ connection failure
- Configuration API not accessible
- Claude Code API token invalid
- Git authentication failure

**Solution**:
```bash
# Check logs
docker-compose logs developer-agent-consumer

# Check RabbitMQ queue
# Access http://localhost:15672 and check develop queue

# Verify Configuration API
curl http://localhost:8081/api/projects

# Check environment variables
docker-compose exec developer-agent-consumer env | grep CLAUDE_CODE_SESSION_TOKEN
```

---

## Security Considerations

### Network Isolation

- All services in Docker internal network
- Only necessary ports exposed to host
- External access only via defined endpoints

### Secrets Management

- Git tokens stored per repository in database
- Claude Code token in environment variable
- **Future**: Use secrets management (Vault, AWS Secrets Manager)

### Authentication

- **Current**: No authentication for internal APIs (Docker network isolation)
- **Future**: Add API key authentication for production

---

## Production Considerations

### Scaling

- **Horizontal Scaling**: JIRA Webhook API, Configuration API (stateless)
- **Vertical Scaling**: Developer Agent Consumer (long-running operations)
- **Database**: MongoDB replica set
- **Message Broker**: RabbitMQ cluster

### Backup

- **MongoDB**: Regular backups of `projects`, `webhook_events`, `developments` collections
- **Git Tokens**: Secure storage and backup
- **RabbitMQ**: Message persistence enabled

### Monitoring

- **APM**: Application Performance Monitoring
- **Alerting**: Failed developments, queue depth
- **Logging**: Centralized log aggregation

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md) - System design
- [Database Schema](DATABASE.md) - MongoDB collections
- [Integration Details](INTEGRATION.md) - External system integration
- [Testing Guide](TESTING.md) - Functional testing approach
