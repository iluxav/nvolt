package services

import (
	"encoding/json"
	"fmt"
	"iluxav/nvolt/internal/crypto"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
	"os"
)

type MachineConfig struct {
	Config      *types.MachineLocalConfig
	OrgUsers    []*types.OrgUser
	Project     string
	Environment string
}

func NewMachineConfigService() *MachineConfig {
	config := tryLoadConfig()
	return &MachineConfig{
		Config: config.Config,
	}
}

func (s *MachineConfig) SavePublicKey() {
	saveConfig(s.Config)
}

func (s *MachineConfig) SaveMachineConfigToServer() error {
	serverURL := helpers.GetEnv("SERVER_BASE_URL", "https://nvolt.io")

	// Step 1: Get user's organizations to determine which org to save the machine to
	var orgID string
	if s.Config.ActiveOrgID != "" {
		orgID = s.Config.ActiveOrgID
	} else {
		// Fetch user orgs and use the first one (for first-time login)
		userOrgsURL := fmt.Sprintf("%s/api/v1/user/orgs", serverURL)
		orgUsers, err := helpers.CallAPI[[]*types.OrgUser](userOrgsURL, "GET", s.Config.JWT_Token)
		if err != nil {
			return fmt.Errorf("failed to fetch user organizations: %w", err)
		}

		if orgUsers == nil || len(*orgUsers) == 0 {
			return fmt.Errorf("user has no organizations")
		}

		// Use the first org
		orgID = (*orgUsers)[0].OrgID.String()

		// Save it as active org for future use
		s.Config.ActiveOrgID = orgID
		saveConfig(s.Config)
	}

	// Step 2: Read private key from file
	privateKey, err := s.GetPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Step 3: Extract public key from private key
	publicKey, err := crypto.ExtractPublicKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	// Step 4: Save machine key to server with org_id in path
	saveMachineKeyURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", serverURL, orgID)

	requestBody := &types.SaveMachinePublicKeyRequestDTO{
		MachineID: s.Config.MachineID,
		Name:      helpers.GetLocalMachineName(),
		PublicKey: publicKey,
	}
	_, err = helpers.CallAPIWithPayload[types.SaveMachinePublicKeyResponseDTO, types.SaveMachinePublicKeyRequestDTO](saveMachineKeyURL, "POST", s.Config.JWT_Token, requestBody)
	if err != nil {
		return fmt.Errorf("failed to save machine key: %w", err)
	}
	return nil
}

// GetPrivateKey reads the private key from ~/.nvolt/private_key.pem
func (s *MachineConfig) GetPrivateKey() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := helpers.GetEnv("NVOLT_CONF", ".nvolt")
	privateKeyPath := fmt.Sprintf("%s/%s/private_key.pem", homeDir, configDir)

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key from %s: %w", privateKeyPath, err)
	}

	return string(privateKeyBytes), nil
}

// SavePrivateKey saves the private key to ~/.nvolt/private_key.pem with secure permissions
func (s *MachineConfig) SavePrivateKey(privateKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := helpers.GetEnv("NVOLT_CONF", ".nvolt")
	privateKeyPath := fmt.Sprintf("%s/%s/private_key.pem", homeDir, configDir)

	// Create directory if it doesn't exist
	dirPath := fmt.Sprintf("%s/%s", homeDir, configDir)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write private key with secure permissions (0600 = rw-------)
	if err := os.WriteFile(privateKeyPath, []byte(privateKey), 0600); err != nil {
		return fmt.Errorf("failed to write private key to %s: %w", privateKeyPath, err)
	}

	return nil
}

func tryLoadConfig() *MachineConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	var config = &MachineConfig{}
	configDir := fmt.Sprintf("%s/%s", homeDir, helpers.GetEnv("NVOLT_CONF", ".nvolt"))
	configPath := fmt.Sprintf("%s/config.json", configDir)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Generate new key pair for first-time setup
		keyPairs, err := crypto.GenerateKeyPair()
		if err != nil {
			fmt.Println("Error generating key pair:", err)
			os.Exit(1)
		}

		config.Config = &types.MachineLocalConfig{}
		config.Config.MachineID = helpers.GetLocalMachineID()
		config.Config.ServerURL = helpers.GetEnv("SERVER_BASE_URL", "https://nvolt.io")
		config.Config.JWT_Token = ""

		// Save private key to separate file with secure permissions
		if err := config.SavePrivateKey(keyPairs.PrivateKey); err != nil {
			fmt.Println("Error saving private key:", err)
			os.Exit(1)
		}

		saveConfig(config.Config)
		return config
	}

	jsonData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}

	err = json.Unmarshal(jsonData, &config.Config)
	if err != nil {
		fmt.Println("Error unmarshalling config:", err)
		os.Exit(1)
	}

	// Note: Private key is now stored in separate file ~/.nvolt/private_key.pem
	// It will be read on-demand when needed
	return config
}

func saveConfig(config *types.MachineLocalConfig) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	configDir := fmt.Sprintf("%s/%s", homeDir, helpers.GetEnv("NVOLT_CONF", ".nvolt"))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Println("Error creating config directory:", err)
		os.Exit(1)
	}

	configPath := fmt.Sprintf("%s/config.json", configDir)
	jsonData, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Error marshalling config:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		fmt.Printf("Error writing config file to %s: %v\n", configPath, err)
		os.Exit(1)
	}
}

func (s *MachineConfig) SaveJWT(token string) error {
	s.Config.JWT_Token = token
	saveConfig(s.Config)
	return nil
}

func (s *MachineConfig) TryResolveLocalDirProjectName() *types.ProjectResolver {
	projectResolver, err := helpers.TryGetProjectName()
	if err != nil {
		return nil
	}
	s.Project = projectResolver.ProjectName
	s.Environment = "default"
	return projectResolver
}

// SaveActiveOrg saves the active organization ID to the config file
func (s *MachineConfig) SaveActiveOrg(orgID string) error {
	s.Config.ActiveOrgID = orgID
	saveConfig(s.Config)
	return nil
}
