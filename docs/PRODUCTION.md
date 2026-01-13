# Production Deployment Guide

Complete guide for deploying SDLC AI Agents to production.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Infrastructure Requirements](#infrastructure-requirements)
- [Security Hardening](#security-hardening)
- [Environment Configuration](#environment-configuration)
- [Deployment Options](#deployment-options)
- [Monitoring & Alerting](#monitoring--alerting)
- [Backup & Recovery](#backup--recovery)
- [Scaling](#scaling)
- [Maintenance](#maintenance)

---

## Prerequisites

### Required Services
- Docker Engine 20.10+
- Docker Compose 2.0+ OR Kubernetes 1.24+
- MongoDB 7.0+ (Standalone or Replica Set)
- RabbitMQ 3.12+ (Standalone or Cluster)

### Required Credentials
- GitHub/GitLab Personal Access Tokens
- Claude Code API session token
- JIRA webhook secret (optional but recommended)
- MongoDB admin credentials
- RabbitMQ admin credentials

### Infrastructure
- Minimum 4 CPU cores
- Minimum 8GB RAM
- Minimum 50GB disk space
- Network connectivity to Git providers
- Network connectivity to Claude Code API

---

## Infrastructure Requirements

### Recommended Specifications

#### Small Deployment (1-10 developers)
- **CPU**: 4 cores
- **RAM**: 8GB
- **Disk**: 100GB SSD
- **Network**: 100 Mbps

#### Medium Deployment (10-50 developers)
- **CPU**: 8 cores
- **RAM**: 16GB
- **Disk**: 250GB SSD
- **Network**: 1 Gbps

#### Large Deployment (50+ developers)
- **CPU**: 16+ cores
- **RAM**: 32GB+
- **Disk**: 500GB+ SSD
- **Network**: 1 Gbps+
- **Load Balancer**: Required
- **Multiple Instances**: Required

### Per-Service Resource Allocation

```yaml
# MongoDB
resources:
  limits:
    cpu: "2"
    memory: 4Gi
  requests:
    cpu: "1"
    memory: 2Gi

# RabbitMQ
resources:
  limits:
    cpu: "1"
    memory: 2Gi
  requests:
    cpu: "500m"
    memory: 1Gi

# Configuration API
resources:
  limits:
    cpu: "500m"
    memory: 512Mi
  requests:
    cpu: "250m"
    memory: 256Mi

# JIRA Webhook API
resources:
  limits:
    cpu: "500m"
    memory: 512Mi
  requests:
    cpu: "250m"
    memory: 256Mi

# Developer Agent Consumer
resources:
  limits:
    cpu: "2"
    memory: 2Gi
  requests:
    cpu: "1"
    memory: 1Gi

# Backoffice UI
resources:
  limits:
    cpu: "250m"
    memory: 256Mi
  requests:
    cpu: "100m"
    memory: 128Mi
```

---

## Security Hardening

### 1. Network Security

**Firewall Rules**
```bash
# Only expose necessary ports
# MongoDB: Internal only (27017)
# RabbitMQ: Internal only (5672, 15672)
# Configuration API: Internal only (8081)
# JIRA Webhook API: External (8080) - with rate limiting
# Backoffice UI: External (443) - HTTPS only
```

**Docker Network Isolation**
```yaml
networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true  # No external access

services:
  backoffice-ui:
    networks:
      - frontend

  configuration-api:
    networks:
      - frontend
      - backend

  mongodb:
    networks:
      - backend  # Not exposed externally
```

### 2. Secrets Management

**Use Docker Secrets (Swarm)**
```yaml
secrets:
  mongodb_password:
    external: true
  rabbitmq_password:
    external: true
  claude_session_token:
    external: true

services:
  mongodb:
    secrets:
      - mongodb_password
    environment:
      MONGO_INITDB_ROOT_PASSWORD_FILE: /run/secrets/mongodb_password
```

**Use Kubernetes Secrets**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sdlc-secrets
type: Opaque
stringData:
  mongodb-password: <base64-encoded>
  rabbitmq-password: <base64-encoded>
  claude-session-token: <base64-encoded>
```

**Use HashiCorp Vault (Recommended)**
```bash
# Store secrets in Vault
vault kv put secret/sdlc-agent \
  mongodb_password="secure-password" \
  rabbitmq_password="secure-password" \
  claude_session_token="your-token"

# Inject into containers
# Use vault-agent or vault-k8s for automatic injection
```

### 3. Database Security

**MongoDB Authentication**
```javascript
// Create admin user
db.createUser({
  user: "sdlc_admin",
  pwd: "STRONG_PASSWORD_HERE",
  roles: [
    { role: "readWrite", db: "sdlc_agent" },
    { role: "dbAdmin", db: "sdlc_agent" }
  ]
})

// Enable authentication
mongod --auth --bind_ip_all
```

**MongoDB Connection String**
```
mongodb://sdlc_admin:PASSWORD@mongodb:27017/sdlc_agent?authSource=admin
```

**MongoDB Encryption at Rest**
```yaml
# MongoDB configuration
security:
  encryption:
    enableEncryption: true
    encryptionKeyFile: /path/to/keyfile
```

### 4. API Security

**Add API Key Authentication**

Configuration API example:
```go
// middleware/auth.go
func APIKeyAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "API key required"})
            c.Abort()
            return
        }

        // Validate API key against database or environment
        if !validateAPIKey(apiKey) {
            c.JSON(403, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**JIRA Webhook Secret Validation**
```go
// Validate JIRA webhook signature
func ValidateJIRASignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

### 5. HTTPS/TLS

**Use Reverse Proxy (Nginx/Traefik)**
```nginx
server {
    listen 443 ssl http2;
    server_name sdlc-backoffice.company.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass http://backoffice-ui:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 443 ssl http2;
    server_name sdlc-webhook.company.com;

    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=webhook:10m rate=10r/s;
    limit_req zone=webhook burst=20;

    location /webhook {
        proxy_pass http://jira-webhook-api:8080;
    }
}
```

---

## Environment Configuration

### Production Environment Variables

Create `.env.production`:
```bash
# MongoDB
MONGO_URL=mongodb://sdlc_admin:PASSWORD@mongodb:27017/sdlc_agent?authSource=admin
MONGO_DATABASE=sdlc_agent

# RabbitMQ
RABBITMQ_URL=amqp://sdlc_user:PASSWORD@rabbitmq:5672/
RABBITMQ_DEFAULT_USER=sdlc_user
RABBITMQ_DEFAULT_PASS=STRONG_PASSWORD

# Configuration API
CONFIGURATION_API_PORT=8081
CONFIGURATION_API_URL=http://configuration-api:8081

# JIRA Webhook API
JIRA_WEBHOOK_API_PORT=8080
JIRA_WEBHOOK_SECRET=your-jira-webhook-secret

# Developer Agent Consumer
CLAUDE_API_URL=https://api.claude.ai/code
CLAUDE_CODE_SESSION_TOKEN=your-session-token

# Backoffice UI
VITE_API_URL=https://api.sdlc.company.com

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Monitoring
ENABLE_METRICS=true
METRICS_PORT=9090
```

### Environment Validation

Add to each service's `main.go`:
```go
func validateEnvironment() error {
    required := []string{
        "MONGO_URL",
        "RABBITMQ_URL",
        "CLAUDE_CODE_SESSION_TOKEN",
    }

    var missing []string
    for _, env := range required {
        if os.Getenv(env) == "" {
            missing = append(missing, env)
        }
    }

    if len(missing) > 0 {
        return fmt.Errorf("missing required environment variables: %s",
            strings.Join(missing, ", "))
    }

    return nil
}

func main() {
    if err := validateEnvironment(); err != nil {
        log.Fatal(err)
    }
    // Continue with startup...
}
```

---

## Deployment Options

### Option 1: Docker Compose (Single Server)

**Production docker-compose.yml**
```yaml
version: '3.8'

services:
  mongodb:
    image: mongo:7.0
    container_name: sdlc-mongodb
    restart: always
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD_FILE: /run/secrets/mongodb_root_password
    secrets:
      - mongodb_root_password
    volumes:
      - /data/mongodb:/data/db
      - ./db/init-db.js:/docker-entrypoint-initdb.d/init-db.js:ro
    networks:
      - backend
    command: mongod --auth
    healthcheck:
      test: mongosh --eval "db.runCommand('ping').ok" --quiet
      interval: 10s
      timeout: 5s
      retries: 5

  rabbitmq:
    image: rabbitmq:3.12-management
    container_name: sdlc-rabbitmq
    restart: always
    environment:
      RABBITMQ_DEFAULT_USER_FILE: /run/secrets/rabbitmq_user
      RABBITMQ_DEFAULT_PASS_FILE: /run/secrets/rabbitmq_password
    secrets:
      - rabbitmq_user
      - rabbitmq_password
    volumes:
      - /data/rabbitmq:/var/lib/rabbitmq
    networks:
      - backend

  configuration-api:
    image: sdlc-configuration-api:${VERSION:-latest}
    restart: always
    environment:
      MONGODB_URL_FILE: /run/secrets/mongodb_url
    secrets:
      - mongodb_url
    networks:
      - backend
      - frontend
    depends_on:
      mongodb:
        condition: service_healthy

  # ... other services

secrets:
  mongodb_root_password:
    file: ./secrets/mongodb_root_password
  mongodb_url:
    file: ./secrets/mongodb_url
  rabbitmq_user:
    file: ./secrets/rabbitmq_user
  rabbitmq_password:
    file: ./secrets/rabbitmq_password

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true

volumes:
  mongodb_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/mongodb
  rabbitmq_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/rabbitmq
```

**Deploy**
```bash
# Create secrets directory
mkdir -p secrets
echo "strong-password" > secrets/mongodb_root_password
echo "mongodb://..." > secrets/mongodb_url
chmod 600 secrets/*

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Monitor
docker-compose -f docker-compose.prod.yml logs -f
```

### Option 2: Kubernetes

**Create namespace**
```bash
kubectl create namespace sdlc-agent
```

**Deploy MongoDB**
```yaml
# mongodb-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
  namespace: sdlc-agent
spec:
  serviceName: mongodb
  replicas: 3
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:7.0
        ports:
        - containerPort: 27017
        env:
        - name: MONGO_INITDB_ROOT_USERNAME
          value: admin
        - name: MONGO_INITDB_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: sdlc-secrets
              key: mongodb-password
        volumeMounts:
        - name: mongodb-data
          mountPath: /data/db
  volumeClaimTemplates:
  - metadata:
      name: mongodb-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 100Gi
```

**Deploy Services**
```bash
# Apply all manifests
kubectl apply -f k8s/

# Check status
kubectl get pods -n sdlc-agent
kubectl get svc -n sdlc-agent

# View logs
kubectl logs -f deployment/developer-agent-consumer -n sdlc-agent
```

### Option 3: Cloud Platforms

**AWS ECS**
- Use ECS Task Definitions for each service
- Use RDS for MongoDB (DocumentDB)
- Use Amazon MQ for RabbitMQ
- Use ALB for load balancing

**Google Cloud Run**
- Deploy each service as Cloud Run service
- Use Cloud SQL for MongoDB
- Use Cloud Pub/Sub instead of RabbitMQ
- Use Cloud Load Balancing

**Azure Container Instances**
- Use Azure Container Instances
- Use Cosmos DB (MongoDB API)
- Use Azure Service Bus
- Use Azure Application Gateway

---

## Monitoring & Alerting

### Prometheus Metrics

**Add metrics endpoint to services**
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
    prometheus.MustRegister(requestDuration)
}

// In main()
http.Handle("/metrics", promhttp.Handler())
```

**Prometheus Configuration**
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'configuration-api'
    static_configs:
      - targets: ['configuration-api:9090']

  - job_name: 'jira-webhook-api'
    static_configs:
      - targets: ['jira-webhook-api:9090']

  - job_name: 'developer-agent-consumer'
    static_configs:
      - targets: ['developer-agent-consumer:9090']
```

### Grafana Dashboards

**Import pre-built dashboards**
- MongoDB: Dashboard ID 2583
- RabbitMQ: Dashboard ID 10991
- Docker: Dashboard ID 893
- Go Applications: Dashboard ID 6671

**Custom Dashboard Panels**
1. Development Requests Per Hour
2. Success vs Failed Developments
3. Average Processing Time
4. Queue Depth Over Time
5. API Response Times
6. Error Rate

### Alerting Rules

```yaml
# alerts.yml
groups:
  - name: sdlc_agent_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"

      - alert: QueueBacklog
        expr: rabbitmq_queue_messages{queue="develop"} > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Development queue has {{ $value }} messages"

      - alert: ServiceDown
        expr: up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
```

### Log Aggregation (ELK Stack)

**Filebeat Configuration**
```yaml
filebeat.inputs:
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'

processors:
  - add_docker_metadata: ~

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "sdlc-agent-%{+yyyy.MM.dd}"
```

---

## Backup & Recovery

See [BACKUP.md](BACKUP.md) for detailed procedures.

### Quick Backup

```bash
# MongoDB backup
docker exec sdlc-mongodb mongodump \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent?authSource=admin" \
  --out=/backup/$(date +%Y%m%d)

# Copy backup from container
docker cp sdlc-mongodb:/backup/$(date +%Y%m%d) ./backups/
```

### Automated Backups

```bash
# Add to cron
0 2 * * * /opt/sdlc-agent/scripts/backup.sh
```

---

## Scaling

### Horizontal Scaling

**Scale Consumer**
```bash
docker-compose up --scale developer-agent-consumer=5
```

**Kubernetes HPA**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: developer-agent-consumer-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: developer-agent-consumer
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Database Scaling

**MongoDB Replica Set**
```yaml
# docker-compose with replica set
services:
  mongodb-primary:
    image: mongo:7.0
    command: mongod --replSet rs0

  mongodb-secondary-1:
    image: mongo:7.0
    command: mongod --replSet rs0

  mongodb-secondary-2:
    image: mongo:7.0
    command: mongod --replSet rs0
```

**RabbitMQ Cluster**
```yaml
services:
  rabbitmq-1:
    image: rabbitmq:3.12-management
    environment:
      RABBITMQ_ERLANG_COOKIE: 'secret-cookie'

  rabbitmq-2:
    image: rabbitmq:3.12-management
    environment:
      RABBITMQ_ERLANG_COOKIE: 'secret-cookie'
    depends_on:
      - rabbitmq-1
```

---

## Maintenance

### Updates

```bash
# Pull latest images
docker-compose pull

# Rolling update
docker-compose up -d --no-deps --build configuration-api

# Verify health
curl http://localhost:8081/health
```

### Database Maintenance

```bash
# Compact collections
docker exec -it sdlc-mongodb mongosh sdlc_agent
db.projects.compact()
db.webhook_events.compact()
db.developments.compact()

# Rebuild indexes
db.developments.reIndex()
```

### Log Rotation

```bash
# Configure Docker log rotation
# /etc/docker/daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

---

## Checklist

Before going to production:

- [ ] All secrets stored securely (not in code)
- [ ] HTTPS enabled with valid certificates
- [ ] API authentication implemented
- [ ] Database authentication enabled
- [ ] Firewall rules configured
- [ ] Backups automated and tested
- [ ] Monitoring and alerting configured
- [ ] Log aggregation set up
- [ ] Resource limits defined
- [ ] Health checks working
- [ ] Documentation updated
- [ ] Disaster recovery plan documented
- [ ] Security audit completed
- [ ] Performance testing done
- [ ] Staging environment tested

---

## See Also

- [Backup & Recovery](BACKUP.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [API Documentation](API.md)
- [Architecture Overview](ARCHITECTURE.md)
