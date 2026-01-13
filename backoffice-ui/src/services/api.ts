import axios, { AxiosInstance } from 'axios';
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
  Repository,
  AddRepositoryRequest,
  UpdateRepositoryRequest,
} from '../types/project';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8081/api',
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  // Project endpoints
  async getAllProjects(): Promise<Project[]> {
    const response = await this.client.get<Project[]>('/projects');
    return response.data;
  }

  async getProjectById(id: string): Promise<Project> {
    const response = await this.client.get<Project>(`/projects/${id}`);
    return response.data;
  }

  async getProjectByJiraKey(jiraProjectKey: string): Promise<Project> {
    const response = await this.client.get<Project>('/projects', {
      params: { jira_project_key: jiraProjectKey },
    });
    return response.data;
  }

  async createProject(data: CreateProjectRequest): Promise<Project> {
    const response = await this.client.post<Project>('/projects', data);
    return response.data;
  }

  async updateProject(id: string, data: UpdateProjectRequest): Promise<void> {
    await this.client.put(`/projects/${id}`, data);
  }

  async deleteProject(id: string): Promise<void> {
    await this.client.delete(`/projects/${id}`);
  }

  // Repository endpoints
  async getRepositories(projectId: string): Promise<Repository[]> {
    const response = await this.client.get<Repository[]>(
      `/projects/${projectId}/repositories`
    );
    return response.data;
  }

  async addRepository(projectId: string, data: AddRepositoryRequest): Promise<void> {
    await this.client.post(`/projects/${projectId}/repositories`, data);
  }

  async updateRepository(
    projectId: string,
    repositoryId: string,
    data: UpdateRepositoryRequest
  ): Promise<void> {
    await this.client.put(`/repositories/${repositoryId}`, data, {
      params: { project_id: projectId },
    });
  }

  async deleteRepository(projectId: string, repositoryId: string): Promise<void> {
    await this.client.delete(`/repositories/${repositoryId}`, {
      params: { project_id: projectId },
    });
  }
}

export const api = new ApiClient();
