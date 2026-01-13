package models

import (
	"encoding/json"
	"testing"
)

func TestJiraWebhookPayload_Unmarshal(t *testing.T) {
	jsonData := `{
		"webhookEvent": "jira:issue_updated",
		"issue_event_type_name": "issue_generic",
		"issue": {
			"id": "10001",
			"key": "PROJ-123",
			"fields": {
				"summary": "Test Issue",
				"description": "Test Description",
				"status": {
					"name": "In Development",
					"id": "3"
				},
				"project": {
					"key": "PROJ",
					"name": "Test Project",
					"self": "https://jira.example.com/rest/api/2/project/10000"
				},
				"issuetype": {
					"name": "Task"
				}
			}
		},
		"changelog": {
			"items": [
				{
					"field": "status",
					"fieldtype": "jira",
					"from": "1",
					"fromString": "To Do",
					"to": "3",
					"toString": "In Development"
				}
			]
		}
	}`

	var payload JiraWebhookPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Validate webhook event
	if payload.WebhookEvent != "jira:issue_updated" {
		t.Errorf("Expected webhookEvent 'jira:issue_updated', got '%s'", payload.WebhookEvent)
	}

	// Validate issue
	if payload.Issue.Key != "PROJ-123" {
		t.Errorf("Expected issue key 'PROJ-123', got '%s'", payload.Issue.Key)
	}

	if payload.Issue.Fields.Summary != "Test Issue" {
		t.Errorf("Expected summary 'Test Issue', got '%s'", payload.Issue.Fields.Summary)
	}

	// Validate status
	if payload.Issue.Fields.Status.Name != "In Development" {
		t.Errorf("Expected status 'In Development', got '%s'", payload.Issue.Fields.Status.Name)
	}

	// Validate project
	if payload.Issue.Fields.Project.Key != "PROJ" {
		t.Errorf("Expected project key 'PROJ', got '%s'", payload.Issue.Fields.Project.Key)
	}

	// Validate changelog
	if payload.Changelog == nil {
		t.Fatal("Expected changelog to be present")
	}

	if len(payload.Changelog.Items) != 1 {
		t.Fatalf("Expected 1 changelog item, got %d", len(payload.Changelog.Items))
	}

	item := payload.Changelog.Items[0]
	if item.Field != "status" {
		t.Errorf("Expected field 'status', got '%s'", item.Field)
	}

	if item.FromString != "To Do" {
		t.Errorf("Expected fromString 'To Do', got '%s'", item.FromString)
	}

	if item.ToString != "In Development" {
		t.Errorf("Expected toString 'In Development', got '%s'", item.ToString)
	}
}

func TestDevelopmentRequest_Marshal(t *testing.T) {
	request := &DevelopmentRequest{
		JiraIssueID:    "10001",
		JiraIssueKey:   "PROJ-123",
		JiraProjectKey: "PROJ",
		Summary:        "Test Issue",
		Description:    "Test Description",
		Repository:     "https://github.com/test/repo",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	var unmarshaled DevelopmentRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled.JiraIssueKey != "PROJ-123" {
		t.Errorf("Expected issue key 'PROJ-123', got '%s'", unmarshaled.JiraIssueKey)
	}

	if unmarshaled.JiraProjectKey != "PROJ" {
		t.Errorf("Expected project key 'PROJ', got '%s'", unmarshaled.JiraProjectKey)
	}

	if unmarshaled.Summary != "Test Issue" {
		t.Errorf("Expected summary 'Test Issue', got '%s'", unmarshaled.Summary)
	}
}

func TestWebhookEvent_Structure(t *testing.T) {
	event := &WebhookEvent{
		JiraIssueID:    "10001",
		JiraIssueKey:   "PROJ-123",
		JiraProjectKey: "PROJ",
		Summary:        "Test Issue",
		Description:    "Test Description",
		Status:         "In Development",
		PreviousStatus: "To Do",
		EventType:      "jira:issue_updated",
	}

	// Validate structure
	if event.JiraIssueKey != "PROJ-123" {
		t.Errorf("Expected issue key 'PROJ-123', got '%s'", event.JiraIssueKey)
	}

	if event.Status != "In Development" {
		t.Errorf("Expected status 'In Development', got '%s'", event.Status)
	}

	if event.PreviousStatus != "To Do" {
		t.Errorf("Expected previous status 'To Do', got '%s'", event.PreviousStatus)
	}
}

func TestJiraWebhookPayload_NoChangelog(t *testing.T) {
	jsonData := `{
		"webhookEvent": "jira:issue_created",
		"issue": {
			"id": "10002",
			"key": "PROJ-124",
			"fields": {
				"summary": "New Issue",
				"description": "New Description",
				"status": {
					"name": "In Development"
				},
				"project": {
					"key": "PROJ"
				},
				"issuetype": {
					"name": "Bug"
				}
			}
		}
	}`

	var payload JiraWebhookPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if payload.Changelog != nil {
		t.Error("Expected no changelog for issue creation")
	}

	if payload.Issue.Key != "PROJ-124" {
		t.Errorf("Expected issue key 'PROJ-124', got '%s'", payload.Issue.Key)
	}
}
