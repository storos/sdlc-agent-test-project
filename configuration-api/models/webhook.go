package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebhookEvent struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JiraIssueID    string             `bson:"jira_issue_id" json:"jira_issue_id"`
	JiraIssueKey   string             `bson:"jira_issue_key" json:"jira_issue_key"`
	JiraProjectKey string             `bson:"jira_project_key" json:"jira_project_key"`
	Summary        string             `bson:"summary" json:"summary"`
	Description    string             `bson:"description" json:"description"`
	Status         string             `bson:"status" json:"status"`
	PreviousStatus string             `bson:"previous_status" json:"previous_status"`
	EventType      string             `bson:"event_type" json:"event_type"`
	ReceivedAt     time.Time          `bson:"received_at" json:"received_at"`
	ProcessedAt    *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
	RawPayload     interface{}        `bson:"raw_payload" json:"raw_payload"`
}
