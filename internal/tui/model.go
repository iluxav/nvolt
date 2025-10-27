package tui

import (
	"fmt"
	"iluxav/nvolt/internal/services"
	"iluxav/nvolt/internal/types"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the main TUI model
type Model struct {
	// Services
	machineConfig *services.MachineConfig
	secretsClient *services.SecretsClient
	aclService    *services.ACLService

	// Data
	organizations      []Organization
	activeOrgIndex     int
	projects           []Project
	activeProjectIndex int
	environments       []Environment
	activeEnvIndex     int
	variables          []EnvVariable
	users              []User

	// UI state
	focusedPanel     FocusedPanel
	activeTab        RightPanelTab
	projectsCursor   int
	variablesCursor  int
	usersCursor      int
	showModal        ModalType
	modalTarget      string // Name of the item to delete
	permissionEditor PermissionEditor
	width            int
	height           int
	err              error
	loading          bool
}

// NewModel creates a new TUI model
func NewModel(
	machineConfig *services.MachineConfig,
	secretsClient *services.SecretsClient,
	aclService *services.ACLService,
	projectName string,
) Model {
	// Load organizations from machineConfig
	organizations := []Organization{}
	activeOrgIndex := 0
	if machineConfig != nil && machineConfig.OrgUsers != nil {
		for i, orgUser := range machineConfig.OrgUsers {
			if orgUser.Org != nil {
				organizations = append(organizations, Organization{
					ID:   orgUser.Org.ID,
					Name: orgUser.Org.Name,
				})
				// Set active org based on ActiveOrgID in config
				if machineConfig.Config != nil && orgUser.Org.ID == machineConfig.Config.ActiveOrgID {
					activeOrgIndex = i
				}
			}
		}
	}

	return Model{
		machineConfig:      machineConfig,
		secretsClient:      secretsClient,
		aclService:         aclService,
		organizations:      organizations,
		activeOrgIndex:     activeOrgIndex,
		projects:           []Project{},
		activeProjectIndex: 0,
		environments:       []Environment{},
		activeEnvIndex:     0,
		variables:          []EnvVariable{},
		users:              []User{},
		focusedPanel:       ProjectsPanel,
		activeTab:          VariablesTab,
		projectsCursor:     0,
		variablesCursor:    0,
		usersCursor:        0,
		showModal:          NoModal,
		loading:            true, // Start with loading state
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Load projects first, which will load environments for the first project
	return m.loadProjects()
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loadProjectsMsg:
		// Projects have been loaded
		if msg.err != nil {
			m.loading = false
			m.err = msg.err
			return m, nil
		}

		// Update projects
		m.projects = msg.projects
		if len(m.projects) > 0 {
			m.activeProjectIndex = 0
			m.projectsCursor = 0

			// Load environments for the first project
			if len(m.projects[0].Environments) > 0 {
				m.environments = make([]Environment, len(m.projects[0].Environments))
				for i, envName := range m.projects[0].Environments {
					m.environments[i] = Environment{
						Name:     envName,
						IsActive: i == 0,
					}
				}
				m.activeEnvIndex = 0
			}
		}

		// Now load data for the first project/environment
		return m, m.loadData()

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
		if msg.users != nil && len(m.projects) > 0 && len(m.environments) > 0 {
			projectName := m.projects[m.activeProjectIndex].Name
			envName := m.environments[m.activeEnvIndex].Name
			m.users = convertToUsers(msg.users, projectName, envName)
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
			// Cycle through panels: Projects -> Right Panel (Variables/Users)
			switch m.focusedPanel {
			case ProjectsPanel:
				m.focusedPanel = RightPanel
			case RightPanel:
				m.focusedPanel = ProjectsPanel
			}
			return m, nil

		case "v":
			// Direct shortcut to Variables tab
			m.activeTab = VariablesTab
			m.focusedPanel = RightPanel
			return m, nil

		case "u":
			// Direct shortcut to Users tab
			m.activeTab = UsersTab
			m.focusedPanel = RightPanel
			return m, nil

		case "left", "right":
			// When in right panel, switch between tabs
			// Otherwise, navigate between environments
			if m.focusedPanel == RightPanel {
				if msg.String() == "left" {
					m.activeTab = VariablesTab
				} else {
					m.activeTab = UsersTab
				}
				return m, nil
			}

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
			switch m.focusedPanel {
			case ProjectsPanel:
				if m.projectsCursor > 0 {
					m.projectsCursor--
				}
			case RightPanel:
				// Navigate based on active tab
				if m.activeTab == VariablesTab {
					if m.variablesCursor > 0 {
						m.variablesCursor--
					}
				} else {
					if m.usersCursor > 0 {
						m.usersCursor--
					}
				}
			}
			return m, nil

		case "down", "j":
			switch m.focusedPanel {
			case ProjectsPanel:
				if m.projectsCursor < len(m.projects)-1 {
					m.projectsCursor++
				}
			case RightPanel:
				// Navigate based on active tab
				if m.activeTab == VariablesTab {
					if m.variablesCursor < len(m.variables)-1 {
						m.variablesCursor++
					}
				} else {
					if m.usersCursor < len(m.users)-1 {
						m.usersCursor++
					}
				}
			}
			return m, nil

		case "enter", " ":
			// Handle Enter based on focused panel
			if m.focusedPanel == ProjectsPanel && len(m.projects) > 0 {
				// Select project and reload environments and data
				oldProjectIndex := m.activeProjectIndex
				m.activeProjectIndex = m.projectsCursor

				if oldProjectIndex != m.activeProjectIndex {
					// Load environments for the new project
					if len(m.projects[m.activeProjectIndex].Environments) > 0 {
						m.environments = make([]Environment, len(m.projects[m.activeProjectIndex].Environments))
						for i, envName := range m.projects[m.activeProjectIndex].Environments {
							m.environments[i] = Environment{
								Name:     envName,
								IsActive: i == 0,
							}
						}
						m.activeEnvIndex = 0
					}

					// Reload data for the new project
					m.loading = true
					m.variablesCursor = 0
					m.usersCursor = 0
					return m, m.loadData()
				}
				return m, nil

			} else if m.focusedPanel == RightPanel && m.activeTab == VariablesTab && len(m.variables) > 0 {
				// Copy variable value to clipboard
				selectedVar := m.variables[m.variablesCursor]
				err := clipboard.WriteAll(selectedVar.Value)
				if err != nil {
					m.err = fmt.Errorf("failed to copy to clipboard: %w", err)
				}
			} else if m.focusedPanel == RightPanel && m.activeTab == UsersTab && len(m.users) > 0 {
				// Open permission editor for selected user
				selectedUser := m.users[m.usersCursor]
				m.showModal = EditUserPermissionsModal

				// Initialize permission editor with current user permissions
				projectPerms := selectedUser.ProjectPermissions
				if projectPerms == nil {
					projectPerms = &types.Permission{Read: false, Write: false, Delete: false}
				}
				envPerms := selectedUser.EnvironmentPermissions
				if envPerms == nil {
					envPerms = &types.Permission{Read: false, Write: false, Delete: false}
				}

				m.permissionEditor = PermissionEditor{
					UserEmail:              selectedUser.Email,
					ProjectPermissions:     *projectPerms,
					EnvironmentPermissions: *envPerms,
					FocusedSection:         0,
					FocusedPermission:      0,
				}
			}
			return m, nil

		case "delete", "d":
			// Show delete modal based on active tab
			if m.focusedPanel == RightPanel {
				if m.activeTab == VariablesTab && len(m.variables) > 0 {
					m.showModal = DeleteVariableModal
					m.modalTarget = m.variables[m.variablesCursor].Name
				} else if m.activeTab == UsersTab && len(m.users) > 0 {
					m.showModal = DeleteUserModal
					m.modalTarget = m.users[m.usersCursor].Email
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// handleModalKeys handles keyboard input when a modal is shown
func (m Model) handleModalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle permission editor separately
	if m.showModal == EditUserPermissionsModal {
		return m.handlePermissionEditorKeys(msg)
	}

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
			// Remove the user from organization via API
			orgID := m.machineConfig.Config.ActiveOrgID
			err := m.aclService.RemoveUserFromOrg(orgID, m.modalTarget)
			if err != nil {
				m.err = fmt.Errorf("failed to remove user: %w", err)
				m.showModal = NoModal
				m.modalTarget = ""
				return m, nil
			}

			// Remove from local list
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

// handlePermissionEditorKeys handles keyboard input in the permission editor modal
func (m Model) handlePermissionEditorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel editing
		m.showModal = NoModal
		return m, nil

	case "tab":
		// Switch between project and environment permissions sections
		m.permissionEditor.FocusedSection = (m.permissionEditor.FocusedSection + 1) % 2
		m.permissionEditor.FocusedPermission = 0 // Reset to first permission
		return m, nil

	case "up", "k":
		// Navigate up in permissions
		if m.permissionEditor.FocusedPermission > 0 {
			m.permissionEditor.FocusedPermission--
		}
		return m, nil

	case "down", "j":
		// Navigate down in permissions
		if m.permissionEditor.FocusedPermission < 2 { // 0=read, 1=write, 2=delete
			m.permissionEditor.FocusedPermission++
		}
		return m, nil

	case " ", "enter":
		// Toggle the focused permission
		if m.permissionEditor.FocusedSection == 0 {
			// Project permissions
			switch m.permissionEditor.FocusedPermission {
			case 0:
				m.permissionEditor.ProjectPermissions.Read = !m.permissionEditor.ProjectPermissions.Read
			case 1:
				m.permissionEditor.ProjectPermissions.Write = !m.permissionEditor.ProjectPermissions.Write
			case 2:
				m.permissionEditor.ProjectPermissions.Delete = !m.permissionEditor.ProjectPermissions.Delete
			}
		} else {
			// Environment permissions
			switch m.permissionEditor.FocusedPermission {
			case 0:
				m.permissionEditor.EnvironmentPermissions.Read = !m.permissionEditor.EnvironmentPermissions.Read
			case 1:
				m.permissionEditor.EnvironmentPermissions.Write = !m.permissionEditor.EnvironmentPermissions.Write
			case 2:
				m.permissionEditor.EnvironmentPermissions.Delete = !m.permissionEditor.EnvironmentPermissions.Delete
			}
		}
		return m, nil

	case "s", "ctrl+s":
		// Save permissions to server
		if len(m.projects) == 0 || len(m.environments) == 0 {
			m.err = fmt.Errorf("no active project or environment")
			m.showModal = NoModal
			return m, nil
		}

		orgID := m.machineConfig.Config.ActiveOrgID
		projectName := m.projects[m.activeProjectIndex].Name
		envName := m.environments[m.activeEnvIndex].Name

		req := &types.ModifyUserPermissionsRequest{
			Email:                  m.permissionEditor.UserEmail,
			ProjectName:            projectName,
			Environment:            envName,
			ProjectPermissions:     &m.permissionEditor.ProjectPermissions,
			EnvironmentPermissions: &m.permissionEditor.EnvironmentPermissions,
		}

		_, err := m.aclService.ModifyUserPermissions(orgID, req)
		if err != nil {
			m.err = fmt.Errorf("failed to save permissions: %w", err)
			m.showModal = NoModal
			return m, nil
		}

		// Update the user in the users list locally
		for i := range m.users {
			if m.users[i].Email == m.permissionEditor.UserEmail {
				m.users[i].ProjectPermissions = &m.permissionEditor.ProjectPermissions
				m.users[i].EnvironmentPermissions = &m.permissionEditor.EnvironmentPermissions
				break
			}
		}

		m.showModal = NoModal
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate panel widths for 2 columns
	// Account for borders and padding between panels
	projectsPanelWidth := m.width / 6                   // ~17% for projects
	rightPanelWidth := m.width - projectsPanelWidth - 6 // Remaining width minus borders/padding

	// Render header
	header := m.renderHeader()

	// Render error banner if any
	var errorBanner string
	if m.err != nil {
		errorBanner = m.renderErrorBanner()
	}

	// Render panels
	projectsPanel := m.renderProjectsPanel(projectsPanelWidth)
	rightPanel := m.renderRightPanel(rightPanelWidth)

	// Combine panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, projectsPanel, rightPanel)

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

	// Organization, project and environment info
	var orgName, projectName, envName string
	if len(m.organizations) > 0 && m.activeOrgIndex < len(m.organizations) {
		orgName = m.organizations[m.activeOrgIndex].Name
	} else {
		orgName = "Loading..."
	}

	if len(m.projects) > 0 && m.activeProjectIndex < len(m.projects) {
		projectName = m.projects[m.activeProjectIndex].Name
	} else {
		projectName = "Loading..."
	}

	if len(m.environments) > 0 && m.activeEnvIndex < len(m.environments) {
		envName = m.environments[m.activeEnvIndex].Name
	} else {
		envName = "Loading..."
	}

	// Organization section
	orgLabel := headerStyle.Render("Organization:")
	orgTag := tagStyle.Render(orgName)
	orgSection := lipgloss.JoinHorizontal(lipgloss.Bottom, orgLabel, " ", orgTag)

	// Project section
	projectLabel := headerStyle.Render("Project:")
	projectTag := tagStyle.Render(projectName)
	projectSection := lipgloss.JoinHorizontal(lipgloss.Bottom, projectLabel, " ", projectTag)

	// Environment section
	envLabel := headerStyle.Render("Environment:")
	envTag := tagStyle.Render(envName)
	envSection := lipgloss.JoinHorizontal(lipgloss.Bottom, envLabel, " ", envTag)

	// Combine all sections with proper spacing
	content := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		logo,
		"   ",
		orgSection,
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
	helpText := "Tab: Switch panels | " +
		"v/u: Jump to Variables/Users | " +
		"←/→: Switch tabs (in right panel) or environments | " +
		"↑/↓: Navigate items | " +
		"Enter: Select project or copy value to clipboard | " +
		"Delete: Remove item | " +
		"q: Quit"

	if m.err != nil {
		helpText += " | Esc: Dismiss error"
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
	if m.showModal == EditUserPermissionsModal {
		return m.renderPermissionEditor()
	}

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

// renderPermissionEditor renders the permission editing modal
func (m Model) renderPermissionEditor() string {
	title := fmt.Sprintf("Edit Permissions for %s", m.permissionEditor.UserEmail)

	// Helper function to render a checkbox
	checkbox := func(checked bool, focused bool) string {
		box := "[ ]"
		if checked {
			box = "[✓]"
		}
		if focused {
			return selectedRowStyle.Render(box)
		}
		return box
	}

	// Project permissions section
	projectFocused := m.permissionEditor.FocusedSection == 0
	projectTitle := "Project Permissions"
	if projectFocused {
		projectTitle = titleStyle.Render("▶ " + projectTitle)
	} else {
		projectTitle = dimTagStyle.Render("  " + projectTitle)
	}

	projectRead := checkbox(
		m.permissionEditor.ProjectPermissions.Read,
		projectFocused && m.permissionEditor.FocusedPermission == 0,
	) + " Read"
	projectWrite := checkbox(
		m.permissionEditor.ProjectPermissions.Write,
		projectFocused && m.permissionEditor.FocusedPermission == 1,
	) + " Write"
	projectDelete := checkbox(
		m.permissionEditor.ProjectPermissions.Delete,
		projectFocused && m.permissionEditor.FocusedPermission == 2,
	) + " Delete"

	// Environment permissions section
	envFocused := m.permissionEditor.FocusedSection == 1
	envTitle := "Environment Permissions"
	if envFocused {
		envTitle = titleStyle.Render("▶ " + envTitle)
	} else {
		envTitle = dimTagStyle.Render("  " + envTitle)
	}

	envRead := checkbox(
		m.permissionEditor.EnvironmentPermissions.Read,
		envFocused && m.permissionEditor.FocusedPermission == 0,
	) + " Read"
	envWrite := checkbox(
		m.permissionEditor.EnvironmentPermissions.Write,
		envFocused && m.permissionEditor.FocusedPermission == 1,
	) + " Write"
	envDelete := checkbox(
		m.permissionEditor.EnvironmentPermissions.Delete,
		envFocused && m.permissionEditor.FocusedPermission == 2,
	) + " Delete"

	prompt := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(title),
		"",
		projectTitle,
		"  " + projectRead,
		"  " + projectWrite,
		"  " + projectDelete,
		"",
		envTitle,
		"  " + envRead,
		"  " + envWrite,
		"  " + envDelete,
		"",
		dimTagStyle.Render("Tab: Switch section | ↑/↓: Navigate | Space/Enter: Toggle | S: Save | Esc: Cancel"),
	)

	return modalStyle.Render(prompt)
}
