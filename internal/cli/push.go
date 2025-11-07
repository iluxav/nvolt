package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/iluxav/nvolt/internal/config"
	"github.com/iluxav/nvolt/internal/crypto"
	"github.com/iluxav/nvolt/internal/git"
	"github.com/iluxav/nvolt/internal/vault"
	"github.com/iluxav/nvolt/pkg/types"
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
		autoGrant, _ := cmd.Flags().GetBool("auto-grant")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		return runPush(envFile, environment, project, keyValues, autoGrant, dryRun)
	},
}

func runPush(envFile, environment, project string, keyValues []string, autoGrant, dryRun bool) error {
	if dryRun {
		fmt.Println("[DRY RUN] Simulating push operation...")
	} else {
		fmt.Println("Pushing secrets to vault...")
	}

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Determine project name and get vault paths
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)

		// Pull latest changes BEFORE doing any work
		fmt.Println("Global mode: pulling latest changes...")
		if err := git.SafePull(repoPath); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}
		fmt.Println("✓ Pulled latest changes from repository")

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

	// Get vault paths with unified logic (projectName is ignored in local mode)
	paths := vault.GetVaultPaths(vaultPath, project)

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
	masterKey, isNew, err := getOrCreateMasterKey(paths, environment)
	if err != nil {
		return err
	}
	// Ensure master key is cleared from memory when done
	defer crypto.ZeroBytes(masterKey)

	if isNew {
		fmt.Println("✓ Generated new master key")
	} else {
		fmt.Println("✓ Using existing master key")
	}

	// Get current machine info for GrantedBy
	machineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		return fmt.Errorf("failed to load machine info: %w", err)
	}

	// ALWAYS wrap master key for all machines (with permission control per environment)
	// This ensures newly added machines can request access to specific environments
	if autoGrant {
		fmt.Println("Wrapping master key for all machines (auto-granting access)...")
	} else {
		fmt.Println("Wrapping master key for machines (will prompt for new machines)...")
	}
	if err := wrapMasterKeyForMachines(paths, environment, masterKey, machineInfo.ID, autoGrant); err != nil {
		return fmt.Errorf("failed to wrap master key: %w", err)
	}
	fmt.Println("✓ Master key wrapping complete")

	// Encrypt and save each secret
	fmt.Printf("Encrypting %d secrets for environment '%s'...\n", len(secrets), environment)
	for key, value := range secrets {
		encrypted, err := vault.EncryptSecret(masterKey, value)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret %s: %w", key, err)
		}

		if !dryRun {
			if err := vault.SaveEncryptedSecret(paths, environment, key, encrypted); err != nil {
				return fmt.Errorf("failed to save secret %s: %w", key, err)
			}
		} else {
			fmt.Printf("  [DRY RUN] Would save secret: %s\n", key)
		}
	}

	fmt.Printf("\n✓ Successfully pushed %d secrets to vault\n", len(secrets))
	fmt.Printf("  Environment: %s\n", environment)
	fmt.Printf("  Vault: %s\n", vaultPath)
	fmt.Println("\nSecrets encrypted:")
	for key := range secrets {
		fmt.Printf("  - %s\n", key)
	}

	// Auto-commit and push in global mode
	if vault.IsGlobalMode(vaultPath) && !dryRun {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("\nGlobal mode: committing and pushing changes...")

		// Generate commit message
		commitMsg := fmt.Sprintf("Update secrets for project '%s' environment '%s'", project, environment)

		// Commit and push project directory (not .nvolt which doesn't exist in global mode)
		if err := git.CommitAndPush(repoPath, commitMsg, project, "machines"); err != nil {
			return fmt.Errorf("failed to commit and push changes: %w", err)
		}

		fmt.Println("✓ Changes committed and pushed to repository")
	} else if vault.IsGlobalMode(vaultPath) && dryRun {
		fmt.Println("\n[DRY RUN] Would commit and push changes to repository")
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] No changes were made")
	}

	return nil
}

// getOrCreateMasterKey gets the existing master key or creates a new one for the specified environment
func getOrCreateMasterKey(paths *vault.Paths, environment string) ([]byte, bool, error) {
	// Try to unwrap existing master key
	masterKey, err := unwrapMasterKey(paths, environment)
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

// unwrapMasterKey attempts to load and unwrap the master key for the current machine in a specific environment
func unwrapMasterKey(paths *vault.Paths, environment string) ([]byte, error) {
	// Get current machine ID
	machineID, err := vault.GetCurrentMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current machine ID: %w", err)
	}

	// Load wrapped key
	wrappedKeyPath := paths.GetWrappedKeyPath(environment, machineID)
	data, err := vault.ReadFile(wrappedKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wrapped key: %w (machine may not have access to '%s' environment)", err, environment)
	}

	var wrappedKeyData types.WrappedKey
	if err := json.Unmarshal(data, &wrappedKeyData); err != nil {
		return nil, fmt.Errorf("failed to parse wrapped key: %w", err)
	}

	// Load private key
	privateKey, err := vault.LoadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Unwrap master key
	wrappedKey, err := base64.StdEncoding.DecodeString(wrappedKeyData.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wrapped key: %w", err)
	}

	masterKey, err := crypto.UnwrapKey(privateKey, wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap master key: %w", err)
	}

	return masterKey, nil
}

// wrapMasterKeyForMachines wraps the master key for all machines in the vault
// This is a wrapper around vault.WrapMasterKeyForMachines
func wrapMasterKeyForMachines(paths *vault.Paths, environment string, masterKey []byte, grantedBy string, autoGrant bool) error {
	return vault.WrapMasterKeyForMachines(paths, environment, masterKey, grantedBy, autoGrant)
}

func init() {
	pushCmd.Flags().StringP("file", "f", "", "Environment file to encrypt")
	pushCmd.Flags().StringP("env", "e", "default", "Environment name")
	pushCmd.Flags().StringP("project", "p", "", "Project name (auto-detected if not specified)")
	pushCmd.Flags().StringSliceP("key", "k", []string{}, "Key=value pairs (can be specified multiple times)")
	pushCmd.Flags().Bool("auto-grant", false, "Automatically grant access to all machines without prompting")
	pushCmd.Flags().Bool("dry-run", false, "Show what would be done without making any changes")
	rootCmd.AddCommand(pushCmd)
}
