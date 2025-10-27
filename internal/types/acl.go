package types

type OrgUser struct {
	ID          string           `json:"id"`
	OrgID       string           `json:"org_id"`
	UserID      string           `json:"user_id"`
	Role        string           `json:"role"`
	Org         *Org             `json:"org"`
	User        *User            `json:"user"`
	Permissions *UserPermissions `json:"-"` // Client-side only, fetched separately
}

// OrgUsersResponse is the response for listing organization users
type OrgUsersResponse struct {
	Users []*OrgUser `json:"users"`
}

type Org struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Permission represents permissions for read, write, delete operations
type Permission struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

// AddUserToOrgRequest is the request body for adding a user to an organization
type AddUserToOrgRequest struct {
	Email                  string      `json:"email"`
	ProjectName            string      `json:"project_name,omitempty"`
	Environment            string      `json:"environment,omitempty"`
	ProjectPermissions     *Permission `json:"project_permissions,omitempty"`
	EnvironmentPermissions *Permission `json:"environment_permissions,omitempty"`
}

// AddUserToOrgResponse is the response for adding a user to an organization
type AddUserToOrgResponse struct {
	Success               bool                 `json:"success"`
	Message               string               `json:"message"`
	User                  *User                `json:"user,omitempty"`
	ProjectEnvironments   []ProjectEnvironment `json:"project_environments,omitempty"`
	RequiresKeyRewrapping bool                 `json:"requires_key_rewrapping"`
}

// ModifyUserPermissionsRequest is the request body for modifying user permissions
type ModifyUserPermissionsRequest struct {
	Email                  string      `json:"email"`
	Role                   string      `json:"role,omitempty"`
	ProjectName            string      `json:"project_name,omitempty"`
	Environment            string      `json:"environment,omitempty"`
	ProjectPermissions     *Permission `json:"project_permissions,omitempty"`
	EnvironmentPermissions *Permission `json:"environment_permissions,omitempty"`
}

// ModifyUserPermissionsResponse is the response for modifying user permissions
type ModifyUserPermissionsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	User    *User  `json:"user,omitempty"`
}

// UserPermissions contains all permissions for a user in an organization
type UserPermissions struct {
	Email       string              `json:"email"`
	Role        string              `json:"role"`
	Projects    []ProjectPermission `json:"projects"`
	AllProjects []string            `json:"all_projects"`
	User        *User               `json:"user,omitempty"`
}

// ProjectPermission represents permissions for a specific project
type ProjectPermission struct {
	ProjectName     string                  `json:"project_name"`
	Permissions     Permission              `json:"permissions"`
	Environments    []EnvironmentPermission `json:"environments"`     // Environments the user has access to
	AllEnvironments []string                `json:"all_environments"` // All environments that exist for this project
}

// EnvironmentPermission represents permissions for a specific environment
type EnvironmentPermission struct {
	Environment string     `json:"environment"`
	Permissions Permission `json:"permissions"`
}
