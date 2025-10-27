package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderProjectsPanel renders the left panel with project list
func (m Model) renderProjectsPanel(width int) string {
	// Panel title
	title := titleStyle.Render("Projects")

	// Table headers
	headers := []string{"Project Name"}
	headerWidths := []int{width - 4}

	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Width(headerWidths[i]).Render(header)
	}

	// Table rows
	rows := []string{}
	for i, project := range m.projects {
		// Determine if this row is selected (cursor) or active (currently loaded)
		isCursor := i == m.projectsCursor && m.focusedPanel == ProjectsPanel
		isActive := i == m.activeProjectIndex

		// Build row with indicators
		prefix := "  "
		if isCursor && isActive {
			prefix = "▶ " // Both cursor and active
		} else if isCursor {
			prefix = "> " // Just cursor
		} else if isActive {
			prefix = "• " // Just active
		}

		var nameCell string
		if isCursor {
			// Cursor on this row - highlight
			nameCell = selectedRowStyle.Width(headerWidths[0]).Render(prefix + project.Name)
		} else if isActive {
			// Active project but not cursor - show with primary color
			nameCell = lipgloss.NewStyle().
				Width(headerWidths[0]).
				Foreground(primaryColor).
				Render(prefix + project.Name)
		} else {
			// Regular row
			nameCell = tableRowStyle.Width(headerWidths[0]).Render(prefix + project.Name)
		}

		rows = append(rows, nameCell)
	}

	// Build table
	var table string
	if m.loading && len(m.projects) == 0 {
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
			table = lipgloss.JoinVertical(lipgloss.Left, table, tableRowStyle.Render("No projects found"))
		}
	}

	// Add environment info for selected project
	var envInfo string
	if len(m.projects) > 0 && m.activeProjectIndex < len(m.projects) {
		envList := strings.Join(m.projects[m.activeProjectIndex].Environments, ", ")
		envInfo = fmt.Sprintf("\n%s",
			infoStyle.Render(fmt.Sprintf("Environments: %s", envList)))
	}

	// Build panel content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", table, envInfo)

	// Apply panel style based on focus
	if m.focusedPanel == ProjectsPanel {
		return activePanelStyle.Width(width).Height(m.height - 10).Render(content)
	}
	return panelStyle.Width(width).Height(m.height - 10).Render(content)
}
