package cli

import (
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organization settings",
	Long:  `Manage organization settings including selecting the active organization`,
}

var orgSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the active organization",
	Long:  `Set the active organization for the current machine. This will be used for all operations unless overridden with the -o flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineConfig := cmd.Context().Value("machine_config").(*services.MachineConfig)
		aclService := cmd.Context().Value("acl_service").(*services.ACLService)

		// Fetch user organizations
		userOrgs, err := aclService.GetUserOrgs()
		if err != nil {
			return fmt.Errorf("failed to fetch organizations: %w", err)
		}

		if len(userOrgs) == 0 {
			return fmt.Errorf("you do not belong to any organization")
		}

		// Show interactive org selector
		selectedOrgID, err := helpers.ShowOrgSelector(userOrgs)
		if err != nil {
			return err
		}

		// Save the selected org
		err = machineConfig.SaveActiveOrg(selectedOrgID)
		if err != nil {
			return fmt.Errorf("failed to save active organization: %w", err)
		}

		// Find org name for display
		var orgName string
		for _, org := range userOrgs {
			if org.OrgID.String() == selectedOrgID {
				orgName = org.Org.Name
				break
			}
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Active organization set to: %s", orgName)))

		return nil
	},
}

func init() {
	orgCmd.AddCommand(orgSetCmd)
	rootCmd.AddCommand(orgCmd)
}
