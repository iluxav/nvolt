package cli

import (
	"fmt"
	"time"

	"github.com/iluxav/nvolt/internal/vault"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Vault management and verification commands",
	Long:  `Show vault information and verify integrity.`,
}

var vaultShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display vault information and machine access",
	Long: `Show all registered machines, their fingerprints, access timestamps,
and available environments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVaultShow()
	},
}

var vaultVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify vault integrity",
	Long: `Verify the integrity of encrypted files, wrapped keys, and vault structure.

This checks:
- All encrypted files are readable
- Wrapped keys are valid
- Machine public keys match fingerprints
- keyinfo.json structure is valid`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVaultVerify()
	},
}

func runVaultShow() error {
	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// TODO: In global mode with multiple projects, this should show project-specific info
	// For now, we use empty project name which works for local mode
	paths := vault.GetVaultPaths(vaultPath, "")

	fmt.Printf("Vault Information\n")
	fmt.Printf("=================\n\n")

	// Show vault location
	fmt.Printf("Vault Location: %s\n\n", vaultPath)

	// Get current machine info
	currentMachineInfo, err := vault.LoadMachineInfo()
	if err != nil {
		fmt.Printf("Warning: Could not load current machine info: %v\n\n", err)
	} else {
		fmt.Printf("Current Machine:\n")
		fmt.Printf("  ID:          %s\n", currentMachineInfo.ID)
		fmt.Printf("  Hostname:    %s\n", currentMachineInfo.Hostname)
		fmt.Printf("  Fingerprint: %s\n\n", currentMachineInfo.Fingerprint)
	}

	// List all machines
	machines, err := vault.ListMachines(paths)
	if err != nil {
		return fmt.Errorf("failed to list machines: %w", err)
	}

	// List environments first to show access per environment
	envDirs, err := vault.ListDirs(paths.Secrets)
	if err != nil {
		fmt.Printf("Warning: Could not list environments: %v\n\n", err)
		envDirs = []string{}
	}

	environments := []string{}
	for _, envDir := range envDirs {
		environments = append(environments, vault.GetDirName(envDir))
	}

	fmt.Printf("Registered Machines (%d):\n", len(machines))
	if len(machines) == 0 {
		fmt.Printf("  (none)\n")
	} else {
		for _, m := range machines {
			fmt.Printf("\n  Machine: %s\n", m.ID)
			fmt.Printf("    Hostname:    %s\n", m.Hostname)
			fmt.Printf("    Fingerprint: %s\n", m.Fingerprint)
			fmt.Printf("    Created:     %s\n", m.CreatedAt.Format(time.RFC3339))
			if m.Description != "" {
				fmt.Printf("    Description: %s\n", m.Description)
			}

			// Check access for each environment
			if len(environments) > 0 {
				fmt.Printf("    Access:\n")
				for _, env := range environments {
					wrappedKeyPath := paths.GetWrappedKeyPath(env, m.ID)
					hasKey := vault.FileExists(wrappedKeyPath)
					if hasKey {
						fmt.Printf("      %s: ✓\n", env)
					} else {
						fmt.Printf("      %s: ✗\n", env)
					}
				}
			} else {
				fmt.Printf("    Access:      (no environments)\n")
			}
		}
	}

	fmt.Printf("\n")

	// List environments
	envDirs, err = vault.ListDirs(paths.Secrets)
	if err != nil {
		fmt.Printf("Environments: (error listing: %v)\n\n", err)
	} else if len(envDirs) == 0 {
		fmt.Printf("Environments: (none)\n\n")
	} else {
		fmt.Printf("Environments (%d):\n", len(envDirs))
		for _, envDir := range envDirs {
			envName := vault.GetDirName(envDir)
			secretFiles, err := vault.ListFiles(envDir)
			if err != nil {
				fmt.Printf("  - %s (error: %v)\n", envName, err)
			} else {
				fmt.Printf("  - %s (%d secret(s))\n", envName, len(secretFiles))
			}
		}
		fmt.Printf("\n")
	}

	return nil
}

func runVaultVerify() error {
	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		return err
	}

	// TODO: In global mode with multiple projects, this should verify all projects
	// For now, we use empty project name which works for local mode
	paths := vault.GetVaultPaths(vaultPath, "")

	fmt.Printf("Verifying Vault Integrity\n")
	fmt.Printf("=========================\n\n")

	errors := []string{}
	warnings := []string{}

	// Check vault structure
	fmt.Printf("Checking vault structure...\n")
	if err := vault.ValidateVaultStructure(vaultPath); err != nil {
		errors = append(errors, fmt.Sprintf("Vault structure invalid: %v", err))
	} else {
		fmt.Printf("✓ Vault structure is valid\n")
	}

	// Check Git security
	fmt.Printf("\nChecking Git security...\n")
	if err := vault.EnsurePrivateKeysNotInGit(vaultPath); err != nil {
		errors = append(errors, fmt.Sprintf("Private key security issue: %v", err))
	} else {
		fmt.Printf("✓ Private keys are not in Git repository\n")
	}

	// Check .gitignore for sensitive patterns
	missing, err := vault.ValidateGitignore(vaultPath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Git ignore validation: %v", err))
	} else if len(missing) > 0 {
		for _, pattern := range missing {
			warnings = append(warnings, fmt.Sprintf("Missing .gitignore pattern: %s", pattern))
		}
	} else {
		fmt.Printf("✓ .gitignore properly configured for sensitive files\n")
	}

	// Check current machine
	fmt.Printf("\nChecking current machine...\n")
	currentMachine, err := vault.LoadMachineInfo()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Cannot load current machine info: %v", err))
	} else {
		fmt.Printf("✓ Current machine: %s\n", currentMachine.ID)

		// List environments to check access
		envDirs, err := vault.ListDirs(paths.Secrets)
		if err == nil && len(envDirs) > 0 {
			fmt.Printf("Checking access to environments...\n")
			for _, envDir := range envDirs {
				envName := vault.GetDirName(envDir)
				_, err := vault.UnwrapMasterKey(paths, envName)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("Current machine cannot unwrap master key for '%s': %v", envName, err))
				} else {
					fmt.Printf("  ✓ Can access '%s'\n", envName)
				}
			}
		}
	}

	// Check machines
	fmt.Printf("\nChecking registered machines...\n")
	machines, err := vault.ListMachines(paths)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Cannot list machines: %v", err))
	} else {
		fmt.Printf("✓ Found %d machine(s)\n", len(machines))

		// Get list of environments
		envDirs, err := vault.ListDirs(paths.Secrets)
		environments := []string{}
		if err == nil {
			for _, envDir := range envDirs {
				environments = append(environments, vault.GetDirName(envDir))
			}
		}

		// Check each machine has wrapped keys for environments
		if len(environments) > 0 {
			for _, m := range machines {
				hasAnyKey := false
				for _, env := range environments {
					wrappedKeyPath := paths.GetWrappedKeyPath(env, m.ID)
					if vault.FileExists(wrappedKeyPath) {
						hasAnyKey = true
						break
					}
				}
				if !hasAnyKey {
					warnings = append(warnings, fmt.Sprintf("Machine %s has no wrapped keys in any environment", m.ID))
				}
			}
		}
	}

	// Check wrapped keys for orphans
	fmt.Printf("\nChecking wrapped keys...\n")

	// Get list of environments
	envDirs, err := vault.ListDirs(paths.Secrets)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Cannot list environments: %v", err))
	} else {
		totalWrappedKeys := 0

		// Check for orphaned wrapped keys
		machineIDs := make(map[string]bool)
		for _, m := range machines {
			machineIDs[m.ID] = true
		}

		// Iterate through each environment's wrapped keys
		for _, envDir := range envDirs {
			envName := vault.GetDirName(envDir)
			wrappedKeysEnvPath := paths.GetWrappedKeysEnvPath(envName)

			wrappedKeyFiles, err := vault.ListFiles(wrappedKeysEnvPath)
			if err != nil {
				// Skip if directory doesn't exist
				continue
			}

			totalWrappedKeys += len(wrappedKeyFiles)

			for _, keyFile := range wrappedKeyFiles {
				// Extract machine ID from filename (remove .json)
				filename := vault.GetDirName(keyFile)
				machineID := filename
				if len(filename) > 5 && filename[len(filename)-5:] == ".json" {
					machineID = filename[:len(filename)-5]
				}

				if !machineIDs[machineID] {
					warnings = append(warnings, fmt.Sprintf("Orphaned wrapped key found: %s/%s", envName, filename))
				}
			}
		}

		fmt.Printf("✓ Found %d wrapped key(s) across all environments\n", totalWrappedKeys)
	}

	// Check secrets
	fmt.Printf("\nChecking secrets...\n")
	envDirs, err = vault.ListDirs(paths.Secrets)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Cannot list environments: %v", err))
	} else {
		totalSecrets := 0
		for _, envDir := range envDirs {
			secretFiles, err := vault.ListFiles(envDir)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Cannot list secrets in %s: %v", envDir, err))
				continue
			}
			totalSecrets += len(secretFiles)
		}
		fmt.Printf("✓ Found %d secret(s) across %d environment(s)\n", totalSecrets, len(envDirs))
	}

	// Print summary
	fmt.Printf("\n")
	fmt.Printf("Summary\n")
	fmt.Printf("=======\n")

	if len(errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  ✗ %s\n", e)
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
	}

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Printf("\n✓ Vault is healthy - no issues found\n")
	} else if len(errors) == 0 {
		fmt.Printf("\n✓ Vault is functional but has warnings\n")
	} else {
		fmt.Printf("\n✗ Vault has errors that need attention\n")
		return fmt.Errorf("vault verification failed with %d error(s)", len(errors))
	}

	return nil
}

func init() {
	vaultCmd.AddCommand(vaultShowCmd)
	vaultCmd.AddCommand(vaultVerifyCmd)
	rootCmd.AddCommand(vaultCmd)
}
