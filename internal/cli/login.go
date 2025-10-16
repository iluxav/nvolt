package cli

import (
	"encoding/json"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with nvolt.io using Google OAuth or private key",
	Long: `Authenticate with nvolt.io using Google OAuth (browser) or silent login (CI/CD).

For interactive login (default):
  nvolt login

For silent login (CI/CD):
  nvolt login --silent --machine ci-runner-prod

Silent login requires ~/.nvolt/private_key.pem file to exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		silent, _ := cmd.Flags().GetBool("silent")
		machineName, _ := cmd.Flags().GetString("machine")

		machineConfig := cmd.Context().Value("machine_config").(*services.MachineConfig)

		if silent {
			if machineName == "" {
				return fmt.Errorf("--machine flag is required with --silent")
			}
			return runSilentLogin(machineConfig, machineName)
		}

		return runLogin(machineConfig)
	},
}

func init() {
	loginCmd.Flags().BoolP("silent", "s", false, "Silent login using private key (for CI/CD)")
	loginCmd.Flags().StringP("machine", "m", "", "Machine name for silent login")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(machineConfig *services.MachineConfig) error {
	fmt.Println("Logging in...")

	loginURL := fmt.Sprintf("%s/login?machine_id=%s", helpers.GetEnv("SERVER_BASE_URL", "https://nvolt.io"), machineConfig.Config.MachineID)
	if err := browser.OpenURL(loginURL); err != nil {
		fmt.Println(warnStyle.Render("⚠  Failed to open browser automatically"))
		fmt.Println(infoStyle.Render("→ Please manually open: " + loginURL))
	}

	fmt.Println("Waiting for authentication...")
	err := pollForToken(machineConfig)
	if err != nil {
		return fmt.Errorf("failed to poll for token: %w", err)
	}

	if machineConfig.Config.JWT_Token == "" {
		return fmt.Errorf("authentication failed")
	}

	// save machine key to server
	err = machineConfig.SaveMachineConfigToServer()
	if err != nil {
		return fmt.Errorf("failed to save machine key: %w", err)
	}

	return nil
}

func pollForToken(machineConfig *services.MachineConfig) error {
	serverURL := helpers.GetEnv("SERVER_BASE_URL", "https://nvolt.io")
	pollURL := fmt.Sprintf("%s/auth/poll?machine_id=%s", serverURL, machineConfig.Config.MachineID)
	client := &http.Client{}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(pollURL)
		if err != nil {
			fmt.Println(warnStyle.Render(fmt.Sprintf("⚠  Poll error: %v", err)))
			time.Sleep(2 * time.Second)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		if result["status"] == "success" {
			token, ok := result["token"].(string)
			if !ok {
				return fmt.Errorf("invalid token format")
			}

			if err := machineConfig.SaveJWT(token); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			fmt.Println(successStyle.Render("\n✓ Successfully authenticated!\n"))
			return nil
		}

		fmt.Printf(".")
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("authentication timeout")
}

func runSilentLogin(machineConfig *services.MachineConfig, machineName string) error {
	fmt.Println(titleStyle.Render(fmt.Sprintf("🔐 Silent login for machine: %s", machineName)))

	// Step 1: Load private key from file

	nvoltConf := helpers.GetEnv("NVOLT_CONF", ".nvolt")
	privateKeyPath := fmt.Sprintf("%s/%s/private_key.pem", os.Getenv("HOME"), nvoltConf)
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key from %s: %w\nPlease ensure the private key exists", privateKeyPath, err)
	}

	// Step 2: Request challenge from server
	authClient := services.NewAuthClient(machineConfig.Config)
	challenge, challengeID, err := authClient.RequestChallenge(machineName)
	if err != nil {
		return fmt.Errorf("failed to request challenge: %w", err)
	}

	// Step 3: Decrypt challenge and sign it
	signature, err := authClient.SignChallenge(string(privateKeyBytes), challenge)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Step 4: Verify signature and get JWT
	token, err := authClient.VerifySignature(machineName, challengeID, signature)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Step 5: Save JWT, machine ID, and private key
	machineConfig.Config.MachineID = machineName

	if err := machineConfig.SaveJWT(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Note: Private key file should already exist from step 1 (it was read from disk)
	// No need to save it again as it's already at ~/.nvolt/private_key.pem

	fmt.Println(successStyle.Render("✓ Successfully authenticated!"))
	return nil
}
