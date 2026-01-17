export interface Development {
  id: string;
  jira_issue_id: string;
  jira_issue_key: string;
  jira_project_key: string;
  repository_url: string;
  branch_name: string;
  pr_mr_url?: string;
  status: 'ready' | 'completed' | 'failed';
  development_details?: string;
  error_message?: string;
  created_at: string;
  completed_at?: string;
}
