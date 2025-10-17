package cli

import (
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull environment variables from nvolt.io",
	Long:  `Pull encrypted environment variables from the server, decrypt them, and output to file or console`,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")
		file, _ := cmd.Flags().GetString("file")
		key, _ := cmd.Flags().GetString("key")

		machineConfig := cmd.Context().Value("machine_config").(*services.MachineConfig)
		aclService := cmd.Context().Value("acl_service").(*services.ACLService)
		if project != "" {
			machineConfig.Project = project
		}
		if environment != "" {
			machineConfig.Environment = environment
		}

		return runPull(machineConfig, aclService, file, key)
	},
}

func init() {
	pullCmd.PersistentFlags().StringP("file", "f", "", "Output file to write decrypted variables (e.g., .env.local)")
	pullCmd.PersistentFlags().StringP("key", "k", "", "Pull a specific key only")
	rootCmd.AddCommand(pullCmd)
}

func runPull(machineConfig *services.MachineConfig, aclService *services.ACLService, outputFile, specificKey string) error {

	activeOrgName, Role, err := aclService.GetActiveOrgName(machineConfig.Config.ActiveOrgID)
	if err != nil {
		return fmt.Errorf("failed to fetch active organization name: %w", err)
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ Project: %s", machineConfig.Project)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Environment: %s", machineConfig.Environment)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Active Organization: %s (%s) [%s]", activeOrgName, Role, machineConfig.Config.ActiveOrgID)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Machine Key ID: %s", machineConfig.Config.MachineID)))

	if specificKey != "" {
		fmt.Println(infoStyle.Render(fmt.Sprintf("→ Pulling specific key: %s", specificKey)))
	}

	secretsClient := services.NewSecretsClient(machineConfig)

	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " " + titleStyle.Render("Pulling secrets from server...")
	s.Start()

	vars, err := secretsClient.PullSecrets(machineConfig.Project, machineConfig.Environment, specificKey)
	s.Stop()
	fmt.Print("\033[K")
	if err != nil {
		// Check if it's a "no wrapped key" error (new machine not synced yet)
		if strings.Contains(err.Error(), "no wrapped key") || strings.Contains(err.Error(), "WrappedKey is empty") {
			fmt.Println("\n" + warnStyle.Render("⚠ No secrets found for this machine"))
			fmt.Println(infoStyle.Render("\nThis machine hasn't been synced yet. To enable access:"))
			fmt.Println(infoStyle.Render("  1. Run 'nvolt sync' from ANY authorized machine, OR"))
			fmt.Println(infoStyle.Render("  2. Run 'nvolt push' from any machine to sync all machines"))
			return nil
		}
		return err
	}

	if len(vars) == 0 {
		fmt.Println("\n" + warnStyle.Render("⚠ No variables found for this scope"))
		return nil
	}

	// If output file is specified, write to file
	if outputFile != "" {
		err := helpers.WriteEnvFile(outputFile, vars)
		if err != nil {
			return fmt.Errorf("failed to write env file: %w", err)
		}

		fmt.Println("\n" + successStyle.Render(fmt.Sprintf("✓ Successfully pulled %d variable(s)!", len(vars))))
		fmt.Println(infoStyle.Render(fmt.Sprintf("→ Saved to: %s", outputFile)))
	} else {
		// Output to console
		fmt.Println("\n" + successStyle.Render(fmt.Sprintf("✓ Successfully pulled %d variable(s)!", len(vars))))
		fmt.Println(titleStyle.Render("\nDecrypted Variables:\n"))

		// Sort keys for consistent output
		keys := make([]string, 0, len(vars))
		for k := range vars {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Print in .env format
		for _, k := range keys {
			v := vars[k]
			// Escape special characters in values if needed
			if strings.ContainsAny(v, " \t\n\"'") {
				v = fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\""))
			}
			fmt.Println(listItemStyle.Render(fmt.Sprintf("%s=%s", k, v)))
		}
		fmt.Println(infoStyle.Render("\n"))
	}

	return nil
}
