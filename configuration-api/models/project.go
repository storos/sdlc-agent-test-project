package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository represents a Git repository configuration
type Repository struct {
	RepositoryID     string `json:"repository_id" bson:"repository_id" binding:"required"`
	URL              string `json:"url" bson:"url" binding:"required,url"`
	Description      string `json:"description" bson:"description" binding:"required"`
	GitAccessToken   string `json:"git_access_token" bson:"git_access_token" binding:"required"`
	BaseBranch       string `json:"base_branch" bson:"base_branch"` // Base branch for PRs (e.g., "main", "master")
}

// Project represents a project configuration
type Project struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name            string             `json:"name" bson:"name" binding:"required"`
	Description     string             `json:"description" bson:"description" binding:"required"`
	Scope           string             `json:"scope" bson:"scope" binding:"required"`
	JiraProjectKey  string             `json:"jira_project_key" bson:"jira_project_key" binding:"required"`
	JiraProjectName string             `json:"jira_project_name" bson:"jira_project_name" binding:"required"`
	JiraProjectURL  string             `json:"jira_project_url" bson:"jira_project_url" binding:"required,url"`
	Repositories    []Repository       `json:"repositories" bson:"repositories"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name            string       `json:"name" binding:"required"`
	Description     string       `json:"description" binding:"required"`
	Scope           string       `json:"scope" binding:"required"`
	JiraProjectKey  string       `json:"jira_project_key" binding:"required"`
	JiraProjectName string       `json:"jira_project_name" binding:"required"`
	JiraProjectURL  string       `json:"jira_project_url" binding:"required,url"`
	Repositories    []Repository `json:"repositories"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Scope           string       `json:"scope"`
	JiraProjectKey  string       `json:"jira_project_key"`
	JiraProjectName string       `json:"jira_project_name"`
	JiraProjectURL  string       `json:"jira_project_url"`
	Repositories    []Repository `json:"repositories"`
}

// AddRepositoryRequest represents the request body for adding a repository
type AddRepositoryRequest struct {
	URL             string `json:"url" binding:"required,url"`
	Description     string `json:"description" binding:"required"`
	GitAccessToken  string `json:"git_access_token" binding:"required"`
	BaseBranch      string `json:"base_branch"` // Base branch for PRs (defaults to "main" if not specified)
}

// UpdateRepositoryRequest represents the request body for updating a repository
type UpdateRepositoryRequest struct {
	URL             string `json:"url" binding:"url"`
	Description     string `json:"description"`
	GitAccessToken  string `json:"git_access_token"`
	BaseBranch      string `json:"base_branch"`
}
