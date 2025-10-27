package tui

import (
	"iluxav/nvolt/internal/types"
	"time"
)

// FocusedPanel represents which panel is currently focused
type FocusedPanel int

const (
	VariablesPanel FocusedPanel = iota
	UsersPanel
)

// EnvVariable represents an environment variable
type EnvVariable struct {
	Name       string
	Value      string
	Created    time.Time
	IsRevealed bool // Track if the value is shown in plain text
}

// User represents a user with permissions
type User struct {
	Name                   string
	Email                  string
	OrgRole                string
	UserID                 string
	ProjectPermissions     *types.Permission
	EnvironmentPermissions *types.Permission
}

// Environment represents an environment (default, staging, production, etc.)
type Environment struct {
	Name     string
	IsActive bool
}

// ModalType represents different types of modals
type ModalType int

const (
	NoModal ModalType = iota
	DeleteVariableModal
	DeleteUserModal
)

// Messages for Bubble Tea
type loadDataMsg struct {
	variables map[string]types.VariableWithMetadata
	users     []*types.OrgUser
	err       error
}

type loadEnvironmentsMsg struct {
	environments []string
	err          error
}

type errorMsg struct {
	err error
}
