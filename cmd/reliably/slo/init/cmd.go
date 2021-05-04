package init

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	iso8601 "github.com/ChannelMeter/iso8601duration"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	manifestPath        string
	supportedExtensions = []string{".yaml", ".json"}
	providersMap        = map[string]string{
		"Amazon Web Services":   "aws",
		"Google Cloud Platform": "gcp",
	}
)

var emptyOptions = []question.AskOpt{}

const iconWarn = "⚠️"

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:     "init",
		Short:   "initialise the slo portion of the manifest",
		Long:    longCommandDescription(),
		Example: examples(),
		RunE:    runE,
	}

	cmd.Flags().StringVarP(&manifestPath, "output", "o", "./reliably.yaml", "store a local copy of the service manifest created")
	return &cmd
}

func runE(_ *cobra.Command, args []string) error {
	log.Debug("fetching internal service manifest")
	if _, err := os.Stat(manifestPath); err == nil {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing local manifest detected (%s); Do you want to overwrite it?", manifestPath), emptyOptions, question.WithNoAsDefault) {
			return nil
		}
	}

	m, err := manifest.Load(manifestPath)
	client := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	if err != nil {
		log.Debugf("error reading local manifest, attempting to retrieve from reliably: %s", err)
		m, err = api.PullManifest(client)
		if err != nil {
			return err
		}

	}

	if m == nil {
		log.Debug("no service manifest detected, creating a new one")
		m = &manifest.Manifest{}
	}

	populateManifestInteractively(m)

	// validate
	if err := m.Validate(); err != nil {
		return err
	}

	// push manifestto backend
	if err := api.PushManifest(client, m); err != nil {
		return fmt.Errorf("an error occurred while push manifest to reliably: %s", err)
	}

	// write file output
	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		return err
	}

	log.Infof("service manifest created at: %s", manifestPath)
	return nil
}

func populateManifestInteractively(m *manifest.Manifest) {

	var s manifest.Service

	s.Name = question.WithStringAnswer("What is the name of the service you want to declare SLOs for?", emptyOptions)

	declareSLOForService(&s)

	m.Services = append(m.Services, &s)
	fmt.Println(color.Green(fmt.Sprintf("Service '%s' added", s.Name)))

	fmt.Println()
	if question.WithBoolAnswer("Do you want to add another Service?", emptyOptions, question.WithNoAsDefault) {
		populateManifestInteractively(m)
	}

}

func declareSLOForService(s *manifest.Service) {
	var sl manifest.ServiceLevel

	slType := question.WithSingleChoiceAnswer("What type of SLO do you want to declare?", emptyOptions, "Availability", "Latency")
	sl.Type = sanitizeString(slType)

	sl.Objective = question.WithFloat64Answer("What is your target for this SLO (in %)?", emptyOptions, 0, 100)

	if sl.Type == "latency" {
		threshold := question.WithDurationAnswer("What is your latency threshold (in milliseconds)?", emptyOptions)
		sl.Criteria = manifest.LatencyCriteria{Threshold: threshold}
	}

	sl.ObservationWindow = getObservationWindow()

	do := question.WithBoolAnswer("Do you want to add a resource for measuring your SLI?", emptyOptions, question.WithYesAsDefault)

	if do {
		providers := []string{}
		for key := range providersMap {
			providers = append(providers, key)
		}
		sort.Strings(providers) // sorts slice in-place

		for do {
			providerFullName := question.WithSingleChoiceAnswer("On which cloud provider?", emptyOptions, providers...)
			provider := providersMap[providerFullName]
			id := getResourceIDForProvider(provider)

			if id != "" { // We're returning empty strings when something fails...
				sl.Indicators = append(sl.Indicators, manifest.ServiceLevelIndicator{
					Provider: provider,
					ID:       id,
				})
			}

			fmt.Println()
			do = question.WithBoolAnswer("Do you want to add another resource for measuring your SLI?", emptyOptions, question.WithNoAsDefault)
		}
	}
	_ = initDefaultSloName(&sl)
	sl.Name = question.WithStringAnswerV2("What is the name of this SLO?", "", sl.Name, emptyOptions)
	s.ServiceLevels = append(s.ServiceLevels, &sl)
	fmt.Println(color.Green(fmt.Sprintf("SLO '%s' added to Service '%s'", sl.Name, s.Name)))

	fmt.Println()
	if question.WithBoolAnswer("Do you want to add another SLO?", emptyOptions, question.WithNoAsDefault) {
		declareSLOForService(s)
	}
}

func getResourceIDForProvider(provider string) string {
	switch provider {
	case "aws":
		return buildAWSArn()
	case "gcp":
		return buildGCPResourceID()
	default:
		return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.", emptyOptions)
	}
}

func getObservationWindow() core.Iso8601Duration {

	const (
		oneHour  = "1 hour"
		oneDay   = "1 day"
		oneWeek  = "1 week"
		oneMonth = "1 month"
		custom   = "custom"
	)

	choices := []string{
		oneHour,
		oneDay,
		oneWeek,
		oneMonth,
		custom,
	}

	var choice string

	p := &survey.Select{
		Message: "What is your observation window for this SLO?",
		Options: choices,
	}

	err := survey.AskOne(p, &choice, survey.WithValidator(survey.Required))
	checkPromptExit(err)

	answers := struct {
		Window core.Iso8601Duration
	}{}

	switch choice {
	case oneHour:
		answers.Window = core.Iso8601Duration{Duration: iso8601.Duration{Hours: 1}}
	case oneDay:
		answers.Window = core.Iso8601Duration{Duration: iso8601.Duration{Hours: 24}}
	case oneWeek:
		answers.Window = core.Iso8601Duration{Duration: iso8601.Duration{Weeks: 1}}
	case oneMonth:
		answers.Window = core.Iso8601Duration{Duration: iso8601.Duration{Days: 30}}
	case custom:

		q := []*survey.Question{
			{
				Name: "window",
				Prompt: &survey.Input{
					Message: "Define your custom observation window",
					Help:    "Must be an iso8601 duration with the following format: P[n]DT[n]H[n]M or P[n]W as (D)ays, (H)ours, (M)inutes, (W)eeks",
				},
				Validate: survey.ComposeValidators(survey.Required, func(val interface{}) error {
					str := strings.ToUpper(val.(string))
					d, err := iso8601.FromString(str)
					if err != nil {
						return fmt.Errorf("Unable to parse your string: %s", err)
					}
					if d.Seconds > 0 && math.Mod(float64(d.Seconds), 60) != 0 {
						return fmt.Errorf("We only support precision to 1 minute. If used, seconds must be a multiple of 60.")
					}
					duration := d.ToDuration()
					if duration == 0 {
						return errors.New("Your duration cannot be zero. Please check your format.")
					}

					if duration > time.Hour*24*365 {
						return errors.New("Your duration cannot exceed 1 year.")
					}
					return nil
				}),
				Transform: survey.ComposeTransformers(
					survey.TransformString(strings.ToUpper),
					func(ans interface{}) interface{} {
						d, _ := iso8601.FromString(ans.(string))
						return core.Iso8601Duration{Duration: *d}
					},
				),
			},
		}

		err := survey.Ask(q, &answers)
		checkPromptExit(err)

	}

	return answers.Window
}

func sanitizeString(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}

func checkPromptExit(err error) {
	if err == terminal.InterruptErr {
		os.Exit(0)
	}
}
