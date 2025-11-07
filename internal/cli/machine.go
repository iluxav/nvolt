package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nvolt/nvolt/internal/crypto"
	"github.com/nvolt/nvolt/internal/git"
	"github.com/nvolt/nvolt/internal/vault"
	"github.com/nvolt/nvolt/pkg/types"
	"github.com/spf13/cobra"
)

var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage machine access and keys",
	Long:  `Add or remove machines from the vault access list.`,
}

var machineAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new machine and generate its keypair",
	Long: `Generate a new keypair for a machine (CI/CD or another device).

Example:
  nvolt machine add ci-server
  nvolt machine add alice-laptop`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]
		return runMachineAdd(machineName)
	},
}

var machineRmCmd = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a machine and revoke its access",
	Long: `Revoke access for a machine by removing its keys and re-wrapping master key.

Example:
  nvolt machine rm ci-server
  nvolt machine rm alice-laptop`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]
		return runMachineRm(machineName)
	},
}

var machineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all machines with access",
	Long:  `Display all machines that have access to the vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMachineList()
	},
}

func runMachineAdd(machineName string) error {
	fmt.Printf("Adding machine: %s\n", machineName)

	// Find vault path (local or global)
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Pull latest changes in global mode BEFORE doing any work
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("Global mode: pulling latest changes...")
		if err := git.SafePull(repoPath); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}
		fmt.Println("✓ Pulled latest changes from repository")
	}

	// Generate keypair for new machine
	fmt.Println("Generating keypair...")
	privateKey, err := crypto.GenerateRSAKeypair()
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	publicKey := &privateKey.PublicKey

	// Encode keys
	privateKeyPEM, err := crypto.EncodePrivateKeyPEM(privateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := crypto.EncodePublicKeyPEM(publicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	fingerprint, err := crypto.GenerateFingerprint(publicKey)
	if err != nil {
		return fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	// Create machine info
	machineID := vault.GenerateMachineID(machineName, fingerprint)
	machineInfo := &types.MachineInfo{
		ID:          machineID,
		PublicKey:   string(publicKeyPEM),
		Fingerprint: fingerprint,
		Hostname:    machineName,
		Description: fmt.Sprintf("Machine: %s", machineName),
		CreatedAt:   time.Now(),
	}

	// Add to vault
	if err := vault.AddMachineToVault(vaultPath, machineInfo); err != nil {
		return fmt.Errorf("failed to add machine to vault: %w", err)
	}

	fmt.Printf("\n✓ Machine added successfully\n")
	fmt.Printf("  Machine ID: %s\n", machineID)
	fmt.Printf("  Fingerprint: %s\n", fingerprint)
	fmt.Printf("\nPrivate key (save this securely for the new machine):\n")
	fmt.Printf("---\n%s---\n", string(privateKeyPEM))
	fmt.Printf("\nTo use this machine:\n")
	fmt.Printf("1. Save the private key to ~/.nvolt/private_key.pem on the target machine\n")
	fmt.Printf("2. Set permissions: chmod 600 ~/.nvolt/private_key.pem\n")
	fmt.Printf("3. Save the machine info to ~/.nvolt/machines/machine-info.json\n")

	// Auto-commit and push in global mode
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("\nGlobal mode: committing and pushing changes...")

		commitMsg := fmt.Sprintf("Add machine: %s", machineName)
		if err := git.CommitAndPush(repoPath, commitMsg, ".nvolt"); err != nil {
			return fmt.Errorf("failed to commit and push changes: %w", err)
		}

		fmt.Println("✓ Changes committed and pushed to repository")
	}

	return nil
}

func runMachineRm(machineName string) error {
	fmt.Printf("Removing machine: %s\n", machineName)

	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// Pull latest changes in global mode BEFORE doing any work
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("Global mode: pulling latest changes...")
		if err := git.SafePull(repoPath); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}
		fmt.Println("✓ Pulled latest changes from repository")
	}

	// List all machines and find matching ones
	machines, err := vault.ListMachines(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to list machines: %w", err)
	}

	// Find machines matching the hostname or ID
	var matchingMachines []*types.MachineInfo
	for _, m := range machines {
		if m.Hostname == machineName || m.ID == machineName {
			matchingMachines = append(matchingMachines, m)
		}
	}

	if len(matchingMachines) == 0 {
		return fmt.Errorf("no machine found with name or ID: %s", machineName)
	}

	var machineID string
	if len(matchingMachines) == 1 {
		machineID = matchingMachines[0].ID
	} else {
		// Multiple matches - let user choose
		fmt.Println("\nMultiple machines found:")
		for i, m := range matchingMachines {
			fmt.Printf("%d. %s (Fingerprint: %s)\n", i+1, m.ID, m.Fingerprint)
		}
		fmt.Print("\nEnter number to remove: ")
		var choice int
		fmt.Scanln(&choice)
		if choice < 1 || choice > len(matchingMachines) {
			return fmt.Errorf("invalid choice")
		}
		machineID = matchingMachines[choice-1].ID
	}

	// Confirm removal
	fmt.Printf("Are you sure you want to remove machine %s? This cannot be undone. (yes/no): ", machineID)
	var response string
	fmt.Scanln(&response)
	if response != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	// Remove machine
	if err := vault.RemoveMachineFromVault(vaultPath, machineID); err != nil {
		return fmt.Errorf("failed to remove machine: %w", err)
	}

	fmt.Printf("✓ Machine %s removed successfully\n", machineID)
	fmt.Println("\nNote: You should re-wrap the master key using 'nvolt sync' to ensure")
	fmt.Println("the removed machine cannot decrypt new secrets.")

	// Auto-commit and push in global mode
	if vault.IsGlobalMode(vaultPath) {
		repoPath := vault.GetRepoPathFromVault(vaultPath)
		fmt.Println("\nGlobal mode: committing and pushing changes...")

		commitMsg := fmt.Sprintf("Remove machine: %s", machineID)
		if err := git.CommitAndPush(repoPath, commitMsg, ".nvolt"); err != nil {
			return fmt.Errorf("failed to commit and push changes: %w", err)
		}

		fmt.Println("✓ Changes committed and pushed to repository")
	}

	return nil
}

func runMachineList() error {
	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// List machines
	machines, err := vault.ListMachines(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to list machines: %w", err)
	}

	if len(machines) == 0 {
		fmt.Println("No machines found in vault.")
		return nil
	}

	fmt.Printf("Machines (%d):\n\n", len(machines))
	for _, m := range machines {
		fmt.Printf("  ID:          %s\n", m.ID)
		fmt.Printf("  Hostname:    %s\n", m.Hostname)
		fmt.Printf("  Fingerprint: %s\n", m.Fingerprint)
		fmt.Printf("  Created:     %s\n", m.CreatedAt.Format(time.RFC3339))
		if m.Description != "" {
			fmt.Printf("  Description: %s\n", m.Description)
		}
		fmt.Println()
	}

	return nil
}

// findVaultPath tries to find the vault in local or global mode
func findVaultPath() (string, error) {
	// Try local mode first
	localPath, err := vault.GetLocalVaultPath()
	if err == nil && vault.IsVaultInitialized(localPath) {
		return localPath, nil
	}

	// Try to find global vault in ~/.nvolt/projects
	homePaths, err := vault.GetHomePaths()
	if err != nil {
		return "", fmt.Errorf("vault not found. Run 'nvolt init' first")
	}

	// Check if projects directory exists
	if !vault.FileExists(homePaths.Projects) {
		return "", fmt.Errorf("vault not found. Run 'nvolt init' first")
	}

	// Scan for vaults in ~/.nvolt/projects/org/repo/.nvolt
	vaultPath, err := findGlobalVault(homePaths.Projects)
	if err != nil {
		return "", fmt.Errorf("vault not found. Run 'nvolt init' first")
	}

	return vaultPath, nil
}

// findGlobalVault searches for a vault in ~/.nvolt/projects/
// Returns the first valid vault found in the structure: ~/.nvolt/projects/org/repo/.nvolt
func findGlobalVault(projectsDir string) (string, error) {
	// List all org directories in ~/.nvolt/projects
	orgDirs, err := vault.ListDirs(projectsDir)
	if err != nil {
		return "", fmt.Errorf("no global vaults found")
	}

	// Scan each org directory for repos
	for _, orgDir := range orgDirs {
		repoDirs, err := vault.ListDirs(orgDir)
		if err != nil {
			continue
		}

		// Check each repo for a .nvolt directory
		for _, repoDir := range repoDirs {
			vaultPath := filepath.Join(repoDir, ".nvolt")
			if vault.IsVaultInitialized(vaultPath) {
				return vaultPath, nil
			}
		}
	}

	return "", fmt.Errorf("no global vaults found")
}

func init() {
	machineCmd.AddCommand(machineAddCmd)
	machineCmd.AddCommand(machineRmCmd)
	machineCmd.AddCommand(machineListCmd)
	rootCmd.AddCommand(machineCmd)
}
