package services

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
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
