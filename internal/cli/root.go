package cli

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Version is set via -ldflags at build time
	Version = "dev"

	// Styles for colorful output - matching nvolt.io branding
	brandColor    = lipgloss.Color("#FFC107") // Yellow/gold brand color
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(brandColor)
	successStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D084"))
	errorStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B"))
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	warnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	listItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))

	// ASCII Logo
	logoText = `
    ███╗   ██╗██╗   ██╗ ██████╗ ██╗  ████████╗
    ████╗  ██║██║   ██║██╔═══██╗██║  ╚══██╔══╝
    ██╔██╗ ██║██║   ██║██║   ██║██║     ██║
    ██║╚██╗██║╚██╗ ██╔╝██║   ██║██║     ██║
    ██║ ╚████║ ╚████╔╝ ╚██████╔╝███████╗██║
    ╚═╝  ╚═══╝  ╚═══╝   ╚═════╝ ╚══════╝╚═╝
    `
)

var rootCmd = &cobra.Command{
	Use:     "nvolt",
	Short:   "Secure environment variable synchronization CLI",
	Long:    renderLogo() + "\n\n" + "Ultra-secure, Zero-Trust, CLI-first environment variable manager",
	Version: Version,
}

func renderLogo() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(brandColor).
		Bold(true)
	return logoStyle.Render(logoText)
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

// resolveAndSetActiveOrg resolves the active organization using the NEW smart logic:
// 1. If user has only one org → auto-select it (no prompt, no save)
// 2. If user has multiple orgs + active_org_id set → use it
// 3. If user has multiple orgs + no active_org_id → prompt user + ask to set as default
func resolveAndSetActiveOrg(machineConfig *services.MachineConfig, aclService *services.ACLService) error {
	// Use the new smart org resolver
	orgID, orgName, shouldSave, err := aclService.ResolveOrgID(machineConfig)
	if err != nil {
		return err
	}

	// Set the active org (in memory for this session)
	machineConfig.Config.ActiveOrgID = orgID

	// Only save to disk if user chose to set as default
	if shouldSave {
		if err := machineConfig.SaveActiveOrg(orgID); err != nil {
			return fmt.Errorf("failed to save default organization: %w", err)
		}
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Set '%s' as default organization", orgName)))
	}

	return nil
}
