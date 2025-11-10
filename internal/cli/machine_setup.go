package cli

import (
	"github.com/iluxav/nvolt/internal/ui"
	"github.com/iluxav/nvolt/internal/vault"
)

// EnsureMachineInitialized ensures the machine is initialized with keypair and name
// Prompts for custom machine name if this is the first initialization
func EnsureMachineInitialized() error {
	// Check if machine is already initialized
	initialized, err := vault.IsMachineInitialized()
	if err != nil {
		return err
	}

	if initialized {
		return nil
	}

	// First time setup - prompt for machine name
	ui.Info("")
	ui.Info(ui.Gold("⚙️  First-time Machine Setup"))
	ui.Info("This machine needs to be initialized before using nvolt.")
	ui.Info("")

	customName, err := ui.PromptMachineName()
	if err != nil {
		return err
	}

	// Initialize machine with custom name (or empty for auto-generated)
	machineInfo, err := vault.InitializeMachine(customName)
	if err != nil {
		return err
	}

	ui.Success("Machine initialized successfully")
	ui.Info(ui.Cyan("Machine ID: ") + ui.Bold(machineInfo.ID))
	ui.Info("")

	return nil
}
