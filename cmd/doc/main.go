package main

import (
	"fmt"
	"os"

	"github.com/reliablyhq/cli/cmd/doc/doc"
	reliably "github.com/reliablyhq/cli/cmd/reliably"
)

func main() {
	reliablyCmd := reliably.NewCmdRoot()

	docCmd := doc.NewCmdDoc(reliablyCmd)
	if err := docCmd.Execute(); err != nil {
		fmt.Println("Unable to generate the doc")
		os.Exit(1)
	}
}
