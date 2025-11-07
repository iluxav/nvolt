package vault

import (
	"strings"
)

// VaultMode represents the mode of operation
type VaultMode int

const (
	// ModeLocal means vault is in current directory (./.nvolt)
	ModeLocal VaultMode = iota
	// ModeGlobal means vault is in a global repo (~/.nvolt/orgs/org/repo)
	ModeGlobal
)

// GetVaultMode determines if the vault is in local or global mode
func GetVaultMode(vaultPath string) VaultMode {
	homePaths, err := GetHomePaths()
	if err != nil {
		return ModeLocal
	}

	// If vault path is under ~/.nvolt/orgs/, it's global mode
	if strings.HasPrefix(vaultPath, homePaths.Orgs) {
		return ModeGlobal
	}

	return ModeLocal
}

// GetRepoPathFromVault extracts the repository path from a vault path
// For global mode: ~/.nvolt/orgs/org/repo -> ~/.nvolt/orgs/org/repo
// For local mode: returns the directory containing .nvolt
func GetRepoPathFromVault(vaultPath string) string {
	return GetRepoRootFromVault(vaultPath)
}

// IsGlobalMode checks if a vault path is in global mode
func IsGlobalMode(vaultPath string) bool {
	return GetVaultMode(vaultPath) == ModeGlobal
}
