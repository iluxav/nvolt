package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/iluxav/nvolt/internal/errors"
)

// ValidateNotEmpty checks that a string is not empty
func ValidateNotEmpty(field, value string) error {
	if value == "" {
		return errors.NewInvalidInput(field, value, "value cannot be empty")
	}
	return nil
}

// ValidateEnvironmentName validates an environment name
func ValidateEnvironmentName(env string) error {
	if env == "" {
		return errors.NewInvalidInput("environment", env, "environment name cannot be empty")
	}

	// Environment names should be alphanumeric with dashes/underscores
	match := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(env)
	if !match {
		return errors.NewInvalidInput("environment", env, "must contain only letters, numbers, dashes, and underscores")
	}

	return nil
}

// ValidateProjectName validates a project name
func ValidateProjectName(project string) error {
	if project == "" {
		return errors.NewInvalidInput("project", project, "project name cannot be empty")
	}

	// Project names can contain alphanumeric, dashes, underscores, slashes, dots
	match := regexp.MustCompile(`^[a-zA-Z0-9_.\-/]+$`).MatchString(project)
	if !match {
		return errors.NewInvalidInput("project", project, "must contain only letters, numbers, dashes, underscores, slashes, and dots")
	}

	return nil
}

// ValidateMachineID validates a machine ID
func ValidateMachineID(machineID string) error {
	if machineID == "" {
		return errors.NewInvalidInput("machine ID", machineID, "machine ID cannot be empty")
	}

	// Machine IDs should start with 'm-' and contain alphanumeric/dashes/underscores
	match := regexp.MustCompile(`^m-[a-zA-Z0-9_-]+$`).MatchString(machineID)
	if !match {
		return errors.NewInvalidInput("machine ID", machineID, "must start with 'm-' and contain only letters, numbers, dashes, and underscores")
	}

	return nil
}

// ValidateFilePath checks if a file exists and is readable
func ValidateFilePath(path string) error {
	if path == "" {
		return errors.NewInvalidInput("file path", path, "path cannot be empty")
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.NewFileNotFound(path)
	}
	if err != nil {
		return errors.Wrap(err, errors.ErrFileNotFound, "failed to access file")
	}

	if info.IsDir() {
		return errors.NewInvalidInput("file path", path, "path is a directory, not a file")
	}

	return nil
}

// ValidateDirectoryPath checks if a directory exists and is accessible
func ValidateDirectoryPath(path string) error {
	if path == "" {
		return errors.NewInvalidInput("directory path", path, "path cannot be empty")
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.NewFileNotFound(path)
	}
	if err != nil {
		return errors.Wrap(err, errors.ErrFileNotFound, "failed to access directory")
	}

	if !info.IsDir() {
		return errors.NewInvalidInput("directory path", path, "path is a file, not a directory")
	}

	return nil
}

// ValidateKeyValuePair validates a KEY=VALUE pair
func ValidateKeyValuePair(pair string) error {
	if pair == "" {
		return errors.NewInvalidInput("key-value pair", pair, "pair cannot be empty")
	}

	// Must contain at least one '='
	if !regexp.MustCompile(`^[A-Z_][A-Z0-9_]*=.*$`).MatchString(pair) {
		return errors.NewInvalidInput("key-value pair", pair, "must be in format KEY=value (key must be uppercase with underscores)")
	}

	return nil
}

// ValidateRepoURL validates a Git repository URL format
func ValidateRepoURL(url string) error {
	if url == "" {
		return errors.NewInvalidInput("repository URL", url, "URL cannot be empty")
	}

	// Should be in format: org/repo or github.com/org/repo
	match := regexp.MustCompile(`^([a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+|github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+)$`).MatchString(url)
	if !match {
		return errors.NewInvalidInput("repository URL", url, "must be in format 'org/repo' or 'github.com/org/repo'")
	}

	return nil
}

// ValidateFilePermissions checks if a file has the expected permissions
func ValidateFilePermissions(path string, expectedPerm os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrap(err, errors.ErrFileNotFound, "failed to check file permissions")
	}

	actualPerm := info.Mode().Perm()
	if actualPerm != expectedPerm {
		return errors.NewInvalidInput("file permissions", fmt.Sprintf("%o", actualPerm),
			fmt.Sprintf("expected %o, file may be readable by others", expectedPerm))
	}

	return nil
}

// ValidateEnvFile validates that an env file exists and has valid format
func ValidateEnvFile(path string) error {
	if err := ValidateFilePath(path); err != nil {
		return err
	}

	// Check file extension
	ext := filepath.Ext(path)
	if ext != ".env" && ext != "" {
		return errors.NewInvalidInput("env file", path, "should have .env extension or no extension")
	}

	// TODO: Could add more validation of file contents here

	return nil
}
