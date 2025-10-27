package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderVariablesContent renders the variables table content (without panel wrapper)
func (m Model) renderVariablesContent(width int) string {
	// Panel title
	title := titleStyle.Render("Environment Variables")

	// Table headers with creation date
	// Adjust widths: smaller created date, more space for name and value
	headers := []string{"Variable Name", "Value", "Created"}
	headerWidths := []int{(width - 6) * 3 / 10, (width - 6) * 5 / 10, (width - 6) * 2 / 10}

	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Width(headerWidths[i]).Render(header)
	}

	// Table rows
	rows := []string{}
	for i, variable := range m.variables {
		// Determine if this row is selected
		isSelected := i == m.variablesCursor && m.focusedPanel == RightPanel && m.activeTab == VariablesTab

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
		// Add spacing after header
		table = lipgloss.JoinVertical(lipgloss.Left, headerRow, "")
		if len(rows) > 0 {
			// Add small vertical spacing between rows
			table = lipgloss.JoinVertical(lipgloss.Left, table, strings.Join(rows, "\n\n"))
		} else {
			table = lipgloss.JoinVertical(lipgloss.Left, table, tableRowStyle.Render("No variables found"))
		}
	}

	// Build content (no panel wrapper - that's handled by renderRightPanel)
	content := lipgloss.JoinVertical(lipgloss.Left, title, "", table)
	return content
}
