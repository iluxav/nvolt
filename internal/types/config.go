package types

type MachineLocalConfig struct {
	MachineID          string `json:"machine_id"`
	ServerURL          string `json:"server_url"`
	JWT_Token          string `json:"jwt_token"`
	ActiveOrgID        string `json:"active_org_id,omitempty"`
	DefaultEnvironment string `json:"default_environment,omitempty"`
}
