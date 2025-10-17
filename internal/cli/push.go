package cli

import (
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/services"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push environment variables to nvolt.io",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")
		file, _ := cmd.Flags().GetString("file")
		keyValues, _ := cmd.Flags().GetStringSlice("key")

		machineConfig := cmd.Context().Value("machine_config").(*services.MachineConfig)
		aclService := cmd.Context().Value("acl_service").(*services.ACLService)
		if project != "" {
			machineConfig.Project = project
		}
		if environment != "" {
			machineConfig.Environment = environment
		}

		var vars map[string]string
		var replaceAll bool

		// Priority: -k flag over -f flag
		if len(keyValues) > 0 {
			// Parse key-value pairs from -k flags
			vars = make(map[string]string)
			for _, kv := range keyValues {
				key, value, err := helpers.ParseKeyValue(kv)
				if err != nil {
					return fmt.Errorf("invalid key-value format '%s': %w", kv, err)
				}
				vars[key] = value
			}
			replaceAll = false // Partial update - only add/update specified keys
		} else if file != "" {
			// Read environment variables from file
			var err error
			vars, err = helpers.ReadEnvFile(file)
			if err != nil {
				return fmt.Errorf("failed to read env file: %w", err)
			}
			replaceAll = true // Full replacement - replace all keys with file contents
		} else {
			return fmt.Errorf("either -f (file) or -k (key-value) flag must be specified")
		}

		return runPush(machineConfig, aclService, vars, replaceAll)
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("project", "p", "", "Project name")
	rootCmd.PersistentFlags().StringP("environment", "e", "", "Environment")
	pushCmd.PersistentFlags().StringP("file", "f", "", ".env or .env.production or .env.staging or .env.development file to push")
	pushCmd.PersistentFlags().StringSliceP("key", "k", []string{}, "Push individual key-value pairs (e.g., -k FOO=bar -k BUZ=qux)")
	rootCmd.AddCommand(pushCmd)
}

func runPush(machineConfig *services.MachineConfig, aclService *services.ACLService, vars map[string]string, replaceAll bool) error {
	activeOrgName, Role, err := aclService.GetActiveOrgName(machineConfig.Config.ActiveOrgID)
	if err != nil {
		return fmt.Errorf("failed to fetch active organization name: %w", err)
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("\n→ Project: %s", machineConfig.Project)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Environment: %s", machineConfig.Environment)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Active Organization: %s (%s) [%s]", activeOrgName, Role, machineConfig.Config.ActiveOrgID)))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Machine Key ID: %s", machineConfig.Config.MachineID)))

	if replaceAll {
		fmt.Println(infoStyle.Render("→ Mode: Full replacement (all existing variables will be replaced)"))
	} else {
		fmt.Println(infoStyle.Render("→ Mode: Partial update (only specified variables will be added/updated)"))
	}

	fmt.Println(infoStyle.Render(fmt.Sprintf("→ Variables to push: %d", len(vars))))

	secretsClient := services.NewSecretsClient(machineConfig)

	fmt.Println("\n" + titleStyle.Render("Pushing secrets to server..."))

	err = secretsClient.PushSecrets(machineConfig.Project, machineConfig.Environment, vars, replaceAll)
	if err != nil {
		return err
	}

	fmt.Println("\n" + successStyle.Render("✓ Successfully pushed secrets!"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("→ %d variables are now securely stored\n", len(vars))))

	return nil
}
