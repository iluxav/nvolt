package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptMachineName prompts the user to enter a custom machine name
// Returns empty string if user presses enter (use default)
func PromptMachineName() (string, error) {
	fmt.Println()
	fmt.Println(Bold("Machine Name Setup"))
	fmt.Println(Gray("Enter a custom name for this machine (or press Enter for auto-generated name):"))
	fmt.Print(Cyan("Machine name: "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Trim whitespace and newline
	name := strings.TrimSpace(input)

	// Validate name if provided
	if name != "" {
		// Remove invalid characters for filenames
		name = sanitizeMachineName(name)
		if name == "" {
			return "", fmt.Errorf("invalid machine name")
		}
	}

	return name, nil
}

// sanitizeMachineName removes invalid characters from machine name
func sanitizeMachineName(name string) string {
	// Replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		" ", "-",
	)

	sanitized := replacer.Replace(name)

	// Remove any leading/trailing dashes
	sanitized = strings.Trim(sanitized, "-")

	return sanitized
}
