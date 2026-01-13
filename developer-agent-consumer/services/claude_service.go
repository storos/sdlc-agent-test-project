package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

type ClaudeService struct {
	apiURL     string
	httpClient *http.Client
	logger     *logrus.Logger
}

func NewClaudeService(apiURL string, logger *logrus.Logger) *ClaudeService {
	return &ClaudeService{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes for code generation
		},
		logger: logger,
	}
}

func (s *ClaudeService) GenerateCode(
	request *models.DevelopmentRequest,
	project *models.Project,
	analysis *models.RepositoryAnalysis,
	repoPath string,
	sessionToken string,
) (*models.ClaudeCodeResponse, error) {
	// Build the prompt
	prompt := s.buildPrompt(request, project, analysis)

	s.logger.WithFields(logrus.Fields{
		"jira_issue_key": request.JiraIssueKey,
		"prompt_length":  len(prompt),
	}).Info("Calling Claude Code API")

	// Create request payload
	claudeRequest := &models.ClaudeCodeRequest{
		Prompt:         prompt,
		ProjectContext: s.buildProjectContext(project, analysis),
		RepositoryPath: repoPath,
		SessionToken:   sessionToken,
	}

	jsonData, err := json.Marshal(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude Code API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var claudeResponse models.ClaudeCodeResponse
	if err := json.Unmarshal(body, &claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !claudeResponse.Success {
		return nil, fmt.Errorf("Claude Code API returned error: %s", claudeResponse.Error)
	}

	s.logger.WithFields(logrus.Fields{
		"jira_issue_key": request.JiraIssueKey,
		"files_changed":  claudeResponse.FilesChanged,
	}).Info("Code generation completed successfully")

	return &claudeResponse, nil
}

func (s *ClaudeService) buildPrompt(
	request *models.DevelopmentRequest,
	project *models.Project,
	analysis *models.RepositoryAnalysis,
) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("# Development Task: %s\n\n", request.JiraIssueKey))
	prompt.WriteString(fmt.Sprintf("## Summary\n%s\n\n", request.Summary))

	if request.Description != "" {
		prompt.WriteString(fmt.Sprintf("## Description\n%s\n\n", request.Description))
	}

	prompt.WriteString("## Project Context\n\n")
	prompt.WriteString(fmt.Sprintf("**Project**: %s\n", project.Name))
	prompt.WriteString(fmt.Sprintf("**Project Type**: %s\n", analysis.ProjectType))

	if len(analysis.Languages) > 0 {
		prompt.WriteString(fmt.Sprintf("**Languages**: %s\n", strings.Join(analysis.Languages, ", ")))
	}

	if project.Scope != "" {
		prompt.WriteString(fmt.Sprintf("**Project Scope**: %s\n\n", project.Scope))
	}

	// Add repository structure
	prompt.WriteString("## Repository Structure\n\n")

	if len(analysis.EntryPoints) > 0 {
		prompt.WriteString("**Entry Points**:\n")
		for _, ep := range analysis.EntryPoints {
			prompt.WriteString(fmt.Sprintf("- %s\n", ep))
		}
		prompt.WriteString("\n")
	}

	if len(analysis.KeyDirectories) > 0 {
		prompt.WriteString("**Key Directories**:\n")
		for _, dir := range analysis.KeyDirectories {
			prompt.WriteString(fmt.Sprintf("- %s\n", dir))
		}
		prompt.WriteString("\n")
	}

	if len(analysis.Patterns) > 0 {
		prompt.WriteString("**Detected Patterns**:\n")
		for key, value := range analysis.Patterns {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		prompt.WriteString("\n")
	}

	// Add instructions
	prompt.WriteString("## Instructions\n\n")
	prompt.WriteString("Please implement the requested changes following these guidelines:\n\n")
	prompt.WriteString("1. Follow the existing code patterns and architecture detected in the repository\n")
	prompt.WriteString("2. Maintain consistency with the project's coding style and conventions\n")
	prompt.WriteString("3. Add appropriate error handling and logging\n")
	prompt.WriteString("4. Ensure the changes integrate seamlessly with existing code\n")
	prompt.WriteString("5. Write clean, maintainable, and well-documented code\n")
	prompt.WriteString("6. Add unit tests if applicable\n\n")

	prompt.WriteString("Please implement the changes and provide a summary of what was modified.\n")

	return prompt.String()
}

func (s *ClaudeService) buildProjectContext(project *models.Project, analysis *models.RepositoryAnalysis) string {
	var context strings.Builder

	context.WriteString(fmt.Sprintf("Project: %s\n", project.Name))
	context.WriteString(fmt.Sprintf("Type: %s\n", analysis.ProjectType))

	if len(analysis.Languages) > 0 {
		context.WriteString(fmt.Sprintf("Languages: %s\n", strings.Join(analysis.Languages, ", ")))
	}

	if len(analysis.DependencyManagers) > 0 {
		context.WriteString(fmt.Sprintf("Dependency Managers: %s\n", strings.Join(analysis.DependencyManagers, ", ")))
	}

	if arch, ok := analysis.Patterns["architecture"]; ok {
		context.WriteString(fmt.Sprintf("Architecture: %s\n", arch))
	}

	return context.String()
}
