# Backup & Recovery Procedures

Comprehensive backup and disaster recovery procedures for SDLC AI Agents.

## Table of Contents

- [Backup Strategy](#backup-strategy)
- [MongoDB Backup](#mongodb-backup)
- [RabbitMQ Backup](#rabbitmq-backup)
- [Configuration Backup](#configuration-backup)
- [Automated Backups](#automated-backups)
- [Recovery Procedures](#recovery-procedures)
- [Disaster Recovery](#disaster-recovery)
- [Testing Backups](#testing-backups)

---

## Backup Strategy

### What to Backup

1. **MongoDB Database** (Critical)
   - projects collection
   - webhook_events collection
   - developments collection

2. **RabbitMQ State** (Important)
   - Queue definitions
   - Exchange bindings
   - Messages in queues (if any)

3. **Configuration Files** (Critical)
   - docker-compose.yml
   - .env files
   - nginx configurations
   - Custom scripts

4. **Secrets** (Critical)
   - API tokens
   - Database passwords
   - Git access tokens
   - Claude Code session token

### Backup Schedule

| Type | Frequency | Retention | Priority |
|------|-----------|-----------|----------|
| MongoDB Full | Daily | 30 days | Critical |
| MongoDB Incremental | Hourly | 7 days | High |
| RabbitMQ Config | Daily | 14 days | Medium |
| Configuration Files | On change | 90 days | High |
| Secrets | On change | Forever | Critical |

### Backup Locations

**Primary**: Local server (`/backups/`)
**Secondary**: Network storage (NAS/NFS)
**Tertiary**: Cloud storage (S3/GCS/Azure Blob)

---

## MongoDB Backup

### Manual Backup

#### Full Database Backup

```bash
#!/bin/bash
# backup-mongodb.sh

BACKUP_DIR="/backups/mongodb"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="$BACKUP_DIR/$TIMESTAMP"

# Create backup directory
mkdir -p $BACKUP_PATH

# Backup with mongodump
docker exec sdlc-mongodb mongodump \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent?authSource=admin" \
  --out=/backup/$TIMESTAMP

# Copy from container
docker cp sdlc-mongodb:/backup/$TIMESTAMP $BACKUP_PATH

# Compress
cd $BACKUP_DIR
tar -czf mongodb_backup_$TIMESTAMP.tar.gz $TIMESTAMP
rm -rf $TIMESTAMP

# Verify backup
if [ -f "mongodb_backup_$TIMESTAMP.tar.gz" ]; then
    echo "✓ Backup created: mongodb_backup_$TIMESTAMP.tar.gz"
    echo "Size: $(du -h mongodb_backup_$TIMESTAMP.tar.gz | cut -f1)"
else
    echo "✗ Backup failed!"
    exit 1
fi

# Upload to cloud (optional)
# aws s3 cp mongodb_backup_$TIMESTAMP.tar.gz s3://your-bucket/backups/mongodb/
```

#### Backup Specific Collection

```bash
# Backup only projects collection
docker exec sdlc-mongodb mongodump \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent?authSource=admin" \
  --collection=projects \
  --out=/backup/projects_$(date +%Y%m%d)
```

#### Export to JSON

```bash
# Export projects to JSON
docker exec sdlc-mongodb mongoexport \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent?authSource=admin" \
  --collection=projects \
  --out=/backup/projects_$(date +%Y%m%d).json \
  --jsonArray \
  --pretty

docker cp sdlc-mongodb:/backup/projects_$(date +%Y%m%d).json ./backups/
```

### Incremental Backup (Oplog)

```bash
# Enable oplog
docker exec sdlc-mongodb mongodump \
  --uri="mongodb://admin:PASSWORD@localhost:27017" \
  --oplog \
  --out=/backup/incremental_$(date +%Y%m%d_%H%M%S)
```

### Backup Verification

```bash
#!/bin/bash
# verify-backup.sh

BACKUP_FILE=$1

# Extract backup
tar -xzf $BACKUP_FILE -C /tmp/verify

# Restore to temporary database
docker exec sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent_verify?authSource=admin" \
  --drop \
  /tmp/verify

# Count documents
PROJECTS_COUNT=$(docker exec sdlc-mongodb mongosh sdlc_agent_verify \
  --eval "db.projects.countDocuments()" --quiet)

echo "Verified: $PROJECTS_COUNT projects in backup"

# Cleanup
docker exec sdlc-mongodb mongosh sdlc_agent_verify --eval "db.dropDatabase()"
rm -rf /tmp/verify
```

---

## RabbitMQ Backup

### Export Definitions

```bash
#!/bin/bash
# backup-rabbitmq.sh

BACKUP_DIR="/backups/rabbitmq"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# Export all definitions (exchanges, queues, bindings)
docker exec sdlc-rabbitmq rabbitmqctl export_definitions /tmp/definitions_$TIMESTAMP.json

# Copy from container
docker cp sdlc-rabbitmq:/tmp/definitions_$TIMESTAMP.json $BACKUP_DIR/

echo "✓ RabbitMQ definitions backed up: definitions_$TIMESTAMP.json"
```

### Backup Messages (if needed)

```bash
# Save messages from develop queue
docker exec sdlc-rabbitmq rabbitmqadmin get queue=develop count=1000 \
  > /backups/rabbitmq/messages_$(date +%Y%m%d).json
```

---

## Configuration Backup

### Backup Script

```bash
#!/bin/bash
# backup-config.sh

BACKUP_DIR="/backups/config"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/config_backup_$TIMESTAMP.tar.gz"

mkdir -p $BACKUP_DIR

# Files to backup
tar -czf $BACKUP_FILE \
  docker-compose.yml \
  docker-compose.prod.yml \
  .env.example \
  db/init-db.js \
  db/seed-data.js \
  nginx/nginx.conf \
  scripts/*.sh \
  k8s/*.yaml

echo "✓ Configuration backed up: $BACKUP_FILE"
```

### Secrets Backup

```bash
#!/bin/bash
# backup-secrets.sh (SECURE THIS FILE!)

BACKUP_DIR="/backups/secrets"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
chmod 700 $BACKUP_DIR

# Encrypt secrets
tar -czf - secrets/ .env | \
  openssl enc -aes-256-cbc -salt -pbkdf2 \
  -out $BACKUP_DIR/secrets_$TIMESTAMP.tar.gz.enc

echo "✓ Secrets encrypted and backed up"
echo "Remember your encryption password!"
```

---

## Automated Backups

### Cron Configuration

```bash
# Edit crontab
crontab -e

# Add backup jobs
# MongoDB full backup daily at 2 AM
0 2 * * * /opt/sdlc-agent/scripts/backup-mongodb.sh >> /var/log/backup-mongodb.log 2>&1

# MongoDB incremental backup every hour
0 * * * * /opt/sdlc-agent/scripts/backup-mongodb-incremental.sh >> /var/log/backup-incremental.log 2>&1

# RabbitMQ definitions daily at 3 AM
0 3 * * * /opt/sdlc-agent/scripts/backup-rabbitmq.sh >> /var/log/backup-rabbitmq.log 2>&1

# Configuration backup on changes
0 4 * * * /opt/sdlc-agent/scripts/backup-config.sh >> /var/log/backup-config.log 2>&1

# Cleanup old backups daily at 5 AM
0 5 * * * /opt/sdlc-agent/scripts/cleanup-old-backups.sh >> /var/log/backup-cleanup.log 2>&1
```

### Comprehensive Backup Script

```bash
#!/bin/bash
# full-backup.sh - Complete system backup

set -e

BACKUP_ROOT="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="/var/log/backup-$TIMESTAMP.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

log "Starting full system backup"

# 1. MongoDB
log "Backing up MongoDB..."
/opt/sdlc-agent/scripts/backup-mongodb.sh
MONGO_STATUS=$?

# 2. RabbitMQ
log "Backing up RabbitMQ..."
/opt/sdlc-agent/scripts/backup-rabbitmq.sh
RABBITMQ_STATUS=$?

# 3. Configuration
log "Backing up configuration..."
/opt/sdlc-agent/scripts/backup-config.sh
CONFIG_STATUS=$?

# 4. Create manifest
MANIFEST_FILE="$BACKUP_ROOT/manifest_$TIMESTAMP.txt"
cat > $MANIFEST_FILE <<EOF
Backup Timestamp: $TIMESTAMP
MongoDB Status: $([ $MONGO_STATUS -eq 0 ] && echo "SUCCESS" || echo "FAILED")
RabbitMQ Status: $([ $RABBITMQ_STATUS -eq 0 ] && echo "SUCCESS" || echo "FAILED")
Config Status: $([ $CONFIG_STATUS -eq 0 ] && echo "SUCCESS" || echo "FAILED")
Backup Location: $BACKUP_ROOT
Total Size: $(du -sh $BACKUP_ROOT | cut -f1)
EOF

log "Backup manifest created: $MANIFEST_FILE"

# 5. Upload to cloud (optional)
if [ "$ENABLE_CLOUD_BACKUP" = "true" ]; then
    log "Uploading to cloud storage..."
    aws s3 sync $BACKUP_ROOT/$TIMESTAMP s3://your-bucket/sdlc-backups/$TIMESTAMP/
    log "Cloud upload complete"
fi

log "Full system backup completed"

# Send notification (optional)
if [ "$SEND_NOTIFICATIONS" = "true" ]; then
    curl -X POST https://your-webhook-url \
      -H "Content-Type: application/json" \
      -d "{\"text\": \"SDLC Agent backup completed: $TIMESTAMP\"}"
fi
```

### Cleanup Old Backups

```bash
#!/bin/bash
# cleanup-old-backups.sh

BACKUP_DIR="/backups"
MONGODB_RETENTION_DAYS=30
RABBITMQ_RETENTION_DAYS=14
CONFIG_RETENTION_DAYS=90

# Delete old MongoDB backups
find $BACKUP_DIR/mongodb -name "*.tar.gz" -mtime +$MONGODB_RETENTION_DAYS -delete
echo "Cleaned up MongoDB backups older than $MONGODB_RETENTION_DAYS days"

# Delete old RabbitMQ backups
find $BACKUP_DIR/rabbitmq -name "*.json" -mtime +$RABBITMQ_RETENTION_DAYS -delete
echo "Cleaned up RabbitMQ backups older than $RABBITMQ_RETENTION_DAYS days"

# Delete old config backups
find $BACKUP_DIR/config -name "*.tar.gz" -mtime +$CONFIG_RETENTION_DAYS -delete
echo "Cleaned up config backups older than $CONFIG_RETENTION_DAYS days"

# Keep at least 3 most recent backups of each type
cd $BACKUP_DIR/mongodb && ls -t *.tar.gz | tail -n +4 | xargs -r rm
cd $BACKUP_DIR/rabbitmq && ls -t *.json | tail -n +4 | xargs -r rm
cd $BACKUP_DIR/config && ls -t *.tar.gz | tail -n +4 | xargs -r rm
```

---

## Recovery Procedures

### Full MongoDB Restore

```bash
#!/bin/bash
# restore-mongodb.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz>"
    exit 1
fi

echo "⚠️  WARNING: This will restore MongoDB and overwrite existing data!"
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Restore cancelled"
    exit 0
fi

# Stop services that use MongoDB
docker-compose stop configuration-api jira-webhook-api developer-agent-consumer

# Extract backup
TEMP_DIR="/tmp/mongo-restore"
mkdir -p $TEMP_DIR
tar -xzf $BACKUP_FILE -C $TEMP_DIR

# Find backup directory
BACKUP_DIR=$(find $TEMP_DIR -type d -name "sdlc_agent" | head -1 | xargs dirname)

# Restore
docker exec -i sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017?authSource=admin" \
  --drop \
  $BACKUP_DIR

# Verify
PROJECTS_COUNT=$(docker exec sdlc-mongodb mongosh sdlc_agent \
  --eval "db.projects.countDocuments()" --quiet)

echo "✓ Restored $PROJECTS_COUNT projects"

# Cleanup
rm -rf $TEMP_DIR

# Restart services
docker-compose start configuration-api jira-webhook-api developer-agent-consumer

echo "✓ MongoDB restore complete"
```

### Point-in-Time Recovery

```bash
#!/bin/bash
# restore-to-point-in-time.sh

FULL_BACKUP=$1
OPLOG_TIMESTAMP=$2  # Format: "2025-01-15T10:00:00"

# Restore full backup
docker exec -i sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017?authSource=admin" \
  --drop \
  /path/to/full/backup

# Replay oplog up to timestamp
docker exec -i sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017?authSource=admin" \
  --oplogReplay \
  --oplogLimit="$OPLOG_TIMESTAMP" \
  /path/to/oplog/backup
```

### Restore Single Collection

```bash
# Restore only projects collection
docker exec -i sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017/sdlc_agent?authSource=admin" \
  --collection=projects \
  --drop \
  /path/to/backup/sdlc_agent/projects.bson
```

### RabbitMQ Restore

```bash
#!/bin/bash
# restore-rabbitmq.sh

DEFINITIONS_FILE=$1

# Import definitions
docker cp $DEFINITIONS_FILE sdlc-rabbitmq:/tmp/definitions.json

docker exec sdlc-rabbitmq rabbitmqctl import_definitions /tmp/definitions.json

echo "✓ RabbitMQ definitions restored"
```

---

## Disaster Recovery

### Complete System Recovery

```bash
#!/bin/bash
# disaster-recovery.sh - Full system recovery

echo "========================================="
echo "SDLC AI Agents - Disaster Recovery"
echo "========================================="

# 1. Verify backup files exist
BACKUP_DATE=$1  # Format: YYYYMMDD

MONGODB_BACKUP="/backups/mongodb/mongodb_backup_${BACKUP_DATE}_*.tar.gz"
RABBITMQ_BACKUP="/backups/rabbitmq/definitions_${BACKUP_DATE}_*.json"
CONFIG_BACKUP="/backups/config/config_backup_${BACKUP_DATE}_*.tar.gz"

if [ ! -f $MONGODB_BACKUP ]; then
    echo "✗ MongoDB backup not found"
    exit 1
fi

echo "✓ Backup files verified"

# 2. Stop all services
echo "Stopping all services..."
docker-compose down

# 3. Restore configuration
echo "Restoring configuration..."
tar -xzf $CONFIG_BACKUP -C /opt/sdlc-agent/

# 4. Start infrastructure services
echo "Starting MongoDB and RabbitMQ..."
docker-compose up -d mongodb rabbitmq

# Wait for services to be ready
sleep 30

# 5. Restore MongoDB
echo "Restoring MongoDB..."
./scripts/restore-mongodb.sh $MONGODB_BACKUP

# 6. Restore RabbitMQ
echo "Restoring RabbitMQ..."
./scripts/restore-rabbitmq.sh $RABBITMQ_BACKUP

# 7. Start application services
echo "Starting application services..."
docker-compose up -d

# 8. Verify system health
echo "Verifying system health..."
sleep 10

curl -f http://localhost:8081/health && echo "✓ Configuration API healthy" || echo "✗ Configuration API failed"
curl -f http://localhost:8080/health && echo "✓ JIRA Webhook API healthy" || echo "✗ JIRA Webhook API failed"
curl -f http://localhost:3000 && echo "✓ Backoffice UI healthy" || echo "✗ Backoffice UI failed"

echo "========================================="
echo "Disaster recovery complete!"
echo "========================================="
```

### Recovery Time Objectives (RTO/RPO)

| Component | RTO | RPO | Notes |
|-----------|-----|-----|-------|
| MongoDB | 30 minutes | 1 hour | Using incremental backups |
| RabbitMQ | 15 minutes | 24 hours | Messages in queue may be lost |
| Services | 10 minutes | 0 | Stateless, quick restart |
| Complete System | 1 hour | 1 hour | Full disaster recovery |

---

## Testing Backups

### Monthly Backup Test

```bash
#!/bin/bash
# test-backup-restore.sh

echo "Starting backup restore test..."

# Use latest backup
LATEST_BACKUP=$(ls -t /backups/mongodb/*.tar.gz | head -1)

# Restore to test database
TEMP_DB="sdlc_agent_test_$(date +%s)"

docker exec -i sdlc-mongodb mongorestore \
  --uri="mongodb://admin:PASSWORD@localhost:27017/$TEMP_DB?authSource=admin" \
  <extracted-backup-path>

# Verify data
PROJECTS=$(docker exec sdlc-mongodb mongosh $TEMP_DB \
  --eval "db.projects.countDocuments()" --quiet)

WEBHOOKS=$(docker exec sdlc-mongodb mongosh $TEMP_DB \
  --eval "db.webhook_events.countDocuments()" --quiet)

DEVELOPMENTS=$(docker exec sdlc-mongodb mongosh $TEMP_DB \
  --eval "db.developments.countDocuments()" --quiet)

echo "Test Results:"
echo "  Projects: $PROJECTS"
echo "  Webhook Events: $WEBHOOKS"
echo "  Developments: $DEVELOPMENTS"

# Cleanup
docker exec sdlc-mongodb mongosh $TEMP_DB --eval "db.dropDatabase()"

if [ $PROJECTS -gt 0 ]; then
    echo "✓ Backup test PASSED"
    exit 0
else
    echo "✗ Backup test FAILED"
    exit 1
fi
```

### Quarterly Disaster Recovery Drill

```bash
# Schedule DR drill
0 0 1 */3 * /opt/sdlc-agent/scripts/dr-drill.sh
```

---

## Backup Checklist

### Daily
- [ ] MongoDB full backup completed
- [ ] Backup verification passed
- [ ] Cloud upload successful
- [ ] Old backups cleaned up

### Weekly
- [ ] Test restore to staging environment
- [ ] Review backup logs for errors
- [ ] Verify backup retention policy

### Monthly
- [ ] Full backup restore test
- [ ] Review and update recovery procedures
- [ ] Test disaster recovery plan
- [ ] Audit backup encryption

### Quarterly
- [ ] Full disaster recovery drill
- [ ] Review RTO/RPO targets
- [ ] Update disaster recovery documentation
- [ ] Train team on recovery procedures

---

## Troubleshooting

### Backup Fails

**Check disk space**
```bash
df -h /backups
```

**Check MongoDB connection**
```bash
docker exec sdlc-mongodb mongosh --eval "db.runCommand('ping')"
```

**Check permissions**
```bash
ls -la /backups
chmod 755 /backups
```

### Restore Fails

**Check backup file integrity**
```bash
tar -tzf backup.tar.gz > /dev/null && echo "OK" || echo "Corrupted"
```

**Check MongoDB is running**
```bash
docker-compose ps mongodb
```

**Check MongoDB logs**
```bash
docker-compose logs mongodb | tail -100
```

---

## See Also

- [Production Deployment Guide](PRODUCTION.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [Architecture Documentation](ARCHITECTURE.md)
