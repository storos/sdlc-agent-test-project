package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
)

type GitService struct {
	logger *logrus.Logger
}

func NewGitService(logger *logrus.Logger) *GitService {
	return &GitService{
		logger: logger,
	}
}

type GitWorkspace struct {
	Path       string
	Repository *git.Repository
	BranchName string
}

func (s *GitService) CloneRepository(repoURL, accessToken, jiraIssueKey string) (*GitWorkspace, error) {
	// Create temporary directory
	tempDir := filepath.Join("/tmp", fmt.Sprintf("sdlc-%s", jiraIssueKey))
	repoPath := filepath.Join(tempDir, "repo")

	// Clean up if directory already exists
	if _, err := os.Stat(tempDir); err == nil {
		s.logger.Warnf("Directory %s already exists, removing", tempDir)
		os.RemoveAll(tempDir)
	}

	// Create directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"repository_url": repoURL,
		"local_path":     repoPath,
	}).Info("Cloning repository")

	// Clone repository with authentication
	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL: repoURL,
		Auth: &http.BasicAuth{
			Username: "git", // Can be anything for PAT
			Password: accessToken,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	s.logger.Info("Repository cloned successfully")

	return &GitWorkspace{
		Path:       repoPath,
		Repository: repo,
	}, nil
}

func (s *GitService) CreateAndCheckoutBranch(workspace *GitWorkspace, jiraIssueKey string) error {
	branchName := fmt.Sprintf("feature/%s", jiraIssueKey)
	workspace.BranchName = branchName

	s.logger.WithFields(logrus.Fields{
		"branch_name": branchName,
	}).Info("Creating and checking out branch")

	// Get HEAD reference
	headRef, err := workspace.Repository.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Create new branch reference
	refName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(refName, headRef.Hash())
	err = workspace.Repository.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Checkout the branch
	worktree, err := workspace.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: refName,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	s.logger.Infof("Checked out branch: %s", branchName)
	return nil
}

func (s *GitService) CommitChanges(workspace *GitWorkspace, jiraIssueKey, summary string) error {
	worktree, err := workspace.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	s.logger.Info("Adding changes to git")

	// Add all changes
	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Check if there are changes to commit
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		s.logger.Warn("No changes to commit")
		return nil
	}

	// Create commit message
	commitMessage := fmt.Sprintf("[%s] %s\n\nAutomated commit by SDLC AI Agent", jiraIssueKey, summary)

	s.logger.WithFields(logrus.Fields{
		"message": commitMessage,
	}).Info("Creating commit")

	// Commit changes
	_, err = worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "SDLC AI Agent",
			Email: "sdlc-agent@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	s.logger.Info("Changes committed successfully")
	return nil
}

func (s *GitService) PushBranch(workspace *GitWorkspace, accessToken string) error {
	s.logger.WithFields(logrus.Fields{
		"branch": workspace.BranchName,
	}).Info("Pushing branch to remote")

	// Push to remote
	err := workspace.Repository.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", workspace.BranchName, workspace.BranchName)),
		},
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	s.logger.Info("Branch pushed successfully")
	return nil
}

func (s *GitService) Cleanup(workspace *GitWorkspace) error {
	if workspace == nil || workspace.Path == "" {
		return nil
	}

	// Get parent directory (/tmp/sdlc-{jira_issue_key})
	tempDir := filepath.Dir(workspace.Path)

	s.logger.WithFields(logrus.Fields{
		"path": tempDir,
	}).Info("Cleaning up temporary directory")

	err := os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}
