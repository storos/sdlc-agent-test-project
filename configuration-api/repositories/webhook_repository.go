package repositories

import (
	"context"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WebhookRepository struct {
	collection *mongo.Collection
}

func NewWebhookRepository(db *mongo.Database) *WebhookRepository {
	return &WebhookRepository{
		collection: db.Collection("webhook_events"),
	}
}

func (r *WebhookRepository) GetAll(ctx context.Context) ([]models.WebhookEvent, error) {
	// Sort by received_at descending (newest first)
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var webhookEvents []models.WebhookEvent
	if err := cursor.All(ctx, &webhookEvents); err != nil {
		return nil, err
	}

	return webhookEvents, nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.WebhookEvent, error) {
	var webhookEvent models.WebhookEvent
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&webhookEvent)
	if err != nil {
		return nil, err
	}
	return &webhookEvent, nil
}

func (r *WebhookRepository) GetByJiraProjectKey(ctx context.Context, jiraProjectKey string) ([]models.WebhookEvent, error) {
	// Sort by received_at descending (newest first)
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"jira_project_key": jiraProjectKey}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var webhookEvents []models.WebhookEvent
	if err := cursor.All(ctx, &webhookEvents); err != nil {
		return nil, err
	}

	return webhookEvents, nil
}
