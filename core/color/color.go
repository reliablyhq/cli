package color

import (
	"fmt"
	"os"
	"strings"

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
	// Magenta returns a purple-colored string
	Magenta = color.New(color.FgMagenta).SprintFunc()
	// Cyan returns a cyan-colored string
	Cyan = color.New(color.FgCyan).SprintFunc()
	// Underline returns an underline string
	Underline = color.New(color.Underline).SprintFunc()
	// Grey returns a dark grey-colored string
	Grey = color.New(color.FgHiBlack).SprintFunc()

	// Background colored string
	BgYellow  = color.New(color.BgYellow).SprintFunc()
	BgMagenta = color.New(color.BgMagenta).SprintFunc()
	BgRed     = color.New(color.BgRed).SprintFunc()
)

// Conditianal coloring - if the condition is true, the color is applied

func IfTrueRed(condition bool, a ...interface{}) string {
	if condition {
		return Red(a...)
	}

	return fmt.Sprint(a...)
}

func IfTrueGreen(condition bool, a ...interface{}) string {
	if condition {
		return Green(a...)
	}

	return fmt.Sprint(a...)
}

func IfTrueMagenta(condition bool, a ...interface{}) string {
	if condition {
		return Magenta(a...)
	}

	return fmt.Sprint(a...)
}

func IfTrueCyan(condition bool, a ...interface{}) string {
	if condition {
		return Cyan(a...)
	}

	return fmt.Sprint(a...)
}

func IfTrueYellow(condition bool, a ...interface{}) string {
	if condition {
		return Yellow(a...)
	}

	return fmt.Sprint(a...)
}

func Is256ColorSupported() bool {
	term := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	return strings.Contains(term, "256") ||
		strings.Contains(term, "24bit") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(colorterm, "256") ||
		strings.Contains(colorterm, "24bit") ||
		strings.Contains(colorterm, "truecolor")
}
