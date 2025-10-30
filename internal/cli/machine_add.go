package cli

import (
	"fmt"
	"iluxav/nvolt/internal/crypto"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var machineAddCmd = &cobra.Command{
	Use:   "add [machine-name]",
	Short: "Generate keys for a new machine (for CI/CD)",
	Long: `Generate RSA key pair for an additional machine and save public key to server.

This command is used to provision CI/CD machines that cannot use browser-based login.
The private key is automatically saved to [machine-name]_key.pem in the current directory.

Example:
  nvolt machine add ci-runner-prod
  # Creates: ci-runner-prod_key.pem

Then securely transfer the key file to the destination machine.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		machineName := args[0]

		machineConfig := services.MachineConfigFromContext(cmd.Context())
		secretsClient := services.SecretsClientFromContext(cmd.Context())

		activeOrgID := machineConfig.Config.ActiveOrgID
		if activeOrgID == "" {
			return fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
		}

		return runMachineAdd(machineConfig, secretsClient, machineName, activeOrgID)
	},
}

func runMachineAdd(machineConfig *services.MachineConfig, secretsClient *services.SecretsClient, machineName string, orgID string) error {
	fmt.Println(titleStyle.Render(fmt.Sprintf("🔑 Generating keys for machine: %s", machineName)))

	// Step 1: Generate RSA key pair
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Step 2: Save public key to server
	fmt.Println(infoStyle.Render("→ Saving public key to server..."))

	machineService := services.NewMachineService(machineConfig.Config)
	err = machineService.SaveMachineKey(&types.SaveMachinePublicKeyRequestDTO{
		MachineID: machineName,
		Name:      machineName,
		PublicKey: keyPair.PublicKey,
	})
	if err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	fmt.Println(successStyle.Render("✓ Public key saved to server"))

	// Step 3: Save private key to file automatically
	keyFileName := fmt.Sprintf("%s_key.pem", machineName)
	err = os.WriteFile(keyFileName, []byte(keyPair.PrivateKey), 0600)
	if err != nil {
		return fmt.Errorf("failed to save private key to %s: %w", keyFileName, err)
	}

	absPath, _ := filepath.Abs(keyFileName)
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Private key saved to: %s", absPath)))
	fmt.Println(infoStyle.Render("→ File permissions set to 600 (owner read/write only)"))

	// Display instructions
	fmt.Println("\n" + titleStyle.Render("📋 Next Steps"))
	fmt.Println(infoStyle.Render("1. Securely transfer the key file to the destination machine:"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   scp %s user@destination:~/.nvolt/private_key.pem", keyFileName)))
	fmt.Println()
	fmt.Println(infoStyle.Render("   OR use pbcopy to copy and paste manually:"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   cat %s | pbcopy", keyFileName)))
	fmt.Println(infoStyle.Render("   # Then on destination machine:"))
	fmt.Println(infoStyle.Render("   pbpaste > ~/.nvolt/private_key.pem && chmod 600 ~/.nvolt/private_key.pem"))
	fmt.Println()
	fmt.Println(infoStyle.Render("2. On the destination machine, authenticate:"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("   nvolt login --silent --machine %s", machineName)))
	fmt.Println()
	fmt.Println(warnStyle.Render(fmt.Sprintf("⚠️  Remember to delete the local key file after transfer: rm %s", keyFileName)))

	// Step 4: Automatically sync keys for all projects/environments
	fmt.Println("\n" + titleStyle.Render("🔄 Syncing keys for all machines..."))

	// Get all project/environment combinations
	projectEnvs, err := secretsClient.GetProjectEnvironments(orgID)
	if err != nil {
		fmt.Println(warnStyle.Render(fmt.Sprintf("⚠ Warning: Could not sync keys: %v", err)))
		fmt.Println(infoStyle.Render("→ You can manually sync later with: nvolt sync --all"))
		return nil
	}

	if len(projectEnvs) == 0 {
		fmt.Println(infoStyle.Render("→ No projects/environments to sync yet"))
		return nil
	}

	// Sync each project/environment
	successCount := 0
	for _, pe := range projectEnvs {
		err := secretsClient.SyncKeys(orgID, pe.ProjectName, pe.Environment)
		if err == nil {
			successCount++
		}
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Synced %d/%d project/environment combination(s)", successCount, len(projectEnvs))))
	fmt.Println(successStyle.Render(fmt.Sprintf("✓ Machine '%s' is ready! All machines can now access secrets.", machineName)))

	return nil
}

func init() {
	machineCmd.AddCommand(machineAddCmd)
}
