package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JiraWebhookPayload represents the incoming JIRA webhook
type JiraWebhookPayload struct {
	WebhookEvent   string      `json:"webhookEvent"`
	IssueEventType string      `json:"issue_event_type_name"`
	Issue          JiraIssue   `json:"issue"`
	Changelog      *Changelog  `json:"changelog,omitempty"`
	User           *JiraUser   `json:"user,omitempty"`
}

// JiraIssue represents JIRA issue details
type JiraIssue struct {
	ID     string          `json:"id"`
	Key    string          `json:"key"`
	Self   string          `json:"self"`
	Fields JiraIssueFields `json:"fields"`
}

// JiraIssueFields contains issue field data
type JiraIssueFields struct {
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Status      JiraStatus   `json:"status"`
	Project     JiraProject  `json:"project"`
	IssueType   JiraIssueType `json:"issuetype"`
}

// JiraStatus represents issue status
type JiraStatus struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// JiraProject represents project information
type JiraProject struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Self string `json:"self"`
}

// JiraIssueType represents issue type
type JiraIssueType struct {
	Name string `json:"name"`
}

// JiraUser represents user information
type JiraUser struct {
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
}

// Changelog represents status change information
type Changelog struct {
	Items []ChangelogItem `json:"items"`
}

// ChangelogItem represents individual change
type ChangelogItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	From       string `json:"from"`
	FromString string `json:"fromString"`
	To         string `json:"to"`
	ToString   string `json:"toString"`
}

// WebhookEvent represents stored webhook event in MongoDB
type WebhookEvent struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JiraIssueID      string             `bson:"jira_issue_id" json:"jira_issue_id"`
	JiraIssueKey     string             `bson:"jira_issue_key" json:"jira_issue_key"`
	JiraProjectKey   string             `bson:"jira_project_key" json:"jira_project_key"`
	Summary          string             `bson:"summary" json:"summary"`
	Description      string             `bson:"description" json:"description"`
	Status           string             `bson:"status" json:"status"`
	PreviousStatus   string             `bson:"previous_status" json:"previous_status"`
	EventType        string             `bson:"event_type" json:"event_type"`
	ReceivedAt       time.Time          `bson:"received_at" json:"received_at"`
	ProcessedAt      *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	RawPayload       interface{}        `bson:"raw_payload" json:"raw_payload"`
}

// DevelopmentRequest represents message sent to RabbitMQ
type DevelopmentRequest struct {
	JiraIssueID    string `json:"jira_issue_id"`
	JiraIssueKey   string `json:"jira_issue_key"`
	JiraProjectKey string `json:"jira_project_key"`
	Summary        string `json:"summary"`
	Description    string `json:"description"`
	Repository     string `json:"repository,omitempty"`
}
