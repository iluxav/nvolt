package cli

import (
	"fmt"

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
		// TODO: Implement vault show logic
		fmt.Printf("Vault Information:\n")
		fmt.Printf("==================\n")

		return fmt.Errorf("not yet implemented")
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
		// TODO: Implement vault verify logic
		fmt.Printf("Verifying vault integrity...\n")

		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	vaultCmd.AddCommand(vaultShowCmd)
	vaultCmd.AddCommand(vaultVerifyCmd)
	rootCmd.AddCommand(vaultCmd)
}
