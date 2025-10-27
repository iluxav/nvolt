package tui

import (
	"fmt"
	"iluxav/nvolt/internal/services"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the main TUI model
type Model struct {
	// Services
	machineConfig  *services.MachineConfig
	secretsClient  *services.SecretsClient
	aclService     *services.ACLService

	// Data
	projectName    string
	environments   []Environment
	activeEnvIndex int
	variables      []EnvVariable
	users          []User

	// UI state
	focusedPanel    FocusedPanel
	variablesCursor int
	usersCursor     int
	showModal       ModalType
	modalTarget     string // Name of the item to delete
	width           int
	height          int
	err             error
	loading         bool
}

// NewModel creates a new TUI model
func NewModel(
	machineConfig *services.MachineConfig,
	secretsClient *services.SecretsClient,
	aclService *services.ACLService,
	projectName string,
) Model {
	// Initialize with temporary default environment - will be replaced by real data
	environments := []Environment{
		{Name: "default", IsActive: true},
	}

	return Model{
		machineConfig:   machineConfig,
		secretsClient:   secretsClient,
		aclService:      aclService,
		projectName:     projectName,
		environments:    environments,
		activeEnvIndex:  0,
		variables:       []EnvVariable{},
		users:           []User{},
		focusedPanel:    VariablesPanel,
		variablesCursor: 0,
		usersCursor:     0,
		showModal:       NoModal,
		loading:         true, // Start with loading state
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Load environments first, then data will load when environments are ready
	return m.loadEnvironments()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loadEnvironmentsMsg:
		// Environments have been loaded
		if msg.err != nil {
			m.loading = false
			m.err = msg.err
			return m, nil
		}

		// Update environments
		if len(msg.environments) > 0 {
			m.environments = make([]Environment, len(msg.environments))
			for i, envName := range msg.environments {
				m.environments[i] = Environment{
					Name:     envName,
					IsActive: i == 0, // First environment is active
				}
			}
			m.activeEnvIndex = 0
		}

		// Now load data for the first environment
		return m, m.loadData()

	case loadDataMsg:
		// Data has been loaded
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		// Convert and update variables
		if msg.variables != nil {
			m.variables = convertToEnvVariables(msg.variables)
		}

		// Convert and update users
		if msg.users != nil {
			envName := m.environments[m.activeEnvIndex].Name
			m.users = convertToUsers(msg.users, m.projectName, envName)
		}

		// Reset cursors if needed
		if m.variablesCursor >= len(m.variables) {
			m.variablesCursor = 0
		}
		if m.usersCursor >= len(m.users) {
			m.usersCursor = 0
		}

		return m, nil

	case tea.KeyMsg:
		// Handle modal first
		if m.showModal != NoModal {
			return m.handleModalKeys(msg)
		}

		// Handle error dismissal
		if m.err != nil && msg.String() == "esc" {
			m.err = nil
			return m, nil
		}

		// Handle normal keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Switch between left and right panels
			if m.focusedPanel == VariablesPanel {
				m.focusedPanel = UsersPanel
			} else {
				m.focusedPanel = VariablesPanel
			}
			return m, nil

		case "left", "right":
			// Navigate between environments
			oldIndex := m.activeEnvIndex
			if msg.String() == "left" && m.activeEnvIndex > 0 {
				m.activeEnvIndex--
			} else if msg.String() == "right" && m.activeEnvIndex < len(m.environments)-1 {
				m.activeEnvIndex++
			}

			// If environment changed, reload data
			if oldIndex != m.activeEnvIndex {
				// Mark active environment
				for i := range m.environments {
					m.environments[i].IsActive = i == m.activeEnvIndex
				}
				// Set loading state and trigger data load
				m.loading = true
				m.variablesCursor = 0 // Reset cursor
				return m, m.loadData()
			}
			return m, nil

		case "up", "k":
			if m.focusedPanel == VariablesPanel {
				if m.variablesCursor > 0 {
					m.variablesCursor--
				}
			} else {
				if m.usersCursor > 0 {
					m.usersCursor--
				}
			}
			return m, nil

		case "down", "j":
			if m.focusedPanel == VariablesPanel {
				if m.variablesCursor < len(m.variables)-1 {
					m.variablesCursor++
				}
			} else {
				if m.usersCursor < len(m.users)-1 {
					m.usersCursor++
				}
			}
			return m, nil

		case "enter", " ":
			// Toggle reveal/hide value for selected variable
			if m.focusedPanel == VariablesPanel && len(m.variables) > 0 {
				m.variables[m.variablesCursor].IsRevealed = !m.variables[m.variablesCursor].IsRevealed
			}
			return m, nil

		case "delete", "d":
			// Show delete modal
			if m.focusedPanel == VariablesPanel && len(m.variables) > 0 {
				m.showModal = DeleteVariableModal
				m.modalTarget = m.variables[m.variablesCursor].Name
			} else if m.focusedPanel == UsersPanel && len(m.users) > 0 {
				m.showModal = DeleteUserModal
				m.modalTarget = m.users[m.usersCursor].Email
			}
			return m, nil
		}
	}

	return m, nil
}

// handleModalKeys handles keyboard input when a modal is shown
func (m Model) handleModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		// Cancel deletion
		m.showModal = NoModal
		m.modalTarget = ""
		return m, nil

	case "y", "enter":
		// Confirm deletion
		if m.showModal == DeleteVariableModal {
			// Remove the variable
			m.variables = append(m.variables[:m.variablesCursor], m.variables[m.variablesCursor+1:]...)
			if m.variablesCursor >= len(m.variables) && m.variablesCursor > 0 {
				m.variablesCursor--
			}
		} else if m.showModal == DeleteUserModal {
			// Remove the user
			m.users = append(m.users[:m.usersCursor], m.users[m.usersCursor+1:]...)
			if m.usersCursor >= len(m.users) && m.usersCursor > 0 {
				m.usersCursor--
			}
		}
		m.showModal = NoModal
		m.modalTarget = ""
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate panel widths
	panelWidth := (m.width - 6) / 2

	// Render header
	header := m.renderHeader()

	// Render error banner if any
	var errorBanner string
	if m.err != nil {
		errorBanner = m.renderErrorBanner()
	}

	// Render panels
	leftPanel := m.renderVariablesPanel(panelWidth)
	rightPanel := m.renderUsersPanel(panelWidth)

	// Combine panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Render help text
	help := m.renderHelp()

	// Combine all sections
	var content string
	if errorBanner != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, header, errorBanner, panels, help)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, header, panels, help)
	}

	// Overlay modal if shown
	if m.showModal != NoModal {
		modal := m.renderModal()
		// Create a dimmed overlay by placing modal in center
		overlay := lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modal,
			lipgloss.WithWhitespaceChars("░"),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
		)
		return overlay
	}

	return content
}

// renderHeader renders the project and environment selector
func (m Model) renderHeader() string {
	// Compact one-line logo
	logo := renderCompactLogo()

	// Project section - label and value aligned
	projectLabel := headerStyle.Render("Project:")
	projectTag := tagStyle.Render(m.projectName)
	projectSection := lipgloss.JoinHorizontal(lipgloss.Bottom, projectLabel, " ", projectTag)

	// Environment section - label and values aligned
	envLabel := headerStyle.Render("Environment:")
	envTags := make([]string, len(m.environments))
	for i, env := range m.environments {
		if env.IsActive {
			envTags[i] = tagStyle.Render(env.Name)
		} else {
			envTags[i] = dimTagStyle.Render(env.Name)
		}
	}
	envTagsJoined := lipgloss.JoinHorizontal(lipgloss.Bottom, envTags...)
	envSection := lipgloss.JoinHorizontal(lipgloss.Bottom, envLabel, " ", envTagsJoined)

	// Combine all sections with proper spacing
	content := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		logo,
		"   ",
		projectSection,
		"   ",
		envSection,
	)

	// Wrap in styled header box with gray border
	headerBox := lipgloss.NewStyle().
		Background(lipgloss.Color("#1f1f1f")).
		Foreground(lipgloss.Color("#ffffff")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1, 2).
		Width(m.width - 4).
		Align(lipgloss.Left)

	return headerBox.Render(content)
}

// renderHelp renders help text at the bottom
func (m Model) renderHelp() string {
	helpText := "Use tab key to jump between the left and right sections | " +
		"Navigate horizontally between Environment names with keys: Arrow Left and Arrow Right | " +
		"Navigate vertically between Environment selection and variable rows with keys: Arrow Up and Arrow Down | " +
		"By pressing delete on selected row Delete prompt should appear | " +
		"Press Enter/Space to reveal/hide value | Press q to quit"

	if m.err != nil {
		helpText += " | Press Esc to dismiss error"
	}

	return helpStyle.Render(helpText)
}

// renderErrorBanner renders an error notification banner
func (m Model) renderErrorBanner() string {
	errorText := fmt.Sprintf("⚠ Error: %v", m.err)

	errorBannerStyle := lipgloss.NewStyle().
		Foreground(errorColor).
		Background(lipgloss.Color("#3D1F1F")).
		Padding(0, 2).
		Width(m.width - 4).
		Bold(true)

	return errorBannerStyle.Render(errorText)
}

// renderModal renders the delete confirmation modal
func (m Model) renderModal() string {
	var title string
	if m.showModal == DeleteVariableModal {
		title = fmt.Sprintf("Delete \"%s\"?", m.modalTarget)
	} else if m.showModal == DeleteUserModal {
		title = fmt.Sprintf("Remove user \"%s\"?", m.modalTarget)
	}

	prompt := lipgloss.JoinVertical(
		lipgloss.Center,
		errorStyle.Render(title),
		"",
		dimTagStyle.Render("✕ Close with Esc"),
		"",
		"Press Y to confirm or N to cancel",
	)

	return modalStyle.Render(prompt)
}
