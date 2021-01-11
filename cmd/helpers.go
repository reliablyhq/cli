package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func er(msg interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Println(red("Error:"), msg)
	os.Exit(1)
}
