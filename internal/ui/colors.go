package ui

import (
	"fmt"
	"os"
)

// Color codes using ANSI escape sequences
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"

	// Bright colors
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"

	// Bold
	ColorBold = "\033[1m"

	// Gold/Orange (256 color)
	ColorGold = "\033[38;5;220m"
)

var colorsEnabled = true

// SetColorsEnabled enables or disables color output
func SetColorsEnabled(enabled bool) {
	colorsEnabled = enabled
}

// IsColorsEnabled returns whether colors are enabled
func IsColorsEnabled() bool {
	return colorsEnabled
}

// DisableColorsIfNotTTY disables colors if stdout is not a TTY
func DisableColorsIfNotTTY() {
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// Not a terminal, disable colors
		colorsEnabled = false
	}
}

// Colorize wraps text with a color code
func Colorize(color, text string) string {
	if !colorsEnabled {
		return text
	}
	return color + text + ColorReset
}

// Yellow returns yellow colored text
func Yellow(text string) string {
	return Colorize(ColorYellow, text)
}

// Gold returns gold colored text
func Gold(text string) string {
	return Colorize(ColorGold, text)
}

// Green returns green colored text
func Green(text string) string {
	return Colorize(ColorGreen, text)
}

// BrightGreen returns bright green colored text
func BrightGreen(text string) string {
	return Colorize(ColorBrightGreen, text)
}

// Red returns red colored text
func Red(text string) string {
	return Colorize(ColorRed, text)
}

// BrightRed returns bright red colored text
func BrightRed(text string) string {
	return Colorize(ColorBrightRed, text)
}

// Blue returns blue colored text
func Blue(text string) string {
	return Colorize(ColorBlue, text)
}

// Cyan returns cyan colored text
func Cyan(text string) string {
	return Colorize(ColorCyan, text)
}

// Gray returns gray colored text
func Gray(text string) string {
	return Colorize(ColorGray, text)
}

// Bold returns bold text
func Bold(text string) string {
	return Colorize(ColorBold, text)
}

// Printf with color support
func Colorf(color, format string, args ...interface{}) {
	fmt.Print(Colorize(color, fmt.Sprintf(format, args...)))
}

// Println with color support
func Colorln(color, text string) {
	fmt.Println(Colorize(color, text))
}

func init() {
	// Auto-detect if we should use colors
	DisableColorsIfNotTTY()

	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		colorsEnabled = false
	}
}
