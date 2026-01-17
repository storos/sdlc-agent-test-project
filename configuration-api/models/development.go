package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Development struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JiraIssueID        string             `bson:"jira_issue_id" json:"jira_issue_id"`
	JiraIssueKey       string             `bson:"jira_issue_key" json:"jira_issue_key"`
	JiraProjectKey     string             `bson:"jira_project_key" json:"jira_project_key"`
	RepositoryURL      string             `bson:"repository_url" json:"repository_url"`
	BranchName         string             `bson:"branch_name" json:"branch_name"`
	PRMRUrl            string             `bson:"pr_mr_url,omitempty" json:"pr_mr_url,omitempty"`
	Status             string             `bson:"status" json:"status"`
	DevelopmentDetails string             `bson:"development_details,omitempty" json:"development_details,omitempty"`
	ErrorMessage       string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	CompletedAt        *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}
