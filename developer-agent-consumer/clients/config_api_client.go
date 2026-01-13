package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

type ConfigAPIClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewConfigAPIClient(baseURL string, logger *logrus.Logger) *ConfigAPIClient {
	return &ConfigAPIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (c *ConfigAPIClient) GetProjectByJiraKey(jiraProjectKey string) (*models.Project, error) {
	// Build URL with query parameter
	endpoint := fmt.Sprintf("%s/api/projects", c.baseURL)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("jira_project_key", jiraProjectKey)
	reqURL.RawQuery = query.Encode()

	c.logger.WithFields(logrus.Fields{
		"url":              reqURL.String(),
		"jira_project_key": jiraProjectKey,
	}).Debug("Fetching project from Configuration API")

	// Make HTTP request
	resp, err := c.httpClient.Get(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project not found for jira_project_key: %s", jiraProjectKey)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (single project object when filtered by jira_project_key)
	var project models.Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"project_id":       project.ID,
		"project_name":     project.Name,
		"repositories":     len(project.Repositories),
		"jira_project_key": jiraProjectKey,
	}).Info("Project fetched successfully")

	return &project, nil
}

func (c *ConfigAPIClient) FindRepositoryInProject(project *models.Project, repoURL string) (*models.Repository, error) {
	// Normalize repository URLs for comparison (remove trailing slashes, .git suffixes)
	normalizedSearchURL := normalizeRepoURL(repoURL)

	for _, repo := range project.Repositories {
		normalizedRepoURL := normalizeRepoURL(repo.URL)
		if normalizedRepoURL == normalizedSearchURL {
			c.logger.WithFields(logrus.Fields{
				"repository_id": repo.RepositoryID,
				"repository_url": repo.URL,
			}).Info("Repository matched in project")
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("repository %s not found in project %s", repoURL, project.Name)
}

func normalizeRepoURL(repoURL string) string {
	// Remove trailing slash
	if len(repoURL) > 0 && repoURL[len(repoURL)-1] == '/' {
		repoURL = repoURL[:len(repoURL)-1]
	}

	// Remove .git suffix
	if len(repoURL) > 4 && repoURL[len(repoURL)-4:] == ".git" {
		repoURL = repoURL[:len(repoURL)-4]
	}

	return repoURL
}
