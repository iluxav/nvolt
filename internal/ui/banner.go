package ui

import (
	"fmt"
	"io"
	"os"
)

const logoPlain = `
    ███╗   ██╗██╗   ██╗ ██████╗ ██╗  ████████╗
    ████╗  ██║██║   ██║██╔═══██╗██║  ╚══██╔══╝
    ██╔██╗ ██║██║   ██║██║   ██║██║     ██║
    ██║╚██╗██║╚██╗ ██╔╝██║   ██║██║     ██║
    ██║ ╚████║ ╚████╔╝ ╚██████╔╝███████╗██║
    ╚═╝  ╚═══╝  ╚═══╝   ╚═════╝ ╚══════╝╚═╝
`

// getLogo returns the logo with color if enabled
func getLogo() string {
	if IsColorsEnabled() {
		return Gold(logoPlain)
	}
	return logoPlain
}

// PrintLogo prints the nvolt logo in yellow/gold
func PrintLogo() {
	PrintLogoTo(os.Stdout)
}

// PrintLogoTo prints the nvolt logo to a specific writer
func PrintLogoTo(w io.Writer) {
	fmt.Fprint(w, getLogo())
}

// PrintLogoWithVersion prints the logo with version information
func PrintLogoWithVersion(version string) {
	PrintLogo()
	fmt.Printf("\n    %s • %s\n\n",
		Cyan("Zero-Trust Secret Manager"),
		Gray("v"+version))
}

// PrintBanner prints a banner with the logo and a message
func PrintBanner(message string) {
	PrintLogo()
	fmt.Printf("\n    %s\n\n", Cyan(message))
}
