package init

import (
	"bufio"
	"os"
	"strings"

	"github.com/reliablyhq/cli/core/cli/question"
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

	m.App.Name = question.WithStringAnswer(scanner, "What is the name of your application?")
	m.App.Owner = question.WithStringAnswer(scanner, "Who owns this app? Its a good idea to put an email address in here!")
	m.App.Repository = question.WithStringAnswer(scanner, "What is the URL to the repository for this app?")

	if question.WithBoolAnswer(scanner, "Are you using Continuous Integration?") {
		m.CI.Type = question.WithStringAnswer(scanner, "What type of CI are you using?")
	} else {
		m.CI = nil
	}

	if question.WithBoolAnswer(scanner, "Are you building something that will be provided to customers 'as a service'?") {
		m.ServiceLevel.Availability = question.WithFloat64Answer(scanner, "What percentage of availability do you want your application to have?", 0, 100)
		m.ServiceLevel.ErrorBudgetPercent = question.WithFloat64Answer(scanner, "What percentage of requests to your service is it ok to have fail? This will be your 'error budget'.", 0, 100)
		m.ServiceLevel.Latency = question.WithDurationAnswer(scanner, "What is the maximum request-response latency you want from this service")
	} else {
		m.ServiceLevel = nil
	}

	if question.WithBoolAnswer(scanner, "Will your app be hosted on a commercial platform?") {
		m.Hosting = &manifest.Hosting{
			Provider: question.WithStringAnswer(scanner, "What is the name of the provider?"),
		}
	} else {
		m.Hosting = nil
	}

	if question.WithBoolAnswer(scanner, "Does your application have 'service level' dependencies?") {
		deps := make([]*manifest.Dependency, 0)

		addMore := true
		for addMore {
			d := &manifest.Dependency{
				Name: question.WithStringAnswer(scanner, "what is the name of the dependency?"),
			}

			deps = append(deps, d)

			addMore = question.WithBoolAnswer(scanner, "Do you want to add another dependency?")
		}

		m.Dependencies = deps
	} else {
		m.Dependencies = nil
	}

	if question.WithBoolAnswer(scanner, "Are you using Infrastucture as Code?") {
		m.IAC = &manifest.IAC{
			Type: question.WithStringAnswer(scanner, "What IAC provider are you using (terraform, ARM templates, CDK, etc...)?"),
			Root: question.WithStringAnswer(scanner, "Where is the root folder for your IAC code?"),
		}
	} else {
		m.IAC = nil
	}
}
