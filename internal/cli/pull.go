package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/iluxav/nvolt/internal/config"
	"github.com/iluxav/nvolt/internal/crypto"
	"github.com/iluxav/nvolt/internal/git"
	"github.com/iluxav/nvolt/internal/vault"
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

		return runPull(environment, project, write)
	},
}

func runPull(environment, project string, write bool) error {
	fmt.Println("Pulling secrets from vault...")

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Auto-pull in global mode
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("Global mode: pulling latest changes from repository...")
		if err := git.SafePull(repoPath); err != nil {
			return fmt.Errorf("failed to pull from repository: %w", err)
		}
		fmt.Println("✓ Repository up to date")

		// Detect or use provided project name
		if project == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			detectedProject, _, err := config.GetProjectName(cwd, "")
			if err != nil {
				return fmt.Errorf("failed to detect project name. Use -p flag to specify: %w", err)
			}
			project = detectedProject
			fmt.Printf("Detected project: %s\n", project)
		}
	}

	// Get vault paths with project name (empty for local mode)
	paths := vault.GetVaultPaths(vaultPath, project)

	// Unwrap master key
	masterKey, err := vault.UnwrapMasterKey(paths, environment)
	if err != nil {
		return fmt.Errorf("failed to unwrap master key: %w\nMake sure you have pushed secrets first", err)
	}
	// Ensure master key is cleared from memory when done
	defer crypto.ZeroBytes(masterKey)

	// Get list of secret files for this environment
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
		// Extract key name from filename (remove .enc.json)
		filename := filepath.Base(secretFile)
		if !filepath.IsAbs(secretFile) {
			secretFile = filepath.Join(secretsDir, filename)
		}

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

	fmt.Printf("✓ Decrypted %d secrets from environment '%s'\n\n", len(secrets), environment)

	// Format output
	output := vault.FormatEnvOutput(secrets)

	if write {
		// Write to .env file
		envFile := ".env"
		if environment != "default" {
			envFile = fmt.Sprintf(".env.%s", environment)
		}

		err := os.WriteFile(envFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}

		fmt.Printf("✓ Written to %s\n", envFile)
	} else {
		// Print to stdout
		fmt.Println("Secrets:")

		// Sort keys for consistent output
		keys := make([]string, 0, len(secrets))
		for k := range secrets {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Printf("%s=%s\n", key, secrets[key])
		}
	}

	return nil
}

func init() {
	pullCmd.Flags().StringP("env", "e", "default", "Environment name")
	pullCmd.Flags().StringP("project", "p", "", "Project name (auto-detected if not specified)")
	pullCmd.Flags().BoolP("write", "w", false, "Write output to .env file")
	rootCmd.AddCommand(pullCmd)
}
