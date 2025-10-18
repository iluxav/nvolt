package services

import (
	"context"
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
)

type MachineService struct {
	config *types.MachineLocalConfig
}

func MachineServiceFromContext(ctx context.Context) *MachineService {
	return ctx.Value(types.MachineServiceKey).(*MachineService)
}

func NewMachineService(config *types.MachineLocalConfig) *MachineService {
	return &MachineService{
		config: config,
	}
}

// SaveMachineKey saves a machine's public key to the server
func (s *MachineService) SaveMachineKey(orgID string, req *types.SaveMachinePublicKeyRequestDTO) error {

	machineKeyURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.config.ServerURL, orgID)

	resp, err := helpers.CallAPIWithPayload[types.SaveMachinePublicKeyResponseDTO](
		machineKeyURL,
		"POST",
		s.config.JWT_Token,
		req,
		s.config.MachineID,
	)
	if err != nil || !resp.Success {
		return fmt.Errorf("failed to save machine key: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Message)
	}

	return nil
}

// GetOrgMachines fetches all machines in an organization
func (s *MachineService) GetOrgMachines(orgID string) ([]types.MachineKeyDTO, error) {
	machinesURL := fmt.Sprintf("%s/api/v1/organizations/%s/machines", s.config.ServerURL, orgID)

	type Response struct {
		Machines []types.MachineKeyDTO `json:"machines"`
	}

	resp, err := helpers.CallAPI[Response](machinesURL, "GET", s.config.JWT_Token, s.config.MachineID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch machines: %w", err)
	}

	return resp.Machines, nil
}
