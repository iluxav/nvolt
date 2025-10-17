package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set default environment and/or organization",
	Long:  `Set the default environment and/or active organization. These will be used for all operations unless overridden with flags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineConfig := services.MachineConfigFromContext(cmd.Context())

		environment, _ := cmd.Flags().GetString("environment")
		orgID, _ := cmd.Flags().GetString("org")
		serverURL, _ := cmd.Flags().GetString("server-url")
		// Check if at least one flag is provided
		if environment == "" && orgID == "" && serverURL == "" {
			return fmt.Errorf("at least one of --environment (-e), --org (-o) or --server-url (-s) must be specified")
		}

		if serverURL != "" {
			if err := machineConfig.SaveServerURL(serverURL); err != nil {
				return fmt.Errorf("failed to save server URL: %w", err)
			}
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Server URL set to: %s", serverURL)))
		}

		// Update environment if provided
		if environment != "" {
			if err := machineConfig.SaveDefaultEnvironment(environment); err != nil {
				return fmt.Errorf("failed to save default environment: %w", err)
			}
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Default environment set to: %s", environment)))
		}

		// Update org if provided
		if orgID != "" {
			if err := machineConfig.SaveActiveOrg(orgID); err != nil {
				return fmt.Errorf("failed to save active organization: %w", err)
			}
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ Active organization set to: %s", orgID)))
		}

		return nil
	},
}

func init() {
	setCmd.Flags().StringP("environment", "e", "", "Default environment to set")
	setCmd.Flags().StringP("org", "o", "", "Active organization ID to set")
	setCmd.Flags().StringP("server-url", "s", "", "Server URL to set")
	rootCmd.AddCommand(setCmd)
}
