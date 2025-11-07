package cli

import (
	"fmt"

	"github.com/nvolt/nvolt/internal/crypto"
	"github.com/nvolt/nvolt/internal/vault"
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
		return runSync(rotate)
	},
}

func runSync(rotate bool) error {
	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Get current machine info
	machineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		return fmt.Errorf("failed to load machine info: %w", err)
	}

	var masterKey []byte

	if rotate {
		fmt.Println("Rotating master key...")

		// Load existing master key first to re-encrypt secrets
		oldMasterKey, err := vault.UnwrapMasterKey(vaultPath)
		if err != nil {
			return fmt.Errorf("failed to unwrap old master key: %w", err)
		}

		// Generate new master key
		masterKey, err = crypto.GenerateAESKey()
		if err != nil {
			return fmt.Errorf("failed to generate new master key: %w", err)
		}

		// Re-encrypt all secrets with new key
		if err := rotateSecretsEncryption(vaultPath, oldMasterKey, masterKey); err != nil {
			return fmt.Errorf("failed to re-encrypt secrets: %w", err)
		}

		fmt.Println("✓ Generated new master key and re-encrypted all secrets")
	} else {
		fmt.Println("Re-wrapping master key for all machines...")

		// Load existing master key
		masterKey, err = vault.UnwrapMasterKey(vaultPath)
		if err != nil {
			return fmt.Errorf("failed to unwrap master key: %w", err)
		}

		fmt.Println("✓ Loaded existing master key")
	}

	// Wrap master key for all machines
	if err := vault.WrapMasterKeyForMachines(vaultPath, masterKey, machineInfo.ID); err != nil {
		return fmt.Errorf("failed to wrap master key: %w", err)
	}

	// List machines to show what was done
	machines, err := vault.ListMachines(vaultPath)
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

	return nil
}

// rotateSecretsEncryption re-encrypts all secrets with a new master key
func rotateSecretsEncryption(vaultPath string, oldKey, newKey []byte) error {
	paths := vault.GetVaultPaths(vaultPath)

	// Get all environment directories
	envDirs, err := vault.ListDirs(paths.Secrets)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	totalSecrets := 0

	// Re-encrypt secrets in each environment
	for _, envDir := range envDirs {
		// Get environment name from directory
		environment := vault.GetDirName(envDir)

		// List all secrets in this environment
		secretFiles, err := vault.ListFiles(paths.GetSecretsPath(environment))
		if err != nil {
			// Skip if directory doesn't exist or is empty
			continue
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

			totalSecrets++
		}
	}

	if totalSecrets == 0 {
		fmt.Println("No secrets found to re-encrypt")
	} else {
		fmt.Printf("✓ Re-encrypted %d secret(s)\n", totalSecrets)
	}

	return nil
}

func init() {
	syncCmd.Flags().Bool("rotate", false, "Rotate the master key")
	rootCmd.AddCommand(syncCmd)
}
