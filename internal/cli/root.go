package cli

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "nvolt",
	Short: "nvolt - GitHub-native, Zero-Trust encrypted environment variable manager",
	Long: `nvolt is a Zero-Trust CLI for managing encrypted environment variables
without a centralized backend, login, or organization model.

All data lives in Git repositories, with encryption/decryption happening
locally using per-machine keypairs. Access control is cryptographically
enforced through wrapped key files.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
}
