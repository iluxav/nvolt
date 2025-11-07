package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nvolt vault and machine keys",
	Long: `Initialize a new nvolt vault and generate machine keypair if needed.

Local mode (current directory):
  nvolt init

Global mode (dedicated GitHub repo):
  nvolt init --repo org/repo

This command will:
- Generate an RSA/Ed25519 keypair for this machine (if not exists)
- Create .nvolt/ directory structure
- Clone the repo (if --repo provided) into ~/.nvolt/projects/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")

		// TODO: Implement init logic
		fmt.Printf("Initializing nvolt vault...\n")
		if repo != "" {
			fmt.Printf("Repository: %s\n", repo)
		} else {
			fmt.Printf("Mode: Local\n")
		}

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	initCmd.Flags().StringP("repo", "r", "", "GitHub repository (org/repo) for global mode")
	rootCmd.AddCommand(initCmd)
}
