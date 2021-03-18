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

	m := manifest.New()

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

	m.App.Name = askQuestion(scanner, "What is the name of your application?")
	m.App.Owner = askQuestion(scanner, "Who owns this app? Its a good idea to put an email address in here!")
	m.App.Repository = askQuestion(scanner, "What is the URL to the repository for this app?")

	if askQuestionWithBoolAnswer(scanner, "Are you using Continuous Integration?") {
		m.CI.Type = askQuestion(scanner, "What type of CI are you using?")
	} else {
		m.CI = nil
	}

	if askQuestionWithBoolAnswer(scanner, "Are you building something that will be provided to customers 'as a service'?") {
		m.ServiceLevel.Availability = askQuestionWithFloat64Answer(scanner, "What percentage of availability do you want your application to have?", 0, 100)
		m.ServiceLevel.ErrorBudgetPercent = askQuestionWithFloat64Answer(scanner, "What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100)
		m.ServiceLevel.Latency = askQuestionWithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service")
	} else {
		m.ServiceLevel = nil
	}

	if askQuestionWithBoolAnswer(scanner, "Will your app be hosted on a commercial platform?") {
		m.Hosting = &manifest.Hosting{
			Provider: askQuestion(scanner, "What is the name of the provider?"),
		}
	} else {
		m.Hosting = nil
	}

	if askQuestionWithBoolAnswer(scanner, "Does your application have 'service level' dependencies?") {
		deps := make([]*manifest.Dependency, 0)

		addMore := true
		for addMore {
			d := &manifest.Dependency{
				Name: askQuestion(scanner, "what is the name of the dependency?"),
			}

			deps = append(deps, d)

			addMore = askQuestionWithBoolAnswer(scanner, "Do you want to add another dependency?")
		}

		m.Dependencies = deps
	} else {
		m.Dependencies = nil
	}

	if askQuestionWithBoolAnswer(scanner, "Are you using Infrastucture as Code?") {
		m.IAC = &manifest.IAC{
			Type: askQuestion(scanner, "What IAC provider are you using (terraform, ARM templates, CDK, etc...)?"),
			Root: askQuestion(scanner, "Where is the root folder for your IAC code?"),
		}
	} else {
		m.IAC = nil
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

func askQuestionWithBoolAnswer(scanner *bufio.Scanner, question string) bool {
	for {
		answer := askQuestion(scanner, question)
		if b, err := strconv.ParseBool(answer); err == nil {
			return b
		} else {
			// do some noddy-level parsing
			lAnswer := strings.ToLower(answer)
			if lAnswer == "y" || lAnswer == "yes" {
				return true
			} else if lAnswer == "n" || lAnswer == "no" {
				return false
			}

			fmt.Println("the answer you gave could not be parsed to a boolean")
		}
	}
}
