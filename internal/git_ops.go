package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitOperations handles all git-related functionality
type GitOperations struct {
	repo     *git.Repository
	repoPath string
}

// NewGitOperations creates a new GitOperations instance
func NewGitOperations(repoPath string) (*GitOperations, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid git repository: %w", err)
	}

	// Check if repo has uncommitted changes
	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	if !status.IsClean() {
		return nil, fmt.Errorf("repository has uncommitted changes: %s", repoPath)
	}

	return &GitOperations{
		repo:     repo,
		repoPath: repoPath,
	}, nil
}

// GetModifiableFiles returns a list of files matching the given patterns
func (g *GitOperations) GetModifiableFiles(patterns []string) ([]string, error) {
	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(g.repoPath, "**", pattern))
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", pattern, err)
		}
		for _, match := range matches {
			// Skip git directory and empty directories
			if !isGitPath(match) && isRegularFile(match) {
				relPath, err := filepath.Rel(g.repoPath, match)
				if err == nil {
					files = append(files, relPath)
				}
			}
		}
	}
	return files, nil
}

// CreateCommit creates a new commit with the given message and files
func (g *GitOperations) CreateCommit(message string, filesToModify []string, timestamp *time.Time) error {
	w, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Stage modified files
	for _, file := range filesToModify {
		_, err := w.Add(file)
		if err != nil {
			return fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}

	// Create commit options
	opts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Dev Metrics",
			Email: "dev@metrics.local",
			When:  time.Now(),
		},
	}

	// Set custom timestamp if provided
	if timestamp != nil {
		opts.Author.When = *timestamp
		opts.Committer = &object.Signature{
			Name:  "Dev Metrics",
			Email: "dev@metrics.local",
			When:  *timestamp,
		}
	}

	// Create commit
	_, err = w.Commit(message, opts)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// VerifyRepoAccess checks if we have proper access to the repository
func (g *GitOperations) VerifyRepoAccess() error {
	// Check if we can read the repo
	_, err := g.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to read repository head: %w", err)
	}

	// Try to create and delete a test file
	testFile := filepath.Join(g.repoPath, ".git_ops_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}

	return nil
}

// ModifyFile modifies a file with new content
func (g *GitOperations) ModifyFile(filePath string, newContent string) error {
	fullPath := filepath.Join(g.repoPath, filePath)
	err := os.WriteFile(fullPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to modify file %s: %w", filePath, err)
	}
	return nil
}

// ReadFile reads the contents of a file
func (g *GitOperations) ReadFile(filePath string) (string, error) {
	fullPath := filepath.Join(g.repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// Helper functions

func isGitPath(path string) bool {
	return filepath.Base(path) == ".git" || filepath.Dir(path) == ".git"
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}
