package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"github.com/storos/sdlc-agent/configuration-api/repositories"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrRepositoryNotFound  = errors.New("repository not found")
	ErrDuplicateProjectKey = errors.New("project with this JIRA key already exists")
)

// ProjectService handles business logic for projects
type ProjectService struct {
	repo *repositories.ProjectRepository
}

// NewProjectService creates a new project service
func NewProjectService(repo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

// GetAllProjects returns all projects
func (s *ProjectService) GetAllProjects(ctx context.Context) ([]models.Project, error) {
	return s.repo.FindAll(ctx)
}

// GetProjectByID returns a project by ID
func (s *ProjectService) GetProjectByID(ctx context.Context, id string) (*models.Project, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	project, err := s.repo.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, ErrProjectNotFound
	}

	return project, nil
}

// GetProjectByJiraKey returns a project by JIRA project key
func (s *ProjectService) GetProjectByJiraKey(ctx context.Context, jiraProjectKey string) (*models.Project, error) {
	project, err := s.repo.FindByJiraProjectKey(ctx, jiraProjectKey)
	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, ErrProjectNotFound
	}

	return project, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest) (*models.Project, error) {
	// Check if project with same JIRA key already exists
	existing, err := s.repo.FindByJiraProjectKey(ctx, req.JiraProjectKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateProjectKey
	}

	project := &models.Project{
		Name:            req.Name,
		Description:     req.Description,
		Scope:           req.Scope,
		JiraProjectKey:  req.JiraProjectKey,
		JiraProjectName: req.JiraProjectName,
		JiraProjectURL:  req.JiraProjectURL,
		Repositories:    req.Repositories,
	}

	// Initialize repositories if nil
	if project.Repositories == nil {
		project.Repositories = []models.Repository{}
	}

	// Generate repository IDs for any repositories provided
	for i := range project.Repositories {
		project.Repositories[i].RepositoryID = primitive.NewObjectID().Hex()
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(ctx context.Context, id string, req *models.UpdateProjectRequest) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Build update document
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Scope != "" {
		update["scope"] = req.Scope
	}
	if req.JiraProjectKey != "" {
		// Check for duplicate key if changing it
		existing, err := s.repo.FindByJiraProjectKey(ctx, req.JiraProjectKey)
		if err != nil {
			return err
		}
		if existing != nil && existing.ID != objectID {
			return ErrDuplicateProjectKey
		}
		update["jira_project_key"] = req.JiraProjectKey
	}
	if req.JiraProjectName != "" {
		update["jira_project_name"] = req.JiraProjectName
	}
	if req.JiraProjectURL != "" {
		update["jira_project_url"] = req.JiraProjectURL
	}
	if req.Repositories != nil {
		update["repositories"] = req.Repositories
	}

	if len(update) == 0 {
		return nil
	}

	err = s.repo.Update(ctx, objectID, update)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrProjectNotFound
	}
	return err
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	err = s.repo.Delete(ctx, objectID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrProjectNotFound
	}
	return err
}

// AddRepository adds a repository to a project
func (s *ProjectService) AddRepository(ctx context.Context, projectID string, req *models.AddRepositoryRequest) error {
	objectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	repo := models.Repository{
		URL:            req.URL,
		Description:    req.Description,
		GitAccessToken: req.GitAccessToken,
	}

	err = s.repo.AddRepository(ctx, objectID, repo)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrProjectNotFound
	}
	return err
}

// UpdateRepository updates a repository in a project
func (s *ProjectService) UpdateRepository(ctx context.Context, projectID, repoID string, req *models.UpdateRepositoryRequest) error {
	objectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	// Build update document
	update := bson.M{}
	if req.URL != "" {
		update["url"] = req.URL
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.GitAccessToken != "" {
		update["git_access_token"] = req.GitAccessToken
	}

	if len(update) == 0 {
		return nil
	}

	err = s.repo.UpdateRepository(ctx, objectID, repoID, update)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrRepositoryNotFound
	}
	return err
}

// DeleteRepository removes a repository from a project
func (s *ProjectService) DeleteRepository(ctx context.Context, projectID, repoID string) error {
	objectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	err = s.repo.DeleteRepository(ctx, objectID, repoID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrRepositoryNotFound
	}
	return err
}
