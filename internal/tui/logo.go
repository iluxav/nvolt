package tui

import "github.com/charmbracelet/lipgloss"

// renderLogo renders the nvolt ASCII art logo (compact version for TUI)
func renderLogo() string {
	// Compact NVOLT logo that fits in TUI header
	logo := `  ███╗   ██╗██╗   ██╗ ██████╗ ██╗  ████████╗
  ████╗  ██║██║   ██║██╔═══██╗██║  ╚══██╔══╝
  ██╔██╗ ██║╚██╗ ██╔╝██║   ██║██║     ██║
  ██║╚████║ ╚████╔╝ ╚██████╔╝███████╗██║   `

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	return logoStyle.Render(logo)
}

// renderFullLogo renders the complete ASCII art logo (for CLI)
func renderFullLogo() string {
	// Full NVOLT logo for CLI display
	logo := `
    ███╗   ██╗██╗   ██╗ ██████╗ ██╗  ████████╗
    ████╗  ██║██║   ██║██╔═══██╗██║  ╚══██╔══╝
    ██╔██╗ ██║██║   ██║██║   ██║██║     ██║
    ██║╚██╗██║╚██╗ ██╔╝██║   ██║██║     ██║
    ██║ ╚████║ ╚████╔╝ ╚██████╔╝███████╗██║
    ╚═╝  ╚═══╝  ╚═══╝   ╚═════╝ ╚══════╝╚═╝   `

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	return logoStyle.Render(logo)
}

// renderCompactLogo renders a compact one-line logo for header
func renderCompactLogo() string {
	logo := "⚡ NVOLT"

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)

	return logoStyle.Render(logo)
}
