package cli

import (
	"fmt"
	"path/filepath"

	"github.com/nvolt/nvolt/internal/git"
	"github.com/nvolt/nvolt/internal/vault"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nvolt vault and machine keys",
	Long: `Initialize a new nvolt vault and generate machine keypair if needed.

Local mode (current directory):
  nvolt init

Global mode (dedicated GitHub repo):
  nvolt init --repo org/repo

This command will:
- Generate an RSA/Ed25519 keypair for this machine (if not exists)
- Create .nvolt/ directory structure
- Clone the repo (if --repo provided) into ~/.nvolt/projects/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")

		return runInit(repo)
	},
}

func runInit(repoSpec string) error {
	fmt.Println("Initializing nvolt vault...")

	// Step 1: Ensure machine keypair exists
	machineInitialized, err := vault.IsMachineInitialized()
	if err != nil {
		return fmt.Errorf("failed to check machine initialization: %w", err)
	}

	if !machineInitialized {
		fmt.Println("Generating machine keypair...")
		machineInfo, err := vault.InitializeMachine()
		if err != nil {
			return fmt.Errorf("failed to initialize machine: %w", err)
		}
		fmt.Printf("✓ Machine keypair generated\n")
		fmt.Printf("  Machine ID: %s\n", machineInfo.ID)
		fmt.Printf("  Fingerprint: %s\n", machineInfo.Fingerprint)
	} else {
		machineInfo, err := vault.LoadMachineInfo()
		if err != nil {
			return fmt.Errorf("failed to load machine info: %w", err)
		}
		fmt.Printf("✓ Using existing machine keypair\n")
		fmt.Printf("  Machine ID: %s\n", machineInfo.ID)
		fmt.Printf("  Fingerprint: %s\n", machineInfo.Fingerprint)
	}

	// Step 2: Determine mode and initialize vault
	if repoSpec != "" {
		// Global mode
		return initGlobalMode(repoSpec)
	}

	// Local mode
	return initLocalMode()
}

func initLocalMode() error {
	fmt.Println("\nMode: Local")

	vaultPath, err := vault.GetLocalVaultPath()
	if err != nil {
		return fmt.Errorf("failed to get local vault path: %w", err)
	}

	// Load current machine info
	machineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		return fmt.Errorf("failed to load machine info: %w", err)
	}

	// Check if vault already exists
	vaultExists := vault.IsVaultInitialized(vaultPath)
	if vaultExists {
		fmt.Printf("\n✓ Vault already initialized at %s\n", vaultPath)

		// Check if current machine is already in the vault
		paths := vault.GetVaultPaths(vaultPath)
		machinePath := paths.GetMachineInfoPath(machineInfo.ID)
		if vault.FileExists(machinePath) {
			fmt.Printf("✓ Machine %s already registered in vault\n", machineInfo.ID)
		} else {
			// Add current machine to existing vault
			if err := vault.AddMachineToVault(vaultPath, machineInfo); err != nil {
				return fmt.Errorf("failed to add machine to vault: %w", err)
			}
			fmt.Printf("✓ Added machine %s to vault\n", machineInfo.ID)
		}

		fmt.Println("\nYou can now:")
		fmt.Println("  - Push secrets: nvolt push -f .env")
		fmt.Println("  - Pull secrets: nvolt pull")
		fmt.Println("  - Add machines: nvolt machine add <name>")
		fmt.Println("\nNote: In local mode, you are responsible for Git operations.")
		return nil
	}

	// Initialize vault directory
	if err := vault.InitializeVaultDirectory(vaultPath); err != nil {
		return fmt.Errorf("failed to initialize vault directory: %w", err)
	}

	// Add current machine to vault
	if err := vault.AddMachineToVault(vaultPath, machineInfo); err != nil {
		return fmt.Errorf("failed to add machine to vault: %w", err)
	}

	fmt.Printf("✓ Vault initialized at %s\n", vaultPath)
	fmt.Println("\nYou can now:")
	fmt.Println("  - Push secrets: nvolt push -f .env")
	fmt.Println("  - Pull secrets: nvolt pull")
	fmt.Println("  - Add machines: nvolt machine add <name>")
	fmt.Println("\nNote: In local mode, you are responsible for Git operations.")

	return nil
}

func initGlobalMode(repoSpec string) error {
	fmt.Println("\nMode: Global")

	// Check if git is available
	if !git.IsGitAvailable() {
		return fmt.Errorf("git is not available in PATH")
	}

	// Parse repository
	org, repo, err := git.GetRepoPath(repoSpec)
	if err != nil {
		return err
	}

	fmt.Printf("Repository: %s/%s\n", org, repo)

	// Get home paths
	homePaths, err := vault.GetHomePaths()
	if err != nil {
		return err
	}

	// Determine target path
	repoPath := filepath.Join(homePaths.Projects, org, repo)
	vaultPath := filepath.Join(repoPath, vault.NvoltDir)

	// Load current machine info
	machineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		return fmt.Errorf("failed to load machine info: %w", err)
	}

	// Check if already cloned
	vaultExists := vault.IsVaultInitialized(vaultPath)
	if vaultExists {
		fmt.Printf("✓ Repository already cloned at %s\n", repoPath)

		// Check if current machine is already in the vault
		paths := vault.GetVaultPaths(vaultPath)
		machinePath := paths.GetMachineInfoPath(machineInfo.ID)
		machineExists := vault.FileExists(machinePath)

		if machineExists {
			fmt.Printf("✓ Machine %s already registered in vault\n", machineInfo.ID)
		} else {
			// Add current machine to existing vault
			if err := vault.AddMachineToVault(vaultPath, machineInfo); err != nil {
				return fmt.Errorf("failed to add machine to vault: %w", err)
			}
			fmt.Printf("✓ Added machine %s to vault\n", machineInfo.ID)
		}

		// Check if there are uncommitted changes (machine might exist but not be committed)
		hasChanges, err := git.HasUncommittedChanges(repoPath)
		if err == nil && hasChanges {
			// Commit and push the machine
			fmt.Println("\nCommitting machine to repository...")
			commitMsg := fmt.Sprintf("Add machine %s to vault", machineInfo.ID)
			if err := git.CommitAndPush(repoPath, commitMsg, ".nvolt"); err != nil {
				return fmt.Errorf("failed to commit and push machine: %w", err)
			}
			fmt.Println("✓ Machine committed and pushed to repository")
		}

		fmt.Println("\nYou can now:")
		fmt.Println("  - Push secrets: nvolt push -f .env")
		fmt.Println("  - Pull secrets: nvolt pull")
		fmt.Println("  - Add machines: nvolt machine add <name>")
		fmt.Println("\nNote: nvolt will automatically commit and push changes to the repository.")

		return nil
	}

	// Clone repository
	repoURL, err := git.ParseRepoURL(repoSpec)
	if err != nil {
		return err
	}

	fmt.Printf("Cloning repository...\n")
	if err := git.Clone(repoURL, repoPath); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("✓ Repository cloned to %s\n", repoPath)

	// Initialize vault if .nvolt doesn't exist in repo
	if !vault.IsVaultInitialized(vaultPath) {
		fmt.Println("Initializing vault structure in repository...")
		if err := vault.InitializeVaultDirectory(vaultPath); err != nil {
			return fmt.Errorf("failed to initialize vault directory: %w", err)
		}
	}

	// Add current machine to vault
	if err := vault.AddMachineToVault(vaultPath, machineInfo); err != nil {
		return fmt.Errorf("failed to add machine to vault: %w", err)
	}

	// Commit and push the machine's public key to repository
	fmt.Println("\nCommitting machine to repository...")
	commitMsg := fmt.Sprintf("Add machine %s to vault", machineInfo.ID)
	if err := git.CommitAndPush(repoPath, commitMsg, ".nvolt"); err != nil {
		return fmt.Errorf("failed to commit and push machine: %w", err)
	}
	fmt.Println("✓ Machine committed and pushed to repository")

	fmt.Println("\n✓ Global vault initialized successfully")
	fmt.Println("\nYou can now:")
	fmt.Println("  - Push secrets: nvolt push -f .env")
	fmt.Println("  - Pull secrets: nvolt pull")
	fmt.Println("  - Add machines: nvolt machine add <name>")
	fmt.Println("\nNote: nvolt will automatically commit and push changes to the repository.")

	return nil
}

func init() {
	initCmd.Flags().StringP("repo", "r", "", "GitHub repository (org/repo) for global mode")
	rootCmd.AddCommand(initCmd)
}
