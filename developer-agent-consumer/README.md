# Developer Agent Consumer

The Developer Agent Consumer is a microservice that processes development requests from RabbitMQ, generates code using Claude Code, and creates pull/merge requests automatically.

## Features

- **RabbitMQ Message Consumption**: Listens for development requests from the `develop` queue
- **Configuration API Integration**: Fetches project and repository configurations
- **Git Operations**: Clones repositories, creates branches, commits, and pushes changes
- **Repository Analysis**: Analyzes repository structure to understand project patterns
- **Claude Code Integration**: Generates code using Claude Code API with project context
- **PR/MR Creation**: Creates pull requests on GitHub or merge requests on GitLab
- **Development Tracking**: Stores development progress in MongoDB

## Architecture

The service follows clean architecture with the following layers:

- **Consumer**: RabbitMQ message consumption
- **Services**: Business logic (Git, Analysis, Claude, PR)
- **Clients**: External API clients (Configuration API)
- **Repositories**: Data access layer (MongoDB)
- **Models**: Data structures

## Workflow

1. Consume message from RabbitMQ `develop` queue
2. Create development record in MongoDB with status "ready"
3. Fetch project configuration from Configuration API
4. Clone repository to temporary directory `/tmp/sdlc-{jira_issue_key}/repo`
5. Analyze repository structure (entry points, directories, patterns)
6. Generate code using Claude Code API with project context
7. Create feature branch `feature/{jira_issue_key}`
8. Commit changes with message `[{jira_issue_key}] {summary}`
9. Push branch to remote repository
10. Create pull/merge request
11. Update development record with status "completed" and PR/MR URL
12. Clean up temporary directory

On failure, the service:
- Updates development record with status "failed" and error message
- Publishes failed message to `develop_error` queue
- Acknowledges RabbitMQ message to prevent reprocessing

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGO_URL` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | MongoDB database name | `sdlc_agent` |
| `RABBITMQ_URL` | RabbitMQ connection string | `amqp://guest:guest@localhost:5672/` |
| `CONFIG_API_URL` | Configuration API base URL | `http://localhost:3000` |
| `CLAUDE_API_URL` | Claude Code API endpoint | `http://localhost:8000/generate` |
| `CLAUDE_SESSION_TOKEN` | Claude Code session token | _(required)_ |

## Dependencies

- **Go 1.20+**
- **MongoDB**: For storing development records
- **RabbitMQ**: For message queue
- **Configuration API**: For project/repository configuration
- **Claude Code API**: For code generation
- **GitHub/GitLab**: For creating PRs/MRs

## Running Locally

```bash
# Set environment variables
export MONGO_URL=mongodb://localhost:27017
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
export CONFIG_API_URL=http://localhost:3000
export CLAUDE_API_URL=http://localhost:8000/generate
export CLAUDE_SESSION_TOKEN=your-session-token

# Run the service
go run main.go
```

## Running with Docker

```bash
# Build image
docker build -t developer-agent-consumer .

# Run container
docker run -d \
  -e MONGO_URL=mongodb://mongo:27017 \
  -e RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/ \
  -e CONFIG_API_URL=http://config-api:3000 \
  -e CLAUDE_API_URL=http://claude-api:8000/generate \
  -e CLAUDE_SESSION_TOKEN=your-session-token \
  --name developer-agent-consumer \
  developer-agent-consumer
```

## MongoDB Collections

### developments

Stores development request processing status.

**Fields**:
- `_id`: ObjectID
- `jira_issue_id`: JIRA issue ID
- `jira_issue_key`: JIRA issue key (e.g., "PROJ-123")
- `jira_project_key`: JIRA project key (e.g., "PROJ")
- `repository_url`: Repository URL
- `branch_name`: Feature branch name
- `pr_mr_url`: Pull/merge request URL (optional)
- `status`: "ready", "completed", or "failed"
- `development_details`: Details from Claude Code (optional)
- `error_message`: Error message if failed (optional)
- `created_at`: Timestamp
- `completed_at`: Timestamp (optional)

## RabbitMQ Queues

### develop (Input)

Receives development requests with the following structure:

```json
{
  "jira_issue_id": "10001",
  "jira_issue_key": "PROJ-123",
  "jira_project_key": "PROJ",
  "summary": "Add user authentication",
  "description": "Implement JWT authentication for API endpoints",
  "repository": "https://github.com/example/repo"
}
```

### develop_error (Output)

Failed messages are published to this queue with error details:

```json
{
  "original_message": "{...}",
  "error": "failed to clone repository: authentication failed",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Error Handling

The service handles various error scenarios:

- **Project not found**: When JIRA project key is not in Configuration API
- **Repository not found**: When repository URL is not in project configuration
- **Git authentication failed**: When Git access token is invalid
- **Claude API error**: When code generation fails
- **PR/MR creation failed**: When GitHub/GitLab API returns an error

All errors are logged with structured logging and stored in the development record.

## Logging

Structured JSON logging with logrus:

```json
{
  "level": "info",
  "msg": "Processing development request",
  "jira_issue_key": "PROJ-123",
  "jira_project_key": "PROJ",
  "time": "2024-01-15T10:30:00Z"
}
```

## Development

### Running Tests

```bash
go test ./... -v
```

### Adding New Services

1. Create service in `services/` directory
2. Initialize in `main.go`
3. Add to message handler dependencies
4. Update README with configuration if needed

## License

Copyright © 2024 SDLC Agent Project
