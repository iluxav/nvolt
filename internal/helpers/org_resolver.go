package helpers

import (
	"fmt"
	"iluxav/nvolt/internal/types"

	"github.com/AlecAivazis/survey/v2"
)

// ResolveActiveOrg determines which organization to use based on the following rules:
// 1. If user has only one org, return it
// 2. If user has multiple orgs and active_org is set in config, return it
// 3. If user has multiple orgs and no active_org, show interactive selector
func ResolveActiveOrg(userOrgs []*types.OrgUser, currentActiveOrgID string) (string, error) {
	if len(userOrgs) == 0 {
		return "", fmt.Errorf("user does not belong to any organization")
	}

	// If user has only one org, use it
	if len(userOrgs) == 1 {
		return userOrgs[0].OrgID, nil
	}

	// If active_org is set and valid, use it
	if currentActiveOrgID != "" {
		for _, org := range userOrgs {
			if org.OrgID == currentActiveOrgID {
				return currentActiveOrgID, nil
			}
		}
		// If active_org is invalid, continue to interactive selection
	}

	// Show interactive org selector
	return ShowOrgSelector(userOrgs)
}

// ShowOrgSelector displays an interactive organization selector
func ShowOrgSelector(userOrgs []*types.OrgUser) (string, error) {
	if len(userOrgs) == 0 {
		return "", fmt.Errorf("no organizations available")
	}

	// Build options list
	orgOptions := make([]string, len(userOrgs))
	orgMap := make(map[string]string)

	for i, orgUser := range userOrgs {
		orgName := orgUser.Org.Name
		if orgName == "" {
			orgName = orgUser.OrgID
		}
		displayName := fmt.Sprintf("%s (%s)", orgName, orgUser.Role)
		orgOptions[i] = displayName
		orgMap[displayName] = orgUser.OrgID
	}

	var selectedOrg string
	prompt := &survey.Select{
		Message: "Select an organization:",
		Options: orgOptions,
	}

	err := survey.AskOne(prompt, &selectedOrg)
	if err != nil {
		return "", fmt.Errorf("organization selection cancelled: %w", err)
	}

	return orgMap[selectedOrg], nil
}
