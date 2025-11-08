package cli

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command can be executed
	err := rootCmd.Execute()
	// Should not error when just printing help
	if err != nil {
		t.Errorf("Root command returned error: %v", err)
	}
}

func TestCommandsRegistered(t *testing.T) {
	commands := []string{"init", "push", "pull", "run", "machine", "vault", "sync"}

	for _, cmdName := range commands {
		cmd, _, err := rootCmd.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Command %s not found: %v", cmdName, err)
			continue
		}
		if cmd == nil {
			t.Errorf("Command %s is nil", cmdName)
			continue
		}
		// cmd.Use may include args like "run [command]", so just check it starts with the name
		if len(cmd.Use) < len(cmdName) || cmd.Use[:len(cmdName)] != cmdName {
			t.Errorf("Expected command to start with %s, got %s", cmdName, cmd.Use)
		}
	}
}
