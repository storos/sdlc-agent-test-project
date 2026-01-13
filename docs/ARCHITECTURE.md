# Architecture Overview

## System Architecture

SDLC AI Agents follows a **microservices architecture** with event-driven patterns using RabbitMQ for asynchronous processing.

```
┌─────────────┐
│    JIRA     │
└──────┬──────┘
       │ Webhook
       ▼
┌──────────────────────┐      ┌─────────────┐
│ JIRA Webhook API     │─────▶│  MongoDB    │
│  (Port 8080)         │      │             │
└──────┬───────────────┘      └─────────────┘
       │ Publish
       ▼
┌──────────────────────┐
│    RabbitMQ          │
│ (Exchange + Queue)   │
└──────┬───────────────┘
       │ Consume
       ▼
┌──────────────────────┐      ┌─────────────┐
│ Developer Agent      │─────▶│Configuration│
│ Consumer             │◀─────│ API         │
└──────┬───────────────┘      │ (Port 8081) │
       │                      └──────▲──────┘
       │                             │
       │ Create PR/MR                │
       ▼                             │
┌──────────────────────┐      ┌─────┴──────┐
│ GitHub/GitLab        │      │ Backoffice │
│                      │      │ UI (3000)  │
└──────────────────────┘      └────────────┘
       ▲
       │ Clone, Push
       │
┌──────┴───────────────┐
│   Claude Code        │
│   (AI Engine)        │
└──────────────────────┘
```

## Core Components

### 1. JIRA Webhook API

**Type**: HTTP REST API
**Technology**: Go + Gin
**Port**: 8080

**Purpose**: Webhook receiver and event processor

**Responsibilities**:
- Receive incoming webhook calls from JIRA
- Parse and validate webhook payload
- Decide which agent to trigger based on JIRA issue status parameter
- Store webhook events to MongoDB (`webhook_events` collection)
- Publish messages to RabbitMQ exchange when status is "In Development"

**Key Characteristics**:
- Stateless and horizontally scalable
- Lightweight processing (parse, store, publish)
- Average response time: < 100ms

**Workflow**:
1. Receive JIRA webhook payload
2. Check if event is `jira:issue_updated` and status changed to "In Development"
3. Extract required fields from webhook payload
4. Store event data to `webhook_events` collection
5. Publish message to RabbitMQ exchange (`webhook.development.request`)
6. Return HTTP response

### 2. Developer Agent Consumer

**Type**: Message Queue Consumer
**Technology**: Go
**Port**: N/A (background worker)

**Purpose**: AI-driven code generation worker

**Responsibilities**:
- Consume messages from RabbitMQ queue
- Call Configuration API to retrieve project details
- Extract repository URL from JIRA issue components field
- Match repository URL with project's configured repositories
- Clone the matched single repository to temporary folder
- Create feature branch from main
- Analyze repository structure
- Build AI prompts with project context
- Call Claude Code to generate/modify code
- Switch to feature branch and commit changes
- Push feature branch to remote repository
- Create pull/merge request
- Update development status in MongoDB
- Delete temporary folder
- Acknowledge or reject RabbitMQ messages

**Key Characteristics**:
- Single consumer (processes one message at a time)
- Long-running operations (minutes per message)
- Stateful during processing (cloned repo in temp folder)
- Cleanup after each message (delete temp folder)
- **Repository Selection**: Each JIRA issue specifies which repository to use via components field
- **Validation**: Fails if specified repository not found in project configuration

**Workflow**: See INTEGRATION.md for complete workflow with repository matching

### 3. Configuration API

**Type**: HTTP REST API
**Technology**: Go + Gin
**Port**: 8081

**Purpose**: Configuration management backend

**Responsibilities**:
- CRUD operations for projects and repositories
- Serve Backoffice UI
- Provide configuration to Developer Agent Consumer
- MongoDB data access layer

**Key Endpoints**:
- `GET /api/projects` - List all projects
- `GET /api/projects?jira_project_key={jira_project_key}` - Get project by JIRA key (used by Consumer)
- `GET /api/projects/:id` - Get specific project
- `POST /api/projects` - Create new project
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project
- `GET /api/projects/:id/repositories` - Get project repositories
- `POST /api/projects/:id/repositories` - Add repository
- `PUT /api/repositories/:id` - Update repository
- `DELETE /api/repositories/:id` - Remove repository

**Key Characteristics**:
- Stateless and horizontally scalable
- RESTful API design
- MongoDB data access through repositories pattern

### 4. Backoffice UI

**Type**: Single Page Application
**Technology**: React + TypeScript + Vite + Material-UI
**Port**: 3000

**Purpose**: Project management interface

**Responsibilities**:
- Configure or add new project definitions
- Manage Git repository associations (GitHub, GitLab, etc.)
- Set up JIRA project mappings
- View project configurations

**Key Screens**:

#### Project Definitions Screen
- Create/Edit/Delete projects
- Configure JIRA project name and project ID
- Add multiple Git repository URLs
- Add repository descriptions
- Manage Git access tokens per repository

**Fields per Project**:
- Project Name
- Project Description
- Project Scope (guidelines, constraints, coding standards for AI)
- JIRA Project ID
- JIRA Project Name
- JIRA Project URL
- Repositories (array):
  - Repository URL
  - Repository Description
  - Git Access Token (Personal Access Token)

### 5. MongoDB

**Type**: NoSQL Database
**Port**: 27017

**Purpose**: Data persistence

**Collections**:
- `projects` - Project configurations
- `webhook_events` - JIRA webhook audit trail
- `developments` - Development tracking

See DATABASE.md for detailed schema.

### 6. RabbitMQ

**Type**: Message Broker
**Port**: 5672 (AMQP), 15672 (Management UI)

**Purpose**: Asynchronous job processing

**Configuration**:
- Exchange: `webhook.development.request` (Topic, Durable)
- Queue: `develop` (Durable)
- Error Queue: `develop_error` (Durable)
- Binding: Queue `develop` binds to exchange
- Message TTL: No expiration
- Concurrency: Single consumer

See INTEGRATION.md for detailed RabbitMQ setup.

## Design Patterns

### Event-Driven Architecture
- JIRA webhooks trigger the development workflow
- RabbitMQ decouples webhook reception from code generation
- Asynchronous processing allows long-running tasks

### Microservices Pattern
- Each service has a single responsibility
- Services communicate via REST APIs and message queues
- Independent deployment and scaling
- Polyglot persistence (could add other databases)

### Repository Pattern
- MongoDB access is abstracted through repositories
- Clear separation: handlers → services → repositories
- Easier to test and maintain

### Message Queue Pattern
- **Exchange Type**: Topic (allows future routing patterns)
- **Queue Binding**: Simple 1:1 binding initially
- **Error Handling**: Dead letter queue for failed messages
- **No Retry**: Manual review of failed messages

## Communication Patterns

### Synchronous (HTTP/REST)
- **Backoffice UI** ↔ **Configuration API**: CRUD operations
- **Developer Agent Consumer** → **Configuration API**: Fetch project config
- **External** → **JIRA Webhook API**: Receive webhooks

### Asynchronous (Message Queue)
- **JIRA Webhook API** → **RabbitMQ** → **Developer Agent Consumer**
- Guarantees delivery even if consumer is down
- Handles backpressure naturally

### External Integrations
- **JIRA**: Webhook-based (push model)
- **GitHub/GitLab**: REST API (pull model for PRs/MRs)
- **Claude Code**: API calls for code generation
- **Git Repositories**: Git protocol with token auth

## Data Flow

### Happy Path
1. **JIRA** sends webhook → **JIRA Webhook API**
2. **JIRA Webhook API** stores to **MongoDB** (`webhook_events`)
3. **JIRA Webhook API** publishes to **RabbitMQ**
4. **Developer Agent Consumer** consumes from **RabbitMQ**
5. **Consumer** creates record in **MongoDB** (`developments`)
6. **Consumer** calls **Configuration API** for project details
7. **Consumer** matches repository URL from JIRA with project repositories
8. **Consumer** clones from **Git Repository**
9. **Consumer** creates feature branch from main
10. **Consumer** analyzes repository structure
11. **Consumer** calls **Claude Code** for code generation
12. **Consumer** switches to feature branch (if not already on it)
13. **Consumer** commits changes to feature branch
14. **Consumer** pushes feature branch to **Git Repository**
15. **Consumer** creates PR/MR via **GitHub/GitLab API**
16. **Consumer** updates **MongoDB** (`developments`)
17. **Consumer** deletes temporary folder
18. **Consumer** acknowledges **RabbitMQ** message

### Error Path
1. Any error in Consumer workflow (including repository not found in project config)
2. **Consumer** updates **MongoDB** with status "failed" and error message
3. **Consumer** publishes to **RabbitMQ** error queue
4. **Consumer** acknowledges original message
5. Manual review and retry possible

**Common Error Scenarios**:
- Repository URL from JIRA components not found in project repositories
- Configuration not found for JIRA project ID
- Git clone/push failures
- Claude Code API failures
- PR/MR creation failures

## Code Organization

### Go Services Structure
```
service-name/
├── main.go              # Entry point
├── handlers/            # HTTP handlers
│   └── entity_handler.go
├── services/            # Business logic
│   └── entity_service.go
├── repositories/        # Data access
│   └── entity_repository.go
├── models/              # Data structures
│   └── entity.go
├── Dockerfile
├── go.mod
└── go.sum
```

### Clean Architecture Layers
- **Handlers**: HTTP request/response handling
- **Services**: Business logic
- **Repositories**: Database operations
- **Models**: Data structures

## Naming Conventions

### Branch Naming
- Pattern: `feature/{jira_issue_key}`
- Example: `feature/PROJ-123`

### Commit Messages
- Pattern: `[{jira_issue_key}] {description}`
- Example: `[PROJ-123] Implemented user authentication feature`

### MongoDB Collections
- Lowercase with underscores: `webhook_events`, `developments`

### API Endpoints
- RESTful: `/api/projects`, `/api/projects/:id/repositories`

## Technology Decisions

- **Go**: Excellent concurrency, fast, strong libraries for MongoDB/RabbitMQ, single binary deployment
- **React + TypeScript**: Type safety, rich ecosystem (MUI), fast development (Vite)
- **MongoDB**: Flexible schema, good Go support, easy Docker setup
- **RabbitMQ**: Reliable messaging, flexible routing, management UI
- **Docker**: Consistent environments, easy testing, production-ready

## API Standards

- **REST**: Proper HTTP methods, meaningful URLs, correct status codes (2xx/4xx/5xx), JSON format
- **Error Handling**: Structured error responses with descriptive messages
- **Logging**: Structured logging with logrus (JSON format, levels: INFO/WARN/ERROR/DEBUG)

## Scalability Considerations

**Current (MVP)**: Webhook/Config APIs (horizontally scalable), Consumer (single instance), MongoDB/RabbitMQ (single instance)

**Future**: Multiple consumers, MongoDB replica set, RabbitMQ cluster, read replicas

## Security Architecture

- **Authentication**: No auth for internal APIs (Docker network isolation), PATs for Git, Claude Code token via env var
- **Network**: Docker internal network, minimal port exposure
- **Data**: Audit trail in MongoDB, no auto-retry (prevents token abuse)

## Performance Considerations

- **JIRA Webhook API**: Lightweight processing (parse, store, publish)
- **Configuration API**: Database queries optimized with indexes
- **Developer Agent Consumer**: Long-running (clone, AI generation, push)
- **Backoffice UI**: Client-side rendering with React

## Related Documentation

- [Database Schema](DATABASE.md) - MongoDB collections and indexes
- [Integration Details](INTEGRATION.md) - JIRA, RabbitMQ, Git, Claude Code
- [Deployment Guide](DEPLOYMENT.md) - Docker Compose and environment variables
- [Testing Guide](TESTING.md) - Functional testing approach
