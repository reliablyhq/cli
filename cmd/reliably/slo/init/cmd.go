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
	"github.com/davecgh/go-spew/spew"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
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

var iconWarn = iostreams.WarningIcon()

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

	fmt.Println("Validate API Key for datadog")
	if ok, err := validateApiKey(); ok {
		fmt.Println("Authenticated to Datadog API")
	} else {
		fmt.Println("Error while validating DD API KEY", err)
	}

	var slosDD []*manifest.ServiceLevel = make([]*manifest.ServiceLevel, 0)

	fmt.Println("List SLOs from datadog")
	if r, err := ListDatadogSLOs(); err != nil {
		fmt.Println("Error with datadog", err)
	} else {
		fmt.Println("we found your SLOs")

		fmt.Println("-> errors", r.Errors)
		fmt.Println("-> metadata", r.Metadata.Page.GetTotalFilteredCount(), r.Metadata.Page.GetTotalCount())
		fmt.Println("-> data", r.GetData())
		for _, slo := range r.GetData() {

			for _, target := range slo.Thresholds {

				n := fmt.Sprintf("%s - %s ", slo.Name, target.Timeframe)
				fmt.Println("Name, Target, Timeframe")
				fmt.Println(n, target.Target, target.Timeframe)

				dur, _ := iso8601.FromString(strings.ToUpper(fmt.Sprintf("P%s", target.Timeframe)))

				slosDD = append(slosDD, &manifest.ServiceLevel{
					Name:      n,
					Type:      "mirror",
					Objective: target.Target,
					Indicators: []manifest.ServiceLevelIndicator{
						{
							ID:       *slo.Id,
							Provider: "datadog",
						},
					},
					ObservationWindow: core.Iso8601Duration{
						Duration: *dur,
					},
				})

			}
			spew.Dump(slo)
		}
	}

	var mDD manifest.Manifest = manifest.Manifest{

		Services: []*manifest.Service{
			{
				Name:          "datadog imported SLOs",
				ServiceLevels: slosDD,
			},
		},
	}

	fmt.Println("Manifest with SLOs imported from datadog")
	spew.Dump(mDD)

	fmt.Println("Get SLO history from datadog")
	for _, svc := range mDD.Services {
		fmt.Println("###", svc.Name)
		for _, slo := range svc.ServiceLevels {
			fmt.Println(">>>", slo.Name, slo.Objective, slo.ObservationWindow)
			sli := slo.Indicators[0]

			to := time.Now()
			from := to.Add(-slo.ObservationWindow.ToDuration())

			fmt.Println("GetSLOHistory (sloId, from, to, target)")
			fmt.Println(sli.ID, from.UTC().Unix(), to.UTC().Unix(), slo.Objective)

			if sloHist, err := GetSLOHistory(sli.ID, from, to, slo.Objective); err != nil {
				fmt.Println("Unable to fetch SLO history", err)
			} else {
				fmt.Println("we found the history -->")
				data := sloHist.GetData()
				var sliValue float64
				if data.Overall.SliValue != nil {
					sliValue = *data.Overall.SliValue
					fmt.Println("SLI", "=", sliValue, "%")

					if data.Overall.ErrorBudgetRemaining != nil {
						for k, v := range *data.Overall.ErrorBudgetRemaining {
							fmt.Println("Error budget for ", k, v, "%")
						}
					} else {
						fmt.Println("error budget is not available")
					}

				} else {
					fmt.Println("No SLI value computed ! ")
				}

				//spew.Dump(data)
			}

		}
	}

	return fmt.Errorf("Skip")

	log.Debugf("checking for existing service manifest: %s", manifestPath)
	if _, err := os.Stat(manifestPath); err == nil {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing local manifest detected (%s); Do you want to overwrite it?", manifestPath), emptyOptions, question.WithNoAsDefault) {
			return nil
		}
	}

	var m manifest.Manifest
	populateManifestInteractively(&m)

	// validate
	if err := m.Validate(); err != nil {
		return err
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

	fmt.Println()
	fmt.Println(iostreams.SuccessIcon(), "Your manifest has been saved to", manifestPath)
	log.Debugf("service manifest created at: %s", manifestPath)
	return nil
}

func populateManifestInteractively(m *manifest.Manifest) {

	var s manifest.Service
	serviceNameValidator := survey.WithValidator(func(v interface{}) error {
		for _, s := range m.Services {
			if s.Name == v.(string) {
				return fmt.Errorf("service mame [%v] already exists", v)
			}
		}
		return nil
	})

	s.Name = question.WithStringAnswerV2(
		"What is the name of the service you want to declare SLOs for?", "",
		s.Name, []survey.AskOpt{serviceNameValidator})

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
