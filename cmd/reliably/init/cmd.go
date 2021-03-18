package init

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	supportedExtensions = []string{".yaml", ".json"}
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init",
		Short: "initialise reliably",
		// Long: "Lorem ipsum...",
		Run: run,
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest-file", "f", "reliably.yaml", "the location of the manifest file")

	return &cmd
}

func run(_ *cobra.Command, _ []string) {
	validateFilePath()

	m := manifest.Empty()

	populateManifest(m)

	f, err := os.Create(manifestPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		log.Fatal(err)
	}
}

func validateFilePath() {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(manifestPath, ext) {
			return
		}
	}

	log.Fatalf("manifest file must have one of the these extensions: %v", supportedExtensions)
}

func populateManifest(m *manifest.Manifest) {
	scanner := bufio.NewScanner(os.Stdin)
	ok := false

	m.ApplicationName = askQuestion(scanner, "what is the name of your application?")
	m.CI.Type = askQuestion(scanner, "What type of CI are you using?")

	for !ok {
		targetAvailabilityStr := askQuestion(scanner, "what percentage of availability do you want your application to have?")
		if f, err := strconv.ParseFloat(targetAvailabilityStr, 32); err != nil {
			fmt.Println("Please make sure you type a numner")
		} else {
			if f < 0 || f > 100 {
				fmt.Println("the value must be between 0 and 100")
			} else {
				m.ServiceLevel.Availability = f
				ok = true
			}
		}
	}
}

func askQuestion(scanner *bufio.Scanner, questionText string) string {
	var text string

	for len(text) == 0 {
		fmt.Println(questionText)
		scanner.Scan()
		text = scanner.Text()
	}

	return text
}
