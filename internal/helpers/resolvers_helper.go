package helpers

import (
	"encoding/json"
	"fmt"
	"iluxav/nvolt/internal/types"

	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func ReadEnvFile(dotEnvFileName string) (map[string]string, error) {
	envVars, err := godotenv.Read(dotEnvFileName)
	if err != nil {
		nvVars, err := godotenv.Read(".env")
		if err != nil {
			return nil, fmt.Errorf("failed to read env file: %w", err)
		}
		return nvVars, nil
	}
	return envVars, nil
}

func TryGetDotEnvFiles() []string {
	dotEnvFiles := []string{}
	// Get all files in the current directory that match the .env.* pattern
	// Set directory of the execution of the command
	dir, err := os.Getwd()
	if err != nil {
		return dotEnvFiles
	}
	files, err := filepath.Glob(filepath.Join(dir, ".env*"))
	if err != nil {
		return dotEnvFiles
	}
	dotEnvFiles = append(dotEnvFiles, files...)
	// Extract filename from the path
	for i, file := range dotEnvFiles {
		dotEnvFiles[i] = filepath.Base(file)
	}
	return dotEnvFiles
}

func TryGetProjectName() (*types.ProjectResolver, error) {
	projectResolver := types.ProjectResolver{}

	if _, err := os.Stat("go.mod"); err == nil {
		name := GetGoModName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "go"
		return &projectResolver, nil
	} else if _, err := os.Stat("package.json"); err == nil {
		name := GetPackageJsonName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "node"
		return &projectResolver, nil
	} else if _, err := os.Stat("pom.xml"); err == nil {
		name := GetMavenProjectName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "maven"
		return &projectResolver, nil
	} else if _, err := os.Stat("pyproject.toml"); err == nil {
		name := GetPythonProjectName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "python"
		return &projectResolver, nil
	} else if _, err := os.Stat("setup.py"); err == nil {
		name := GetPythonSetupName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "python"
		return &projectResolver, nil
	} else if _, err := os.Stat("Gemfile"); err == nil {
		name := GetRubyProjectName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "ruby"
		return &projectResolver, nil
	} else if _, err := os.Stat("Cargo.toml"); err == nil {
		name := GetRustProjectName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "rust"
		return &projectResolver, nil
	} else if _, err := os.Stat("composer.json"); err == nil {
		name := GetComposerProjectName()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "php"
		return &projectResolver, nil
	} else if _, err := os.Stat("requirements.txt"); err == nil {
		name := DetectProject()
		projectResolver.ProjectName = name
		projectResolver.ProjectType = "python"
		return &projectResolver, nil
	}
	projectResolver.ProjectName = DetectProject()
	projectResolver.ProjectType = "unknown"
	return &projectResolver, fmt.Errorf("failed to detect project name and type")
}

func GetGoModName() string {
	goMod, err := os.ReadFile("go.mod")
	if err != nil {
		return DetectProject()
	}
	lines := strings.Split(string(goMod), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return DetectProject()
}

func GetPackageJsonName() string {
	packageJson, err := os.ReadFile("package.json")
	if err != nil {
		return DetectProject()
	}
	var packageJsonData map[string]interface{}
	json.Unmarshal(packageJson, &packageJsonData)
	return packageJsonData["name"].(string)
}

func GetEnvFileName(environment string) string {
	switch environment {
	case "production", "prod":
		return ".env.production"
	case "staging":
		return ".env.staging"
	case "development":
		return ".env.development"
	}
	return ".env"
}

func DetectProject() string {
	// In real implementation, would detect from git or config
	dir, _ := filepath.Abs(".")
	return strings.TrimPrefix(filepath.Base(dir), "/")
}

// WriteEnvFile writes environment variables to a .env file
func WriteEnvFile(filename string, vars map[string]string) error {
	// Build content
	var content strings.Builder
	for key, value := range vars {
		// Escape quotes in value
		escapedValue := strings.ReplaceAll(value, `"`, `\"`)

		// Check if value contains special characters that need quoting
		needsQuotes := strings.ContainsAny(value, " \t\n\r#")

		if needsQuotes {
			content.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, escapedValue))
		} else {
			content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}

	// Write to file
	if err := os.WriteFile(filename, []byte(content.String()), 0600); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}

// GetMavenProjectName extracts project name from pom.xml
func GetMavenProjectName() string {
	pomXml, err := os.ReadFile("pom.xml")
	if err != nil {
		return DetectProject()
	}

	// Try to extract artifactId (preferred) or name
	artifactIdRegex := regexp.MustCompile(`<artifactId>(.*?)</artifactId>`)
	matches := artifactIdRegex.FindStringSubmatch(string(pomXml))
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}

	// Fallback to name tag
	nameRegex := regexp.MustCompile(`<name>(.*?)</name>`)
	matches = nameRegex.FindStringSubmatch(string(pomXml))
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}

	return DetectProject()
}

// GetPythonProjectName extracts project name from pyproject.toml
func GetPythonProjectName() string {
	pyprojectToml, err := os.ReadFile("pyproject.toml")
	if err != nil {
		return DetectProject()
	}

	// Simple regex to extract name from [project] or [tool.poetry] section
	nameRegex := regexp.MustCompile(`(?m)^name\s*=\s*["']([^"']+)["']`)
	matches := nameRegex.FindStringSubmatch(string(pyprojectToml))
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}

	return DetectProject()
}

// GetPythonSetupName extracts project name from setup.py
func GetPythonSetupName() string {
	setupPy, err := os.ReadFile("setup.py")
	if err != nil {
		return DetectProject()
	}

	// Try to extract name from setup() call
	nameRegex := regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`)
	matches := nameRegex.FindStringSubmatch(string(setupPy))
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}

	return DetectProject()
}

// GetRubyProjectName extracts project name from Gemfile or .gemspec
func GetRubyProjectName() string {
	// Check for .gemspec files first
	files, err := filepath.Glob("*.gemspec")
	if err == nil && len(files) > 0 {
		gemspecContent, err := os.ReadFile(files[0])
		if err == nil {
			nameRegex := regexp.MustCompile(`\.name\s*=\s*["']([^"']+)["']`)
			matches := nameRegex.FindStringSubmatch(string(gemspecContent))
			if len(matches) > 1 && matches[1] != "" {
				return strings.TrimSpace(matches[1])
			}
		}
	}

	// Fallback to directory name as Gemfile doesn't typically contain project name
	return DetectProject()
}

// GetRustProjectName extracts project name from Cargo.toml
func GetRustProjectName() string {
	cargoToml, err := os.ReadFile("Cargo.toml")
	if err != nil {
		return DetectProject()
	}

	// Extract name from [package] section
	nameRegex := regexp.MustCompile(`(?m)^\s*name\s*=\s*["']([^"']+)["']`)
	matches := nameRegex.FindStringSubmatch(string(cargoToml))
	if len(matches) > 1 && matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}

	return DetectProject()
}

// GetComposerProjectName extracts project name from composer.json (PHP)
func GetComposerProjectName() string {
	composerJson, err := os.ReadFile("composer.json")
	if err != nil {
		return DetectProject()
	}

	var composerData map[string]interface{}
	if err := json.Unmarshal(composerJson, &composerData); err != nil {
		return DetectProject()
	}

	// Try to get name from composer.json
	if name, ok := composerData["name"].(string); ok && name != "" {
		// Composer names are typically vendor/package format, extract package name
		parts := strings.Split(name, "/")
		if len(parts) > 1 {
			return parts[1]
		}
		return name
	}

	return DetectProject()
}
