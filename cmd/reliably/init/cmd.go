package init

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/reliablyhq/cli/core"
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

	m.ApplicationName = askQuestion(scanner, "What is the name of your application?")
	m.CI.Type = askQuestion(scanner, "What type of CI are you using?")

	m.ServiceLevel.Availability = askQuestionWithFloat64Answer(scanner, "What percentage of availability do you want your application to have?", 0, 100)
	m.ServiceLevel.ErrorBudgetPercent = askQuestionWithFloat64Answer(scanner, "What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100)
	m.ServiceLevel.Latency = askQuestionWithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service")
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

func askQuestionWithFloat64Answer(scanner *bufio.Scanner, question string, min, max float64) float64 {
	for {
		answer := askQuestion(scanner, question)
		if f, err := strconv.ParseFloat(answer, 32); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			if f < min || f > max {
				fmt.Printf("the value must be between %.2f and %.2f\n", min, max)
			} else {
				return f
			}
		}
	}
}

func askQuestionWithDurationAnswer(scanner *bufio.Scanner, question string) core.Duration {
	for {
		answer := askQuestion(scanner, question)
		if d, err := time.ParseDuration(answer); err != nil {
			fmt.Println("The value you entered could not be parsed to a duration.")
		} else {
			return core.Duration{Duration: d}
		}
	}
}
