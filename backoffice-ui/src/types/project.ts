export interface Repository {
  repository_id: string;
  url: string;
  description: string;
  git_access_token: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  scope: string;
  jira_project_key: string;
  jira_project_name: string;
  jira_project_url: string;
  repositories: Repository[];
  created_at: string;
  updated_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description: string;
  scope: string;
  jira_project_key: string;
  jira_project_name: string;
  jira_project_url: string;
  repositories?: Repository[];
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  scope?: string;
  jira_project_key?: string;
  jira_project_name?: string;
  jira_project_url?: string;
  repositories?: Repository[];
}

export interface AddRepositoryRequest {
  url: string;
  description: string;
  git_access_token: string;
}

export interface UpdateRepositoryRequest {
  url?: string;
  description?: string;
  git_access_token?: string;
}
