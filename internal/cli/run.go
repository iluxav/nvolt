package cli

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a command with environment variables injected from nvolt.io",
	Long:  `Pull encrypted environment variables from the server, decrypt them, and inject them into a subprocess. No .env file is created.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")
		command, _ := cmd.Flags().GetString("command")

		// If -c flag is provided, use it
		// Otherwise, use all remaining args as the command
		if command == "" {
			if len(args) == 0 {
				return fmt.Errorf("command is required. Use -c \"command\" or provide command as arguments")
			}
			command = strings.Join(args, " ")
		}

		machineConfig := services.MachineConfigFromContext(cmd.Context())
		machineConfig.TryOverrideWithFlags(project, environment)

		return runWithSecrets(machineConfig, command)
	},
}

func init() {
	runCmd.PersistentFlags().StringP("command", "c", "", "Command to execute with injected environment variables (optional if command provided as args)")
	rootCmd.AddCommand(runCmd)
}

func runWithSecrets(machineConfig *services.MachineConfig, command string) error {
	fmt.Println(titleStyle.Render("🔐 nvolt run"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Project: %s", machineConfig.GetProject())))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Environment: %s", machineConfig.GetEnvironment())))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Command: %s", command)))

	// Step 1: Pull and decrypt secrets from server
	fmt.Println("\n" + titleStyle.Render("Pulling secrets from server..."))

	secretsClient := services.NewSecretsClient(machineConfig)

	vars, err := secretsClient.PullSecrets(machineConfig.GetProject(), machineConfig.GetEnvironment(), "")
	if err != nil {
		// Check if it's a permission error (403 Forbidden)
		if strings.Contains(err.Error(), "status: 403") || strings.Contains(err.Error(), "Forbidden") {
			fmt.Println("\n" + warnStyle.Render("⚠ Permission Denied"))
			fmt.Println(infoStyle.Render("\nYou don't have read permission for this environment."))
			fmt.Println(infoStyle.Render(fmt.Sprintf("  Project: %s", machineConfig.GetProject())))
			fmt.Println(infoStyle.Render(fmt.Sprintf("  Environment: %s", machineConfig.GetEnvironment())))
			fmt.Println(infoStyle.Render("\nPlease contact your organization admin to grant you access."))
			return nil
		}
		return fmt.Errorf("failed to pull secrets: %w", err)
	}

	if len(vars) == 0 {
		fmt.Println("\n" + warnStyle.Render("⚠ No secrets found for this scope"))
		fmt.Println(infoStyle.Render("Running command without injected variables..."))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("✓ Successfully pulled %d secret(s)!", len(vars))))
	}

	// Clear decrypted secrets from memory after execution
	defer func() {
		for k := range vars {
			vars[k] = ""
			delete(vars, k)
		}
	}()

	// Step 2: Prepare command execution
	fmt.Println("\n" + titleStyle.Render("Executing command..."))
	fmt.Println(infoStyle.Render("─────────────────────────────────────────────────────"))

	// Determine shell to use based on OS
	var shell string
	var shellFlag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		shellFlag = "/C"
	} else {
		// Unix-like systems (Linux, macOS, etc.)
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		shellFlag = "-c"
	}

	// Create command
	execCmd := exec.Command(shell, shellFlag, command)

	// Step 3: Inject environment variables
	// Start with current environment
	execCmd.Env = os.Environ()

	// Add/override with pulled secrets
	for key, value := range vars {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Step 4: Connect stdin/stdout/stderr for interactive commands
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Step 5: Execute the command
	err = execCmd.Run()

	fmt.Println(infoStyle.Render("─────────────────────────────────────────────────────"))

	if err != nil {
		// Check if it's an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Command exited with code %d", exitErr.ExitCode())))
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Command completed successfully"))
	return nil
}

// Helper function to parse shell command (handles quotes, etc.)
func parseCommand(cmdString string) (string, []string) {
	parts := strings.Fields(cmdString)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
