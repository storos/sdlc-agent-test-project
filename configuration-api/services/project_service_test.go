package services

import (
	"context"
	"errors"
	"testing"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockProjectRepository is a mock implementation of the repository
type MockProjectRepository struct {
	projects    map[string]*models.Project
	shouldError bool
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[string]*models.Project),
	}
}

func (m *MockProjectRepository) Create(ctx context.Context, project *models.Project) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	project.ID = primitive.NewObjectID()
	m.projects[project.ID.Hex()] = project
	return nil
}

func (m *MockProjectRepository) FindAll(ctx context.Context) ([]*models.Project, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	projects := make([]*models.Project, 0, len(m.projects))
	for _, p := range m.projects {
		projects = append(projects, p)
	}
	return projects, nil
}

func (m *MockProjectRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Project, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	project, exists := m.projects[id.Hex()]
	if !exists {
		return nil, nil
	}
	return project, nil
}

func (m *MockProjectRepository) FindByJiraProjectKey(ctx context.Context, key string) (*models.Project, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	for _, p := range m.projects {
		if p.JiraProjectKey == key {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockProjectRepository) Update(ctx context.Context, id primitive.ObjectID, project *models.Project) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	if _, exists := m.projects[id.Hex()]; !exists {
		return nil
	}
	project.ID = id
	m.projects[id.Hex()] = project
	return nil
}

func (m *MockProjectRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	if _, exists := m.projects[id.Hex()]; !exists {
		return nil
	}
	delete(m.projects, id.Hex())
	return nil
}

func TestCreateProject(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err := service.CreateProject(context.Background(), project)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if project.ID.IsZero() {
		t.Error("Expected project ID to be set")
	}
}

func TestCreateProject_DuplicateKey(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project1 := &models.Project{
		Name:            "Test Project 1",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err := service.CreateProject(context.Background(), project1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	project2 := &models.Project{
		Name:            "Test Project 2",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err = service.CreateProject(context.Background(), project2)
	if err != ErrDuplicateProjectKey {
		t.Errorf("Expected ErrDuplicateProjectKey, got %v", err)
	}
}

func TestGetAllProjects(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	service.CreateProject(context.Background(), project)

	projects, err := service.GetAllProjects(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

func TestGetProjectByID(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	service.CreateProject(context.Background(), project)

	found, err := service.GetProjectByID(context.Background(), project.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found.Name != project.Name {
		t.Errorf("Expected name %s, got %s", project.Name, found.Name)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	id := primitive.NewObjectID()
	_, err := service.GetProjectByID(context.Background(), id)
	if err != ErrProjectNotFound {
		t.Errorf("Expected ErrProjectNotFound, got %v", err)
	}
}

func TestGetProjectByJiraKey(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	service.CreateProject(context.Background(), project)

	found, err := service.GetProjectByJiraKey(context.Background(), "TEST")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found.JiraProjectKey != "TEST" {
		t.Errorf("Expected JIRA key TEST, got %s", found.JiraProjectKey)
	}
}

func TestUpdateProject(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	service.CreateProject(context.Background(), project)

	project.Name = "Updated Project"
	err := service.UpdateProject(context.Background(), project.ID, project)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	found, _ := service.GetProjectByID(context.Background(), project.ID)
	if found.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got %s", found.Name)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	id := primitive.NewObjectID()
	project := &models.Project{
		Name:            "Updated Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err := service.UpdateProject(context.Background(), id, project)
	if err != ErrProjectNotFound {
		t.Errorf("Expected ErrProjectNotFound, got %v", err)
	}
}

func TestDeleteProject(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	service.CreateProject(context.Background(), project)

	err := service.DeleteProject(context.Background(), project.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	_, err = service.GetProjectByID(context.Background(), project.ID)
	if err != ErrProjectNotFound {
		t.Errorf("Expected ErrProjectNotFound after deletion, got %v", err)
	}
}

func TestAddRepository(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
		Repositories:    []models.Repository{},
	}

	service.CreateProject(context.Background(), project)

	repo := &models.Repository{
		URL:            "https://github.com/test/repo",
		Description:    "Test Repo",
		GitAccessToken: "token123",
	}

	err := service.AddRepository(context.Background(), project.ID, repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if repo.RepositoryID == "" {
		t.Error("Expected repository ID to be set")
	}
}

func TestUpdateRepository(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
		Repositories: []models.Repository{
			{
				RepositoryID:   "repo1",
				URL:            "https://github.com/test/repo",
				Description:    "Test Repo",
				GitAccessToken: "token123",
			},
		},
	}

	service.CreateProject(context.Background(), project)

	updatedRepo := &models.Repository{
		URL:            "https://github.com/test/repo-updated",
		Description:    "Updated Repo",
		GitAccessToken: "token456",
	}

	err := service.UpdateRepository(context.Background(), project.ID, "repo1", updatedRepo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	found, _ := service.GetProjectByID(context.Background(), project.ID)
	if found.Repositories[0].Description != "Updated Repo" {
		t.Errorf("Expected description 'Updated Repo', got %s", found.Repositories[0].Description)
	}
}

func TestDeleteRepository(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := NewProjectService(mockRepo)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
		Repositories: []models.Repository{
			{
				RepositoryID:   "repo1",
				URL:            "https://github.com/test/repo",
				Description:    "Test Repo",
				GitAccessToken: "token123",
			},
		},
	}

	service.CreateProject(context.Background(), project)

	err := service.DeleteRepository(context.Background(), project.ID, "repo1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	found, _ := service.GetProjectByID(context.Background(), project.ID)
	if len(found.Repositories) != 0 {
		t.Errorf("Expected 0 repositories, got %d", len(found.Repositories))
	}
}
