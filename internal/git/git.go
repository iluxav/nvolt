package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Clone clones a Git repository to the specified path
func Clone(repoURL, targetPath string) error {
	cmd := exec.Command("git", "clone", repoURL, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Pull performs a git pull in the specified directory
func Pull(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Add stages files for commit in the specified directory
func Add(repoPath string, paths ...string) error {
	args := append([]string{"-C", repoPath, "add"}, paths...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Commit creates a commit in the specified directory
func Commit(repoPath, message string) error {
	cmd := exec.Command("git", "-C", repoPath, "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if there's nothing to commit
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Push pushes commits to remote in the specified directory
func Push(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "push")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// Status returns the git status in the specified directory
func Status(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// IsGitAvailable checks if git is available in PATH
func IsGitAvailable() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// ParseRepoURL parses org/repo format and returns GitHub URL
func ParseRepoURL(orgRepo string) (string, error) {
	parts := strings.Split(orgRepo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: expected 'org/repo', got '%s'", orgRepo)
	}

	org := strings.TrimSpace(parts[0])
	repo := strings.TrimSpace(parts[1])

	if org == "" || repo == "" {
		return "", fmt.Errorf("invalid repository format: org and repo cannot be empty")
	}

	return fmt.Sprintf("git@github.com:%s/%s.git", org, repo), nil
}

// GetRepoPath extracts org and repo from org/repo string
func GetRepoPath(orgRepo string) (org string, repo string, err error) {
	parts := strings.Split(orgRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: expected 'org/repo', got '%s'", orgRepo)
	}

	org = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])

	if org == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository format: org and repo cannot be empty")
	}

	return org, repo, nil
}
