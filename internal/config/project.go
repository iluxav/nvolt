package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ProjectInfo contains detected project information
type ProjectInfo struct {
	Name   string
	Source string // Where the name was detected from
}

// PackageJSON represents a Node.js package.json file
type PackageJSON struct {
	Name string `json:"name"`
}

// GoMod represents a Go go.mod file structure
type GoMod struct {
	Module string
}

// CargoToml represents a Rust Cargo.toml file
type CargoToml struct {
	Package struct {
		Name string `toml:"name"`
	} `toml:"package"`
}

// PyProjectToml represents a Python pyproject.toml file
type PyProjectToml struct {
	Project struct {
		Name string `toml:"name"`
	} `toml:"project"`
}

// DetectProjectName attempts to detect the project name from various sources
func DetectProjectName(basePath string) (*ProjectInfo, error) {
	// Priority chain:
	// 1. package.json (Node.js)
	// 2. go.mod (Go)
	// 3. Cargo.toml (Rust)
	// 4. pyproject.toml (Python)
	// 5. Current directory name

	// Try package.json
	if name, err := detectFromPackageJSON(basePath); err == nil {
		return &ProjectInfo{Name: name, Source: "package.json"}, nil
	}

	// Try go.mod
	if name, err := detectFromGoMod(basePath); err == nil {
		return &ProjectInfo{Name: name, Source: "go.mod"}, nil
	}

	// Try Cargo.toml
	if name, err := detectFromCargoToml(basePath); err == nil {
		return &ProjectInfo{Name: name, Source: "Cargo.toml"}, nil
	}

	// Try pyproject.toml
	if name, err := detectFromPyProjectToml(basePath); err == nil {
		return &ProjectInfo{Name: name, Source: "pyproject.toml"}, nil
	}

	// Fallback to directory name
	name := detectFromDirectoryName(basePath)
	return &ProjectInfo{Name: name, Source: "directory"}, nil
}

// detectFromPackageJSON detects project name from package.json
func detectFromPackageJSON(basePath string) (string, error) {
	path := filepath.Join(basePath, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package.json: %w", err)
	}

	if pkg.Name == "" {
		return "", fmt.Errorf("package.json has no name field")
	}

	// Remove npm scope if present (e.g., "@org/package" -> "package")
	name := pkg.Name
	if strings.HasPrefix(name, "@") {
		parts := strings.Split(name, "/")
		if len(parts) == 2 {
			name = parts[1]
		}
	}

	return sanitizeProjectName(name), nil
}

// detectFromGoMod detects project name from go.mod
func detectFromGoMod(basePath string) (string, error) {
	path := filepath.Join(basePath, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Simple parser for go.mod module line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimPrefix(line, "module ")
			module = strings.TrimSpace(module)

			// Use full module path for uniqueness
			// e.g., "github.com/iluxav/nvolt" -> "github.com-iluxav-nvolt"
			return sanitizeProjectName(module), nil
		}
	}

	return "", fmt.Errorf("no module declaration found in go.mod")
}

// detectFromCargoToml detects project name from Cargo.toml
func detectFromCargoToml(basePath string) (string, error) {
	path := filepath.Join(basePath, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var cargo CargoToml
	if err := toml.Unmarshal(data, &cargo); err != nil {
		return "", fmt.Errorf("failed to parse Cargo.toml: %w", err)
	}

	if cargo.Package.Name == "" {
		return "", fmt.Errorf("Cargo.toml has no package.name field")
	}

	return sanitizeProjectName(cargo.Package.Name), nil
}

// detectFromPyProjectToml detects project name from pyproject.toml
func detectFromPyProjectToml(basePath string) (string, error) {
	path := filepath.Join(basePath, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var pyproject PyProjectToml
	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return "", fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}

	if pyproject.Project.Name == "" {
		return "", fmt.Errorf("pyproject.toml has no project.name field")
	}

	return sanitizeProjectName(pyproject.Project.Name), nil
}

// detectFromDirectoryName uses the current directory name as project name
func detectFromDirectoryName(basePath string) string {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		abs = basePath
	}

	name := filepath.Base(abs)
	return sanitizeProjectName(name)
}

// sanitizeProjectName cleans up a project name
func sanitizeProjectName(name string) string {
	// Remove common suffixes/prefixes
	name = strings.TrimSpace(name)

	// Replace spaces and special characters with hyphens
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)

	// Remove consecutive hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	// Convert to lowercase
	name = strings.ToLower(name)

	return name
}

// GetProjectName gets the project name, using override if provided
func GetProjectName(basePath, override string) (string, string, error) {
	if override != "" {
		return sanitizeProjectName(override), "override", nil
	}

	info, err := DetectProjectName(basePath)
	if err != nil {
		return "", "", err
	}

	return info.Name, info.Source, nil
}
