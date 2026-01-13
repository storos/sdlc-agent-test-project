package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PRService struct {
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewPRService(logger *logrus.Logger) *PRService {
	return &PRService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

type RepoInfo struct {
	Platform string // "github" or "gitlab"
	Owner    string
	Repo     string
	BaseURL  string // For GitLab self-hosted instances
}

func (s *PRService) CreatePullRequest(
	repoURL, branchName, jiraIssueKey, summary, description, accessToken string,
) (string, error) {
	// Parse repository URL to determine platform
	repoInfo, err := s.parseRepoURL(repoURL)
	if err != nil {
		return "", err
	}

	s.logger.WithFields(logrus.Fields{
		"platform":       repoInfo.Platform,
		"owner":          repoInfo.Owner,
		"repo":           repoInfo.Repo,
		"branch":         branchName,
		"jira_issue_key": jiraIssueKey,
	}).Info("Creating pull/merge request")

	if repoInfo.Platform == "github" {
		return s.createGitHubPR(repoInfo, branchName, jiraIssueKey, summary, description, accessToken)
	} else if repoInfo.Platform == "gitlab" {
		return s.createGitLabMR(repoInfo, branchName, jiraIssueKey, summary, description, accessToken)
	}

	return "", fmt.Errorf("unsupported platform: %s", repoInfo.Platform)
}

func (s *PRService) parseRepoURL(repoURL string) (*RepoInfo, error) {
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	info := &RepoInfo{}

	// Determine platform
	if strings.Contains(parsedURL.Host, "github.com") {
		info.Platform = "github"
		info.BaseURL = "https://api.github.com"
	} else if strings.Contains(parsedURL.Host, "gitlab.com") {
		info.Platform = "gitlab"
		info.BaseURL = "https://gitlab.com/api/v4"
	} else if strings.Contains(parsedURL.Host, "gitlab") {
		// Self-hosted GitLab
		info.Platform = "gitlab"
		info.BaseURL = fmt.Sprintf("%s://%s/api/v4", parsedURL.Scheme, parsedURL.Host)
	} else {
		return nil, fmt.Errorf("could not determine platform from URL: %s", repoURL)
	}

	// Parse owner and repo from path
	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository path: %s", path)
	}

	info.Owner = parts[0]
	info.Repo = parts[1]

	return info, nil
}

func (s *PRService) createGitHubPR(
	repoInfo *RepoInfo,
	branchName, jiraIssueKey, summary, description, accessToken string,
) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/pulls", repoInfo.BaseURL, repoInfo.Owner, repoInfo.Repo)

	title := fmt.Sprintf("[%s] %s", jiraIssueKey, summary)
	body := s.buildPRBody(jiraIssueKey, description)

	payload := map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  branchName,
		"base":  "main", // Default to main, could be configurable
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}
	defer resp.Body.Close()

	body_bytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body_bytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body_bytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	prURL := result["html_url"].(string)

	s.logger.WithFields(logrus.Fields{
		"pr_url": prURL,
	}).Info("GitHub PR created successfully")

	return prURL, nil
}

func (s *PRService) createGitLabMR(
	repoInfo *RepoInfo,
	branchName, jiraIssueKey, summary, description, accessToken string,
) (string, error) {
	// Get project ID first
	projectPath := fmt.Sprintf("%s/%s", repoInfo.Owner, repoInfo.Repo)
	encodedPath := url.PathEscape(projectPath)
	projectURL := fmt.Sprintf("%s/projects/%s", repoInfo.BaseURL, encodedPath)

	req, err := http.NewRequest("GET", projectURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitLab API returned status %d: %s", resp.StatusCode, string(body))
	}

	var project map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return "", fmt.Errorf("failed to parse project: %w", err)
	}

	projectID := int(project["id"].(float64))

	// Create MR
	apiURL := fmt.Sprintf("%s/projects/%d/merge_requests", repoInfo.BaseURL, projectID)

	title := fmt.Sprintf("[%s] %s", jiraIssueKey, summary)
	mrDescription := s.buildPRBody(jiraIssueKey, description)

	payload := map[string]interface{}{
		"source_branch": branchName,
		"target_branch": "main", // Default to main, could be configurable
		"title":         title,
		"description":   mrDescription,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err = http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create MR: %w", err)
	}
	defer resp.Body.Close()

	body_bytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitLab API returned status %d: %s", resp.StatusCode, string(body_bytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body_bytes, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	mrURL := result["web_url"].(string)

	s.logger.WithFields(logrus.Fields{
		"mr_url": mrURL,
	}).Info("GitLab MR created successfully")

	return mrURL, nil
}

func (s *PRService) buildPRBody(jiraIssueKey, description string) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("## JIRA Issue: %s\n\n", jiraIssueKey))

	if description != "" {
		body.WriteString("## Description\n\n")
		body.WriteString(description)
		body.WriteString("\n\n")
	}

	body.WriteString("---\n\n")
	body.WriteString("*This pull request was automatically generated by SDLC AI Agent*\n")

	return body.String()
}
