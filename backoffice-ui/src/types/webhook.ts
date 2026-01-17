export interface WebhookEvent {
  id: string;
  jira_issue_id: string;
  jira_issue_key: string;
  jira_project_key: string;
  summary: string;
  description: string;
  status: string;
  previous_status: string;
  event_type: string;
  received_at: string;
  processed_at?: string;
  raw_payload?: any;
}
