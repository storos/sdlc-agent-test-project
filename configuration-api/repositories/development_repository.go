package repositories

import (
	"context"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DevelopmentRepository struct {
	collection *mongo.Collection
}

func NewDevelopmentRepository(db *mongo.Database) *DevelopmentRepository {
	return &DevelopmentRepository{
		collection: db.Collection("developments"),
	}
}

func (r *DevelopmentRepository) GetAll(ctx context.Context) ([]models.Development, error) {
	// Sort by created_at descending (newest first)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developments []models.Development
	if err := cursor.All(ctx, &developments); err != nil {
		return nil, err
	}

	return developments, nil
}

func (r *DevelopmentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Development, error) {
	var development models.Development
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&development)
	if err != nil {
		return nil, err
	}
	return &development, nil
}

func (r *DevelopmentRepository) GetByJiraProjectKey(ctx context.Context, jiraProjectKey string) ([]models.Development, error) {
	// Sort by created_at descending (newest first)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"jira_project_key": jiraProjectKey}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var developments []models.Development
	if err := cursor.All(ctx, &developments); err != nil {
		return nil, err
	}

	return developments, nil
}
