package types

import "encoding/json"

type SaveMachinePublicKeyRequestDTO struct {
	MachineID string `json:"machine_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type SaveMachinePublicKeyResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CLISessionContext struct {
	Project     string `json:"project"`
	Environment string `json:"environment"`
}

type ProjectResolver struct {
	ProjectName string
	ProjectType string
}

type MachineKeyDTO struct {
	ID        string `json:"id"`
	MachineID string `json:"machine_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	UserID    string `json:"user_id"`
	User      *User  `json:"user"`
	CreatedAt string `json:"created_at"`
}

type GetMachinesResponseDTO struct {
	Success  bool            `json:"success"`
	Machines []MachineKeyDTO `json:"machines"`
}

type PushSecretsRequestDTO struct {
	OrgID        string            `json:"org_id"` // User specifies which org
	MachineKeyID string            `json:"machine_key_id"`
	ProjectName  string            `json:"project_name"`
	Environment  string            `json:"environment"`
	Variables    map[string]string `json:"variables"`
	WrappedKeys  map[string]string `json:"wrapped_keys"`
	ReplaceAll   bool              `json:"replace_all"` // true = full replacement (-f), false = partial update (-k)
}

type PushSecretsResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VariableWithMetadata struct {
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

// UnmarshalJSON handles backward compatibility - if we receive a plain string, treat it as the Value
func (v *VariableWithMetadata) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a plain string first (old format)
	var plainString string
	if err := json.Unmarshal(data, &plainString); err == nil {
		v.Value = plainString
		v.CreatedAt = ""
		return nil
	}

	// Otherwise, unmarshal as the full metadata object (new format)
	type Alias VariableWithMetadata
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(v),
	}
	return json.Unmarshal(data, &aux)
}

type PullSecretsResponseDTO struct {
	Variables  map[string]VariableWithMetadata `json:"variables"`   // key -> encrypted value with metadata
	WrappedKey string                          `json:"wrapped_key"` // wrapped master key for this machine
}

type ProjectEnvironment struct {
	ProjectName string `json:"project_name"`
	Environment string `json:"environment"`
}

type GetProjectEnvironmentsResponseDTO struct {
	ProjectEnvironments []ProjectEnvironment `json:"project_environments"`
}
