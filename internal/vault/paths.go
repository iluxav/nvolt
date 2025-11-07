package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// NvoltDir is the name of the vault directory
	NvoltDir = ".nvolt"

	// HomeNvoltDir is the nvolt directory in user's home
	HomeNvoltDir = ".nvolt"

	// SubDirs
	SecretsDir     = "secrets"
	WrappedKeysDir = "wrapped_keys"
	MachinesDir    = "machines"
	ProjectsDir    = "projects"

	// Files
	PrivateKeyFile  = "private_key.pem"
	MachineInfoFile = "machine-info.json"
	KeyInfoFile     = "keyinfo.json"
	ConfigFile      = "config.json"
)

// Paths holds all vault-related paths
type Paths struct {
	// Root is the root directory of the vault (.nvolt or ~/.nvolt/projects/org/repo)
	Root string

	// Secrets directory
	Secrets string

	// Wrapped keys directory
	WrappedKeys string

	// Machines directory
	Machines string

	// KeyInfo file
	KeyInfo string

	// Config file
	Config string
}

// HomePaths holds paths in the home directory
type HomePaths struct {
	// Root is ~/.nvolt
	Root string

	// PrivateKey is the machine's private key
	PrivateKey string

	// MachineInfo is the machine metadata
	MachineInfo string

	// Machines directory for storing machine public keys
	Machines string

	// Projects directory for global repos
	Projects string
}

// GetHomePaths returns the home directory paths
func GetHomePaths() (*HomePaths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	root := filepath.Join(homeDir, HomeNvoltDir)

	return &HomePaths{
		Root:        root,
		PrivateKey:  filepath.Join(root, PrivateKeyFile),
		MachineInfo: filepath.Join(root, MachinesDir, MachineInfoFile),
		Machines:    filepath.Join(root, MachinesDir),
		Projects:    filepath.Join(root, ProjectsDir),
	}, nil
}

// GetVaultPaths returns the vault directory paths
func GetVaultPaths(vaultRoot string) *Paths {
	return &Paths{
		Root:        vaultRoot,
		Secrets:     filepath.Join(vaultRoot, SecretsDir),
		WrappedKeys: filepath.Join(vaultRoot, WrappedKeysDir),
		Machines:    filepath.Join(vaultRoot, MachinesDir),
		KeyInfo:     filepath.Join(vaultRoot, KeyInfoFile),
		Config:      filepath.Join(vaultRoot, ConfigFile),
	}
}

// GetLocalVaultPath returns the vault path in the current directory
func GetLocalVaultPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return filepath.Join(cwd, NvoltDir), nil
}

// GetGlobalVaultPath returns the vault path for a global repo
func GetGlobalVaultPath(org, repo string) (*Paths, error) {
	homePaths, err := GetHomePaths()
	if err != nil {
		return nil, err
	}

	vaultRoot := filepath.Join(homePaths.Projects, org, repo, NvoltDir)
	return GetVaultPaths(vaultRoot), nil
}

// GetSecretsPath returns the path for an environment's secrets
func (p *Paths) GetSecretsPath(environment string) string {
	return filepath.Join(p.Secrets, environment)
}

// GetSecretFilePath returns the full path for a secret file
func (p *Paths) GetSecretFilePath(environment, key string) string {
	return filepath.Join(p.Secrets, environment, fmt.Sprintf("%s.enc.json", key))
}

// GetWrappedKeyPath returns the path for a machine's wrapped key
func (p *Paths) GetWrappedKeyPath(machineID string) string {
	return filepath.Join(p.WrappedKeys, fmt.Sprintf("%s.json", machineID))
}

// GetMachineInfoPath returns the path for a machine's info file
func (p *Paths) GetMachineInfoPath(machineID string) string {
	return filepath.Join(p.Machines, fmt.Sprintf("%s.json", machineID))
}

// IsVaultInitialized checks if a vault exists at the given path
func IsVaultInitialized(vaultPath string) bool {
	info, err := os.Stat(vaultPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// IsMachineInitialized checks if the machine keypair exists
func IsMachineInitialized() (bool, error) {
	homePaths, err := GetHomePaths()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(homePaths.PrivateKey)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check private key: %w", err)
	}

	return true, nil
}
