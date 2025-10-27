package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - matching nvolt.io website
	primaryColor    = lipgloss.Color("#FFC107") // Yellow/gold from logo
	successColor    = lipgloss.Color("#00D084")
	errorColor      = lipgloss.Color("#FF6B6B")
	infoColor       = lipgloss.Color("#4ECDC4")
	warnColor       = lipgloss.Color("#FFD93D")
	selectedColor   = lipgloss.Color("#FFC107") // Yellow for selected items
	borderColor     = lipgloss.Color("#333333") // Dark borders like website
	dimColor        = lipgloss.Color("#666666") // Lighter dim for better contrast
	backgroundColor = lipgloss.Color("#1a1a1a") // Dark background like website

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 1)

	// Table header style
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(dimColor).
				Padding(0, 1)

	// Table row styles
	tableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(selectedColor).
				Bold(true).
				Padding(0, 1)

	// Tag styles (for project/environment badges)
	tagStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	dimTagStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(0, 1)

	// Modal styles
	modalOverlayStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#000000"))

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Background(lipgloss.Color("#3D1F1F")).
			Padding(2, 4).
			Width(60)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 2)

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor)

	// Success style
	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	// Info style
	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor)
)
