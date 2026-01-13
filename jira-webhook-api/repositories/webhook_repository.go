package repositories

import (
	"context"
	"time"

	"github.com/storos/sdlc-agent/jira-webhook-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// WebhookRepository handles webhook event persistence
type WebhookRepository struct {
	collection *mongo.Collection
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *mongo.Database) *WebhookRepository {
	return &WebhookRepository{
		collection: db.Collection("webhook_events"),
	}
}

// Create stores a new webhook event
func (r *WebhookRepository) Create(ctx context.Context, event *models.WebhookEvent) error {
	event.ID = primitive.NewObjectID()
	event.ReceivedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, event)
	return err
}

// FindByJiraIssueKey retrieves webhook events by JIRA issue key
func (r *WebhookRepository) FindByJiraIssueKey(ctx context.Context, issueKey string) ([]*models.WebhookEvent, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"jira_issue_key": issueKey})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*models.WebhookEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// MarkProcessed updates the processed_at timestamp
func (r *WebhookRepository) MarkProcessed(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"processed_at": now,
		},
	}

	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}

// FindUnprocessed retrieves all unprocessed webhook events
func (r *WebhookRepository) FindUnprocessed(ctx context.Context) ([]*models.WebhookEvent, error) {
	filter := bson.M{
		"processed_at": bson.M{"$exists": false},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*models.WebhookEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
