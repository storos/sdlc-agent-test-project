package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	collection *mongo.Collection
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *mongo.Database) *ProjectRepository {
	return &ProjectRepository{
		collection: db.Collection("projects"),
	}
}

// FindAll returns all projects
func (r *ProjectRepository) FindAll(ctx context.Context) ([]models.Project, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []models.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// FindByID returns a project by ID
func (r *ProjectRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Project, error) {
	var project models.Project
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

// FindByJiraProjectKey returns a project by JIRA project key
func (r *ProjectRepository) FindByJiraProjectKey(ctx context.Context, jiraProjectKey string) (*models.Project, error) {
	var project models.Project
	err := r.collection.FindOne(ctx, bson.M{"jira_project_key": jiraProjectKey}).Decode(&project)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &project, nil
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, project)
	if err != nil {
		return err
	}

	project.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// Update updates an existing project
func (r *ProjectRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Delete deletes a project
func (r *ProjectRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// AddRepository adds a repository to a project
func (r *ProjectRepository) AddRepository(ctx context.Context, projectID primitive.ObjectID, repo models.Repository) error {
	// Generate repository ID
	repo.RepositoryID = primitive.NewObjectID().Hex()

	update := bson.M{
		"$push": bson.M{"repositories": repo},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": projectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// UpdateRepository updates a repository in a project
func (r *ProjectRepository) UpdateRepository(ctx context.Context, projectID primitive.ObjectID, repoID string, update bson.M) error {
	// Build the update for the specific repository in the array
	setUpdate := bson.M{"updated_at": time.Now()}
	for key, value := range update {
		setUpdate["repositories.$."+key] = value
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":                     projectID,
			"repositories.repository_id": repoID,
		},
		bson.M{"$set": setUpdate},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// DeleteRepository removes a repository from a project
func (r *ProjectRepository) DeleteRepository(ctx context.Context, projectID primitive.ObjectID, repoID string) error {
	update := bson.M{
		"$pull": bson.M{"repositories": bson.M{"repository_id": repoID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": projectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
