package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/iluxav/nvolt/internal/config"
	"github.com/iluxav/nvolt/internal/ui"
	"github.com/iluxav/nvolt/internal/vault"
	"github.com/spf13/cobra"
)

var (
	statusProject     string
	statusEnvironment string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault status and configuration overview",
	Long: `Display a tree view of your vault structure including:
- Mode (local or global)
- Organizations and projects
- Environments and secrets
- Registered machines

The current project and environment are highlighted.

Examples:
  nvolt status
  nvolt status -p myproject
  nvolt status -e production
  nvolt status -p myproject -e staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

// Styles for the status display
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220")). // Gold
			MarginTop(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")) // Cyan

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")) // Light gray
)

func runStatus() error {
	// Find vault path
	vaultPath, err := findVaultPath()
	if err != nil {
		fmt.Println(ui.Gray("No vault initialized"))
		fmt.Println("\nRun 'nvolt init' to get started")
		return nil
	}

	mode := vault.GetVaultMode(vaultPath)
	isGlobal := mode == vault.ModeGlobal

	// Get current machine info
	currentMachine, err := vault.LoadMachineInfo()
	machineID := "(unknown)"
	if err == nil {
		machineID = currentMachine.ID
	}

	// Print header
	fmt.Println()
	fmt.Println(headerStyle.Render("📊 Vault Status"))

	// Display store and machine info
	modeStr := "Local"
	if isGlobal {
		modeStr = "Global"
	}
	fmt.Printf("%s %s (%s)\n", labelStyle.Render("Store:"), valueStyle.Render(modeStr), valueStyle.Render(vaultPath))
	fmt.Printf("%s %s\n\n", labelStyle.Render("Machine:"), valueStyle.Render(machineID))

	if isGlobal {
		return displayGlobalStatus(vaultPath)
	}
	return displayLocalStatus(vaultPath)
}

func displayLocalStatus(vaultPath string) error {
	paths := vault.GetVaultPaths(vaultPath, "")

	// Get project name
	cwd, _ := os.Getwd()
	projectName, _, err := config.GetProjectName(cwd, "")
	if err != nil {
		projectName = "local"
	}

	// List environments
	envDirs, err := vault.ListDirs(paths.Secrets)
	if err != nil || len(envDirs) == 0 {
		fmt.Println(valueStyle.Render("No environments found"))
		return nil
	}

	sort.Strings(envDirs)

	// Build flat table data
	var rows [][]string

	for _, envDir := range envDirs {
		envName := vault.GetDirName(envDir)

		// Filter by environment if specified
		if statusEnvironment != "" && envName != statusEnvironment {
			continue
		}

		secretFiles, err := vault.ListFiles(envDir)
		if err != nil {
			continue
		}

		if len(secretFiles) == 0 {
			// Show environment even if no secrets
			rows = append(rows, []string{projectName, envName, valueStyle.Render("(no secrets)")})
		} else {
			// Add a row for each secret
			for _, secretFile := range secretFiles {
				secretKey := vault.GetSecretKeyFromFilename(secretFile)
				rows = append(rows, []string{projectName, envName, secretKey})
			}
		}
	}

	// Render table
	fmt.Println(headerStyle.Render("Secrets"))
	if len(rows) == 0 {
		fmt.Println(valueStyle.Render("No secrets found matching the filter criteria"))
	} else {
		fmt.Println(renderTable([]string{"Project Name", "Environment", "Env Var"}, rows))
	}
	fmt.Println()

	// Display machine access
	if err := displayMachineAccess(paths, projectName); err != nil {
		return err
	}

	return displayMachines(paths)
}

func displayGlobalStatus(vaultPath string) error {
	// List all projects in the global vault
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return fmt.Errorf("failed to read vault directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "machines" && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, entry.Name())
		}
	}

	if len(projects) == 0 {
		fmt.Println(valueStyle.Render("No projects found"))
		return nil
	}

	sort.Strings(projects)

	// Build flat table data for all projects
	var rows [][]string
	for _, project := range projects {
		// Filter by project if specified
		if statusProject != "" && project != statusProject {
			continue
		}

		paths := vault.GetVaultPaths(vaultPath, project)
		envDirs, err := vault.ListDirs(paths.Secrets)
		if err != nil {
			continue
		}

		sort.Strings(envDirs)

		for _, envDir := range envDirs {
			envName := vault.GetDirName(envDir)

			// Filter by environment if specified
			if statusEnvironment != "" && envName != statusEnvironment {
				continue
			}

			secretFiles, err := vault.ListFiles(envDir)
			if err != nil {
				continue
			}

			if len(secretFiles) == 0 {
				// Show environment even if no secrets
				rows = append(rows, []string{project, envName, valueStyle.Render("(no secrets)")})
			} else {
				// Add a row for each secret
				for _, secretFile := range secretFiles {
					key := vault.GetSecretKeyFromFilename(secretFile)
					rows = append(rows, []string{project, envName, key})
				}
			}
		}
	}

	// Render table
	fmt.Println(headerStyle.Render("Secrets"))
	if len(rows) == 0 {
		fmt.Println(valueStyle.Render("No secrets found matching the filter criteria"))
	} else {
		fmt.Println(renderTable([]string{"Project Name", "Environment", "Env Var"}, rows))
	}
	fmt.Println()

	// Display machine access for all projects
	if err := displayGlobalMachineAccess(vaultPath, projects); err != nil {
		return err
	}

	// Show machines using the root-level machines directory
	paths := vault.GetVaultPaths(vaultPath, "")
	return displayMachines(paths)
}

func displayMachineAccess(paths *vault.Paths, projectName string) error {
	// Get all machines
	machines, err := vault.ListMachines(paths)
	if err != nil {
		return err
	}

	if len(machines) == 0 {
		return nil
	}

	// Build machine access table
	var rows [][]string

	// Get all environments to check for wrapped keys
	envDirs, err := vault.ListDirs(paths.Secrets)
	if err != nil {
		return err
	}

	for _, machine := range machines {
		var accessibleEnvs []string

		// Check each environment for this machine's wrapped key
		for _, envDir := range envDirs {
			envName := vault.GetDirName(envDir)

			// Filter by environment if specified
			if statusEnvironment != "" && envName != statusEnvironment {
				continue
			}

			wrappedKeyPath := paths.GetWrappedKeyPath(envName, machine.ID)
			if _, err := os.Stat(wrappedKeyPath); err == nil {
				accessibleEnvs = append(accessibleEnvs, envName)
			}
		}

		if len(accessibleEnvs) > 0 {
			envList := strings.Join(accessibleEnvs, ", ")
			rows = append(rows, []string{machine.ID, projectName, envList})
		}
	}

	if len(rows) > 0 {
		fmt.Println(headerStyle.Render("Machine Access"))
		fmt.Println(renderTable([]string{"Machine ID", "Project", "Environments"}, rows))
		fmt.Println()
	}

	return nil
}

func displayGlobalMachineAccess(vaultPath string, projects []string) error {
	// Get all machines from root machines directory
	rootPaths := vault.GetVaultPaths(vaultPath, "")
	machines, err := vault.ListMachines(rootPaths)
	if err != nil {
		return err
	}

	if len(machines) == 0 {
		return nil
	}

	// Build machine access table
	var rows [][]string

	for _, machine := range machines {
		for _, project := range projects {
			// Filter by project if specified
			if statusProject != "" && project != statusProject {
				continue
			}

			paths := vault.GetVaultPaths(vaultPath, project)
			var accessibleEnvs []string

			// Get all environments for this project
			envDirs, err := vault.ListDirs(paths.Secrets)
			if err != nil {
				continue
			}

			// Check each environment for this machine's wrapped key
			for _, envDir := range envDirs {
				envName := vault.GetDirName(envDir)

				// Filter by environment if specified
				if statusEnvironment != "" && envName != statusEnvironment {
					continue
				}

				wrappedKeyPath := paths.GetWrappedKeyPath(envName, machine.ID)
				if _, err := os.Stat(wrappedKeyPath); err == nil {
					accessibleEnvs = append(accessibleEnvs, envName)
				}
			}

			if len(accessibleEnvs) > 0 {
				envList := strings.Join(accessibleEnvs, ", ")
				rows = append(rows, []string{machine.ID, project, envList})
			}
		}
	}

	if len(rows) > 0 {
		fmt.Println(headerStyle.Render("Machine Access"))
		fmt.Println(renderTable([]string{"Machine ID", "Project", "Environments"}, rows))
		fmt.Println()
	}

	return nil
}

func displayMachines(paths *vault.Paths) error {
	machines, err := vault.ListMachines(paths)
	if err != nil {
		return err
	}

	currentMachine, err := vault.LoadMachineInfo()
	var currentID string
	if err == nil {
		currentID = currentMachine.ID
	}

	if len(machines) == 0 {
		fmt.Println(valueStyle.Render("No machines registered"))
		return nil
	}

	// Sort machines, put current first
	sort.Slice(machines, func(i, j int) bool {
		if machines[i].ID == currentID {
			return true
		}
		if machines[j].ID == currentID {
			return false
		}
		return machines[i].ID < machines[j].ID
	})

	// Build machines table
	var rows [][]string
	for _, m := range machines {
		isCurrent := m.ID == currentID

		machineDisplay := m.ID
		statusDisplay := "registered"

		if isCurrent {
			statusDisplay = "current"
		}

		rows = append(rows, []string{machineDisplay, statusDisplay})
	}

	// Render table
	fmt.Println(headerStyle.Render("Registered Machines"))
	fmt.Println(renderTable([]string{"Machine ID", "Status"}, rows))

	return nil
}

// Helper to strip ANSI codes for accurate length calculation
func stripAnsiCodes(s string) string {
	// Remove ANSI escape sequences
	result := ""
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result += string(s[i])
	}
	return result
}

// renderTable renders a proper table with borders using bubbles/table
func renderTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	// Create table columns
	columns := make([]table.Column, len(headers))
	for i, header := range headers {
		// Calculate column width based on actual content (stripping ANSI codes)
		width := len(header)
		for _, row := range rows {
			if i < len(row) {
				// Strip ANSI codes to get actual text length
				plainText := stripAnsiCodes(row[i])
				if len(plainText) > width {
					width = len(plainText)
				}
			}
		}
		columns[i] = table.Column{
			Title: header,
			Width: width + 2,
		}
	}

	// Convert rows to table.Row
	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}

	// Create table style
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("220")) // Gold
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("0")).
		Bold(false)

	// Create and configure table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(false),
		table.WithStyles(s),
	)

	// Set height after creation to fit content
	// Need extra row for proper display (accounts for header rendering)
	t.SetHeight(len(tableRows) + 2)

	return t.View()
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVarP(&statusProject, "project", "p", "", "Filter by project name")
	statusCmd.Flags().StringVarP(&statusEnvironment, "environment", "e", "", "Filter by environment name")
}
