package cli

import (
	"fmt"

	"github.com/nvolt/nvolt/internal/crypto"
	"github.com/nvolt/nvolt/internal/vault"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Encrypt and push secrets to vault",
	Long: `Encrypt secrets from a .env file or command-line and store them in the vault.

Examples:
  nvolt push -f .env
  nvolt push -k FOO=bar -k BAZ=qux
  nvolt push -f .env -k OVERRIDE=value
  nvolt push -f .env.production -e production
  nvolt push -f .env -p myproject -e staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		envFile, _ := cmd.Flags().GetString("file")
		environment, _ := cmd.Flags().GetString("env")
		project, _ := cmd.Flags().GetString("project")
		keyValues, _ := cmd.Flags().GetStringSlice("key")

		return runPush(envFile, environment, project, keyValues)
	},
}

func runPush(envFile, environment, project string, keyValues []string) error {
	fmt.Println("Pushing secrets to vault...")

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	paths := vault.GetVaultPaths(vaultPath)

	// Collect secrets from file and/or command line
	secrets := make(map[string]string)

	// Load from file if specified and exists
	if envFile != "" && vault.FileExists(envFile) {
		fileSecrets, err := vault.ParseEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to parse env file: %w", err)
		}
		fmt.Printf("Loaded %d secrets from %s\n", len(fileSecrets), envFile)
		for k, v := range fileSecrets {
			secrets[k] = v
		}
	} else if envFile != "" && envFile != ".env" {
		// Only error if a specific file was requested
		return fmt.Errorf("file not found: %s", envFile)
	}

	// Add/override with command-line key=value pairs
	if len(keyValues) > 0 {
		kvSecrets, err := vault.ParseKeyValuePairs(keyValues)
		if err != nil {
			return fmt.Errorf("failed to parse key=value pairs: %w", err)
		}
		fmt.Printf("Adding %d secrets from command line\n", len(kvSecrets))
		for k, v := range kvSecrets {
			secrets[k] = v
		}
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets to push. Use -f to specify a file or -k to add key=value pairs")
	}

	// Check if master key exists or generate new one
	masterKey, isNew, err := getOrCreateMasterKey(vaultPath)
	if err != nil {
		return err
	}

	if isNew {
		fmt.Println("✓ Generated new master key")

		// Get current machine info for GrantedBy
		machineInfo, err := vault.LoadMachineInfo()
		if err != nil {
			return fmt.Errorf("failed to load machine info: %w", err)
		}

		// Wrap master key for all machines
		fmt.Println("Wrapping master key for all machines...")
		if err := vault.WrapMasterKeyForMachines(vaultPath, masterKey, machineInfo.ID); err != nil {
			return fmt.Errorf("failed to wrap master key: %w", err)
		}
		fmt.Println("✓ Master key wrapped for all machines")
	} else {
		fmt.Println("✓ Using existing master key")
	}

	// Encrypt and save each secret
	fmt.Printf("Encrypting %d secrets for environment '%s'...\n", len(secrets), environment)
	for key, value := range secrets {
		encrypted, err := vault.EncryptSecret(masterKey, value)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret %s: %w", key, err)
		}

		if err := vault.SaveEncryptedSecret(paths, environment, key, encrypted); err != nil {
			return fmt.Errorf("failed to save secret %s: %w", key, err)
		}
	}

	fmt.Printf("\n✓ Successfully pushed %d secrets to vault\n", len(secrets))
	fmt.Printf("  Environment: %s\n", environment)
	fmt.Printf("  Vault: %s\n", vaultPath)
	fmt.Println("\nSecrets encrypted:")
	for key := range secrets {
		fmt.Printf("  - %s\n", key)
	}

	return nil
}

// getOrCreateMasterKey gets the existing master key or creates a new one
func getOrCreateMasterKey(vaultPath string) ([]byte, bool, error) {
	// Try to unwrap existing master key
	masterKey, err := vault.UnwrapMasterKey(vaultPath)
	if err == nil {
		return masterKey, false, nil
	}

	// Generate new master key
	masterKey, err = crypto.GenerateAESKey()
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate master key: %w", err)
	}

	return masterKey, true, nil
}

func init() {
	pushCmd.Flags().StringP("file", "f", "", "Environment file to encrypt")
	pushCmd.Flags().StringP("env", "e", "default", "Environment name")
	pushCmd.Flags().StringP("project", "p", "", "Project name (auto-detected if not specified)")
	pushCmd.Flags().StringSliceP("key", "k", []string{}, "Key=value pairs (can be specified multiple times)")
	rootCmd.AddCommand(pushCmd)
}
