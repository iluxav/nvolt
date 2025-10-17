package types

type contextKey string

const (
	MachineConfigKey  contextKey = "machine_config"
	ACLServiceKey     contextKey = "acl_service"
	SecretsClientKey  contextKey = "secrets_client"
	MachineServiceKey contextKey = "machine_service"
)
