package initAlpha

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
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	//"github.com/reliablyhq/cli/api"
	initCmd "github.com/reliablyhq/cli/cmd/reliably/slo/init"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	//v "github.com/reliablyhq/cli/version"
)

var (
	emptyOptions = []question.AskOpt{}
	iconWarn     = iostreams.WarningIcon()
	providersMap = map[string]string{
		"Amazon Web Services":   "aws",
		"Google Cloud Platform": "gcp",
	}
)

func AlpaInitRun(opts *initCmd.InitOptions) error {

	manifestPath := opts.ManifestPath
	var entities []entities.Entity = make([]entities.Entity, 0)

	log.Debugf("checking for existing service manifest: %s", manifestPath)
	if _, err := os.Stat(manifestPath); err == nil {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing local manifest detected (%s); Do you want to overwrite it?", manifestPath), emptyOptions, question.WithNoAsDefault) {
			return nil
		}
	}

	objectives, err := promptForServices(opts.IO)
	if err != nil {
		return err
	}

	for _, o := range objectives {
		entities = append(entities, o)
	}

	/*
		// validate
		if err := m.Validate(); err != nil {
			return err
		}
	*/

	// write file output
	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()

	ye := yaml.NewEncoder(f)
	for _, e := range entities {
		if err := ye.Encode(e); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println(iostreams.SuccessIcon(), "Your manifest has been saved to", manifestPath)
	log.Debugf("manifest created at: %s", manifestPath)
	return nil

	//////
	/*
		var sliSelect entities.Labels = entities.Labels{
			"abc": "123",
		}

		slo := entities.Objective{
			TypeMeta: entities.TypeMeta{APIVersion: "v1", Kind: "Objective"},
			Metadata: entities.Metadata{
				Name: "SLO for tests",
				Labels: map[string]string{
					"name": "SLO for tests",
					"env":  "test",
				},
				RelatedTo: []map[string]string{
					{"any": "entity"},
					{"more": "complex", "relation": "entity"},
				},
			},

			Spec: entities.ObjectiveSpec{
				IndicatorSelector: entities.Selector(sliSelect),
				ObjectivePercent:  90,
				Window:            core.Duration{Duration: time.Duration(24 * time.Hour)},
			},
		}

		hostname := core.Hostname()
		entityHost := core.Hostname()
		if v.IsDevVersion() {
			if hostFromEnv := os.Getenv("RELIABLY_ENTITY_HOST"); hostFromEnv != "" {
				entityHost = hostFromEnv
			}
		}

		apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))
		org, _ := api.CurrentUserOrganization(apiClient, hostname)

		return api.CreateEntity(apiClient, entityHost, org.Name, slo)
	*/

	//return errors.New("This is alpha version of the 'slo init' command")
}

func promptForServices(io *iostreams.IOStreams) ([]entities.Objective, error) {

	var objectives []entities.Objective = make([]entities.Objective, 0)
	w := io.Out

	askForService := true
	for askForService {
		name := question.WithStringAnswer(
			"What is the name of the service you want to declare SLOs for?", emptyOptions)

		serviceObjectives, err := promptForObjectives(io, name)
		if err != nil {
			return objectives, err
		}
		objectives = append(objectives, serviceObjectives...)

		fmt.Fprintln(w, color.Green(fmt.Sprintf("Service '%s' added", name)))

		fmt.Fprintln(w)
		askForService = question.WithBoolAnswer("Do you want to add another Service?", emptyOptions, question.WithNoAsDefault)
	}

	return objectives, nil
}

func promptForObjectives(io *iostreams.IOStreams, serviceName string) ([]entities.Objective, error) {
	var objectives []entities.Objective = make([]entities.Objective, 0)
	w := io.Out

	askForObjective := true
	for askForObjective {
		objective := *entities.NewObjective()

		// ask users for objective data
		objective.Spec.ObjectivePercent = question.WithFloat64Answer("What is your target for this SLO (in %)?", emptyOptions, 0, 100)

		slType := question.WithSingleChoiceAnswer("What type of SLO do you want to declare?", emptyOptions, "Availability", "Latency")
		slType = sanitizeString(slType)
		objective.Spec.IndicatorSelector["category"] = slType

		if slType == "latency" {
			threshold := question.WithDurationAnswer("What is your latency threshold (in milliseconds)?", emptyOptions)
			objective.Spec.IndicatorSelector["latency_target"] = fmt.Sprint(threshold)

			// should we prompt for the percentile as well ? ...
			objective.Spec.IndicatorSelector["percentile"] = "99"
		}

		objective.Spec.Window = core.Duration{Duration: getObservationWindow().ToDuration()}

		providerLabels, _ := promptForProvider(io)
		for k, v := range providerLabels {
			objective.Spec.IndicatorSelector[k] = v
		}

		defaultSloName := generateDefaultSloName(objective)
		name := question.WithStringAnswerV2("What is the name of this SLO?", "", defaultSloName, emptyOptions)
		objective.Metadata.Name = name
		objective.Metadata.Labels["name"] = name

		objectives = append(objectives, objective)

		fmt.Println(color.Green(fmt.Sprintf("SLO '%s' added to Service '%s'", name, serviceName)))

		fmt.Fprintln(w)
		askForObjective = question.WithBoolAnswer("Do you want to add another SLO?", emptyOptions, question.WithNoAsDefault)
	}

	return objectives, nil
}

/*
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
*/

func promptForProvider(io *iostreams.IOStreams) (entities.Labels, error) {

	var labels entities.Labels = make(entities.Labels)

	providers := []string{}
	for key := range providersMap {
		providers = append(providers, key)
	}
	sort.Strings(providers) // sorts slice in-place

	providerFullName := question.WithSingleChoiceAnswer("Which cloud provider are you targeting?", emptyOptions, providers...)
	provider := providersMap[providerFullName]
	id := getResourceIDForProvider(provider)

	labels["provider"] = provider
	labels["id"] = id

	return labels, nil
}

/*
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
*/

func getResourceIDForProvider(provider string) string {
	switch provider {
	case "aws":
		return promptAWSArn()
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
