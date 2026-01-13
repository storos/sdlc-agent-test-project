# API Documentation

Complete API reference for SDLC AI Agents system.

## Configuration API

Base URL: `http://localhost:8081`

### Projects

#### List All Projects

```http
GET /api/projects
```

**Response** `200 OK`
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "name": "E-Commerce Platform",
    "description": "Main e-commerce application",
    "scope": "Focus on API development, follow REST principles",
    "jira_project_key": "ECOM",
    "jira_project_name": "E-Commerce",
    "jira_project_url": "https://jira.company.com/projects/ECOM",
    "repositories": [
      {
        "repository_id": "repo-1",
        "url": "https://github.com/company/ecommerce-api",
        "description": "Backend API",
        "git_access_token": "ghp_***"
      }
    ],
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-01-15T10:00:00Z"
  }
]
```

#### Get Project by ID

```http
GET /api/projects/:id
```

**Parameters**
- `id` (path) - Project MongoDB ObjectID

**Response** `200 OK` - Returns single project object
**Response** `404 Not Found` - Project not found

#### Query Projects by JIRA Key

```http
GET /api/projects?jira_project_key={key}
```

**Parameters**
- `jira_project_key` (query) - JIRA project key (e.g., "ECOM")

**Response** `200 OK` - Returns array of matching projects

#### Create Project

```http
POST /api/projects
Content-Type: application/json
```

**Request Body**
```json
{
  "name": "E-Commerce Platform",
  "description": "Main e-commerce application",
  "scope": "Focus on API development, follow REST principles",
  "jira_project_key": "ECOM",
  "jira_project_name": "E-Commerce",
  "jira_project_url": "https://jira.company.com/projects/ECOM",
  "repositories": [
    {
      "repository_id": "repo-1",
      "url": "https://github.com/company/ecommerce-api",
      "description": "Backend API",
      "git_access_token": "ghp_xxxxxxxxxxxx"
    }
  ]
}
```

**Validation Rules**
- `name` - Required, string, 1-200 characters
- `jira_project_key` - Required, string, uppercase letters
- `repositories` - Required, array with at least 1 repository
- `repositories[].url` - Required, valid Git URL
- `repositories[].git_access_token` - Required, non-empty string

**Response** `201 Created`
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "E-Commerce Platform",
  ...
}
```

**Response** `400 Bad Request` - Validation error
```json
{
  "error": "Validation failed: name is required"
}
```

#### Update Project

```http
PUT /api/projects/:id
Content-Type: application/json
```

**Parameters**
- `id` (path) - Project MongoDB ObjectID

**Request Body** - Same as Create Project

**Response** `200 OK` - Returns updated project
**Response** `404 Not Found` - Project not found
**Response** `400 Bad Request` - Validation error

#### Delete Project

```http
DELETE /api/projects/:id
```

**Parameters**
- `id` (path) - Project MongoDB ObjectID

**Response** `204 No Content` - Project deleted successfully
**Response** `404 Not Found` - Project not found

### Repositories

#### List Project Repositories

```http
GET /api/projects/:id/repositories
```

**Parameters**
- `id` (path) - Project MongoDB ObjectID

**Response** `200 OK`
```json
[
  {
    "repository_id": "repo-1",
    "url": "https://github.com/company/ecommerce-api",
    "description": "Backend API",
    "git_access_token": "ghp_***"
  }
]
```

#### Add Repository to Project

```http
POST /api/projects/:id/repositories
Content-Type: application/json
```

**Request Body**
```json
{
  "repository_id": "repo-2",
  "url": "https://github.com/company/ecommerce-frontend",
  "description": "Frontend application",
  "git_access_token": "ghp_xxxxxxxxxxxx"
}
```

**Response** `201 Created` - Returns updated project
**Response** `400 Bad Request` - Validation error

#### Update Repository

```http
PUT /api/repositories/:repository_id
Content-Type: application/json
```

**Parameters**
- `repository_id` (path) - Repository ID

**Request Body**
```json
{
  "url": "https://github.com/company/ecommerce-frontend",
  "description": "Updated description",
  "git_access_token": "ghp_new_token"
}
```

**Response** `200 OK` - Returns updated project
**Response** `404 Not Found` - Repository not found

#### Delete Repository

```http
DELETE /api/repositories/:repository_id
```

**Parameters**
- `repository_id` (path) - Repository ID

**Response** `200 OK` - Returns updated project
**Response** `404 Not Found` - Repository not found

### Health Check

```http
GET /health
```

**Response** `200 OK`
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

---

## JIRA Webhook API

Base URL: `http://localhost:8080`

### Webhooks

#### Receive JIRA Webhook

```http
POST /webhook
Content-Type: application/json
```

**Request Body** - JIRA webhook payload

```json
{
  "webhookEvent": "jira:issue_updated",
  "issue_event_type_name": "issue_generic",
  "issue": {
    "id": "10001",
    "key": "ECOM-123",
    "fields": {
      "summary": "Add payment gateway integration",
      "description": "Integrate Stripe payment gateway with checkout flow",
      "status": {
        "name": "In Development",
        "id": "3"
      },
      "project": {
        "key": "ECOM",
        "name": "E-Commerce"
      },
      "issuetype": {
        "name": "Task"
      }
    }
  },
  "changelog": {
    "items": [
      {
        "field": "status",
        "fieldtype": "jira",
        "from": "1",
        "fromString": "To Do",
        "to": "3",
        "toString": "In Development"
      }
    ]
  }
}
```

**Processing Logic**
1. Validates webhook payload structure
2. Checks if status changed to "In Development"
3. Stores event in MongoDB `webhook_events` collection
4. Publishes message to RabbitMQ `develop` queue

**Response** `200 OK`
```json
{
  "message": "Webhook processed successfully"
}
```

**Response** `200 OK` (Ignored)
```json
{
  "message": "Webhook received but ignored"
}
```
*Returned when status is not "In Development"*

**Response** `400 Bad Request`
```json
{
  "error": "Invalid webhook payload"
}
```

**Response** `500 Internal Server Error`
```json
{
  "error": "Failed to process webhook"
}
```

### Health Check

```http
GET /health
```

**Response** `200 OK`
```json
{
  "status": "healthy",
  "mongodb": "connected",
  "rabbitmq": "connected"
}
```

---

## RabbitMQ Messages

### Development Request Queue

**Queue**: `develop` (Durable)
**Exchange**: `sdlc_events` (Topic, Durable)
**Routing Key**: `develop`

**Message Format**
```json
{
  "jira_issue_id": "10001",
  "jira_issue_key": "ECOM-123",
  "jira_project_key": "ECOM",
  "summary": "Add payment gateway integration",
  "description": "Integrate Stripe payment gateway with checkout flow",
  "repository": "https://github.com/company/ecommerce-api"
}
```

**Consumer**: Developer Agent Consumer
**Prefetch**: 1
**Acknowledgment**: Manual

### Error Queue

**Queue**: `develop_error` (Durable)

**Message Format**
```json
{
  "original_message": "{...}",
  "error": "Failed to clone repository: authentication failed",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

---

## MongoDB Collections

### projects

**Indexes**
- `_id` (unique)
- `jira_project_key` (unique)

**Document Schema**
```javascript
{
  _id: ObjectId,
  name: String,
  description: String,
  scope: String,
  jira_project_key: String,
  jira_project_name: String,
  jira_project_url: String,
  repositories: [
    {
      repository_id: String,
      url: String,
      description: String,
      git_access_token: String
    }
  ],
  created_at: ISODate,
  updated_at: ISODate
}
```

### webhook_events

**Indexes**
- `_id` (unique)
- `jira_issue_key`
- `received_at`

**Document Schema**
```javascript
{
  _id: ObjectId,
  jira_issue_id: String,
  jira_issue_key: String,
  jira_project_key: String,
  summary: String,
  description: String,
  status: String,
  previous_status: String,
  event_type: String,
  received_at: ISODate,
  processed_at: ISODate (optional)
}
```

### developments

**Indexes**
- `_id` (unique)
- `jira_issue_key` (unique)
- `status`
- `created_at`

**Document Schema**
```javascript
{
  _id: ObjectId,
  jira_issue_id: String,
  jira_issue_key: String,
  jira_project_key: String,
  repository_url: String,
  branch_name: String,
  pr_mr_url: String (optional),
  status: String, // "ready", "completed", "failed"
  development_details: String (optional),
  error_message: String (optional),
  created_at: ISODate,
  completed_at: ISODate (optional)
}
```

---

## Error Codes

### Configuration API

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Resource created |
| 204 | Resource deleted |
| 400 | Bad request / Validation error |
| 404 | Resource not found |
| 500 | Internal server error |

### JIRA Webhook API

| Code | Description |
|------|-------------|
| 200 | Webhook processed or ignored |
| 400 | Invalid webhook payload |
| 500 | Processing error |

---

## Rate Limiting

Currently no rate limiting is implemented. For production deployment, consider:
- Configuration API: 1000 requests/hour per IP
- JIRA Webhook API: 100 requests/minute per IP

---

## Authentication

### Current Status
- Configuration API: No authentication (internal use)
- JIRA Webhook API: No authentication (webhook validation recommended)

### Recommendations for Production
1. Add API key authentication for Configuration API
2. Add JIRA webhook secret validation
3. Add JWT tokens for service-to-service communication
4. Use mutual TLS for internal communication

---

## Examples

### Complete Workflow Example

```bash
# 1. Create project
curl -X POST http://localhost:8081/api/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "jira_project_key": "MYPROJ",
    "jira_project_name": "My Project",
    "jira_project_url": "https://jira.example.com/projects/MYPROJ",
    "scope": "Backend API development",
    "repositories": [{
      "repository_id": "main-repo",
      "url": "https://github.com/myorg/myproject",
      "description": "Main repository",
      "git_access_token": "ghp_xxxxxxxxxxxx"
    }]
  }'

# 2. Send webhook (simulate JIRA)
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "webhookEvent": "jira:issue_updated",
    "issue": {
      "id": "10001",
      "key": "MYPROJ-123",
      "fields": {
        "summary": "Add user authentication",
        "description": "Implement JWT authentication",
        "status": {"name": "In Development"},
        "project": {"key": "MYPROJ"}
      }
    },
    "changelog": {
      "items": [{
        "field": "status",
        "fromString": "To Do",
        "toString": "In Development"
      }]
    }
  }'

# 3. Check development status
docker exec sdlc-mongodb mongosh sdlc_agent \
  --eval 'db.developments.find({jira_issue_key: "MYPROJ-123"}).pretty()'
```

---

## See Also

- [Architecture Documentation](ARCHITECTURE.md)
- [Database Schema](DATABASE.md)
- [Integration Guide](INTEGRATION.md)
- [Testing Guide](TESTING.md)
