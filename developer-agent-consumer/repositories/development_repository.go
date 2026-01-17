package repositories

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

type DevelopmentRepository struct {
	collection *mongo.Collection
}

func NewDevelopmentRepository(db *mongo.Database) *DevelopmentRepository {
	return &DevelopmentRepository{
		collection: db.Collection("developments"),
	}
}

func (r *DevelopmentRepository) Create(ctx context.Context, dev *models.Development) error {
	dev.ID = primitive.NewObjectID()
	dev.CreatedAt = time.Now()
	dev.Status = "ready"

	_, err := r.collection.InsertOne(ctx, dev)
	if err != nil {
		return fmt.Errorf("failed to insert development: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) MarkCompleted(ctx context.Context, id primitive.ObjectID, prURL, details string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":              "completed",
			"pr_mr_url":           prURL,
			"development_details": details,
			"completed_at":        &now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) UpdateRepositoryInfo(ctx context.Context, id primitive.ObjectID, repositoryURL, branchName string) error {
	update := bson.M{
		"$set": bson.M{
			"repository_url": repositoryURL,
			"branch_name":    branchName,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update repository info: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) UpdatePrompt(ctx context.Context, id primitive.ObjectID, prompt string) error {
	update := bson.M{
		"$set": bson.M{
			"prompt": prompt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) MarkFailed(ctx context.Context, id primitive.ObjectID, errorMsg string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":        "failed",
			"error_message": errorMsg,
			"completed_at":  &now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}

	return nil
}

func (r *DevelopmentRepository) FindByJiraIssueKey(ctx context.Context, jiraIssueKey string) (*models.Development, error) {
	var dev models.Development
	err := r.collection.FindOne(ctx, bson.M{"jira_issue_key": jiraIssueKey}).Decode(&dev)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find development: %w", err)
	}

	return &dev, nil
}

func (r *DevelopmentRepository) FindByStatus(ctx context.Context, status string) ([]models.Development, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, fmt.Errorf("failed to find developments: %w", err)
	}
	defer cursor.Close(ctx)

	var developments []models.Development
	if err := cursor.All(ctx, &developments); err != nil {
		return nil, fmt.Errorf("failed to decode developments: %w", err)
	}

	return developments, nil
}
