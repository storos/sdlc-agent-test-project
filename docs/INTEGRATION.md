# Integration Guide

## Overview

This document covers all external integrations and the complete end-to-end workflow:
- JIRA webhook integration
- RabbitMQ message broker configuration
- Git repository integration (GitHub/GitLab)
- Claude Code AI integration
- Complete workflow from JIRA issue to pull/merge request

---

## JIRA Integration

### Webhook Setup

JIRA sends webhook events when issue status changes to "In Development".

**Webhook Event**: `jira:issue_updated`

### Webhook Payload

JIRA sends webhook payloads containing issue details. Key fields extracted:

- `issue.id` / `issue.key` - Issue identifiers
- `issue.fields.summary` / `description` - Task details
- `issue.fields.project.*` - Project information
- `issue.fields.components[]` - Must include repository URL (name: "repository")
- `changelog.items[]` - Status change from/to values

**Important**: The repository URL in components must match a configured repository in the project settings.

### Status Detection

Development workflow triggers when:
- Webhook event is `jira:issue_updated`
- Status changed to "In Development" (via changelog)

---

## RabbitMQ Configuration

### Exchange

**Name**: `webhook.development.request`
**Type**: Topic
**Durable**: Yes
**Created by**: JIRA Webhook API on startup

**Purpose**: Receive messages from JIRA Webhook API and route to development queue.

### Queues

#### Main Queue
**Name**: `develop`
**Durable**: Yes
**Created by**: Developer Agent Consumer on startup
**Binding**: Bound to exchange `webhook.development.request`

#### Error Queue
**Name**: `develop_error`
**Durable**: Yes
**Created by**: Developer Agent Consumer on startup
**Purpose**: Store failed messages for manual review (no automatic retry)

### Message Format

```json
{
  "jira_issue_id": "10001",
  "jira_issue_key": "PROJ-123",
  "jira_project_key": "PROJ",
  "summary": "Issue summary text",
  "description": "Issue description details",
  "repository": "https://github.com/company/ecom-api"
}
```

**Note**: The `repository` field contains the repository URL from JIRA components field. The Developer Agent Consumer will match this URL against project repositories.

### Message Properties

- **Delivery Mode**: Persistent (messages survive broker restart)
- **TTL**: No expiration (messages stay until consumed)
- **Prefetch**: 1 (consumer fetches one message at a time)
- **Concurrency**: Single consumer

### Error Handling

**On Processing Failure**:
1. Update development record in MongoDB (status: `failed`, error message)
2. Publish message to `develop_error` queue
3. Acknowledge original message from `develop` queue
4. **No automatic retry** (prevents infinite loops)

**Failed Message Review**:
- Messages in `develop_error` queue can be reviewed manually
- Can be republished to `develop` queue after fixing issues

---

## Git Repository Integration

### Supported Platforms

- GitHub
- GitLab
- Any Git hosting platform with HTTP(S) access

### Authentication

**Method**: Personal Access Tokens (PAT)
**Storage**: Per-repository in MongoDB (`projects.repositories[].git_access_token`)

**Required Permissions**:
- **GitHub**: `repo`, `workflow` (format: `ghp_...`)
- **GitLab**: `api`, `read_repository`, `write_repository` (format: `glpat-...`)

### Git Operations

```bash
# Clone repository
git clone https://{token}@github.com/owner/repo.git /tmp/sdlc-{issue_key}/repo

# Create and switch to feature branch
git checkout -b feature/{jira_issue_key}

# Commit changes
git add .
git commit -m "[{jira_issue_key}] {description}"

# Push branch
git push origin feature/{jira_issue_key}
```

**Branch Pattern**: `feature/PROJ-123`
**Commit Pattern**: `[PROJ-123] Implemented user authentication feature`

### Pull/Merge Request Creation

**GitHub**: `POST /repos/{owner}/{repo}/pulls`
- Auth: `Authorization: token {git_access_token}`
- Title: `[{jira_issue_key}] {summary}`
- Head: `feature/{jira_issue_key}`, Base: `main`

**GitLab**: `POST /api/v4/projects/{project_id}/merge_requests`
- Auth: `PRIVATE-TOKEN: {git_access_token}`
- Source: `feature/{jira_issue_key}`, Target: `main`

---

## Claude Code Integration

### Authentication

**Method**: Session token from browser

**Storage**: Environment variable `CLAUDE_CODE_SESSION_TOKEN`

**Format**: `sk-xxxxxxxxxxxxx`

### Prompt Building Strategy

The Consumer builds structured prompts combining:

**1. Project Context**: Name, description, scope/guidelines (from Configuration API)
**2. Task Requirements**: JIRA issue key, summary, description
**3. Repository Analysis**: Structure, patterns, key files

**Prompt Structure**:
```markdown
# Development Task: {jira_issue_key}
## Project Context
{project_name}, {project_description}, {project_scope}

## Task Requirements
{jira_issue_key}: {summary}
{description}

## Repository Information
{repository_url}: {repository_description}
Structure: {analyzed_structure}

## Instructions
Implement changes following project guidelines and existing patterns.
Provide summary of changes.
```

**Repository Analysis** identifies:
- Entry points (main.go, index.ts)
- Directory structure (handlers/, models/, services/)
- Code patterns and conventions

Response summary stored in `development_details` for audit trail.

---

## End-to-End Workflow

### Workflow Phases

**Phase 1: Webhook Reception (Steps 1-7)**
1. Developer updates JIRA issue → "In Development"
2. JIRA webhook → JIRA Webhook API
3. API validates payload, stores to `webhook_events`, publishes to RabbitMQ
4. Returns HTTP 200 OK

**Phase 2: Message Routing (Steps 8-9)**
5. RabbitMQ routes message to `develop` queue
6. Developer Agent Consumer picks up message (prefetch: 1)

**Phase 3: Configuration & Setup (Steps 10-14)**
7. Create `developments` record (status: `ready`)
8. Fetch project config from Configuration API (lookup by `jira_project_key`)
9. Match repository URL from JIRA with configured repositories
10. Clone matched repository to `/tmp/sdlc-{jira_issue_key}/repo`

**Phase 4: Code Generation (Steps 15-18)**
11. Create feature branch: `feature/{jira_issue_key}`
12. Analyze repository structure and patterns
13. Build prompt with project context + task + repo analysis
14. Call Claude Code API for code generation

**Phase 5: Git Operations (Steps 19-22)**
15. Switch to feature branch
16. Commit changes: `[{jira_issue_key}] {description}`
17. Push branch to remote
18. Create PR/MR via GitHub/GitLab API

**Phase 6: Cleanup & Complete (Steps 23-26)**
19. Update `developments` record (status: `completed`)
20. Delete temporary folder
21. Acknowledge RabbitMQ message
22. PR/MR ready for review

### Error Handling

On any error during processing:
1. Update `developments` record → status: `failed` + error message
2. Publish to `develop_error` queue (no automatic retry)
3. Acknowledge original message, delete temp folder

**Common Errors**: Repository not found, configuration missing, Git auth failure, Claude Code API failure, PR/MR creation failure

---

## Workflow Diagram

```
┌────────────────────────────────────────────────────────────┐
│                     JIRA Issue Status                      │
│                   Changed to "In Development"              │
└───────────────────────────┬────────────────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ JIRA Webhook  │
                    └───────┬───────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  JIRA Webhook API             │
            │  1. Validate payload          │
            │  2. Store to webhook_events   │
            │  3. Publish to RabbitMQ       │
            └───────────────┬───────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │   RabbitMQ    │
                    │ develop queue │
                    └───────┬───────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │ Developer Agent Consumer      │
            │  1. Consume message           │
            │  2. Create development record │
            │  3. Get project config        │
            │  4. Match repository URL      │
            │  5. Clone repository          │
            │  6. Create branch             │
            │  7. Analyze structure         │
            │  8. Build prompt              │
            │  9. Call Claude Code          │
            │ 10. Switch to feature branch  │
            │ 11. Commit changes            │
            │ 12. Push branch               │
            │ 13. Create PR/MR              │
            │ 14. Update development        │
            │ 15. Cleanup                   │
            │ 16. Acknowledge message       │
            └───────────────┬───────────────┘
                            │
                ┌───────────┴───────────┐
                ▼                       ▼
        ┌───────────────┐       ┌──────────────┐
        │ GitHub/GitLab │       │   MongoDB    │
        │  Pull Request │       │ developments │
        └───────────────┘       └──────────────┘
```

---

## Performance Considerations

- **JIRA Webhook API**: < 100ms response, lightweight, horizontally scalable
- **Developer Agent Consumer**: Minutes per message, single consumer (prevents Git conflicts), resource-intensive
- **RabbitMQ**: Persistent messages, prevents backpressure, single consumer

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md) - System design
- [Database Schema](DATABASE.md) - MongoDB collections
- [Deployment Guide](DEPLOYMENT.md) - Docker Compose setup
- [Testing Guide](TESTING.md) - Functional testing approach
