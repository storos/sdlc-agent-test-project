# Database Schema

## MongoDB Collections

The system uses MongoDB with three main collections for data persistence.

---

## Projects Collection

**Collection Name**: `projects`

**Purpose**: Store project configurations including JIRA mappings, repository details, and AI development guidelines.

### Schema

```json
{
  "_id": "ObjectId",
  "name": "string",
  "description": "string",
  "scope": "string",
  "jira_project_key": "string",
  "jira_project_name": "string",
  "jira_project_url": "string",
  "repositories": [
    {
      "repository_id": "string",
      "url": "string",
      "description": "string",
      "git_access_token": "string"
    }
  ],
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Key Fields

- **`scope`**: Development guidelines for AI (used in Claude Code prompts)
- **`jira_project_key`**: Unique project key for matching webhooks (used by Consumer to find project config)
- **`repositories[]`**: Each repository has its own `git_access_token`

### Indexes

```javascript
db.projects.createIndex({ "jira_project_key": 1 }, { unique: true })
db.projects.createIndex({ "name": 1 })
db.projects.createIndex({ "created_at": -1 })
```

### Example

```json
{
  "name": "E-Commerce Platform",
  "scope": "Go backend, Gin framework, clean architecture...",
  "jira_project_key": "ECOM",
  "repositories": [
    {
      "url": "https://github.com/company/ecom-api",
      "description": "Main API backend",
      "git_access_token": "ghp_xxxxxxxxxxxx"
    }
  ]
}
```

---

## Webhook Events Collection

**Collection Name**: `webhook_events`

**Purpose**: Audit trail of all JIRA webhook events received by the system.

### Schema

```json
{
  "_id": "ObjectId",
  "jira_issue_id": "string",
  "jira_issue_key": "string",
  "jira_project_key": "string",
  "summary": "string",
  "description": "string",
  "issue_status": "string",
  "issue_type": "string",
  "repository": "string",
  "webhook_payload": "object",
  "created_at": "datetime"
}
```

### Key Fields

- **`repository`**: URL from JIRA components field (used for matching)
- **`webhook_payload`**: Complete payload for audit trail
- **`issue_status`**: Typically "In Development"

### Indexes

```javascript
db.webhook_events.createIndex({ "jira_issue_key": 1 })
db.webhook_events.createIndex({ "jira_project_key": 1 })
db.webhook_events.createIndex({ "created_at": -1 })
db.webhook_events.createIndex({ "issue_status": 1, "created_at": -1 })
```

### Example

```json
{
  "jira_issue_key": "ECOM-123",
  "summary": "Add user authentication endpoint",
  "issue_status": "In Development",
  "repository": "https://github.com/company/ecom-api",
  "webhook_payload": { ... }
}
```

---

## Developments Collection

**Collection Name**: `developments`

**Purpose**: Track all AI-driven development activities, including status, generated code details, and PR/MR links.

### Schema

```json
{
  "_id": "ObjectId",
  "project_id": "ObjectId",
  "jira_issue_id": "string",
  "jira_issue_key": "string",
  "jira_project_key": "string",
  "summary": "string",
  "description": "string",
  "repository_url": "string",
  "git_access_token": "string",
  "branch_name": "string",
  "status": "string",
  "pr_mr_url": "string",
  "development_details": "string",
  "error_message": "string",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Key Fields

- **`project_id`**: Reference to projects collection
- **`status`**: `ready` → `completed` or `failed`
- **`repository_url`**: Matched repository from JIRA components
- **`pr_mr_url`**: Generated PR/MR link (when completed)
- **`development_details`**: Claude Code summary (when completed)
- **`error_message`**: Failure details (when failed)

### Indexes

```javascript
db.developments.createIndex({ "project_id": 1 })
db.developments.createIndex({ "jira_issue_key": 1 })
db.developments.createIndex({ "jira_project_key": 1 })
db.developments.createIndex({ "status": 1, "created_at": -1 })
db.developments.createIndex({ "project_id": 1, "status": 1, "created_at": -1 })
```

### Examples

**Completed**:
```json
{
  "jira_issue_key": "ECOM-123",
  "branch_name": "feature/ECOM-123",
  "status": "completed",
  "pr_mr_url": "https://github.com/company/ecom-api/pull/42",
  "development_details": "Created /api/auth/login endpoint..."
}
```

**Failed**:
```json
{
  "jira_issue_key": "ECOM-124",
  "status": "failed",
  "error_message": "Repository not found in project configuration"
}
```

---

## Database Initialization

### Creating Indexes

Run these commands after MongoDB is running:

```javascript
// Connect to MongoDB
use sdlc_agents

// Projects collection indexes
db.projects.createIndex({ "jira_project_key": 1 }, { unique: true })
db.projects.createIndex({ "name": 1 })
db.projects.createIndex({ "created_at": -1 })

// Webhook Events collection indexes
db.webhook_events.createIndex({ "jira_issue_id": 1 })
db.webhook_events.createIndex({ "jira_issue_key": 1 })
db.webhook_events.createIndex({ "jira_project_key": 1 })
db.webhook_events.createIndex({ "created_at": -1 })
db.webhook_events.createIndex({ "issue_status": 1, "created_at": -1 })

// Developments collection indexes
db.developments.createIndex({ "project_id": 1 })
db.developments.createIndex({ "jira_issue_id": 1 })
db.developments.createIndex({ "jira_issue_key": 1 })
db.developments.createIndex({ "jira_project_key": 1 })
db.developments.createIndex({ "status": 1, "created_at": -1 })
db.developments.createIndex({ "project_id": 1, "status": 1, "created_at": -1 })
```

### Database Connection

**Connection String Format**:
```
mongodb://mongodb:27017/sdlc_agents
```

**Environment Variable**:
```bash
MONGODB_URL=mongodb://mongodb:27017/sdlc_agents
```

---

## Query Examples

### Find Project by JIRA Project Key

```javascript
db.projects.findOne({ "jira_project_key": "ECOM" })
```

**Used by**: Developer Agent Consumer to retrieve project configuration.

### Find All Webhook Events for an Issue

```javascript
db.webhook_events.find({ "jira_issue_key": "ECOM-123" }).sort({ "created_at": -1 })
```

**Used for**: Audit trail and debugging.

### Find Active Developments

```javascript
db.developments.find({ "status": "ready" }).sort({ "created_at": -1 })
```

**Used for**: Monitoring processing queue.

### Find Failed Developments for a Project

```javascript
db.developments.find({
  "project_id": ObjectId("507f1f77bcf86cd799439011"),
  "status": "failed"
}).sort({ "created_at": -1 })
```

**Used for**: Error monitoring and manual retry.

### Find Recent Developments

```javascript
db.developments.find().sort({ "created_at": -1 }).limit(10)
```

**Used for**: Dashboard and monitoring.

---

## Data Relationships

```
projects (1) ──< (N) developments
    │
    └─ jira_project_key matches jira_project_key in webhook_events
                                   and developments
```

- One **project** can have many **developments**
- **projects** are found by `jira_project_key` (used by Consumer for lookup)
- **webhook_events** and **developments** are linked by `jira_issue_key`
- **projects** and **developments** are linked by `project_id` (ObjectId reference)

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md) - System design and components
- [Integration Details](INTEGRATION.md) - External system integration
- [Deployment Guide](DEPLOYMENT.md) - Docker Compose and environment variables
- [Testing Guide](TESTING.md) - Functional testing approach
