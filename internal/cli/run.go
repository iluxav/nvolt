package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a command with decrypted secrets as environment variables",
	Long: `Load decrypted secrets into environment and execute a command.

Examples:
  nvolt run npm start
  nvolt run -e production ./app
  nvolt run -c "go test ./..."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		environment, _ := cmd.Flags().GetString("env")
		command, _ := cmd.Flags().GetString("command")

		var execCmd string
		if command != "" {
			execCmd = command
		} else if len(args) > 0 {
			execCmd = fmt.Sprintf("%v", args)
		}

		// TODO: Implement run logic
		fmt.Printf("Running command with secrets...\n")
		fmt.Printf("Environment: %s\n", environment)
		fmt.Printf("Command: %s\n", execCmd)

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	runCmd.Flags().StringP("env", "e", "default", "Environment name")
	runCmd.Flags().StringP("command", "c", "", "Command to execute")
	rootCmd.AddCommand(runCmd)
}
