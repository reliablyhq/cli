package cmd

import (
	"fmt"
	"os"

	"github.com/reliablyhq/cli/core/color"
)

func er(msg interface{}) {
	fmt.Println(color.Red("Error:"), msg)
	os.Exit(1)
}
