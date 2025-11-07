package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage machine access and keys",
	Long:  `Add or remove machines from the vault access list.`,
}

var machineAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new machine and generate its keypair",
	Long: `Generate a new keypair for a machine (CI/CD or another device).

Example:
  nvolt machine add ci-server
  nvolt machine add alice-laptop`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]

		// TODO: Implement machine add logic
		fmt.Printf("Adding machine: %s\n", machineName)

		return fmt.Errorf("not yet implemented")
	},
}

var machineRmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a machine and revoke its access",
	Long: `Revoke access for a machine by removing its keys and re-wrapping master key.

Example:
  nvolt machine rm ci-server
  nvolt machine rm alice-laptop`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]

		// TODO: Implement machine remove logic
		fmt.Printf("Removing machine: %s\n", machineName)

		return fmt.Errorf("not yet implemented")
	},
}

var machineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all machines with access",
	Long:  `Display all machines that have access to the vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement machine list logic
		fmt.Printf("Listing machines...\n")

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	machineCmd.AddCommand(machineAddCmd)
	machineCmd.AddCommand(machineRmCmd)
	machineCmd.AddCommand(machineListCmd)
	rootCmd.AddCommand(machineCmd)
}
