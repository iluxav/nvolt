package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nvolt/nvolt/internal/crypto"
	"github.com/nvolt/nvolt/pkg/types"
)

// InitializeMachine creates a new machine keypair and stores it in ~/.nvolt
func InitializeMachine() (*types.MachineInfo, error) {
	homePaths, err := GetHomePaths()
	if err != nil {
		return nil, err
	}

	// Check if machine already initialized
	if FileExists(homePaths.PrivateKey) {
		return nil, fmt.Errorf("machine already initialized: %s exists", homePaths.PrivateKey)
	}

	// Ensure home directory exists
	if err := InitializeHomeDirectory(); err != nil {
		return nil, fmt.Errorf("failed to initialize home directory: %w", err)
	}

	// Generate keypair
	privateKey, err := crypto.GenerateRSAKeypair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	publicKey := &privateKey.PublicKey

	// Encode private key
	privateKeyPEM, err := crypto.EncodePrivateKeyPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	// Encode public key
	publicKeyPEM, err := crypto.EncodePublicKeyPEM(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	// Generate fingerprint
	fingerprint, err := crypto.GenerateFingerprint(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create machine info
	machineInfo := &types.MachineInfo{
		ID:          GenerateMachineID(hostname),
		PublicKey:   string(publicKeyPEM),
		Fingerprint: fingerprint,
		Hostname:    hostname,
		Description: fmt.Sprintf("Machine: %s", hostname),
		CreatedAt:   time.Now(),
	}

	// Save private key with restricted permissions
	if err := WriteFileAtomic(homePaths.PrivateKey, privateKeyPEM, PrivateKeyPerm); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	// Save machine info
	if err := SaveMachineInfo(homePaths.MachineInfo, machineInfo); err != nil {
		// Clean up private key on error
		DeleteFile(homePaths.PrivateKey)
		return nil, fmt.Errorf("failed to save machine info: %w", err)
	}

	return machineInfo, nil
}

// LoadMachineInfo loads the current machine's info from ~/.nvolt
func LoadMachineInfo() (*types.MachineInfo, error) {
	homePaths, err := GetHomePaths()
	if err != nil {
		return nil, err
	}

	return LoadMachineInfoFromFile(homePaths.MachineInfo)
}

// LoadMachineInfoFromFile loads machine info from a specific file
func LoadMachineInfoFromFile(path string) (*types.MachineInfo, error) {
	data, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read machine info: %w", err)
	}

	var machineInfo types.MachineInfo
	if err := json.Unmarshal(data, &machineInfo); err != nil {
		return nil, fmt.Errorf("failed to parse machine info: %w", err)
	}

	return &machineInfo, nil
}

// SaveMachineInfo saves machine info to a file
func SaveMachineInfo(path string, machineInfo *types.MachineInfo) error {
	data, err := json.MarshalIndent(machineInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal machine info: %w", err)
	}

	if err := WriteFileAtomic(path, data, FilePerm); err != nil {
		return fmt.Errorf("failed to write machine info: %w", err)
	}

	return nil
}

// LoadPrivateKey loads the machine's private key from ~/.nvolt
func LoadPrivateKey() (*crypto.RSAPrivateKey, error) {
	homePaths, err := GetHomePaths()
	if err != nil {
		return nil, err
	}

	privateKeyPEM, err := ReadFile(homePaths.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := crypto.DecodePrivateKeyPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return privateKey, nil
}

// GenerateMachineID generates a machine ID from hostname
func GenerateMachineID(hostname string) string {
	if hostname == "" || hostname == "unknown" {
		return fmt.Sprintf("m-%d", time.Now().Unix())
	}
	return fmt.Sprintf("m-%s", hostname)
}

// AddMachineToVault adds a new machine's public key to the vault
func AddMachineToVault(vaultPath string, machineInfo *types.MachineInfo) error {
	paths := GetVaultPaths(vaultPath)
	machinePath := paths.GetMachineInfoPath(machineInfo.ID)

	// Check if machine already exists
	if FileExists(machinePath) {
		return fmt.Errorf("machine %s already exists in vault", machineInfo.ID)
	}

	return SaveMachineInfo(machinePath, machineInfo)
}

// RemoveMachineFromVault removes a machine from the vault
func RemoveMachineFromVault(vaultPath string, machineID string) error {
	paths := GetVaultPaths(vaultPath)

	// Remove machine info
	machinePath := paths.GetMachineInfoPath(machineID)
	if err := DeleteFile(machinePath); err != nil {
		return fmt.Errorf("failed to remove machine info: %w", err)
	}

	// Remove wrapped key
	wrappedKeyPath := paths.GetWrappedKeyPath(machineID)
	if err := DeleteFile(wrappedKeyPath); err != nil {
		return fmt.Errorf("failed to remove wrapped key: %w", err)
	}

	return nil
}

// ListMachines lists all machines in the vault
func ListMachines(vaultPath string) ([]*types.MachineInfo, error) {
	paths := GetVaultPaths(vaultPath)

	files, err := ListFiles(paths.Machines)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}

	var machines []*types.MachineInfo
	for _, file := range files {
		machineInfo, err := LoadMachineInfoFromFile(file)
		if err != nil {
			// Skip invalid files
			continue
		}
		machines = append(machines, machineInfo)
	}

	return machines, nil
}

// GetCurrentMachineID returns the current machine's ID
func GetCurrentMachineID() (string, error) {
	machineInfo, err := LoadMachineInfo()
	if err != nil {
		return "", err
	}
	return machineInfo.ID, nil
}
