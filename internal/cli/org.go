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
	RunE: func(cmd *cobra.Command, args []string) error {
		machineConfig := services.MachineConfigFromContext(cmd.Context())
		aclService := services.ACLServiceFromContext(cmd.Context())

		// Fetch user organizations
		userOrgs, err := aclService.GetUserOrgs()
		if err != nil {
			return fmt.Errorf("failed to fetch organizations: %w", err)
		}

		if len(userOrgs) == 0 {
			return fmt.Errorf("you do not belong to any organization")
		}

		// Get current active org
		currentOrgID := machineConfig.Config.ActiveOrgID
		var currentOrgName string
		for _, org := range userOrgs {
			if org.OrgID == currentOrgID {
				currentOrgName = org.Org.Name
				break
			}
		}

		// If user has only one org, just display it
		if len(userOrgs) == 1 {
			orgName := userOrgs[0].Org.Name
			if orgName == "" {
				orgName = userOrgs[0].OrgID
			}
			fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ Active Organization: %s (%s)", orgName, userOrgs[0].Role)))
			return nil
		}

		// User has multiple orgs - show current and allow switching
		fmt.Println(titleStyle.Render("\n Organization Management\n"))

		if currentOrgName != "" {
			fmt.Println(infoStyle.Render(fmt.Sprintf("→ Current active organization: %s", currentOrgName)))
		} else {
			fmt.Println(warnStyle.Render("→ No active organization set"))
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ You belong to %d organization(s)\n", len(userOrgs))))

		// Show interactive org selector
		selectedOrgID, err := helpers.ShowOrgSelector(userOrgs)
		if err != nil {
			return err
		}

		// If user selected the same org, just confirm
		if selectedOrgID == currentOrgID {
			fmt.Println(infoStyle.Render("\n✓ Keeping current organization"))
			return nil
		}

		// Save the newly selected org
		err = machineConfig.SaveActiveOrg(selectedOrgID)
		if err != nil {
			return fmt.Errorf("failed to save active organization: %w", err)
		}

		// Find org name for display
		var newOrgName string
		for _, org := range userOrgs {
			if org.OrgID == selectedOrgID {
				newOrgName = org.Org.Name
				break
			}
		}

		fmt.Println(successStyle.Render(fmt.Sprintf("\n✓ Active organization changed to: %s", newOrgName)))

		return nil
	},
}

var orgSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the active organization",
	Long:  `Set the active organization for the current machine. This will be used for all operations unless overridden with the -o flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		machineConfig := services.MachineConfigFromContext(cmd.Context())
		aclService := services.ACLServiceFromContext(cmd.Context())

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
			if org.OrgID == selectedOrgID {
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
