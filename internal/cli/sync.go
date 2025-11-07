package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-wrap or rotate master keys",
	Long: `Synchronize and re-wrap keys for all machines.

Use --rotate to generate a new master key and re-encrypt all secrets.

Examples:
  nvolt sync           # Re-wrap existing keys
  nvolt sync --rotate  # Rotate master key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rotate, _ := cmd.Flags().GetBool("rotate")

		// TODO: Implement sync logic
		fmt.Printf("Syncing vault...\n")
		if rotate {
			fmt.Printf("Rotating master key...\n")
		}

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	syncCmd.Flags().Bool("rotate", false, "Rotate the master key")
	rootCmd.AddCommand(syncCmd)
}
