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

// renderUsersContent renders the users table content (without panel wrapper)
func (m Model) renderUsersContent(width int) string {
	// Panel title with current project and environment
	var projectName, envName string
	if len(m.projects) > 0 {
		projectName = m.projects[m.activeProjectIndex].Name
	} else {
		projectName = "..."
	}
	if len(m.environments) > 0 {
		envName = m.environments[m.activeEnvIndex].Name
	} else {
		envName = "..."
	}
	title := titleStyle.Render(fmt.Sprintf("Users in %s, %s", projectName, envName))

	// Table headers
	// Adjust to fit within panel width
	headers := []string{"Name", "Email", "Project Perms", "Env Perms", "Org Role"}
	usableWidth := width - 6
	headerWidths := []int{
		usableWidth / 5,      // Name: 20%
		usableWidth * 3 / 10, // Email: 30%
		usableWidth / 6,      // Project Perms: ~17%
		usableWidth / 6,      // Env Perms: ~17%
		usableWidth / 6,      // Org Role: ~17%
	}

	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Width(headerWidths[i]).Render(header)
	}

	// Table rows
	rows := []string{}
	for i, user := range m.users {
		// Determine if this row is selected
		isSelected := i == m.usersCursor && m.focusedPanel == RightPanel && m.activeTab == UsersTab

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
		// Add spacing after header
		table = lipgloss.JoinVertical(lipgloss.Left, headerRow, "")
		if len(rows) > 0 {
			// Add small vertical spacing between rows
			table = lipgloss.JoinVertical(lipgloss.Left, table, strings.Join(rows, "\n\n"))
		} else {
			table = lipgloss.JoinVertical(lipgloss.Left, table, tableRowStyle.Render("No users found"))
		}
	}

	// Build content (no panel wrapper - that's handled by renderRightPanel)
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", table)
	return content
}
