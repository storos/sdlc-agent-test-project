package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/storos/sdlc-agent/developer-agent-consumer/models"
)

type ClaudeService struct {
	claudePath string
	logger     *logrus.Logger
}

func NewClaudeService(claudePath string, logger *logrus.Logger) *ClaudeService {
	// Default to 'claude' command if not specified
	if claudePath == "" {
		claudePath = "claude"
	}
	return &ClaudeService{
		claudePath: claudePath,
		logger:     logger,
	}
}

// BuildPrompt creates the prompt that will be sent to Claude CLI
func (s *ClaudeService) BuildPrompt(
	request *models.DevelopmentRequest,
	project *models.Project,
	analysis *models.RepositoryAnalysis,
) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("# Development Task: %s\n\n", request.JiraIssueKey))
	prompt.WriteString(fmt.Sprintf("## Summary\n%s\n\n", request.Summary))
	prompt.WriteString(fmt.Sprintf("## Description\n%s\n\n", request.Description))

	if project.Scope != "" {
		prompt.WriteString(fmt.Sprintf("## Project Scope\n%s\n\n", project.Scope))
	}

	prompt.WriteString("## Repository Context\n")
	prompt.WriteString(fmt.Sprintf("- Project Type: %s\n", analysis.ProjectType))
	prompt.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(analysis.Languages, ", ")))
	if len(analysis.EntryPoints) > 0 {
		prompt.WriteString(fmt.Sprintf("- Entry Points: %s\n", strings.Join(analysis.EntryPoints, ", ")))
	}
	if len(analysis.KeyDirectories) > 0 {
		prompt.WriteString(fmt.Sprintf("- Key Directories: %s\n", strings.Join(analysis.KeyDirectories, ", ")))
	}

	prompt.WriteString("\n## Instructions\n")
	prompt.WriteString("Please implement the changes described above in the repository.\n")
	prompt.WriteString("Make sure to:\n")
	prompt.WriteString("1. Follow existing code patterns and conventions\n")
	prompt.WriteString("2. Write clean, maintainable code\n")
	prompt.WriteString("3. Add appropriate error handling\n")
	prompt.WriteString("4. Update any relevant documentation\n")

	return prompt.String()
}

func (s *ClaudeService) GenerateCode(
	request *models.DevelopmentRequest,
	project *models.Project,
	analysis *models.RepositoryAnalysis,
	repoPath string,
) (*models.ClaudeCodeResponse, error) {
	// Build the prompt
	prompt := s.BuildPrompt(request, project, analysis)

	s.logger.WithFields(logrus.Fields{
		"jira_issue_key": request.JiraIssueKey,
		"prompt_length":  len(prompt),
		"repo_path":      repoPath,
	}).Info("Calling Claude Code CLI with screen")

	// Create unique session name and log file
	sessionName := fmt.Sprintf("claude-%s", request.JiraIssueKey)
	logFile := filepath.Join(repoPath, ".claude-output.log")
	doneFile := filepath.Join(repoPath, ".claude-done")

	// Clean up any previous files
	os.Remove(logFile)
	os.Remove(doneFile)

	// Escape single quotes in prompt for bash
	escapedPrompt := strings.ReplaceAll(prompt, "'", "'\\''")

	// Create bash command that runs Claude and signals completion
	bashCmd := fmt.Sprintf(
		"cd '%s' && '%s' --add-dir '%s' --permission-mode acceptEdits '%s' > '%s' 2>&1; echo $? > '%s'",
		repoPath,
		s.claudePath,
		repoPath,
		escapedPrompt,
		logFile,
		doneFile,
	)

	s.logger.WithFields(logrus.Fields{
		"session_name": sessionName,
		"log_file":     logFile,
	}).Info("Starting screen session for Claude CLI")

	// Start screen session in detached mode with logging
	cmd := exec.Command("screen", "-dmS", sessionName, "bash", "-c", bashCmd)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to start screen session: %w", err)
	}

	s.logger.Info("Screen session started, waiting for Claude to complete...")

	// Wait for completion with timeout (10 minutes max)
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Kill the screen session on timeout
			exec.Command("screen", "-S", sessionName, "-X", "quit").Run()
			return nil, fmt.Errorf("Claude CLI timed out after 10 minutes")

		case <-ticker.C:
			// Check if done file exists
			if _, err := os.Stat(doneFile); err == nil {
				s.logger.Info("Claude CLI completed")
				goto completed
			}

			// Also check if screen session still exists
			checkCmd := exec.Command("screen", "-ls", sessionName)
			output, _ := checkCmd.Output()
			if !bytes.Contains(output, []byte(sessionName)) {
				// Session ended but no done file - might have crashed
				if _, err := os.Stat(logFile); err == nil {
					s.logger.Warn("Screen session ended without done file, checking log")
					goto completed
				}
				return nil, fmt.Errorf("screen session ended unexpectedly without output")
			}

			s.logger.Debug("Still waiting for Claude to complete...")
		}
	}

completed:
	// Read the output log
	outputBytes, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude output: %w", err)
	}
	outputStr := string(outputBytes)

	// Check exit code if available
	if exitCodeBytes, err := os.ReadFile(doneFile); err == nil {
		exitCode := strings.TrimSpace(string(exitCodeBytes))
		if exitCode != "0" {
			s.logger.WithFields(logrus.Fields{
				"exit_code": exitCode,
				"output":    outputStr,
			}).Error("Claude CLI failed with non-zero exit code")
			return nil, fmt.Errorf("Claude CLI failed with exit code %s\nOutput: %s", exitCode, outputStr)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"jira_issue_key": request.JiraIssueKey,
		"output_length":  len(outputStr),
	}).Info("Claude CLI completed successfully via screen")

	// Clean up temp files
	os.Remove(logFile)
	os.Remove(doneFile)

	// Count files that were modified
	filesChanged := s.countChangedFiles(repoPath)

	return &models.ClaudeCodeResponse{
		Success:            true,
		Message:            "Code generated successfully via Claude CLI in screen session",
		FilesChanged:       filesChanged,
		DevelopmentDetails: fmt.Sprintf("Generated code using Claude CLI.\n\nClaude Output:\n%s", outputStr),
	}, nil
}

func (s *ClaudeService) countChangedFiles(repoPath string) int {
	// Execute: git status --short
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get git status")
		return 0
	}

	// Count lines (each line is a changed file)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}
