package services

import (
	"context"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"github.com/storos/sdlc-agent/configuration-api/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebhookService struct {
	repo *repositories.WebhookRepository
}

func NewWebhookService(repo *repositories.WebhookRepository) *WebhookService {
	return &WebhookService{
		repo: repo,
	}
}

func (s *WebhookService) GetAll(ctx context.Context) ([]models.WebhookEvent, error) {
	return s.repo.GetAll(ctx)
}

func (s *WebhookService) GetByID(ctx context.Context, id string) (*models.WebhookEvent, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, objectID)
}

func (s *WebhookService) GetByJiraProjectKey(ctx context.Context, jiraProjectKey string) ([]models.WebhookEvent, error) {
	return s.repo.GetByJiraProjectKey(ctx, jiraProjectKey)
}
