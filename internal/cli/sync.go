package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync encryption keys for all machines in the organization",
	Long: `Synchronize encryption keys across all machines in your organization.

This command is useful after adding a new machine to your organization.
It re-wraps the master encryption key for all machines without modifying secrets.

Examples:
  # Sync all project/environment combinations you have access to
  nvolt sync --all

  # Sync a specific project/environment
  nvolt sync -p my-project -e production

This allows new machines to access existing secrets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		machineConfig := services.MachineConfigFromContext(cmd.Context())
		machineConfig.TryOverrideWithFlags(project, environment)

		secretsClient := services.SecretsClientFromContext(cmd.Context())

		activeOrgID := machineConfig.Config.ActiveOrgID
		if activeOrgID == "" {
			return fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
		}

		if all {
			return runSyncAll(secretsClient, activeOrgID)
		}

		// Validate required context for single sync
		if machineConfig.GetProject() == "" || machineConfig.GetEnvironment() == "" {
			return fmt.Errorf("could not determine project or environment. Use --all flag or ensure you're in a project directory")
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("🔄 Syncing keys for %s/%s...", machineConfig.GetProject(), machineConfig.GetEnvironment())))

		// Call the sync operation in secrets service
		err := secretsClient.SyncKeys(activeOrgID, machineConfig.GetProject(), machineConfig.GetEnvironment())
		if err != nil {
			return fmt.Errorf("failed to sync keys: %w", err)
		}

		fmt.Println(successStyle.Render("✓ Keys synchronized successfully!"))
		fmt.Println(infoStyle.Render("→ All machines in your organization can now access these secrets"))

		return nil
	},
}

func runSyncAll(secretsClient *services.SecretsClient, orgID string) error {
	fmt.Println(titleStyle.Render("🔄 Syncing all project/environment combinations..."))

	// Fetch all project/environment combinations
	projectEnvs, err := secretsClient.GetProjectEnvironments(orgID)
	if err != nil {
		return fmt.Errorf("failed to fetch project environments: %w", err)
	}

	if len(projectEnvs) == 0 {
		fmt.Println(warnStyle.Render("⚠ No project/environment combinations found"))
		return nil
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Found %d project/environment combination(s)", len(projectEnvs))))

	// Sync each project/environment
	successCount := 0
	failedCount := 0
	for i, pe := range projectEnvs {
		fmt.Printf("\n[%d/%d] Syncing %s/%s...", i+1, len(projectEnvs), pe.ProjectName, pe.Environment)

		err := secretsClient.SyncKeys(orgID, pe.ProjectName, pe.Environment)
		if err != nil {
			fmt.Println(warnStyle.Render(fmt.Sprintf(" ✗ Failed: %v", err)))
			failedCount++
		} else {
			fmt.Println(successStyle.Render(" ✓"))
			successCount++
		}
	}

	// Summary
	fmt.Println("\n" + titleStyle.Render("Summary:"))
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Successfully synced: %d", successCount)))
	if failedCount > 0 {
		fmt.Println(warnStyle.Render(fmt.Sprintf("✗ Failed: %d", failedCount)))
	}
	fmt.Println(infoStyle.Render("→ All machines can now access synced secrets"))

	return nil
}

func init() {
	syncCmd.Flags().BoolP("all", "a", false, "Sync all project/environment combinations")
	rootCmd.AddCommand(syncCmd)
}
