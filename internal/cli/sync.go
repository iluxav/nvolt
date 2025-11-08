package cli

import (
	"fmt"
	"os"

	"github.com/iluxav/nvolt/internal/config"
	"github.com/iluxav/nvolt/internal/crypto"
	"github.com/iluxav/nvolt/internal/git"
	"github.com/iluxav/nvolt/internal/vault"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-wrap or rotate master keys",
	Long: `Synchronize and re-wrap keys for all machines in a specific environment.

Use --rotate to generate a new master key and re-encrypt all secrets.

Examples:
  nvolt sync                    # Re-wrap existing keys for default environment
  nvolt sync -e production      # Re-wrap keys for production environment
  nvolt sync --rotate           # Rotate master key for default environment
  nvolt sync -e prod --rotate   # Rotate master key for production environment`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rotate, _ := cmd.Flags().GetBool("rotate")
		environment, _ := cmd.Flags().GetString("env")
		autoGrant, _ := cmd.Flags().GetBool("auto-grant")
		return runSync(rotate, environment, autoGrant)
	},
}

func runSync(rotate bool, environment string, autoGrant bool) error {
	// Ensure machine is initialized
	if err := EnsureMachineInitialized(); err != nil {
		return err
	}

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Pull latest changes in global mode BEFORE doing any work
	// This ensures we have the latest machine keys from other machines
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("Global mode: pulling latest changes...")
		if err := git.SafePull(repoPath); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}
		fmt.Println("✓ Pulled latest changes from repository")
	}

	// Get current machine info
	machineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		return fmt.Errorf("failed to load machine info: %w", err)
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

	var masterKey []byte

	if rotate {
		fmt.Printf("Rotating master key for environment '%s'...\n", environment)

		// Load existing master key first to re-encrypt secrets
		oldMasterKey, err := vault.UnwrapMasterKey(paths, environment)
		if err != nil {
			return fmt.Errorf("failed to unwrap old master key: %w", err)
		}

		// Generate new master key
		masterKey, err = crypto.GenerateAESKey()
		if err != nil {
			return fmt.Errorf("failed to generate new master key: %w", err)
		}

		// Re-encrypt secrets in this environment with new key
		if err := rotateSecretsEncryption(paths, environment, oldMasterKey, masterKey); err != nil {
			return fmt.Errorf("failed to re-encrypt secrets: %w", err)
		}

		fmt.Println("✓ Generated new master key and re-encrypted all secrets")
	} else {
		fmt.Printf("Re-wrapping master key for environment '%s' for all machines...\n", environment)

		// Load existing master key
		masterKey, err = vault.UnwrapMasterKey(paths, environment)
		if err != nil {
			return fmt.Errorf("failed to unwrap master key: %w", err)
		}

		fmt.Println("✓ Loaded existing master key")
	}

	// Wrap master key for all machines
	if autoGrant {
		fmt.Println("Auto-granting access to all machines...")
	} else {
		fmt.Println("Wrapping master key for machines (will prompt for new machines)...")
	}
	if err := vault.WrapMasterKeyForMachines(paths, environment, masterKey, machineInfo.ID, autoGrant); err != nil {
		return fmt.Errorf("failed to wrap master key: %w", err)
	}

	// List machines to show what was done
	machines, err := vault.ListMachines(paths)
	if err != nil {
		return fmt.Errorf("failed to list machines: %w", err)
	}

	fmt.Printf("\n✓ Master key wrapped for %d machine(s):\n", len(machines))
	for _, m := range machines {
		fmt.Printf("  - %s (%s)\n", m.ID, m.Hostname)
	}

	if rotate {
		fmt.Println("\nNote: The master key has been rotated. All machines can now decrypt secrets with the new key.")
	} else {
		fmt.Println("\nNote: All machines now have access to decrypt secrets.")
	}

	// Auto-commit and push in global mode
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("\nGlobal mode: committing and pushing changes...")

		// Generate commit message
		var commitMsg string
		if rotate {
			commitMsg = "Rotate master key and re-encrypt all secrets"
		} else {
			commitMsg = "Re-wrap master key for all machines"
		}

		// Commit and push
		if err := git.CommitAndPush(repoPath, commitMsg, ".nvolt"); err != nil {
			return fmt.Errorf("failed to commit and push changes: %w", err)
		}

		fmt.Println("✓ Changes committed and pushed to repository")
	}

	return nil
}

// rotateSecretsEncryption re-encrypts all secrets in a specific environment with a new master key
func rotateSecretsEncryption(paths *vault.Paths, environment string, oldKey, newKey []byte) error {
	// List all secrets in this environment
	secretFiles, err := vault.ListFiles(paths.GetSecretsPath(environment))
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if len(secretFiles) == 0 {
		fmt.Printf("No secrets found in environment '%s' to re-encrypt\n", environment)
		return nil
	}

	fmt.Printf("Re-encrypting %d secret(s) in environment '%s'...\n", len(secretFiles), environment)

	for _, secretFile := range secretFiles {
		// Extract key name from filename
		key := vault.GetSecretKeyFromFilename(secretFile)

		// Load encrypted secret
		encrypted, err := vault.LoadEncryptedSecret(paths, environment, key)
		if err != nil {
			return fmt.Errorf("failed to load secret %s: %w", key, err)
		}

		// Decrypt with old key
		plaintext, err := vault.DecryptSecret(oldKey, encrypted)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret %s with old key: %w", key, err)
		}

		// Re-encrypt with new key
		newEncrypted, err := vault.EncryptSecret(newKey, plaintext)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret %s with new key: %w", key, err)
		}

		// Save re-encrypted secret
		if err := vault.SaveEncryptedSecret(paths, environment, key, newEncrypted); err != nil {
			return fmt.Errorf("failed to save re-encrypted secret %s: %w", key, err)
		}
	}

	fmt.Printf("✓ Re-encrypted %d secret(s)\n", len(secretFiles))

	return nil
}

func init() {
	syncCmd.Flags().Bool("rotate", false, "Rotate the master key")
	syncCmd.Flags().StringP("env", "e", "default", "Environment name")
	syncCmd.Flags().Bool("auto-grant", false, "Automatically grant access to all machines without prompting")
	rootCmd.AddCommand(syncCmd)
}
