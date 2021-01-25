package color

import (
	"github.com/fatih/color"
)

var (
	// Bold returns a bold string
	Bold = color.New(color.Bold).SprintFunc()
	// Red returns a red-colored string
	Red = color.New(color.FgRed).SprintFunc()
	// Green returns a green-colored string
	Green = color.New(color.FgGreen).SprintFunc()
	// Yellow returns a yellow-colored string
	Yellow = color.New(color.FgYellow).SprintFunc()
)
