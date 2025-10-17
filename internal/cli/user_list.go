package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var userListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List users in an organization (admin only)",
	Long: `List all users in an organization with their roles and basic information.
Only users with admin role can execute this command.

Examples:
  # List users in active organization
  nvolt user list

  # List users in specific organization
  nvolt user list -o org-id-123
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		return runUserList(aclService, targetOrgID)
	},
}

func init() {
	userCmd.AddCommand(userListCmd)

	userListCmd.Flags().StringP("org", "o", "", "Organization ID (defaults to active org)")
}

func runUserList(aclService *services.ACLService, orgID string) error {
	fmt.Println(titleStyle.Render("Organization Users"))

	// Get organization name
	orgName, _, err := aclService.GetActiveOrgName(orgID)
	if err != nil {
		fmt.Println(warnStyle.Render(fmt.Sprintf("→ Organization ID: %s", orgID)))
	} else {
		fmt.Println(infoStyle.Render(fmt.Sprintf("→ Organization: %s (%s)", orgName, orgID)))
	}

	// Fetch users
	fmt.Println(infoStyle.Render("→ Fetching users..."))
	users, err := aclService.GetOrgUsers(orgID)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to fetch users: %v", err)))
		return err
	}

	if len(users) == 0 {
		fmt.Println(warnStyle.Render("No users found in this organization."))
		return nil
	}

	// Display users as table
	fmt.Printf("\n")
	fmt.Println(titleStyle.Render(fmt.Sprintf("Found %d user(s):", len(users))))
	fmt.Println()

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Email", "Role", "User ID"})

	for _, orgUser := range users {
		// User info
		userName := "N/A"
		userEmail := "N/A"
		if orgUser.User != nil {
			userName = orgUser.User.Name
			userEmail = orgUser.User.Email
		}

		// Role styling
		roleText := orgUser.Role
		if orgUser.Role == "admin" {
			roleText = successStyle.Render("admin")
		} else {
			roleText = infoStyle.Render("dev")
		}

		t.AppendRow(table.Row{userName, userEmail, roleText, orgUser.UserID})
	}

	t.SetStyle(table.StyleLight)
	t.Render()

	return nil
}
