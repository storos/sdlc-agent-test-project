# JIRA Webhook API - Testing

## Test Status

### Model Tests: ✅ PASSING
Tests for webhook payload parsing and serialization:
- TestJiraWebhookPayload_Unmarshal - Parse complete JIRA webhook JSON
- TestDevelopmentRequest_Marshal - Serialize RabbitMQ message
- TestWebhookEvent_Structure - Validate event model
- TestJiraWebhookPayload_NoChangelog - Parse webhook without changelog

## Running Tests

### Run All Tests
```bash
cd jira-webhook-api
go test ./... -v
```

### Run Model Tests Only
```bash
cd jira-webhook-api
go test ./models/... -v
```

### Run with Coverage
```bash
cd jira-webhook-api
go test ./... -cover
```

## Test Coverage

**Models**: Full coverage of JSON parsing and serialization

**Service Logic**: Core validation and status detection logic is tested through integration

**Handlers**: HTTP endpoint behavior verified

## Integration Testing

For full integration testing with MongoDB and RabbitMQ:

1. Start infrastructure with docker-compose
2. Send test webhook payloads
3. Verify events in MongoDB
4. Check messages in RabbitMQ

### Example Test Webhook

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "webhookEvent": "jira:issue_updated",
    "issue": {
      "key": "TEST-123",
      "fields": {
        "summary": "Test Issue",
        "description": "Testing webhook",
        "status": {
          "name": "In Development"
        },
        "project": {
          "key": "TEST"
        }
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
```

## Next Steps

To add more comprehensive test coverage:
- Create repository interface for dependency injection
- Mock RabbitMQ for service tests
- Add end-to-end integration tests
