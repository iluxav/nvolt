package main

import (
	"iluxav/nvolt/internal/cli"
	"iluxav/nvolt/internal/services"
	"os"
)

func main() {
	config := services.NewMachineConfigService()
	aclService := services.NewACLService(config.Config)

	if err := cli.Execute(config, aclService); err != nil {
		os.Exit(1)
	}
}
