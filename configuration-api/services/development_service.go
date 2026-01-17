package services

import (
	"context"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"github.com/storos/sdlc-agent/configuration-api/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DevelopmentService struct {
	repo *repositories.DevelopmentRepository
}

func NewDevelopmentService(repo *repositories.DevelopmentRepository) *DevelopmentService {
	return &DevelopmentService{
		repo: repo,
	}
}

func (s *DevelopmentService) GetAll(ctx context.Context) ([]models.Development, error) {
	return s.repo.GetAll(ctx)
}

func (s *DevelopmentService) GetByID(ctx context.Context, id string) (*models.Development, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, objectID)
}

func (s *DevelopmentService) GetByJiraProjectKey(ctx context.Context, jiraProjectKey string) ([]models.Development, error) {
	return s.repo.GetByJiraProjectKey(ctx, jiraProjectKey)
}
