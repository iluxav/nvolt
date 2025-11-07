package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Decrypt and pull secrets from vault",
	Long: `Decrypt secrets from the vault and output them in .env format.

Examples:
  nvolt pull
  nvolt pull -e production
  nvolt pull -e staging -p myproject
  nvolt pull --write  # Write to .env file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		environment, _ := cmd.Flags().GetString("env")
		project, _ := cmd.Flags().GetString("project")
		write, _ := cmd.Flags().GetBool("write")

		// TODO: Implement pull logic
		fmt.Printf("Pulling secrets...\n")
		fmt.Printf("Environment: %s\n", environment)
		if project != "" {
			fmt.Printf("Project: %s\n", project)
		}
		if write {
			fmt.Printf("Will write to .env file\n")
		}

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	pullCmd.Flags().StringP("env", "e", "default", "Environment name")
	pullCmd.Flags().StringP("project", "p", "", "Project name (auto-detected if not specified)")
	pullCmd.Flags().BoolP("write", "w", false, "Write output to .env file")
	rootCmd.AddCommand(pullCmd)
}
