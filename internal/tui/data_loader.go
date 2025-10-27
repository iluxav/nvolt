package tui

import (
	"fmt"
	"iluxav/nvolt/internal/types"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// loadEnvironments fetches all environments for the current project
func (m Model) loadEnvironments() tea.Cmd {
	return func() tea.Msg {
		orgID := m.machineConfig.Config.ActiveOrgID

		// Fetch all project/environment combinations
		projectEnvs, err := m.secretsClient.GetProjectEnvironments(orgID)
		if err != nil {
			return loadEnvironmentsMsg{
				environments: nil,
				err:          fmt.Errorf("failed to fetch environments: %w", err),
			}
		}

		// Filter environments for this specific project
		envMap := make(map[string]bool)
		for _, pe := range projectEnvs {
			if pe.ProjectName == m.projectName {
				envMap[pe.Environment] = true
			}
		}

		// Convert map to slice
		environments := make([]string, 0, len(envMap))
		for env := range envMap {
			environments = append(environments, env)
		}

		// Sort environments (default first, then alphabetically)
		sortedEnvs := make([]string, 0, len(environments))
		for _, env := range environments {
			if env == "default" {
				sortedEnvs = append(sortedEnvs, "default")
				break
			}
		}

		// Simple bubble sort for remaining environments
		others := make([]string, 0)
		for _, env := range environments {
			if env != "default" {
				others = append(others, env)
			}
		}

		for i := 0; i < len(others)-1; i++ {
			for j := i + 1; j < len(others); j++ {
				if others[i] > others[j] {
					others[i], others[j] = others[j], others[i]
				}
			}
		}

		sortedEnvs = append(sortedEnvs, others...)

		// If no environments found, return default
		if len(sortedEnvs) == 0 {
			sortedEnvs = []string{"default"}
		}

		return loadEnvironmentsMsg{
			environments: sortedEnvs,
			err:          nil,
		}
	}
}

// loadData fetches variables and users for the current environment
func (m Model) loadData() tea.Cmd {
	return func() tea.Msg {
		// Get the active environment name
		envName := m.environments[m.activeEnvIndex].Name

		// Fetch variables with metadata
		varsWithMeta, err := m.secretsClient.PullSecretsWithMetadata(m.projectName, envName)
		if err != nil {
			// Check if it's a "no data" error vs actual error
			if strings.Contains(err.Error(), "no wrapped key") ||
			   strings.Contains(err.Error(), "WrappedKey is empty") ||
			   strings.Contains(err.Error(), "no secrets found") {
				// No data yet, that's OK - return empty data
				varsWithMeta = make(map[string]types.VariableWithMetadata)
			} else {
				// Real error
				return loadDataMsg{
					variables: nil,
					users:     nil,
					err:       fmt.Errorf("failed to load variables: %w", err),
				}
			}
		}

		// Fetch users with their permissions
		orgID := m.machineConfig.Config.ActiveOrgID
		users, err := m.aclService.GetOrgUsers(orgID)
		if err != nil {
			return loadDataMsg{
				variables: varsWithMeta,
				users:     nil,
				err:       fmt.Errorf("failed to load users: %w", err),
			}
		}

		// Fetch detailed permissions for each user
		usersWithPermissions := make([]*types.OrgUser, 0, len(users))
		for _, user := range users {
			if user.User == nil {
				continue
			}

			// Fetch user permissions to get project/environment level details
			userPerms, err := m.aclService.GetUserPermissions(orgID, user.User.Email)
			if err != nil {
				// If we can't fetch permissions, skip this user or use basic info
				usersWithPermissions = append(usersWithPermissions, user)
				continue
			}

			// Attach permissions to user for later processing
			user.Permissions = userPerms
			usersWithPermissions = append(usersWithPermissions, user)
		}

		return loadDataMsg{
			variables: varsWithMeta,
			users:     usersWithPermissions,
			err:       nil,
		}
	}
}

// convertToEnvVariables converts the map of variables with metadata to EnvVariable structs
func convertToEnvVariables(vars map[string]types.VariableWithMetadata) []EnvVariable {
	result := make([]EnvVariable, 0, len(vars))
	for key, varMeta := range vars {
		// Parse creation date
		createdAt, err := time.Parse(time.RFC3339, varMeta.CreatedAt)
		if err != nil {
			createdAt = time.Time{} // If parsing fails, use zero time
		}

		result = append(result, EnvVariable{
			Name:       key,
			Value:      varMeta.Value,
			IsRevealed: false,
			Created:    createdAt,
		})
	}

	// Sort by name for consistent display
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Name > result[j].Name {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// convertToUsers converts OrgUser slice to User slice for TUI
// projectName and envName are used to extract specific project/environment permissions
func convertToUsers(orgUsers []*types.OrgUser, projectName, envName string) []User {
	result := make([]User, 0, len(orgUsers))
	for _, orgUser := range orgUsers {
		if orgUser.User == nil {
			continue
		}

		var projectPerms *types.Permission
		var envPerms *types.Permission

		// Extract project and environment permissions if available
		if orgUser.Permissions != nil {
			// Find permissions for the current project
			for _, proj := range orgUser.Permissions.Projects {
				if proj.ProjectName == projectName {
					projectPerms = &proj.Permissions

					// Find permissions for the current environment
					for _, env := range proj.Environments {
						if env.Environment == envName {
							envPerms = &env.Permissions
							break
						}
					}
					break
				}
			}
		}

		result = append(result, User{
			Name:                   orgUser.User.Name,
			Email:                  orgUser.User.Email,
			OrgRole:                orgUser.Role,
			UserID:                 orgUser.UserID,
			ProjectPermissions:     projectPerms,
			EnvironmentPermissions: envPerms,
		})
	}
	return result
}
