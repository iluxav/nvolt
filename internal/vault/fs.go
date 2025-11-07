package vault

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	// File permissions
	DirPerm        = 0755 // rwxr-xr-x
	FilePerm       = 0644 // rw-r--r--
	PrivateKeyPerm = 0600 // rw-------
)

// InitializeHomeDirectory creates the ~/.nvolt directory structure
func InitializeHomeDirectory() error {
	homePaths, err := GetHomePaths()
	if err != nil {
		return err
	}

	dirs := []string{
		homePaths.Root,
		homePaths.Machines,
		homePaths.Orgs,
	}

	for _, dir := range dirs {
		if err := ensureDir(dir, DirPerm); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	return nil
}

// InitializeVaultDirectory creates the .nvolt directory structure
func InitializeVaultDirectory(vaultPath string) error {
	// Use empty project name - this initializes base structure
	paths := GetVaultPaths(vaultPath, "")

	dirs := []string{
		paths.Root,
		paths.Secrets,
		paths.WrappedKeys,
		paths.Machines,
	}

	for _, dir := range dirs {
		if err := ensureDir(dir, DirPerm); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	return nil
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(path string, perm fs.FileMode) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, perm)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	return nil
}

// WriteFileAtomic writes data to a file atomically using write-then-rename
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := ensureDir(dir, DirPerm); err != nil {
		return err
	}

	// Write to temporary file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename to final path (atomic on most systems)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up temp file on error
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// ReadFile reads a file and returns its contents
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// DeleteFile deletes a file
func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ListFiles lists all files in a directory
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// ListDirs lists all directories in a directory
func ListDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(dir, entry.Name()))
		}
	}

	return dirs, nil
}

// EnsureSecretsDir creates a directory for an environment's secrets
func EnsureSecretsDir(paths *Paths, environment string) error {
	secretsPath := paths.GetSecretsPath(environment)
	return ensureDir(secretsPath, DirPerm)
}

// ValidateVaultStructure verifies that all required vault directories exist
func ValidateVaultStructure(vaultPath string) error {
	// Use empty project name - this validates base structure
	paths := GetVaultPaths(vaultPath, "")

	requiredDirs := []string{
		paths.Root,
		paths.Secrets,
		paths.WrappedKeys,
		paths.Machines,
	}

	for _, dir := range requiredDirs {
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return fmt.Errorf("required directory missing: %s", dir)
		}
		if err != nil {
			return fmt.Errorf("failed to stat directory %s: %w", dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", dir)
		}
	}

	return nil
}

// GetDirName returns the base name of a directory path
func GetDirName(dirPath string) string {
	return filepath.Base(dirPath)
}

// GetSecretKeyFromFilename extracts the key name from a secret filename
// Example: "API_KEY.enc.json" -> "API_KEY"
func GetSecretKeyFromFilename(filename string) string {
	base := filepath.Base(filename)
	if len(base) > 9 && base[len(base)-9:] == ".enc.json" {
		return base[:len(base)-9]
	}
	return base
}
