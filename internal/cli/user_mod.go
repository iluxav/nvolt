package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var userModCmd = &cobra.Command{
	Use:   "mod [email]",
	Short: "Modify user permissions in an organization (admin only)",
	Long: `Modify user permissions in an organization with interactive prompts for role, project, and environment permissions.
Only users with admin role can execute this command.

Examples:
  # Modify user permissions interactively
  nvolt user mod john@example.com

  # Modify user permissions for specific organization
  nvolt user mod john@example.com -o org-id-123
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]
		orgID, _ := cmd.Flags().GetString("org")

		machineConfig := services.MachineConfigFromContext(cmd.Context())
		aclService := services.ACLServiceFromContext(cmd.Context())

		// Determine which org to use
		targetOrgID := orgID
		if targetOrgID == "" {
			// Use active org from config
			if machineConfig.Config.ActiveOrgID == "" {
				return fmt.Errorf("no active organization set. Use -o flag or run 'nvolt org set' first")
			}
			targetOrgID = machineConfig.Config.ActiveOrgID
		}

		return runUserModify(aclService, targetOrgID, email)
	},
}

func init() {
	userCmd.AddCommand(userModCmd)
	userModCmd.Flags().StringP("org", "o", "", "Organization ID (defaults to active org)")
}

func runUserModify(aclService *services.ACLService, orgID, email string) error {
	fmt.Println(titleStyle.Render("Modifying User Permissions"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Email: %s", email)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Organization ID: %s", orgID)))

	// Fetch current user permissions
	fmt.Println(infoStyle.Render("→ Fetching current permissions..."))
	currentPerms, err := aclService.GetUserPermissions(orgID, email)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to fetch user permissions: %v", err)))
		return err
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Current Role: %s", currentPerms.Role)))

	// Step 1: Show interactive selection for org roles (admin or dev)
	newRole, err := promptForRole(currentPerms.Role)
	if err != nil {
		return err
	}

	// Step 2: Show interactive single selection for all the org projects + option "All"
	selectedProject, err := promptForProjectSelection(currentPerms.AllProjects)
	if err != nil {
		return err
	}

	var projectPermissions *types.Permission
	var envPermissions *types.Permission
	var selectedEnvironment string

	// Handle "All" projects selection
	if selectedProject == "All" {
		fmt.Println(warnStyle.Render("⚠ WARNING! This will grant permissions for ALL projects and ALL environments in this org"))

		// Prompt for project permissions
		projectPermissions, err = promptForPermissionsWithDefaults("Project Permissions (applies to ALL projects)", nil)
		if err != nil {
			return err
		}

		// Automatically apply same permissions to all environments
		envPermissions = projectPermissions

		// Apply to all projects - we'll need to iterate and update each one
		// For simplicity, we'll send one request per project
		for _, proj := range currentPerms.AllProjects {
			req := &types.ModifyUserPermissionsRequest{
				Email:                  email,
				Role:                   newRole,
				ProjectName:            proj,
				ProjectPermissions:     projectPermissions,
				EnvironmentPermissions: envPermissions,
			}

			fmt.Println(infoStyle.Render(fmt.Sprintf("→ Updating permissions for project: %s", proj)))
			_, err := aclService.ModifyUserPermissions(orgID, req)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to modify permissions for project %s: %v", proj, err)))
				return err
			}
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Successfully updated permissions for user %s across all projects", email)))
		return nil
	}

	// Step 3: Get current project permissions if user has them
	var currentProjectPerms *types.Permission
	for _, proj := range currentPerms.Projects {
		if proj.ProjectName == selectedProject {
			currentProjectPerms = &proj.Permissions
			break
		}
	}

	// Prompt for project permissions
	projectPermissions, err = promptForPermissionsWithDefaults(fmt.Sprintf("Project Permissions for '%s'", selectedProject), currentProjectPerms)
	if err != nil {
		return err
	}

	// Step 4: Show interactive single selection of Environments for the selected project + option "All"
	var availableEnvs []string
	for _, proj := range currentPerms.Projects {
		if proj.ProjectName == selectedProject {
			// Use AllEnvironments instead of just environments user has access to
			availableEnvs = proj.AllEnvironments
			break
		}
	}

	selectedEnvironment, err = promptForEnvironmentSelection(availableEnvs)
	if err != nil {
		return err
	}

	// Handle "All" environments selection
	if selectedEnvironment == "All" {
		// Prompt for environment permissions
		envPermissions, err = promptForPermissionsWithDefaults("Environment Permissions (applies to ALL environments)", nil)
		if err != nil {
			return err
		}

		// Apply to all environments in this project
		if len(availableEnvs) > 0 {
			for _, env := range availableEnvs {
				req := &types.ModifyUserPermissionsRequest{
					Email:                  email,
					Role:                   newRole,
					ProjectName:            selectedProject,
					Environment:            env,
					ProjectPermissions:     projectPermissions,
					EnvironmentPermissions: envPermissions,
				}

				fmt.Println(infoStyle.Render(fmt.Sprintf("→ Updating permissions for environment: %s", env)))
				_, err := aclService.ModifyUserPermissions(orgID, req)
				if err != nil {
					fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to modify permissions for environment %s: %v", env, err)))
					return err
				}
			}
		} else {
			// No environments yet, just update project permissions
			req := &types.ModifyUserPermissionsRequest{
				Email:              email,
				Role:               newRole,
				ProjectName:        selectedProject,
				ProjectPermissions: projectPermissions,
			}

			_, err := aclService.ModifyUserPermissions(orgID, req)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to modify permissions: %v", err)))
				return err
			}
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Successfully updated permissions for user %s", email)))
		return nil
	}

	// Get current environment permissions if user has them
	var currentEnvPerms *types.Permission
	for _, proj := range currentPerms.Projects {
		if proj.ProjectName == selectedProject {
			for _, env := range proj.Environments {
				if env.Environment == selectedEnvironment {
					currentEnvPerms = &env.Permissions
					break
				}
			}
			break
		}
	}

	// Prompt for environment permissions
	envPermissions, err = promptForPermissionsWithDefaults(fmt.Sprintf("Environment Permissions for '%s'", selectedEnvironment), currentEnvPerms)
	if err != nil {
		return err
	}

	// Build request
	req := &types.ModifyUserPermissionsRequest{
		Email:                  email,
		Role:                   newRole,
		ProjectName:            selectedProject,
		Environment:            selectedEnvironment,
		ProjectPermissions:     projectPermissions,
		EnvironmentPermissions: envPermissions,
	}

	// Call API
	fmt.Println(infoStyle.Render("→ Sending request to server..."))
	response, err := aclService.ModifyUserPermissions(orgID, req)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to modify user permissions: %v", err)))
		return err
	}

	if response.Success {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s", response.Message)))
		if response.User != nil {
			fmt.Println(infoStyle.Render(fmt.Sprintf("  User: %s (%s)", response.User.Name, response.User.Email)))
		}
	} else {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %s", response.Message)))
	}

	return nil
}

// promptForRole prompts the user to select a role with the current role preselected
func promptForRole(currentRole string) (string, error) {
	roles := []string{"admin", "dev"}

	prompt := &survey.Select{
		Message: "Select user role:",
		Options: roles,
		Default: currentRole,
	}

	var selectedRole string
	err := survey.AskOne(prompt, &selectedRole)
	if err != nil {
		return "", err
	}

	return selectedRole, nil
}

// promptForProjectSelection prompts the user to select a project or "All"
func promptForProjectSelection(allProjects []string) (string, error) {
	// Add "All" option at the beginning
	options := append([]string{"All (WARNING! Grant permissions for ALL projects and all project Environments in this org)"}, allProjects...)

	prompt := &survey.Select{
		Message: "Select project:",
		Options: options,
	}

	var selected string
	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return "", err
	}

	// Return "All" if the warning option was selected
	if selected == options[0] {
		return "All", nil
	}

	return selected, nil
}

// promptForEnvironmentSelection prompts the user to select an environment or "All"
func promptForEnvironmentSelection(availableEnvs []string) (string, error) {
	// Add "All" option at the beginning
	options := append([]string{"All (Grant permissions for ALL environments in this project)"}, availableEnvs...)

	prompt := &survey.Select{
		Message: "Select environment:",
		Options: options,
	}

	var selected string
	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return "", err
	}

	// Return "All" if the warning option was selected
	if selected == options[0] {
		return "All", nil
	}

	return selected, nil
}

// promptForPermissionsWithDefaults shows an interactive multi-select for permissions with defaults
func promptForPermissionsWithDefaults(title string, currentPerms *types.Permission) (*types.Permission, error) {
	options := []string{"read", "write", "delete"}
	var defaultSelected []string

	// Set defaults based on current permissions
	if currentPerms != nil {
		if currentPerms.Read {
			defaultSelected = append(defaultSelected, "read")
		}
		if currentPerms.Write {
			defaultSelected = append(defaultSelected, "write")
		}
		if currentPerms.Delete {
			defaultSelected = append(defaultSelected, "delete")
		}
	} else {
		// Default to read-only if no current permissions
		defaultSelected = []string{"read"}
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: title + " (use space to select, enter to confirm):",
		Options: options,
		Default: defaultSelected,
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	// Convert selected options to Permission struct
	perm := &types.Permission{
		Read:   false,
		Write:  false,
		Delete: false,
	}

	for _, sel := range selected {
		switch sel {
		case "read":
			perm.Read = true
		case "write":
			perm.Write = true
		case "delete":
			perm.Delete = true
		}
	}

	return perm, nil
}
