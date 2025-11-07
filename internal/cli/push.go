package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Encrypt and push secrets to vault",
	Long: `Encrypt secrets from a .env file and store them in the vault.

Examples:
  nvolt push -f .env
  nvolt push -f .env.production -e production
  nvolt push -f .env -p myproject -e staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		envFile, _ := cmd.Flags().GetString("file")
		environment, _ := cmd.Flags().GetString("env")
		project, _ := cmd.Flags().GetString("project")

		// TODO: Implement push logic
		fmt.Printf("Pushing secrets...\n")
		fmt.Printf("File: %s\n", envFile)
		fmt.Printf("Environment: %s\n", environment)
		if project != "" {
			fmt.Printf("Project: %s\n", project)
		}

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	pushCmd.Flags().StringP("file", "f", ".env", "Environment file to encrypt")
	pushCmd.Flags().StringP("env", "e", "default", "Environment name")
	pushCmd.Flags().StringP("project", "p", "", "Project name (auto-detected if not specified)")
	rootCmd.AddCommand(pushCmd)
}
