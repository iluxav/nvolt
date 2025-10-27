package services

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/crypto"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
	"net/url"
)

type SecretsClient struct {
	machineConfig *MachineConfig
}

func SecretsClientFromContext(ctx context.Context) *SecretsClient {
	return ctx.Value(types.SecretsClientKey).(*SecretsClient)
}

func NewSecretsClient(machineConfig *MachineConfig) *SecretsClient {
	return &SecretsClient{
		machineConfig: machineConfig,
	}
}

func (s *SecretsClient) PushSecrets(projectName, environment string, variables map[string]string, replaceAll bool) error {
	// Get active org ID from config
	orgID := s.machineConfig.Config.ActiveOrgID
	if orgID == "" {
		return fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
	}

	// Step 1: Fetch all machine public keys from server
	machinesURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.machineConfig.Config.ServerURL, orgID)
	machinesResp, err := helpers.CallAPI[types.GetMachinesResponseDTO](machinesURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return fmt.Errorf("failed to fetch machines: %w", err)
	}

	// Get current machine key ID
	var currentMachineKeyID string
	for _, machine := range machinesResp.Machines {
		if machine.MachineID == s.machineConfig.Config.MachineID {
			currentMachineKeyID = machine.ID
			break
		}
	}

	if currentMachineKeyID == "" {
		return fmt.Errorf("current machine not found in authorized machines")
	}

	// Step 2: For partial updates, fetch existing encrypted vars and merge with new ones
	varsToEncrypt := variables
	if !replaceAll {
		// Fetch existing secrets to merge with new ones
		pullURL := fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets?machine_key_id=%s",
			s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment), currentMachineKeyID)

		pullResp, err := helpers.CallAPI[types.PullSecretsResponseDTO](pullURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
		if err != nil {
			// If no existing secrets found, that's OK - just use new variables
			fmt.Printf("Note: No existing secrets found, creating new project scope\n")
		} else if pullResp.WrappedKey != "" {
			// Unwrap the existing master key
			privateKey, err := s.machineConfig.GetPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get private key: %w", err)
			}
			oldMasterKey, err := crypto.UnwrapMasterKey(privateKey, pullResp.WrappedKey)
			if err != nil {
				return fmt.Errorf("failed to unwrap existing master key: %w", err)
			}
			defer func() {
				// Clear old master key from memory
				for i := range oldMasterKey {
					oldMasterKey[i] = 0
				}
			}()

			// Decrypt all existing variables
			existingVars := make(map[string]string)
			for key, varMeta := range pullResp.Variables {
				decryptedValue, err := crypto.DecryptWithMasterKey(oldMasterKey, varMeta.Value)
				if err != nil {
					return fmt.Errorf("failed to decrypt existing variable %s: %w", key, err)
				}
				existingVars[key] = decryptedValue
			}

			// Merge: new variables override existing ones
			varsToEncrypt = existingVars
			for k, v := range variables {
				varsToEncrypt[k] = v // Override or add new
			}

			fmt.Printf("\nMerged %d existing variable(s) with %d new variable(s)\n", len(existingVars), len(variables))
		}
	}

	// Step 3: Generate a NEW master key for encryption (key rotation on every push)
	masterKey, err := crypto.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}
	defer func() {
		// Clear master key from memory
		for i := range masterKey {
			masterKey[i] = 0
		}
	}()

	// Step 4: Wrap the NEW master key with each machine's public key
	wrappedKeys := make(map[string]string)
	for _, machine := range machinesResp.Machines {
		wrappedKey, err := crypto.WrapMasterKey(machine.PublicKey, masterKey)
		if err != nil {
			// Log warning but continue with other machines
			fmt.Printf("Warning: Failed to wrap key for machine %s: %v\n", machine.Name, err)
			continue
		}
		wrappedKeys[machine.ID] = wrappedKey
	}

	if len(wrappedKeys) == 0 {
		return fmt.Errorf("failed to wrap master key for any machines")
	}

	// Step 5: Encrypt ALL variables (existing + new) with the NEW master key
	encryptedVars := make(map[string]string)
	for key, value := range varsToEncrypt {
		encryptedValue, err := crypto.EncryptWithMasterKey(masterKey, value)
		if err != nil {
			return fmt.Errorf("failed to encrypt variable %s: %w", key, err)
		}
		encryptedVars[key] = encryptedValue
	}

	// Step 6: Push to server with transaction (server handles concurrency control)
	pushURL := fmt.Sprintf(
		"%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets",
		s.machineConfig.Config.ServerURL,
		orgID,
		url.PathEscape(projectName),
		url.PathEscape(environment),
	)

	payload := types.PushSecretsRequestDTO{
		MachineKeyID: currentMachineKeyID,
		Variables:    encryptedVars,
		WrappedKeys:  wrappedKeys,
		ReplaceAll:   true, // Always do full replacement now since we re-encrypted everything
	}

	pushResp, err := helpers.CallAPIWithPayload[types.PushSecretsResponseDTO, types.PushSecretsRequestDTO](
		pushURL,
		"POST",
		s.machineConfig.Config.JWT_Token,
		&payload,
		s.machineConfig.Config.MachineID,
	)
	if err != nil {
		return fmt.Errorf("failed to push secrets: %w", err)
	}

	if !pushResp.Success {
		return fmt.Errorf("push failed: %s", pushResp.Message)
	}

	return nil
}

func (s *SecretsClient) PullSecrets(projectName, environment, specificKey string) (map[string]string, error) {
	// Get active org ID from config
	orgID := s.machineConfig.Config.ActiveOrgID
	if orgID == "" {
		return nil, fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
	}

	// Get current machine key ID
	machinesURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.machineConfig.Config.ServerURL, orgID)
	machinesResp, err := helpers.CallAPI[types.GetMachinesResponseDTO](machinesURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch machines: %w", err)
	}

	var currentMachineKeyID string
	for _, machine := range machinesResp.Machines {
		if machine.MachineID == s.machineConfig.Config.MachineID {
			currentMachineKeyID = machine.ID
			break
		}
	}

	if currentMachineKeyID == "" {
		return nil, fmt.Errorf("current machine not found in authorized machines")
	}

	// Build pull URL with proper RESTful structure
	var pullURL string
	if specificKey != "" {
		// Pull specific secret
		pullURL = fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets/%s?machine_key_id=%s",
			s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment), url.PathEscape(specificKey), currentMachineKeyID)
	} else {
		// Pull all secrets
		pullURL = fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets?machine_key_id=%s",
			s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment), currentMachineKeyID)
	}

	// Fetch encrypted secrets from server
	pullResp, err := helpers.CallAPI[types.PullSecretsResponseDTO](pullURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to pull secrets: %w", err)
	}

	// Unwrap the master key using machine's private key
	privateKey, err := s.machineConfig.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}
	masterKey, err := crypto.UnwrapMasterKey(privateKey, pullResp.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap master key: %w", err)
	}
	defer func() {
		// Clear master key from memory
		for i := range masterKey {
			masterKey[i] = 0
		}
	}()

	// Decrypt all variables using the master key
	decryptedVars := make(map[string]string)
	for key, varMeta := range pullResp.Variables {
		decryptedValue, err := crypto.DecryptWithMasterKey(masterKey, varMeta.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt variable %s: %w", key, err)
		}
		decryptedVars[key] = decryptedValue
	}

	return decryptedVars, nil
}

// SyncKeys re-wraps the master key for all machines in the org without modifying secrets
// This is useful after adding a new machine to enable it to access existing secrets
func (s *SecretsClient) SyncKeys(orgID string, projectName string, environment string) error {
	// Step 1: Fetch all machines for the org
	machinesURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.machineConfig.Config.ServerURL, orgID)
	machinesResp, err := helpers.CallAPI[types.GetMachinesResponseDTO](machinesURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return fmt.Errorf("failed to fetch machines: %w", err)
	}

	// Get current machine key ID
	var currentMachineKeyID string
	for _, machine := range machinesResp.Machines {
		if machine.MachineID == s.machineConfig.Config.MachineID {
			currentMachineKeyID = machine.ID
			break
		}
	}

	if currentMachineKeyID == "" {
		return fmt.Errorf("current machine not found in authorized machines")
	}

	// Step 2: Pull existing secrets to get the current master key
	pullURL := fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets?machine_key_id=%s",
		s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment), currentMachineKeyID)

	pullResp, err := helpers.CallAPI[types.PullSecretsResponseDTO](pullURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing secrets: %w", err)
	}

	if pullResp.WrappedKey == "" {
		return fmt.Errorf("no secrets found for this project/environment. Nothing to sync")
	}

	// Step 3: Unwrap the current master key
	privateKey, err := s.machineConfig.GetPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get private key: %w", err)
	}
	masterKey, err := crypto.UnwrapMasterKey(privateKey, pullResp.WrappedKey)
	if err != nil {
		return fmt.Errorf("failed to unwrap master key: %w", err)
	}
	defer func() {
		// Clear master key from memory
		for i := range masterKey {
			masterKey[i] = 0
		}
	}()

	// Step 4: Re-wrap the master key for ALL machines (including newly added ones)
	wrappedKeys := make(map[string]string)
	for _, machine := range machinesResp.Machines {
		wrappedKey, err := crypto.WrapMasterKey(machine.PublicKey, masterKey)
		if err != nil {
			fmt.Printf("Warning: Failed to wrap key for machine %s: %v\n", machine.Name, err)
			continue
		}
		wrappedKeys[machine.ID] = wrappedKey
	}

	if len(wrappedKeys) == 0 {
		return fmt.Errorf("failed to wrap master key for any machines")
	}

	fmt.Printf("Re-wrapped keys for %d machine(s)\n", len(wrappedKeys))

	// Step 5: Upload ONLY the wrapped keys (no secret changes)
	// We'll reuse the existing encrypted variables (extract just values without metadata)
	existingVars := make(map[string]string)
	for key, varMeta := range pullResp.Variables {
		existingVars[key] = varMeta.Value
	}

	pushURL := fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets", s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment))
	requestDTO := types.PushSecretsRequestDTO{
		Variables:    existingVars, // Reuse existing encrypted variables
		WrappedKeys:  wrappedKeys,
		ReplaceAll:   true,
		MachineKeyID: currentMachineKeyID,
	}

	_, err = helpers.CallAPIWithPayload[types.PushSecretsResponseDTO](pushURL, "POST", s.machineConfig.Config.JWT_Token, &requestDTO, s.machineConfig.Config.MachineID)
	if err != nil {
		return fmt.Errorf("failed to upload wrapped keys: %w", err)
	}

	return nil
}

// PullSecretsWithMetadata returns decrypted variables along with their metadata (creation date)
func (s *SecretsClient) PullSecretsWithMetadata(projectName, environment string) (map[string]types.VariableWithMetadata, error) {
	// Get active org ID from config
	orgID := s.machineConfig.Config.ActiveOrgID
	if orgID == "" {
		return nil, fmt.Errorf("no active organization set. Please run 'nvolt org set' first")
	}

	// Get current machine key ID
	machinesURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.machineConfig.Config.ServerURL, orgID)
	machinesResp, err := helpers.CallAPI[types.GetMachinesResponseDTO](machinesURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch machines: %w", err)
	}

	var currentMachineKeyID string
	for _, machine := range machinesResp.Machines {
		if machine.MachineID == s.machineConfig.Config.MachineID {
			currentMachineKeyID = machine.ID
			break
		}
	}

	if currentMachineKeyID == "" {
		return nil, fmt.Errorf("current machine not found in authorized machines")
	}

	// Pull all secrets
	pullURL := fmt.Sprintf("%s/api/v1/organizations/%s/projects/%s/environments/%s/secrets?machine_key_id=%s",
		s.machineConfig.Config.ServerURL, orgID, url.PathEscape(projectName), url.PathEscape(environment), currentMachineKeyID)

	// Fetch encrypted secrets from server
	pullResp, err := helpers.CallAPI[types.PullSecretsResponseDTO](pullURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to pull secrets: %w", err)
	}

	// Unwrap the master key using machine's private key
	privateKey, err := s.machineConfig.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}
	masterKey, err := crypto.UnwrapMasterKey(privateKey, pullResp.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap master key: %w", err)
	}
	defer func() {
		// Clear master key from memory
		for i := range masterKey {
			masterKey[i] = 0
		}
	}()

	// Decrypt all variables using the master key and preserve metadata
	decryptedVars := make(map[string]types.VariableWithMetadata)
	for key, varMeta := range pullResp.Variables {
		decryptedValue, err := crypto.DecryptWithMasterKey(masterKey, varMeta.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt variable %s: %w", key, err)
		}
		decryptedVars[key] = types.VariableWithMetadata{
			Value:     decryptedValue,
			CreatedAt: varMeta.CreatedAt,
		}
	}

	return decryptedVars, nil
}

// GetProjectEnvironments fetches all project/environment combinations the user has access to
func (s *SecretsClient) GetProjectEnvironments(orgID string) ([]types.ProjectEnvironment, error) {
	projectEnvsURL := fmt.Sprintf("%s/api/v1/organizations/%s/environments", s.machineConfig.Config.ServerURL, orgID)

	resp, err := helpers.CallAPI[types.GetProjectEnvironmentsResponseDTO](projectEnvsURL, "GET", s.machineConfig.Config.JWT_Token, s.machineConfig.Config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project environments: %w", err)
	}

	return resp.ProjectEnvironments, nil
}
