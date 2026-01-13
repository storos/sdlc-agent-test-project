package repositories

import (
	"context"
	"testing"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockCollection simulates MongoDB collection operations
type MockCollection struct {
	data map[string]*models.Project
}

func NewMockCollection() *MockCollection {
	return &MockCollection{
		data: make(map[string]*models.Project),
	}
}

// MockProjectRepository for testing
type MockProjectRepository struct {
	collection *MockCollection
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		collection: NewMockCollection(),
	}
}

func (m *MockProjectRepository) Create(ctx context.Context, project *models.Project) error {
	project.ID = primitive.NewObjectID()
	m.collection.data[project.ID.Hex()] = project
	return nil
}

func (m *MockProjectRepository) FindAll(ctx context.Context) ([]*models.Project, error) {
	projects := make([]*models.Project, 0, len(m.collection.data))
	for _, p := range m.collection.data {
		projects = append(projects, p)
	}
	return projects, nil
}

func (m *MockProjectRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Project, error) {
	project, exists := m.collection.data[id.Hex()]
	if !exists {
		return nil, nil
	}
	return project, nil
}

func (m *MockProjectRepository) FindByJiraProjectKey(ctx context.Context, key string) (*models.Project, error) {
	for _, p := range m.collection.data {
		if p.JiraProjectKey == key {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockProjectRepository) Update(ctx context.Context, id primitive.ObjectID, project *models.Project) error {
	if _, exists := m.collection.data[id.Hex()]; !exists {
		return nil
	}
	project.ID = id
	m.collection.data[id.Hex()] = project
	return nil
}

func (m *MockProjectRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if _, exists := m.collection.data[id.Hex()]; !exists {
		return nil
	}
	delete(m.collection.data, id.Hex())
	return nil
}

func TestRepositoryCreate(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err := repo.Create(ctx, project)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if project.ID.IsZero() {
		t.Error("Expected project ID to be set")
	}

	// Verify project was stored
	stored, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stored == nil {
		t.Error("Expected project to be found")
	}

	if stored.Name != project.Name {
		t.Errorf("Expected name %s, got %s", project.Name, stored.Name)
	}
}

func TestRepositoryFindAll(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	// Create multiple projects
	for i := 1; i <= 3; i++ {
		project := &models.Project{
			Name:            "Test Project",
			Description:     "Test Description",
			Scope:           "Test Scope",
			JiraProjectKey:  "TEST",
			JiraProjectName: "Test Project",
			JiraProjectURL:  "https://jira.example.com/projects/TEST",
		}
		repo.Create(ctx, project)
	}

	projects, err := repo.FindAll(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func TestRepositoryFindByID(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	repo.Create(ctx, project)

	found, err := repo.FindByID(ctx, project.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found == nil {
		t.Error("Expected project to be found")
	}

	if found.ID != project.ID {
		t.Errorf("Expected ID %s, got %s", project.ID.Hex(), found.ID.Hex())
	}
}

func TestRepositoryFindByID_NotFound(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	id := primitive.NewObjectID()
	found, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found != nil {
		t.Error("Expected project to not be found")
	}
}

func TestRepositoryFindByJiraProjectKey(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST123",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	repo.Create(ctx, project)

	found, err := repo.FindByJiraProjectKey(ctx, "TEST123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found == nil {
		t.Error("Expected project to be found")
	}

	if found.JiraProjectKey != "TEST123" {
		t.Errorf("Expected JIRA key TEST123, got %s", found.JiraProjectKey)
	}
}

func TestRepositoryFindByJiraProjectKey_NotFound(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	found, err := repo.FindByJiraProjectKey(ctx, "NONEXISTENT")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if found != nil {
		t.Error("Expected project to not be found")
	}
}

func TestRepositoryUpdate(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	repo.Create(ctx, project)

	// Update project
	project.Name = "Updated Project"
	project.Description = "Updated Description"

	err := repo.Update(ctx, project.ID, project)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify update
	found, _ := repo.FindByID(ctx, project.ID)
	if found.Name != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got %s", found.Name)
	}

	if found.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %s", found.Description)
	}
}

func TestRepositoryUpdate_NotFound(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	id := primitive.NewObjectID()
	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	err := repo.Update(ctx, id, project)
	if err != nil {
		t.Errorf("Expected no error for non-existent update, got %v", err)
	}
}

func TestRepositoryDelete(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}

	repo.Create(ctx, project)

	// Verify project exists
	found, _ := repo.FindByID(ctx, project.ID)
	if found == nil {
		t.Error("Expected project to exist before deletion")
	}

	// Delete project
	err := repo.Delete(ctx, project.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify project was deleted
	found, _ = repo.FindByID(ctx, project.ID)
	if found != nil {
		t.Error("Expected project to be deleted")
	}
}

func TestRepositoryDelete_NotFound(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

	id := primitive.NewObjectID()
	err := repo.Delete(ctx, id)
	if err != nil {
		t.Errorf("Expected no error for non-existent delete, got %v", err)
	}
}

func TestRepositoryWithRepositories(t *testing.T) {
	repo := NewMockProjectRepository()
	ctx := context.Background()

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
				URL:            "https://github.com/test/repo1",
				Description:    "Repository 1",
				GitAccessToken: "token1",
			},
			{
				RepositoryID:   "repo2",
				URL:            "https://github.com/test/repo2",
				Description:    "Repository 2",
				GitAccessToken: "token2",
			},
		},
	}

	repo.Create(ctx, project)

	// Verify repositories were stored
	found, _ := repo.FindByID(ctx, project.ID)
	if len(found.Repositories) != 2 {
		t.Errorf("Expected 2 repositories, got %d", len(found.Repositories))
	}

	if found.Repositories[0].URL != "https://github.com/test/repo1" {
		t.Errorf("Expected URL https://github.com/test/repo1, got %s", found.Repositories[0].URL)
	}
}
