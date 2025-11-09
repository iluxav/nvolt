package ui

import "fmt"

// PrintBasicUsage prints usage instructions for local mode
func PrintBasicUsage() {
	Info(Gray("  • nvolt push -f .env"))
	Info(Gray("  • nvolt push -k ENV_KEY=<value>"))
	Info(Gray("  • nvolt pull"))
	fmt.Println()
	Info(Yellow("  Note: ") + Gray("In local mode, you manage Git operations manually"))
}

// PrintGlobalUsage prints usage instructions for global mode
func PrintGlobalUsage() {
	Info(Gray("  • nvolt push -f .env -p <project>"))
	Info(Gray("  • nvolt push -k ENV_KEY=<value> -p <project>"))
	Info(Gray("  • nvolt pull -p <project>"))
	fmt.Println()
	Info(Yellow("  Note: ") + Gray("nvolt automatically commits and pushes to the repository"))
}

// Section prints a section header
func Section(message string) {
	fmt.Printf("\n%s\n", Cyan(message))
}

// Substep prints a substep with indentation
func Substep(message string) {
	Info(Gray("  • ") + message)
}

// PrintDetected prints a detected value (e.g., project name)
func PrintDetected(label, value string) {
	Info(fmt.Sprintf("  %s: %s", Gray(label), Cyan(value)))
}

// PrintKeyValue prints a key-value pair
func PrintKeyValue(key, value string) {
	Info(fmt.Sprintf("  %s: %s", key, value))
}

// PrintModeInfo prints mode information
func PrintModeInfo(mode string) {
	Info(fmt.Sprintf("\n%s: %s", Gray("Mode"), Cyan(mode)))
}
