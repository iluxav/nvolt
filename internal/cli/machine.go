package cli

import (
	"github.com/spf13/cobra"
)

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage machines for silent authentication",
	Long:  `Manage machines for silent authentication (CI/CD workflows)`,
}

func init() {
	rootCmd.AddCommand(machineCmd)
}
