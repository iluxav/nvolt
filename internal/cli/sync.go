package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync org-level encryption key for all machines in the organization",
	Long: `Synchronize the organization's master encryption key across all machines.

This command is useful after adding a new machine to your organization.
It re-wraps the org-level master key for all machines without modifying secrets.

With the simplified org-level encryption model, this command only needs to run once
to give all machines access to ALL projects and environments in the organization.

Examples:
  # Sync org-level master key (recommended after adding new machines)
  nvolt sync

This allows new machines to access all existing secrets in the organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineConfig := services.MachineConfigFromContext(cmd.Context())
		secretsClient := services.SecretsClientFromContext(cmd.Context())

		activeOrgID := machineConfig.Config.ActiveOrgID
		if activeOrgID == "" {
			return fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
		}

		fmt.Println(titleStyle.Render("🔄 Syncing org-level master key..."))

		// Call the org-level sync operation
		err := secretsClient.SyncOrgKeys(activeOrgID)
		if err != nil {
			return fmt.Errorf("failed to sync keys: %w", err)
		}

		fmt.Println(successStyle.Render("\n✓ Keys synchronized successfully!"))
		fmt.Println(infoStyle.Render("→ All machines in your organization can now access all secrets"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
