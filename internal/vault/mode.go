package vault

import (
	"path/filepath"
	"strings"
)

// VaultMode represents the mode of operation
type VaultMode int

const (
	// ModeLocal means vault is in current directory (./.nvolt)
	ModeLocal VaultMode = iota
	// ModeGlobal means vault is in a global repo (~/.nvolt/projects/org/repo/.nvolt)
	ModeGlobal
)

// GetVaultMode determines if the vault is in local or global mode
func GetVaultMode(vaultPath string) VaultMode {
	homePaths, err := GetHomePaths()
	if err != nil {
		return ModeLocal
	}

	// If vault path is under ~/.nvolt/projects/, it's global mode
	if strings.HasPrefix(vaultPath, homePaths.Projects) {
		return ModeGlobal
	}

	return ModeLocal
}

// GetRepoPathFromVault extracts the repository path from a vault path
// For example: ~/.nvolt/projects/org/repo/.nvolt -> ~/.nvolt/projects/org/repo
func GetRepoPathFromVault(vaultPath string) string {
	// Remove the .nvolt suffix
	return filepath.Dir(vaultPath)
}

// IsGlobalMode checks if a vault path is in global mode
func IsGlobalMode(vaultPath string) bool {
	return GetVaultMode(vaultPath) == ModeGlobal
}
