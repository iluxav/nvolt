package types

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

type PullSecretsResponseDTO struct {
	Variables  map[string]string `json:"variables"`   // key -> encrypted value
	WrappedKey string            `json:"wrapped_key"` // wrapped master key for this machine
}

type ProjectEnvironment struct {
	ProjectName string `json:"project_name"`
	Environment string `json:"environment"`
}

type GetProjectEnvironmentsResponseDTO struct {
	ProjectEnvironments []ProjectEnvironment `json:"project_environments"`
}
