package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/storos/sdlc-agent/configuration-api/models"
	"github.com/storos/sdlc-agent/configuration-api/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockProjectRepository is a mock implementation for testing
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

// Setup test router
func setupTestRouter(handler *ProjectHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/api/projects", handler.GetAllProjects)
	router.GET("/api/projects/:id", handler.GetProjectByID)
	router.POST("/api/projects", handler.CreateProject)
	router.PUT("/api/projects/:id", handler.UpdateProject)
	router.DELETE("/api/projects/:id", handler.DeleteProject)
	router.GET("/api/projects/:id/repositories", handler.GetRepositories)
	router.POST("/api/projects/:id/repositories", handler.AddRepository)
	router.PUT("/api/repositories/:id", handler.UpdateRepository)
	router.DELETE("/api/repositories/:id", handler.DeleteRepository)

	return router
}

func TestCreateProject_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	projectData := map[string]interface{}{
		"name":              "Test Project",
		"description":       "Test Description",
		"scope":             "Test Scope",
		"jira_project_key":  "TEST",
		"jira_project_name": "Test Project",
		"jira_project_url":  "https://jira.example.com/projects/TEST",
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response models.Project
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %s", response.Name)
	}
}

func TestCreateProject_InvalidData(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	projectData := map[string]interface{}{
		"name": "Test Project",
		// Missing required fields
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAllProjects(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	// Create a test project
	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}
	service.CreateProject(nil, project)

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var projects []models.Project
	json.Unmarshal(w.Body.Bytes(), &projects)

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

func TestGetAllProjects_WithJiraKeyFilter(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	// Create test projects
	project1 := &models.Project{
		Name:            "Test Project 1",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST1",
		JiraProjectName: "Test Project 1",
		JiraProjectURL:  "https://jira.example.com/projects/TEST1",
	}
	service.CreateProject(nil, project1)

	project2 := &models.Project{
		Name:            "Test Project 2",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST2",
		JiraProjectName: "Test Project 2",
		JiraProjectURL:  "https://jira.example.com/projects/TEST2",
	}
	service.CreateProject(nil, project2)

	req, _ := http.NewRequest("GET", "/api/projects?jira_project_key=TEST1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.Project
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.JiraProjectKey != "TEST1" {
		t.Errorf("Expected JIRA key TEST1, got %s", response.JiraProjectKey)
	}
}

func TestGetProjectByID_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}
	service.CreateProject(nil, project)

	req, _ := http.NewRequest("GET", "/api/projects/"+project.ID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.Project
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got %s", response.Name)
	}
}

func TestGetProjectByID_InvalidID(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	req, _ := http.NewRequest("GET", "/api/projects/invalid-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	id := primitive.NewObjectID()
	req, _ := http.NewRequest("GET", "/api/projects/"+id.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUpdateProject_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}
	service.CreateProject(nil, project)

	updateData := map[string]interface{}{
		"name":              "Updated Project",
		"description":       "Updated Description",
		"scope":             "Updated Scope",
		"jira_project_key":  "TEST",
		"jira_project_name": "Updated Project",
		"jira_project_url":  "https://jira.example.com/projects/TEST",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/api/projects/"+project.ID.Hex(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteProject_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
	}
	service.CreateProject(nil, project)

	req, _ := http.NewRequest("DELETE", "/api/projects/"+project.ID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestAddRepository_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

	project := &models.Project{
		Name:            "Test Project",
		Description:     "Test Description",
		Scope:           "Test Scope",
		JiraProjectKey:  "TEST",
		JiraProjectName: "Test Project",
		JiraProjectURL:  "https://jira.example.com/projects/TEST",
		Repositories:    []models.Repository{},
	}
	service.CreateProject(nil, project)

	repoData := map[string]interface{}{
		"url":              "https://github.com/test/repo",
		"description":      "Test Repo",
		"git_access_token": "token123",
	}

	jsonData, _ := json.Marshal(repoData)
	req, _ := http.NewRequest("POST", "/api/projects/"+project.ID.Hex()+"/repositories", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestGetRepositories_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

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
	service.CreateProject(nil, project)

	req, _ := http.NewRequest("GET", "/api/projects/"+project.ID.Hex()+"/repositories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var repos []models.Repository
	json.Unmarshal(w.Body.Bytes(), &repos)

	if len(repos) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(repos))
	}
}

func TestUpdateRepository_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

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
	service.CreateProject(nil, project)

	updateData := map[string]interface{}{
		"url":              "https://github.com/test/repo-updated",
		"description":      "Updated Repo",
		"git_access_token": "token456",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/api/repositories/repo1?project_id="+project.ID.Hex(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteRepository_Success(t *testing.T) {
	mockRepo := NewMockProjectRepository()
	service := services.NewProjectService(mockRepo)
	handler := NewProjectHandler(service)
	router := setupTestRouter(handler)

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
	service.CreateProject(nil, project)

	req, _ := http.NewRequest("DELETE", "/api/repositories/repo1?project_id="+project.ID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}
