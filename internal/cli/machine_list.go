package cli

import (
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var machineListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all machines in an organization",
	Long: `List all machines in an organization with their details.

Examples:
  # List machines in active organization
  nvolt machine list

  # List machines in specific organization
  nvolt machine list -o org-id-123
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		orgID, _ := cmd.Flags().GetString("org")

		machineConfig := services.MachineConfigFromContext(cmd.Context())
		machineService := services.MachineServiceFromContext(cmd.Context())
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

		return runMachineList(aclService, machineService, targetOrgID)
	},
}

func init() {
	machineCmd.AddCommand(machineListCmd)

	machineListCmd.Flags().StringP("org", "o", "", "Organization ID (defaults to active org)")
}

func runMachineList(aclService *services.ACLService, machineService *services.MachineService, orgID string) error {
	fmt.Println(titleStyle.Render("Organization Machines"))

	// Get organization name
	orgName, _, err := aclService.GetActiveOrgName(orgID)
	if err != nil {
		fmt.Println(warnStyle.Render(fmt.Sprintf("→ Organization ID: %s", orgID)))
	} else {
		fmt.Println(infoStyle.Render(fmt.Sprintf("→ Organization: %s (%s)", orgName, orgID)))
	}

	// Fetch machines
	fmt.Println(infoStyle.Render("→ Fetching machines..."))
	machines, err := machineService.GetOrgMachines(orgID)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to fetch machines: %v", err)))
		return err
	}

	if len(machines) == 0 {
		fmt.Println(warnStyle.Render("No machines found in this organization."))
		return nil
	}

	// Display machines as table
	fmt.Printf("\n")
	fmt.Println(titleStyle.Render(fmt.Sprintf("Found %d machine(s):", len(machines))))
	fmt.Println()

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Machine Name", "User Name", "User Email", "Created At", "Machine ID"})

	for _, machine := range machines {
		userName := "N/A"
		userEmail := "N/A"
		if machine.User != nil {
			userName = machine.User.Name
			userEmail = machine.User.Email
		}

		// Format the timestamp nicely
		createdAt := helpers.FormatTimestamp(machine.CreatedAt)

		t.AppendRow(table.Row{machine.Name, userName, userEmail, createdAt, machine.MachineID})
	}

	t.SetStyle(table.StyleLight)
	t.Render()

	return nil
}
