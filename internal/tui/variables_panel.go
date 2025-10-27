package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderVariablesPanel renders the left panel with environment variables
func (m Model) renderVariablesPanel(width int) string {
	// Panel title
	title := titleStyle.Render("Environment Variables")

	// Table headers with creation date
	headers := []string{"Variable Name", "Value", "Created"}
	headerWidths := []int{width * 3 / 10, width * 5 / 10, width * 2 / 10}

	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Width(headerWidths[i]).Render(header)
	}

	// Table rows
	rows := []string{}
	for i, variable := range m.variables {
		// Determine if this row is selected
		isSelected := i == m.variablesCursor && m.focusedPanel == VariablesPanel

		// Show actual decrypted value
		value := variable.Value

		// Format creation date
		created := "-"
		if !variable.Created.IsZero() {
			created = variable.Created.Format("02/01/2006 15:04")
		}

		// Build row with created date
		var nameCell, valueCell, createdCell string
		if isSelected {
			// Selected row - show arrow and highlight
			nameCell = selectedRowStyle.Width(headerWidths[0]).Render("> " + variable.Name)
			valueCell = selectedRowStyle.Width(headerWidths[1]).Render(value)
			createdCell = selectedRowStyle.Width(headerWidths[2]).Render(created)
		} else {
			nameCell = tableRowStyle.Width(headerWidths[0]).Render("  " + variable.Name)
			valueCell = tableRowStyle.Width(headerWidths[1]).Render(value)
			createdCell = tableRowStyle.Width(headerWidths[2]).Render(created)
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left, nameCell, valueCell, createdCell)
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
			table = lipgloss.JoinVertical(lipgloss.Left, table, tableRowStyle.Render("No variables found"))
		}
	}

	// Build panel content
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", table)

	// Apply panel style based on focus
	if m.focusedPanel == VariablesPanel {
		return activePanelStyle.Width(width).Height(m.height - 10).Render(content)
	}
	return panelStyle.Width(width).Height(m.height - 10).Render(content)
}
