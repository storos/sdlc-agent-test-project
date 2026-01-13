# SDLC AI Agents - Documentation

## Overview

SDLC AI Agents is an automated development system that integrates JIRA with Git repositories (GitHub, GitLab, etc.) using AI agents to automatically develop code based on JIRA issues. The system uses Claude Code to generate and modify code, managing the entire workflow from issue assignment to pull/merge request creation.

## Quick Start

```bash
# Start all services
docker-compose up --build

# Access services
# - Backoffice UI: http://localhost:3000
# - JIRA Webhook API: http://localhost:8080
# - Configuration API: http://localhost:8081
# - RabbitMQ Management: http://localhost:15672

# Run functional tests (after services are running)
cd tests
./run-tests.sh
```

## Technology Stack

- **Backend**: Go
- **Frontend**: React with TypeScript + Vite
- **UI Library**: Material-UI (MUI)
- **Database**: MongoDB
- **Message Broker**: RabbitMQ
- **Containerization**: Docker
- **AI Engine**: Claude Code

## Services Overview

| Service | Port | Type | Description |
|---------|------|------|-------------|
| JIRA Webhook API | 8080 | REST API | Receives JIRA webhooks |
| Configuration API | 8081 | REST API | Project configuration backend |
| Backoffice UI | 3000 | Web UI | Project management interface |
| Developer Agent Consumer | N/A | Worker | Code generation agent |
| MongoDB | 27017 | Database | Data persistence |
| RabbitMQ | 5672, 15672 | Message Broker | Async job processing |

## Documentation Structure

### Core Documentation
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design, components, and communication patterns
- **[DATABASE.md](DATABASE.md)** - MongoDB collections, indexes, and data models
- **[INTEGRATION.md](INTEGRATION.md)** - JIRA, RabbitMQ, Git, Claude Code integration, and end-to-end workflow
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Docker Compose, environment variables, and monitoring
- **[TESTING.md](TESTING.md)** - Functional testing approach and test execution

## How It Works

1. **Configure Project** - Use Backoffice UI to add project, repositories, and Git tokens
2. **JIRA Issue** - Create/update JIRA issue and set status to "In Development"
3. **Webhook Trigger** - JIRA sends webhook to JIRA Webhook API
4. **Message Queue** - Webhook API publishes message to RabbitMQ
5. **Code Generation** - Developer Agent Consumer picks up message and calls Claude Code
6. **Pull/Merge Request** - Agent creates PR/MR in GitHub/GitLab
7. **Code Review** - Developer reviews and merges the changes

## Key Features

- **Automated Development**: AI generates code based on JIRA issue descriptions
- **Multi-Repository Support**: Handle multiple repositories per project
- **Platform Agnostic**: Works with both GitHub and GitLab
- **Asynchronous Processing**: RabbitMQ ensures reliable message delivery
- **Context-Aware Prompts**: Analyzes repository structure for better code generation
- **Audit Trail**: Tracks all webhooks, developments, and errors in MongoDB

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for UI development)
- Access to JIRA instance
- GitHub or GitLab account with Personal Access Token
- Claude Code session token

## Project Structure

```
/
├── docker-compose.yml
├── requirements.md
├── docs/                          # This documentation
├── jira-webhook-api/             # JIRA webhook receiver
├── developer-agent-consumer/     # Code generation worker
├── configuration-api/            # Configuration backend
├── backoffice-ui/                # Management UI
└── tests/                        # Functional test application
```

## Getting Help

- Check the [Architecture documentation](ARCHITECTURE.md) for system design
- Review [Database documentation](DATABASE.md) for data models
- See [Integration documentation](INTEGRATION.md) for external system setup
- Consult [Deployment documentation](DEPLOYMENT.md) for configuration options
- Read [Testing documentation](TESTING.md) for running functional tests

## Contributing

When developing:
1. Read the relevant documentation files
2. Follow the coding standards in [ARCHITECTURE.md](ARCHITECTURE.md)
3. Update documentation when making changes
4. Run functional tests before submitting PRs

## License

[Add your license information here]
