package cli

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Version is set via -ldflags at build time
	Version = "dev"

	// Styles for colorful output
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D4CDB"))
	successStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D084"))
	errorStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B"))
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	warnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	listItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
)

var rootCmd = &cobra.Command{
	Use:     "nvolt",
	Short:   "Secure environment variable synchronization CLI",
	Version: Version,
}

func Execute(machineConfig *services.MachineConfig, aclService *services.ACLService) error {
	// Skip org resolution for version and help flags
	isVersionFlag := len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version" || os.Args[1] == "version")
	isHelpFlag := len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help")

	// Resolve active organization (only if user is logged in)
	// Skip for login command to avoid chicken-and-egg issue during first login
	if machineConfig.Config.JWT_Token != "" && !isVersionFlag && !isHelpFlag {
		// Check if the command being run is 'login' - if so, skip org resolution
		isLoginCommand := len(os.Args) > 1 && os.Args[1] == "login"
		if !isLoginCommand {
			err := resolveAndSetActiveOrg(machineConfig, aclService)
			if err != nil {
				// Don't fail the entire execution, just warn
				// This allows commands like 'nvolt org set' to work even if org resolution fails
				fmt.Println(warnStyle.Render(fmt.Sprintf("Warning: %v", err)))
			}
		}
	}

	err := machineConfig.TryResolveLocalDirProjectNameAndEnvironment()
	if err != nil {
		fmt.Println(warnStyle.Render(fmt.Sprintf("Warning: %v", err)))
	}

	// Create SecretsClient and MachineService for commands that need it
	secretsClient := services.NewSecretsClient(machineConfig)
	machineService := services.NewMachineService(machineConfig.Config)

	ctx := context.WithValue(context.Background(), types.MachineConfigKey, machineConfig)
	ctx = context.WithValue(ctx, types.ACLServiceKey, aclService)
	ctx = context.WithValue(ctx, types.SecretsClientKey, secretsClient)
	ctx = context.WithValue(ctx, types.MachineServiceKey, machineService)
	return rootCmd.ExecuteContext(ctx)
}

func init() {

	rootCmd.AddCommand(loginCmd)
}

// resolveAndSetActiveOrg resolves the active organization based on the logic:
// 1. If user has only one org, use it
// 2. If user has multiple orgs and active_org is set, use it
// 3. If user has multiple orgs and no active_org, show interactive selector
func resolveAndSetActiveOrg(machineConfig *services.MachineConfig, aclService *services.ACLService) error {
	// Fetch user organizations
	userOrgs, err := aclService.GetUserOrgs()
	if err != nil {
		return fmt.Errorf("failed to fetch organizations: %w", err)
	}

	if len(userOrgs) == 0 {
		return fmt.Errorf("user does not belong to any organization")
	}

	// Store fetched orgs in MachineConfig for later use
	machineConfig.OrgUsers = userOrgs

	// Resolve active org
	activeOrgID, err := helpers.ResolveActiveOrg(userOrgs, machineConfig.Config.ActiveOrgID)
	if err != nil {
		return err
	}

	// Save if it's different from current or not set
	if machineConfig.Config.ActiveOrgID != activeOrgID {
		err = machineConfig.SaveActiveOrg(activeOrgID)
		if err != nil {
			return fmt.Errorf("failed to save active organization: %w", err)
		}
	}

	return nil
}
