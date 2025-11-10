package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iluxav/nvolt/internal/config"
	"github.com/iluxav/nvolt/internal/crypto"
	"github.com/iluxav/nvolt/internal/ui"
	"github.com/iluxav/nvolt/internal/vault"
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

		var execArgs []string
		if command != "" {
			// Use shell to execute string command
			execArgs = []string{"sh", "-c", command}
		} else if len(args) > 0 {
			// Use args directly
			execArgs = args
		} else {
			return fmt.Errorf("no command specified")
		}

		return runWithSecrets(environment, execArgs)
	},
}

func runWithSecrets(environment string, cmdArgs []string) error {
	// Ensure machine is initialized
	if err := EnsureMachineInitialized(); err != nil {
		return err
	}

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Detect project name in global mode
	var project string
	if vault.IsGlobalMode(vaultPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		detectedProject, _, err := config.GetProjectName(cwd, "")
		if err != nil {
			return fmt.Errorf("failed to detect project name: %w", err)
		}
		project = detectedProject
	}

	paths := vault.GetVaultPaths(vaultPath, project)

	// Unwrap master key
	masterKey, err := vault.UnwrapMasterKey(paths, environment)
	if err != nil {
		return fmt.Errorf("failed to unwrap master key: %w", err)
	}
	// Ensure master key is cleared from memory when done
	defer crypto.ZeroBytes(masterKey)

	// Get list of secret files
	secretsDir := paths.GetSecretsPath(environment)
	secretFiles, err := vault.ListFiles(secretsDir)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if len(secretFiles) == 0 {
		return fmt.Errorf("no secrets found for environment '%s'", environment)
	}

	// Decrypt all secrets
	secrets := make(map[string]string)
	for _, secretFile := range secretFiles {
		filename := filepath.Base(secretFile)

		// Get key name
		key := filename
		if len(key) > 9 && key[len(key)-9:] == ".enc.json" {
			key = key[:len(key)-9]
		}

		// Load and decrypt secret
		encrypted, err := vault.LoadEncryptedSecret(paths, environment, key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load secret %s: %v\n", key, err)
			continue
		}

		value, err := vault.DecryptSecret(masterKey, encrypted)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to decrypt secret %s: %v\n", key, err)
			continue
		}

		secrets[key] = value
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets could be decrypted")
	}

	ui.Success(fmt.Sprintf("Loaded %d secrets from environment '%s'", len(secrets), ui.Cyan(environment)))
	ui.Info(fmt.Sprintf("Running: %s\n", ui.Gray(strings.Join(cmdArgs, " "))))

	// Prepare environment
	env := os.Environ()
	for key, value := range secrets {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute command
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Command exited with non-zero status
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

func init() {
	runCmd.Flags().StringP("env", "e", "default", "Environment name")
	runCmd.Flags().StringP("command", "c", "", "Command to execute")
	rootCmd.AddCommand(runCmd)
}
