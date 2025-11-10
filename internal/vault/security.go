package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateGitignore checks if sensitive files are properly ignored in .gitignore
func ValidateGitignore(vaultPath string) ([]string, error) {
	// Get the root directory (parent of .nvolt or the vault root in global mode)
	repoRoot := GetRepoRootFromVault(vaultPath)

	gitignorePath := filepath.Join(repoRoot, ".gitignore")

	// List of sensitive patterns that should be in .gitignore
	requiredPatterns := []string{
		"private_key.pem",
		"*.pem",
		".env",
		".env.*",
	}

	// Check if .gitignore exists
	if !FileExists(gitignorePath) {
		return requiredPatterns, fmt.Errorf(".gitignore not found at %s", gitignorePath)
	}

	// Read .gitignore
	file, err := os.Open(gitignorePath)
	if err != nil {
		return requiredPatterns, fmt.Errorf("failed to read .gitignore: %w", err)
	}
	defer file.Close()

	// Parse .gitignore patterns
	ignorePatterns := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignorePatterns[line] = true
	}

	if err := scanner.Err(); err != nil {
		return requiredPatterns, fmt.Errorf("error reading .gitignore: %w", err)
	}

	// Check for missing patterns
	var missing []string
	for _, pattern := range requiredPatterns {
		if !ignorePatterns[pattern] {
			missing = append(missing, pattern)
		}
	}

	return missing, nil
}

// EnsurePrivateKeysNotInGit checks if private keys might be tracked by Git
func EnsurePrivateKeysNotInGit(vaultPath string) error {
	repoRoot := GetRepoRootFromVault(vaultPath)

	// Check if we're in a git repository
	gitDir := filepath.Join(repoRoot, ".git")
	if !FileExists(gitDir) {
		// Not a git repo, nothing to check
		return nil
	}

	// Check for common private key locations
	homePaths, err := GetHomePaths()
	if err != nil {
		return err
	}

	privateKeyPath := homePaths.PrivateKey

	// Ensure the private key path is NOT under the repo root
	rel, err := filepath.Rel(repoRoot, privateKeyPath)
	if err == nil && !strings.HasPrefix(rel, "..") {
		// Private key is under repo root - this is dangerous!
		return fmt.Errorf("WARNING: private key at %s is inside Git repository at %s", privateKeyPath, repoRoot)
	}

	return nil
}

// CheckSensitiveFilesNotCommitted validates that sensitive files are not tracked by Git
func CheckSensitiveFilesNotCommitted(vaultPath string) ([]string, error) {
	repoRoot := GetRepoRootFromVault(vaultPath)

	// Check if we're in a git repository
	gitDir := filepath.Join(repoRoot, ".git")
	if !FileExists(gitDir) {
		// Not a git repo, nothing to check
		return nil, nil
	}

	// Note: This is a basic check. A more complete implementation would use
	// `git ls-files` to check if these patterns match tracked files
	// For now, we'll warn if .gitignore doesn't contain required patterns

	missing, err := ValidateGitignore(vaultPath)
	if err != nil || len(missing) > 0 {
		return missing, fmt.Errorf("sensitive patterns missing from .gitignore: %v", missing)
	}

	return nil, nil
}
