package services

import (
	"fmt"
	"iluxav/nvolt/internal/helpers"
	"iluxav/nvolt/internal/types"
)

type MachineService struct {
	config *types.MachineLocalConfig
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
	)
	if err != nil {
		return fmt.Errorf("failed to save machine key: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Message)
	}

	return nil
}
