package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/configuration-api/models"
	"github.com/storos/sdlc-agent/configuration-api/services"
)

// ProjectHandler handles HTTP requests for projects
type ProjectHandler struct {
	service *services.ProjectService
	logger  *logrus.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service *services.ProjectService, logger *logrus.Logger) *ProjectHandler {
	return &ProjectHandler{
		service: service,
		logger:  logger,
	}
}

// GetProjects returns all projects or filters by JIRA project key
// GET /api/projects
// GET /api/projects?jira_project_key=ECOM
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	jiraProjectKey := c.Query("jira_project_key")

	if jiraProjectKey != "" {
		// Filter by JIRA project key
		project, err := h.service.GetProjectByJiraKey(c.Request.Context(), jiraProjectKey)
		if err != nil {
			if errors.Is(err, services.ErrProjectNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
				return
			}
			h.logger.WithError(err).Error("Failed to get project by JIRA key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, project)
		return
	}

	// Return all projects
	projects, err := h.service.GetAllProjects(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get all projects")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject returns a project by ID
// GET /api/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")

	project, err := h.service.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// CreateProject creates a new project
// POST /api/projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.service.CreateProject(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateProjectKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "Project with this JIRA key already exists"})
			return
		}
		h.logger.WithError(err).Error("Failed to create project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithField("project_id", project.ID.Hex()).Info("Project created successfully")
	c.JSON(http.StatusCreated, project)
}

// UpdateProject updates an existing project
// PUT /api/projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateProject(c.Request.Context(), id, &req); err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		if errors.Is(err, services.ErrDuplicateProjectKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "Project with this JIRA key already exists"})
			return
		}
		h.logger.WithError(err).Error("Failed to update project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithField("project_id", id).Info("Project updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully"})
}

// DeleteProject deletes a project
// DELETE /api/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteProject(c.Request.Context(), id); err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to delete project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithField("project_id", id).Info("Project deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// GetRepositories returns all repositories for a project
// GET /api/projects/:id/repositories
func (h *ProjectHandler) GetRepositories(c *gin.Context) {
	id := c.Param("id")

	project, err := h.service.GetProjectByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get project")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, project.Repositories)
}

// AddRepository adds a repository to a project
// POST /api/projects/:id/repositories
func (h *ProjectHandler) AddRepository(c *gin.Context) {
	id := c.Param("id")

	var req models.AddRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddRepository(c.Request.Context(), id, &req); err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to add repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithField("project_id", id).Info("Repository added successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Repository added successfully"})
}

// UpdateRepository updates a repository in a project
// PUT /api/repositories/:id
func (h *ProjectHandler) UpdateRepository(c *gin.Context) {
	repoID := c.Param("id")
	projectID := c.Query("project_id")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id query parameter is required"})
		return
	}

	var req models.UpdateRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateRepository(c.Request.Context(), projectID, repoID, &req); err != nil {
		if errors.Is(err, services.ErrRepositoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to update repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":    projectID,
		"repository_id": repoID,
	}).Info("Repository updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Repository updated successfully"})
}

// DeleteRepository deletes a repository from a project
// DELETE /api/repositories/:id
func (h *ProjectHandler) DeleteRepository(c *gin.Context) {
	repoID := c.Param("id")
	projectID := c.Query("project_id")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id query parameter is required"})
		return
	}

	if err := h.service.DeleteRepository(c.Request.Context(), projectID, repoID); err != nil {
		if errors.Is(err, services.ErrRepositoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to delete repository")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"project_id":    projectID,
		"repository_id": repoID,
	}).Info("Repository deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
}
