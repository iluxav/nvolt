package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/tui"
	"iluxav/nvolt/internal/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive Terminal UI for managing environment variables",
	Long: `Launch an interactive Terminal User Interface (TUI) to browse and manage
environment variables, users, and permissions across different environments.

Features:
- Browse environment variables across multiple environments
- View and manage user permissions
- Interactive navigation with keyboard shortcuts
- Delete confirmation modals
- Reveal/hide encrypted values`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get services from context
		machineConfig := cmd.Context().Value(types.MachineConfigKey).(*services.MachineConfig)
		secretsClient := cmd.Context().Value(types.SecretsClientKey).(*services.SecretsClient)
		aclService := cmd.Context().Value(types.ACLServiceKey).(*services.ACLService)

		// Check if user is logged in
		if machineConfig.Config.JWT_Token == "" {
			return fmt.Errorf("not logged in. Please run 'nvolt login' first")
		}

		// Get project name from flag or resolve from directory
		projectName, _ := cmd.Flags().GetString("project")
		if projectName == "" {
			projectName = machineConfig.GetProject()
			if projectName == "" {
				return fmt.Errorf("project name not specified. Use -p flag or run from a git repository")
			}
		}

		// Create TUI model
		model := tui.NewModel(machineConfig, secretsClient, aclService, projectName)

		// Run the TUI
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("error running TUI: %w", err)
		}

		return nil
	},
}

func init() {
	tuiCmd.Flags().StringP("project", "p", "", "Project name")
	tuiCmd.Flags().StringP("environment", "e", "default", "Environment name")
	tuiCmd.Flags().StringP("org", "o", "", "Organization ID")

	rootCmd.AddCommand(tuiCmd)
}
