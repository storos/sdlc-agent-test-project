// MongoDB Initialization Script for SDLC AI Agents
// This script creates all collections and indexes

db = db.getSiblingDB('sdlc_agents');

print('Creating SDLC AI Agents database and collections...');

// ============================================
// Projects Collection
// ============================================
print('Setting up projects collection...');

db.createCollection('projects');

// Create indexes for projects
db.projects.createIndex({ "jira_project_key": 1 }, { unique: true, name: "idx_jira_project_key_unique" });
db.projects.createIndex({ "name": 1 }, { name: "idx_name" });
db.projects.createIndex({ "created_at": -1 }, { name: "idx_created_at_desc" });

print('✓ Projects collection created with indexes');

// ============================================
// Webhook Events Collection
// ============================================
print('Setting up webhook_events collection...');

db.createCollection('webhook_events');

// Create indexes for webhook_events
db.webhook_events.createIndex({ "jira_issue_id": 1 }, { name: "idx_jira_issue_id" });
db.webhook_events.createIndex({ "jira_issue_key": 1 }, { name: "idx_jira_issue_key" });
db.webhook_events.createIndex({ "jira_project_key": 1 }, { name: "idx_jira_project_key" });
db.webhook_events.createIndex({ "created_at": -1 }, { name: "idx_created_at_desc" });
db.webhook_events.createIndex({ "issue_status": 1, "created_at": -1 }, { name: "idx_status_created_at" });

print('✓ Webhook events collection created with indexes');

// ============================================
// Developments Collection
// ============================================
print('Setting up developments collection...');

db.createCollection('developments');

// Create indexes for developments
db.developments.createIndex({ "project_id": 1 }, { name: "idx_project_id" });
db.developments.createIndex({ "jira_issue_id": 1 }, { name: "idx_jira_issue_id" });
db.developments.createIndex({ "jira_issue_key": 1 }, { name: "idx_jira_issue_key" });
db.developments.createIndex({ "jira_project_key": 1 }, { name: "idx_jira_project_key" });
db.developments.createIndex({ "status": 1, "created_at": -1 }, { name: "idx_status_created_at" });
db.developments.createIndex({ "project_id": 1, "status": 1, "created_at": -1 }, { name: "idx_project_status_created_at" });

print('✓ Developments collection created with indexes');

// ============================================
// Verify Setup
// ============================================
print('\nVerifying setup...');

const collections = db.getCollectionNames();
print('Collections: ' + collections.join(', '));

print('\nProjects indexes:');
db.projects.getIndexes().forEach(idx => print('  - ' + idx.name));

print('\nWebhook events indexes:');
db.webhook_events.getIndexes().forEach(idx => print('  - ' + idx.name));

print('\nDevelopments indexes:');
db.developments.getIndexes().forEach(idx => print('  - ' + idx.name));

print('\n✓ Database initialization completed successfully!');
