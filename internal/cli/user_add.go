package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var userAddCmd = &cobra.Command{
	Use:   "add [email]",
	Short: "Add a user to an organization (admin only)",
	Long: `Add a user to an organization with optional project and environment permissions.
Only users with admin role can execute this command.

Examples:
  # Add user with default permissions
  nvolt user add john@example.com

  # Add user with project permissions
  nvolt user add john@example.com -p my-project -pp read=true,write=true,delete=false

  # Add user with project and environment permissions
  nvolt user add john@example.com -p my-project -e production -pp read=true,write=true -pe read=true,write=false

  # Add user to a specific organization
  nvolt user add john@example.com -o org-id-123
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")
		orgID, _ := cmd.Flags().GetString("org")
		projectPerms, _ := cmd.Flags().GetString("project-permissions")
		envPerms, _ := cmd.Flags().GetString("environment-permissions")

		machineConfig := cmd.Context().Value("machine_config").(*services.MachineConfig)
		aclService := cmd.Context().Value("acl_service").(*services.ACLService)

		// Determine which org to use
		targetOrgID := orgID
		if targetOrgID == "" {
			// Use active org from config
			if machineConfig.Config.ActiveOrgID == "" {
				return fmt.Errorf("no active organization set. Use -o flag or run 'nvolt org set' first")
			}
			targetOrgID = machineConfig.Config.ActiveOrgID
		}

		// If project not provided via flag, prompt for it
		if project == "" {
			projectInput, err := promptForProject(machineConfig.Project)
			if err != nil {
				return err
			}
			project = projectInput
		}

		// Override project in machine config
		if project != "" {
			machineConfig.Project = project
		}

		// If environment not provided via flag and project is set, prompt for it
		if environment == "" && project != "" {
			envInput, err := promptForEnvironment(machineConfig.Environment)
			if err != nil {
				return err
			}
			environment = envInput
		}

		// Override environment in machine config
		if environment != "" {
			machineConfig.Environment = environment
		}

		return runUserAdd(aclService, targetOrgID, email, machineConfig.Project, machineConfig.Environment, projectPerms, envPerms)
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users in an organization",
	Long:  `Manage users in an organization. Admin only commands.`,
}

func init() {
	userCmd.AddCommand(userAddCmd)

	userAddCmd.Flags().StringP("org", "o", "", "Organization ID (defaults to active org)")
	userAddCmd.Flags().StringP("project", "p", "", "Project name")
	userAddCmd.Flags().StringP("environment", "e", "", "Environment name")
	userAddCmd.Flags().StringP("project-permissions", "", "", "Project permissions (e.g., read=true,write=false,delete=false)")
	userAddCmd.Flags().StringP("environment-permissions", "", "", "Environment permissions (e.g., read=true,write=false,delete=false)")

	rootCmd.AddCommand(userCmd)
}

func runUserAdd(aclService *services.ACLService, orgID, email, project, environment, projectPermsStr, envPermsStr string) error {
	fmt.Println(titleStyle.Render("Adding User to Organization"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Email: %s", email)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Organization ID: %s", orgID)))

	// Parse or prompt for project permissions
	var projectPermissions *types.Permission
	if project != "" {
		if projectPermsStr != "" {
			// Parse from flag
			perms, err := parsePermissions(projectPermsStr)
			if err != nil {
				return fmt.Errorf("invalid project permissions: %w", err)
			}
			projectPermissions = perms
		} else {
			// Interactive selector
			fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ Project: %s", project)))
			perms, err := promptForPermissions("Project Permissions")
			if err != nil {
				return err
			}
			projectPermissions = perms
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("  Permissions: read=%t, write=%t, delete=%t",
			projectPermissions.Read, projectPermissions.Write, projectPermissions.Delete)))
	}

	// Parse or prompt for environment permissions
	var envPermissions *types.Permission
	if environment != "" && project != "" {
		if envPermsStr != "" {
			// Parse from flag
			perms, err := parsePermissions(envPermsStr)
			if err != nil {
				return fmt.Errorf("invalid environment permissions: %w", err)
			}
			envPermissions = perms
		} else {
			// Interactive selector
			fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ Environment: %s", environment)))
			perms, err := promptForPermissions("Environment Permissions")
			if err != nil {
				return err
			}
			envPermissions = perms
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("  Permissions: read=%t, write=%t, delete=%t",
			envPermissions.Read, envPermissions.Write, envPermissions.Delete)))
	}

	// Validate: if environment is provided, project must be specified
	if environment != "" && project == "" {
		return fmt.Errorf("project name (-p) is required when environment (-e) is specified")
	}

	// Build request
	req := &types.AddUserToOrgRequest{
		Email:                  email,
		ProjectName:            project,
		Environment:            environment,
		ProjectPermissions:     projectPermissions,
		EnvironmentPermissions: envPermissions,
	}

	// Call API
	fmt.Println(infoStyle.Render("→ Sending request to server..."))
	response, err := aclService.AddUserToOrg(orgID, req)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to add user: %v", err)))
		return err
	}

	if response.Success {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s", response.Message)))
		if response.User != nil {
			fmt.Println(infoStyle.Render(fmt.Sprintf("  User ID: %s", response.User.ID)))
			fmt.Println(infoStyle.Render(fmt.Sprintf("  Name: %s", response.User.Name)))
		}
	} else {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %s", response.Message)))
	}

	return nil
}

// promptForProject prompts the user to enter a project name (optional)
func promptForProject(defaultProject string) (string, error) {
	var project string
	prompt := &survey.Input{
		Message: "Project name (leave empty to skip):",
		Default: defaultProject,
	}

	err := survey.AskOne(prompt, &project)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(project), nil
}

// promptForEnvironment prompts the user to enter an environment name (optional)
func promptForEnvironment(defaultEnv string) (string, error) {
	var environment string
	prompt := &survey.Input{
		Message: "Environment name (leave empty to skip):",
		Default: defaultEnv,
	}

	err := survey.AskOne(prompt, &environment)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(environment), nil
}

// promptForPermissions shows an interactive multi-select for permissions
func promptForPermissions(title string) (*types.Permission, error) {
	options := []string{"read", "write", "delete"}
	var selected []string

	prompt := &survey.MultiSelect{
		Message: title + " (use space to select, enter to confirm):",
		Options: options,
		Default: []string{"read"}, // Default to read-only
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

// parsePermissions parses permission string like "read=true,write=false,delete=false"
// Returns Permission struct. If an operator is not mentioned, defaults to false
func parsePermissions(permStr string) (*types.Permission, error) {
	perm := &types.Permission{
		Read:   false,
		Write:  false,
		Delete: false,
	}

	if permStr == "" {
		return perm, nil
	}

	pairs := strings.Split(permStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid permission format: %s (expected key=value)", pair)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.ToLower(strings.TrimSpace(parts[1]))

		boolValue := value == "true" || value == "1" || value == "yes"

		switch key {
		case "read":
			perm.Read = boolValue
		case "write":
			perm.Write = boolValue
		case "delete":
			perm.Delete = boolValue
		default:
			return nil, fmt.Errorf("unknown permission key: %s (allowed: read, write, delete)", key)
		}
	}

	return perm, nil
}
