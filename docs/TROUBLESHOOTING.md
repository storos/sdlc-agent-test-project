# Troubleshooting Guide

Common issues and solutions for SDLC AI Agents system.

## Table of Contents

- [Services Won't Start](#services-wont-start)
- [MongoDB Issues](#mongodb-issues)
- [RabbitMQ Issues](#rabbitmq-issues)
- [Configuration API Issues](#configuration-api-issues)
- [JIRA Webhook API Issues](#jira-webhook-api-issues)
- [Developer Agent Consumer Issues](#developer-agent-consumer-issues)
- [Backoffice UI Issues](#backoffice-ui-issues)
- [Performance Issues](#performance-issues)
- [Debugging Tips](#debugging-tips)

---

## Services Won't Start

### Problem: Docker Compose fails to start

**Symptoms**
```
ERROR: Service 'xyz' failed to build
ERROR: for xyz  Container exited with code 1
```

**Solutions**

1. **Check Docker is running**
```bash
docker ps
# If error, start Docker Desktop or Docker daemon
```

2. **Check port conflicts**
```bash
# Check if ports are already in use
lsof -i :27017  # MongoDB
lsof -i :5672   # RabbitMQ
lsof -i :8080   # JIRA Webhook API
lsof -i :8081   # Configuration API
lsof -i :3000   # Backoffice UI
lsof -i :15672  # RabbitMQ Management

# Kill process using port
kill -9 <PID>
```

3. **Clean start**
```bash
docker-compose down -v
docker-compose up --build
```

4. **Check logs**
```bash
docker-compose logs <service-name>
```

### Problem: Service health check failing

**Symptoms**
```
service is unhealthy
```

**Solutions**

1. **Check service logs**
```bash
docker-compose logs <service-name>
```

2. **Test health endpoint**
```bash
curl http://localhost:8081/health  # Configuration API
curl http://localhost:8080/health  # JIRA Webhook API
```

3. **Verify dependencies**
```bash
# MongoDB must be healthy before other services
docker-compose ps
```

4. **Increase health check timeout**
Edit `docker-compose.yml`:
```yaml
healthcheck:
  interval: 10s
  timeout: 10s  # Increase this
  retries: 10   # Increase this
```

---

## MongoDB Issues

### Problem: Cannot connect to MongoDB

**Symptoms**
```
Failed to connect to MongoDB: connection timeout
Failed to ping MongoDB
```

**Solutions**

1. **Check MongoDB is running**
```bash
docker ps | grep mongodb
docker-compose ps mongodb
```

2. **Check MongoDB logs**
```bash
docker-compose logs mongodb
```

3. **Test MongoDB connection**
```bash
docker exec -it sdlc-mongodb mongosh --eval "db.runCommand({ping: 1})"
```

4. **Check connection string**
- Inside containers: `mongodb://mongodb:27017`
- From host: `mongodb://localhost:27017`

5. **Restart MongoDB**
```bash
docker-compose restart mongodb
```

### Problem: Database not initialized

**Symptoms**
```
Collection 'projects' does not exist
```

**Solutions**

1. **Check init script ran**
```bash
docker-compose logs mongodb | grep init-db
```

2. **Manually initialize**
```bash
docker exec -i sdlc-mongodb mongosh sdlc_agent < db/init-db.js
docker exec -i sdlc-mongodb mongosh sdlc_agent < db/seed-data.js
```

3. **Verify collections exist**
```bash
docker exec -it sdlc-mongodb mongosh sdlc_agent --eval "show collections"
```

### Problem: MongoDB out of disk space

**Symptoms**
```
No space left on device
```

**Solutions**

1. **Check disk usage**
```bash
docker system df
```

2. **Clean up Docker resources**
```bash
docker system prune -a --volumes
```

3. **Remove unused volumes**
```bash
docker volume ls
docker volume rm <volume-name>
```

---

## RabbitMQ Issues

### Problem: Cannot connect to RabbitMQ

**Symptoms**
```
Failed to connect to RabbitMQ
dial tcp: connection refused
```

**Solutions**

1. **Check RabbitMQ is running**
```bash
docker ps | grep rabbitmq
docker-compose ps rabbitmq
```

2. **Check RabbitMQ logs**
```bash
docker-compose logs rabbitmq
```

3. **Test RabbitMQ connection**
```bash
docker exec sdlc-rabbitmq rabbitmq-diagnostics ping
```

4. **Check RabbitMQ Management UI**
Open http://localhost:15672 (guest/guest)

5. **Restart RabbitMQ**
```bash
docker-compose restart rabbitmq
```

### Problem: Messages not being consumed

**Symptoms**
- Messages stuck in queue
- Consumer not processing

**Solutions**

1. **Check queue status**
Visit http://localhost:15672 → Queues tab

2. **Check consumer is connected**
```bash
docker-compose logs developer-agent-consumer | grep "consumer started"
```

3. **Check for errors**
```bash
docker-compose logs developer-agent-consumer | grep ERROR
```

4. **Purge queue (development only)**
```bash
docker exec sdlc-rabbitmq rabbitmqctl purge_queue develop
```

5. **Check message format**
Ensure message matches expected schema (see API.md)

### Problem: Error queue filling up

**Symptoms**
- `develop_error` queue has many messages

**Solutions**

1. **View error messages**
Open http://localhost:15672 → Queues → develop_error → Get Messages

2. **Common error causes**
- Invalid message format
- Missing project configuration
- Git authentication failure
- Claude Code API error

3. **Fix root cause and reprocess**
- Fix the configuration issue
- Move messages back to main queue manually or delete them

---

## Configuration API Issues

### Problem: API returns 404 for existing project

**Symptoms**
```json
{"error": "Project not found"}
```

**Solutions**

1. **Verify project exists**
```bash
docker exec -it sdlc-mongodb mongosh sdlc_agent \
  --eval 'db.projects.find().pretty()'
```

2. **Check project ID format**
Must be valid MongoDB ObjectID (24 hex characters)

3. **Use JIRA key query instead**
```bash
curl "http://localhost:8081/api/projects?jira_project_key=MYPROJ"
```

### Problem: Cannot create project

**Symptoms**
```json
{"error": "Validation failed: ..."}
```

**Solutions**

1. **Check required fields**
- name, jira_project_key, repositories[] are required

2. **Validate repository URL**
Must be valid Git URL (https://...)

3. **Check JSON format**
Use a JSON validator to ensure proper formatting

4. **View API logs**
```bash
docker-compose logs configuration-api
```

### Problem: API slow to respond

**Solutions**

1. **Check MongoDB connection**
```bash
docker-compose logs configuration-api | grep "MongoDB"
```

2. **Check database indexes**
```bash
docker exec -it sdlc-mongodb mongosh sdlc_agent \
  --eval 'db.projects.getIndexes()'
```

3. **Monitor resource usage**
```bash
docker stats
```

---

## JIRA Webhook API Issues

### Problem: Webhook returns 400 Bad Request

**Symptoms**
```json
{"error": "Invalid webhook payload"}
```

**Solutions**

1. **Verify JIRA webhook payload**
Check payload matches expected structure (see API.md)

2. **Test with curl**
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d @test-webhook.json
```

3. **Check API logs**
```bash
docker-compose logs jira-webhook-api | tail -50
```

### Problem: Webhook received but not processing

**Symptoms**
```json
{"message": "Webhook received but ignored"}
```

**Causes**
- Status is not "In Development"
- No changelog showing status change
- Status changed from something other than expected

**Solutions**

1. **Verify status in payload**
```json
{
  "issue": {
    "fields": {
      "status": {
        "name": "In Development"  // Must be exactly this
      }
    }
  }
}
```

2. **Check changelog**
```json
{
  "changelog": {
    "items": [{
      "field": "status",
      "toString": "In Development"
    }]
  }
}
```

### Problem: Webhook processed but not in database

**Solutions**

1. **Check MongoDB connection**
```bash
docker-compose logs jira-webhook-api | grep MongoDB
```

2. **Verify webhook_events collection**
```bash
docker exec -it sdlc-mongodb mongosh sdlc_agent \
  --eval 'db.webhook_events.find().sort({received_at: -1}).limit(5).pretty()'
```

3. **Check for errors**
```bash
docker-compose logs jira-webhook-api | grep ERROR
```

---

## Developer Agent Consumer Issues

### Problem: Consumer not processing messages

**Symptoms**
- Messages sit in queue
- No log activity

**Solutions**

1. **Check consumer is running**
```bash
docker-compose ps developer-agent-consumer
```

2. **Check consumer logs**
```bash
docker-compose logs developer-agent-consumer | tail -100
```

3. **Verify RabbitMQ connection**
```bash
docker-compose logs developer-agent-consumer | grep RabbitMQ
```

4. **Check queue binding**
Visit http://localhost:15672 → Queues → develop → Bindings

### Problem: Development fails with "project not found"

**Solutions**

1. **Verify project exists**
```bash
curl "http://localhost:8081/api/projects?jira_project_key=MYPROJ"
```

2. **Check JIRA project key matches**
Webhook `issue.fields.project.key` must match project `jira_project_key`

3. **Check Configuration API is accessible**
```bash
docker exec sdlc-developer-agent-consumer curl http://configuration-api:8081/health
```

### Problem: Git clone fails

**Symptoms**
```
Failed to clone repository: authentication failed
```

**Solutions**

1. **Verify Git access token**
- Token must have repository access
- Token must not be expired
- For GitHub: Personal Access Token with repo scope
- For GitLab: Project Access Token with write_repository scope

2. **Test token manually**
```bash
git clone https://oauth2:TOKEN@github.com/user/repo.git
```

3. **Check repository URL format**
- Must be HTTPS (not SSH)
- Must be complete URL

4. **Update token in database**
```bash
docker exec -it sdlc-mongodb mongosh sdlc_agent
db.projects.updateOne(
  {jira_project_key: "MYPROJ"},
  {$set: {"repositories.0.git_access_token": "new-token"}}
)
```

### Problem: Claude Code API timeout

**Symptoms**
```
Failed to call Claude Code API: timeout
Context deadline exceeded
```

**Solutions**

1. **Check Claude Code API is accessible**
```bash
curl -X POST http://claude-api:8000/generate \
  -H "Content-Type: application/json" \
  -d '{"prompt": "test"}'
```

2. **Verify session token**
```bash
echo $CLAUDE_CODE_SESSION_TOKEN
```

3. **Increase timeout** (if needed)
Edit `developer-agent-consumer/services/claude_service.go`:
```go
httpClient: &http.Client{
    Timeout: 600 * time.Second, // 10 minutes
}
```

### Problem: PR/MR creation fails

**Symptoms**
```
Failed to create PR: 401 Unauthorized
Failed to create MR: 403 Forbidden
```

**Solutions**

1. **Verify Git access token has PR/MR permissions**
- GitHub: repo scope
- GitLab: api scope or write_repository

2. **Test API access manually**
```bash
# GitHub
curl -H "Authorization: token YOUR_TOKEN" \
  https://api.github.com/user

# GitLab
curl -H "PRIVATE-TOKEN: YOUR_TOKEN" \
  https://gitlab.com/api/v4/user
```

3. **Check repository URL is correct**
Must match format: https://github.com/owner/repo or https://gitlab.com/owner/repo

---

## Backoffice UI Issues

### Problem: UI not loading

**Symptoms**
- Blank page
- "Cannot GET /" error

**Solutions**

1. **Check service is running**
```bash
docker-compose ps backoffice-ui
```

2. **Check logs**
```bash
docker-compose logs backoffice-ui
```

3. **Verify port mapping**
Should be accessible at http://localhost:3000

4. **Clear browser cache**
Hard refresh: Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)

### Problem: API calls failing

**Symptoms**
- "Network Error"
- CORS errors in console

**Solutions**

1. **Check Configuration API is accessible**
```bash
curl http://localhost:8081/api/projects
```

2. **Verify API URL configuration**
Check `backoffice-ui/.env` or build args in `docker-compose.yml`:
```yaml
args:
  VITE_API_URL: http://localhost:8081
```

3. **Check nginx proxy configuration**
In `backoffice-ui/nginx.conf`, verify `/api` proxy_pass

4. **Rebuild UI**
```bash
docker-compose up --build backoffice-ui
```

---

## Performance Issues

### Problem: System running slow

**Solutions**

1. **Check resource usage**
```bash
docker stats
```

2. **Check for memory leaks**
```bash
docker-compose logs | grep "out of memory"
```

3. **Optimize MongoDB queries**
```bash
# Check slow queries
docker exec -it sdlc-mongodb mongosh sdlc_agent \
  --eval 'db.setProfilingLevel(1, {slowms: 100})'
```

4. **Add database indexes**
See `db/init-db.js` for index creation

5. **Scale consumer**
```bash
docker-compose up --scale developer-agent-consumer=3
```

### Problem: High CPU usage

**Solutions**

1. **Identify culprit**
```bash
docker stats --no-stream
```

2. **Check for infinite loops in logs**
```bash
docker-compose logs | grep ERROR
```

3. **Reduce RabbitMQ prefetch**
Edit `developer-agent-consumer/consumer/rabbitmq_consumer.go`:
```go
prefetchCount = 1  // Lower value
```

---

## Debugging Tips

### Enable Debug Logging

**Configuration API**
```yaml
environment:
  LOG_LEVEL: debug
```

**JIRA Webhook API**
```yaml
environment:
  LOG_LEVEL: debug
```

**Developer Agent Consumer**
```yaml
environment:
  LOG_LEVEL: debug
```

### View Real-time Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f developer-agent-consumer

# Filter for errors
docker-compose logs | grep ERROR

# Filter for specific issue key
docker-compose logs | grep "MYPROJ-123"
```

### Database Debugging

```bash
# Access MongoDB shell
docker exec -it sdlc-mongodb mongosh sdlc_agent

# Query recent webhook events
db.webhook_events.find().sort({received_at: -1}).limit(10).pretty()

# Query developments by status
db.developments.find({status: "failed"}).pretty()

# Count documents
db.projects.countDocuments()
db.webhook_events.countDocuments()
db.developments.countDocuments()

# Find by JIRA key
db.developments.find({jira_issue_key: "MYPROJ-123"}).pretty()
```

### RabbitMQ Debugging

```bash
# List queues
docker exec sdlc-rabbitmq rabbitmqctl list_queues

# List consumers
docker exec sdlc-rabbitmq rabbitmqctl list_consumers

# List connections
docker exec sdlc-rabbitmq rabbitmqctl list_connections

# Check cluster status
docker exec sdlc-rabbitmq rabbitmqctl cluster_status
```

### Network Debugging

```bash
# Test service connectivity from inside container
docker exec sdlc-developer-agent-consumer ping mongodb
docker exec sdlc-developer-agent-consumer ping rabbitmq
docker exec sdlc-developer-agent-consumer ping configuration-api

# Test HTTP endpoints
docker exec sdlc-developer-agent-consumer \
  curl http://configuration-api:8081/health
```

### Reset Everything

```bash
# Nuclear option - complete reset
docker-compose down -v
docker system prune -a
rm -rf node_modules  # If applicable
docker-compose up --build
```

---

## Getting Help

If none of these solutions work:

1. **Check GitHub Issues**: https://github.com/anthropics/sdlc-agent/issues
2. **Review logs carefully**: Often the error message contains the solution
3. **Create minimal reproduction**: Isolate the problem to specific steps
4. **Gather information**:
   - Docker version: `docker --version`
   - Docker Compose version: `docker-compose --version`
   - OS: `uname -a`
   - Service logs: `docker-compose logs`

5. **File an issue** with:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs
   - Environment details

---

## See Also

- [Setup Guide](../SETUP.md)
- [API Documentation](API.md)
- [Testing Guide](TESTING.md)
- [Architecture Documentation](ARCHITECTURE.md)
