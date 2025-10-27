package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// renderRightPanel renders the combined Variables/Users panel with tabs
func (m Model) renderRightPanel(width int) string {
	// Render tab bar
	tabBar := m.renderTabBar(width)

	// Render content based on active tab
	var content string
	if m.activeTab == VariablesTab {
		content = m.renderVariablesContent(width)
	} else {
		content = m.renderUsersContent(width)
	}

	// Combine tab bar and content
	panelContent := lipgloss.JoinVertical(lipgloss.Left, tabBar, content)

	// Apply panel style based on focus
	if m.focusedPanel == RightPanel {
		return activePanelStyle.Width(width).Height(m.height - 10).Render(panelContent)
	}
	return panelStyle.Width(width).Height(m.height - 10).Render(panelContent)
}

// renderTabBar renders the tab selector at the top of the right panel
func (m Model) renderTabBar(width int) string {
	// Tab styles
	activeTabStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(dimColor).
		Padding(0, 2)

	// Render tabs
	var variablesTab, usersTab string
	if m.activeTab == VariablesTab {
		variablesTab = activeTabStyle.Render("● Variables")
		usersTab = inactiveTabStyle.Render("  Users  ")
	} else {
		variablesTab = inactiveTabStyle.Render("  Variables")
		usersTab = activeTabStyle.Render("● Users")
	}

	// Combine tabs
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, variablesTab, usersTab)

	// Add separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render(lipgloss.NewStyle().Width(width - 4).Render("─"))

	return lipgloss.JoinVertical(lipgloss.Left, tabs, separator)
}
