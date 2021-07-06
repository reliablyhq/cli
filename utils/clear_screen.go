package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func ClearScreen() {
	var c *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "cls")
	default:
		// clear should work for UNIX & linux based systems
		c = exec.Command("clear")

		// hide cursor on unix based systems
		fmt.Print("\033[?25l")
	}

	c.Stdout = os.Stdout
	c.Run()
}
