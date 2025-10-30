package services

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"

	"github.com/AlecAivazis/survey/v2"
)

type ACLService struct {
	config *types.MachineLocalConfig
}

func ACLServiceFromContext(ctx context.Context) *ACLService {
	return ctx.Value(types.ACLServiceKey).(*ACLService)
}

func NewACLService(config *types.MachineLocalConfig) *ACLService {
	return &ACLService{
		config: config,
	}
}

// GetUserOrgs fetches all organizations the current user belongs to
func (s *ACLService) GetUserOrgs() ([]*types.OrgUser, error) {

	userOrgsURL := fmt.Sprintf("%s/api/v1/user/orgs", s.config.ServerURL)

	orgUsers, err := helpers.CallAPI[[]*types.OrgUser](userOrgsURL, "GET", s.config.JWT_Token, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user organizations: %w", err)
	}

	return *orgUsers, nil
}

func (s *ACLService) GetActiveOrgName(orgID string) (string, string, error) {
	userOrgs, err := s.GetUserOrgs()
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch organizations: %w", err)
	}
	for _, org := range userOrgs {
		if org.OrgID == orgID {

			return org.Org.Name, org.Role, nil
		}
	}
	return "", "", fmt.Errorf("organization with ID %s not found", orgID)
}

// AddUserToOrg adds a user to an organization with optional permissions
func (s *ACLService) AddUserToOrg(orgID string, req *types.AddUserToOrgRequest) (*types.AddUserToOrgResponse, error) {
	addUserURL := fmt.Sprintf("%s/api/v1/organizations/%s/users", s.config.ServerURL, orgID)

	response, err := helpers.CallAPIWithPayload[types.AddUserToOrgResponse, types.AddUserToOrgRequest](addUserURL, "POST", s.config.JWT_Token, req, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	return response, nil
}

// RemoveUserFromOrg removes a user from an organization
func (s *ACLService) RemoveUserFromOrg(orgID string, userEmail string) error {

	// First, list all users to find the user_id by email
	users, err := s.GetOrgUsers(orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization users: %w", err)
	}

	var userID string
	for _, orgUser := range users {
		if orgUser.User != nil && orgUser.User.Email == userEmail {
			userID = orgUser.User.ID
			break
		}
	}

	if userID == "" {
		return fmt.Errorf("user with email %s not found in organization", userEmail)
	}

	removeUserURL := fmt.Sprintf("%s/api/v1/organizations/%s/users/%s", s.config.ServerURL, orgID, userID)

	_, err = helpers.CallAPI[map[string]interface{}](removeUserURL, "DELETE", s.config.JWT_Token, s.config.MachineID)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}

	return nil
}

// GetOrgUsers lists all users in an organization
func (s *ACLService) GetOrgUsers(orgID string) ([]*types.OrgUser, error) {

	listUsersURL := fmt.Sprintf("%s/api/v1/organizations/%s/users", s.config.ServerURL, orgID)

	response, err := helpers.CallAPI[types.OrgUsersResponse](listUsersURL, "GET", s.config.JWT_Token, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization users: %w", err)
	}

	return response.Users, nil
}

// GetUserPermissions retrieves all permissions for a user in an organization
func (s *ACLService) GetUserPermissions(orgID string, userEmail string) (*types.UserPermissions, error) {

	// First, list all users to find the user_id by email
	users, err := s.GetOrgUsers(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization users: %w", err)
	}

	var userID string
	for _, orgUser := range users {
		if orgUser.User != nil && orgUser.User.Email == userEmail {
			userID = orgUser.User.ID
			break
		}
	}

	if userID == "" {
		return nil, fmt.Errorf("user with email %s not found in organization", userEmail)
	}

	getUserPermsURL := fmt.Sprintf("%s/api/v1/organizations/%s/users/%s", s.config.ServerURL, orgID, userID)

	permissions, err := helpers.CallAPI[types.UserPermissions](getUserPermsURL, "GET", s.config.JWT_Token, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// ModifyUserPermissions modifies user permissions in an organization
func (s *ACLService) ModifyUserPermissions(orgID string, req *types.ModifyUserPermissionsRequest) (*types.ModifyUserPermissionsResponse, error) {
	// First, list all users to find the user_id by email
	users, err := s.GetOrgUsers(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization users: %w", err)
	}

	var userID string
	for _, orgUser := range users {
		if orgUser.User != nil && orgUser.User.Email == req.Email {
			userID = orgUser.User.ID
			break
		}
	}

	if userID == "" {
		return nil, fmt.Errorf("user with email %s not found in organization", req.Email)
	}

	modifyUserURL := fmt.Sprintf("%s/api/v1/organizations/%s/users/%s", s.config.ServerURL, orgID, userID)

	response, err := helpers.CallAPIWithPayload[types.ModifyUserPermissionsResponse, types.ModifyUserPermissionsRequest](modifyUserURL, "PATCH", s.config.JWT_Token, req, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to modify user permissions: %w", err)
	}

	return response, nil
}

// ResolveOrgID smartly resolves which org to use based on user's orgs and active_org_id
// Returns: (orgID, orgName, shouldSave, error)
// - If user has 1 org: auto-select it, don't save as default
// - If user has multiple orgs + active_org_id set: use it
// - If user has multiple orgs + no active_org_id: prompt user and ask if they want to set as default
func (s *ACLService) ResolveOrgID(machineConfig *MachineConfig) (string, string, bool, error) {
	// Fetch all user orgs
	userOrgs, err := s.GetUserOrgs()
	if err != nil {
		return "", "", false, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	if len(userOrgs) == 0 {
		return "", "", false, fmt.Errorf("user has no organizations")
	}

	// Case 1: User has only one org - auto-select it (no prompt, no save)
	if len(userOrgs) == 1 {
		return userOrgs[0].OrgID, userOrgs[0].Org.Name, false, nil
	}

	// Case 2: Multiple orgs + active_org_id is set - use it
	if machineConfig.Config.ActiveOrgID != "" {
		// Verify the active org still exists
		for _, org := range userOrgs {
			if org.OrgID == machineConfig.Config.ActiveOrgID {
				return org.OrgID, org.Org.Name, false, nil
			}
		}
		// If active_org_id doesn't exist anymore, fall through to prompt
		fmt.Println("⚠ Your default organization no longer exists. Please select a new one.")
	}

	// Case 3: Multiple orgs + no active_org_id - prompt user
	orgNames := make([]string, len(userOrgs))
	orgMap := make(map[string]*types.OrgUser)
	for i, org := range userOrgs {
		displayName := fmt.Sprintf("%s (%s)", org.Org.Name, org.Role)
		orgNames[i] = displayName
		orgMap[displayName] = org
	}

	var selectedOrgName string
	prompt := &survey.Select{
		Message: "Select an organization:",
		Options: orgNames,
	}
	if err := survey.AskOne(prompt, &selectedOrgName); err != nil {
		return "", "", false, fmt.Errorf("organization selection cancelled")
	}

	selectedOrg := orgMap[selectedOrgName]

	// Ask if user wants to set as default
	var setAsDefault bool
	confirmPrompt := &survey.Confirm{
		Message: "Set this organization as default? (You won't be prompted again)",
		Default: false,
	}
	if err := survey.AskOne(confirmPrompt, &setAsDefault); err != nil {
		// If they cancel, just continue without setting default
		setAsDefault = false
	}

	return selectedOrg.OrgID, selectedOrg.Org.Name, setAsDefault, nil
}
