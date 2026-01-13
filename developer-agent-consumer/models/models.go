package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DevelopmentRequest represents incoming message from RabbitMQ
type DevelopmentRequest struct {
	JiraIssueID    string `json:"jira_issue_id"`
	JiraIssueKey   string `json:"jira_issue_key"`
	JiraProjectKey string `json:"jira_project_key"`
	Summary        string `json:"summary"`
	Description    string `json:"description"`
	Repository     string `json:"repository,omitempty"`
}

// Development represents development record in MongoDB
type Development struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JiraIssueID       string             `bson:"jira_issue_id" json:"jira_issue_id"`
	JiraIssueKey      string             `bson:"jira_issue_key" json:"jira_issue_key"`
	JiraProjectKey    string             `bson:"jira_project_key" json:"jira_project_key"`
	RepositoryURL     string             `bson:"repository_url" json:"repository_url"`
	BranchName        string             `bson:"branch_name" json:"branch_name"`
	PRMRUrl           string             `bson:"pr_mr_url,omitempty" json:"pr_mr_url,omitempty"`
	Status            string             `bson:"status" json:"status"` // ready, completed, failed
	DevelopmentDetails string            `bson:"development_details,omitempty" json:"development_details,omitempty"`
	ErrorMessage      string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	CompletedAt       *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// Project represents project configuration from Configuration API
type Project struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Scope           string       `json:"scope"`
	JiraProjectKey  string       `json:"jira_project_key"`
	JiraProjectName string       `json:"jira_project_name"`
	JiraProjectURL  string       `json:"jira_project_url"`
	Repositories    []Repository `json:"repositories"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
}

// Repository represents repository configuration
type Repository struct {
	RepositoryID   string `json:"repository_id"`
	URL            string `json:"url"`
	Description    string `json:"description"`
	GitAccessToken string `json:"git_access_token"`
}

// RepositoryAnalysis represents analyzed repository structure
type RepositoryAnalysis struct {
	EntryPoints        []string          `json:"entry_points"`
	KeyDirectories     []string          `json:"key_directories"`
	ConfigFiles        []string          `json:"config_files"`
	Languages          []string          `json:"languages"`
	Patterns           map[string]string `json:"patterns"`
	ProjectType        string            `json:"project_type"`
	DependencyManagers []string          `json:"dependency_managers"`
}

// ClaudeCodeRequest represents request to Claude Code API
type ClaudeCodeRequest struct {
	Prompt           string `json:"prompt"`
	ProjectContext   string `json:"project_context,omitempty"`
	RepositoryPath   string `json:"repository_path,omitempty"`
	SessionToken     string `json:"session_token"`
}

// ClaudeCodeResponse represents response from Claude Code API
type ClaudeCodeResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	FilesChanged     int    `json:"files_changed"`
	DevelopmentDetails string `json:"development_details"`
	Error            string `json:"error,omitempty"`
}
