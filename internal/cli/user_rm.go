package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var userRmCmd = &cobra.Command{
	Use:   "rm [email]",
	Short: "Remove a user from an organization (admin only)",
	Long: `Remove a user from an organization and all related permissions.
Only users with admin role can execute this command.

Examples:
  # Remove user from active organization
  nvolt user rm john@example.com

  # Remove user from specific organization
  nvolt user rm john@example.com -o org-id-123
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

		return runUserRemove(aclService, targetOrgID, email)
	},
}

func init() {
	userCmd.AddCommand(userRmCmd)
	userRmCmd.Flags().StringP("org", "o", "", "Organization ID (defaults to active org)")
}

func runUserRemove(aclService *services.ACLService, orgID, email string) error {
	fmt.Println(titleStyle.Render("Removing User from Organization"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Email: %s", email)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Organization ID: %s", orgID)))

	// Confirm removal
	fmt.Println(warnStyle.Render("⚠ This will remove the user from the organization and delete all their permissions."))
	fmt.Print(infoStyle.Render("Are you sure? (yes/no): "))

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" && confirmation != "y" {
		fmt.Println(infoStyle.Render("→ Cancelled"))
		return nil
	}

	// Call API
	fmt.Println(infoStyle.Render("→ Sending request to server..."))
	err := aclService.RemoveUserFromOrg(orgID, email)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to remove user: %v", err)))
		return err
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ User %s successfully removed from organization", email)))
	return nil
}
