package tui

import (
	"fmt"
	"iluxav/nvolt/internal/types"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatPermissions formats Permission struct into a readable string
func formatPermissions(perms *types.Permission) string {
	if perms == nil {
		return "No Access"
	}

	parts := []string{}
	if perms.Read {
		parts = append(parts, "R")
	}
	if perms.Write {
		parts = append(parts, "W")
	}
	if perms.Delete {
		parts = append(parts, "D")
	}

	if len(parts) == 0 {
		return "No Access"
	}

	return strings.Join(parts, "|")
}

// renderUsersPanel renders the right panel with users
func (m Model) renderUsersPanel(width int) string {
	// Panel title
	title := titleStyle.Render(fmt.Sprintf("Users in %s, %s", m.projectName, m.environments[m.activeEnvIndex].Name))

	// Table headers
	headers := []string{"Name", "Email", "Project Perms", "Env Perms", "Org Role"}
	headerWidths := []int{width / 4, width / 3, width / 7, width / 7, width / 10}

	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Width(headerWidths[i]).Render(header)
	}

	// Table rows
	rows := []string{}
	for i, user := range m.users {
		// Determine if this row is selected
		isSelected := i == m.usersCursor && m.focusedPanel == UsersPanel

		// Format project permissions
		projectPerms := formatPermissions(user.ProjectPermissions)

		// Format environment permissions
		envPerms := formatPermissions(user.EnvironmentPermissions)

		// Build row
		var nameCell, emailCell, projectPermsCell, envPermsCell, roleCell string
		if isSelected {
			// Selected row - highlight
			nameCell = selectedRowStyle.Width(headerWidths[0]).Render(user.Name)
			emailCell = selectedRowStyle.Width(headerWidths[1]).Render(user.Email)
			projectPermsCell = selectedRowStyle.Width(headerWidths[2]).Render(projectPerms)
			envPermsCell = selectedRowStyle.Width(headerWidths[3]).Render(envPerms)
			roleCell = selectedRowStyle.Width(headerWidths[4]).Render(user.OrgRole)
		} else {
			nameCell = tableRowStyle.Width(headerWidths[0]).Render(user.Name)
			emailCell = tableRowStyle.Width(headerWidths[1]).Render(user.Email)
			projectPermsCell = tableRowStyle.Width(headerWidths[2]).Render(projectPerms)
			envPermsCell = tableRowStyle.Width(headerWidths[3]).Render(envPerms)
			roleCell = tableRowStyle.Width(headerWidths[4]).Render(user.OrgRole)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left, nameCell, emailCell, projectPermsCell, envPermsCell, roleCell)
		rows = append(rows, row)
	}

	// Build table
	var table string
	if m.loading {
		// Show loading indicator in table area
		loadingText := infoStyle.Render("⏳ Loading...")
		table = lipgloss.JoinVertical(lipgloss.Left, headerRow, "", loadingText)
	} else {
		table = lipgloss.JoinVertical(lipgloss.Left, headerRow)
		if len(rows) > 0 {
			table = lipgloss.JoinVertical(lipgloss.Left, table, strings.Join(rows, "\n"))
		} else {
			table = lipgloss.JoinVertical(lipgloss.Left, table, tableRowStyle.Render("No users found"))
		}
	}

	// Build panel content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", table)

	// Apply panel style based on focus
	if m.focusedPanel == UsersPanel {
		return activePanelStyle.Width(width).Height(m.height - 10).Render(content)
	}
	return panelStyle.Width(width).Height(m.height - 10).Render(content)
}
