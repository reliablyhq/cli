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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/entities"
	"github.com/reliablyhq/cli/core/iostreams"
	"github.com/reliablyhq/cli/core/metrics/datadog"
)

var (
	emptyOptions        = []question.AskOpt{}
	iconWarn            = iostreams.WarningIcon()
	supportedExtensions = []string{".yaml", ".json"}
	providersMap        = map[string]string{
		"Amazon Web Services":   "aws",
		"Google Cloud Platform": "gcp",
		"Datadog":               "datadog",
	}
)

type InitOptions struct {
	IO *iostreams.IOStreams

	ManifestPath string
}

func NewCommand(runF func(*InitOptions) error) *cobra.Command {
	opts := &InitOptions{
		IO: iostreams.System(),
	}

	cmd := cobra.Command{
		Use:     "init",
		Short:   "initialise the slo portion of the manifest",
		Long:    longCommandDescription(),
		Example: examples(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return initRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.ManifestPath, "output", "o", "./reliably.yaml", "store a local copy of the service manifest created")
	return &cmd
}

func initRun(opts *InitOptions) error {

	fmt.Println("Validate API Key for datadog")
	if ok, err := datadog.ValidateApiKey(); ok {
		fmt.Println("Authenticated to Datadog API")
	} else {
		fmt.Println("Error while validating DD API KEY", err)
	}

	var query string

	/*
		//query = "sum:gcp.loadbalancing.https.backend_request_count{response_code_class:200,backend_name:staging-api}"
		query = "sum:gcp.loadbalancing.https.backend_request_count{response_code_class:200,backend_name:staging-api}.as_count()+sum:gcp.loadbalancing.https.backend_request_count{response_code_class:300,backend_name:staging-api}.as_count()+sum:gcp.loadbalancing.https.backend_request_count{response_code_class:400,backend_name:staging-api}.as_count()"
		_ = QueryMetrics(query)

		query = "sum:gcp.loadbalancing.https.backend_request_count{backend_name:staging-api}.as_count()"
		_ = QueryMetrics(query)
	*/

	/*
		query = "sum:gcp.loadbalancing.https.backend_request_count{response_code_class:200,backend_name:reliablyalpha1}.as_count()+sum:gcp.loadbalancing.https.backend_request_count{response_code_class:300,backend_name:reliablyalpha1}.as_count()"
		_ = QueryMetrics(query)

		query = "sum:gcp.loadbalancing.https.backend_request_count{backend_name:reliablyalpha1}"
		_ = QueryMetrics(query)

		query = "(sum:gcp.loadbalancing.https.backend_request_count{response_code_class:200,backend_name:reliablyalpha1}.as_count()+sum:gcp.loadbalancing.https.backend_request_count{response_code_class:300,backend_name:reliablyalpha1}.as_count()) / sum:gcp.loadbalancing.https.backend_request_count{backend_name:reliablyalpha1}"
		_ = QueryMetrics(query)
	*/

	query = "sum:gcp.loadbalancing.https.backend_request_count{backend_name:reliablyalpha1,response_code_class:200}.as_count()"
	//	_ = QueryMetrics(query)
	num := query

	query = "sum:gcp.loadbalancing.https.backend_request_count{backend_name:reliablyalpha1}.as_count()"
	//_ = QueryMetrics(query)
	denom := query

	//query = "(sum:gcp.loadbalancing.https.backend_request_count{response_code_class:200,backend_name:reliablyalpha1}.as_count()+sum:gcp.loadbalancing.https.backend_request_count{response_code_class:300,backend_name:reliablyalpha1}.as_count()) / sum:gcp.loadbalancing.https.backend_request_count{backend_name:reliablyalpha1}"
	//_ = QueryMetrics(query)
	/*
		fmt.Println("#####")

		_ = ImportSLOsFromDatadog()
	*/

	fmt.Println("----> COMPUTE SLO from numerator/denumerator queries")

	slo, err := datadog.ComputeSloFromQueryMetrics(num, denom)
	if err != nil {
		fmt.Println("we have an error", err)
		return nil
	}
	fmt.Println("SLO ->", slo)

	//return nil

	manifestPath := opts.ManifestPath

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

	// write file output
	f, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer f.Close()

	ye := yaml.NewEncoder(f)
	for _, o := range objectives {
		if err := ye.Encode(o); err != nil {
			return err
		}
	}

	fmt.Println()
	fmt.Println(iostreams.SuccessIcon(), "Your manifest has been saved to", manifestPath)
	log.Debugf("manifest created at: %s", manifestPath)
	return nil
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

		provider, providerLabels, _ := promptForProvider(io)
		for k, v := range providerLabels {
			objective.Spec.IndicatorSelector[k] = v
		}

		switch provider {
		case "gcp", "aws":
			slType := question.WithSingleChoiceAnswer("What type of SLO do you want to declare?", emptyOptions, "Availability", "Latency")
			slType = sanitizeString(slType)
			objective.Spec.IndicatorSelector["category"] = slType

			if slType == "latency" {
				threshold := question.WithDurationAnswer("What is your latency threshold (in milliseconds)?", emptyOptions)
				objective.Spec.IndicatorSelector["latency_target"] = fmt.Sprint(threshold)

				// should we prompt for the percentile as well ? ...
				objective.Spec.IndicatorSelector["percentile"] = "99"
			}

		case "datadog":
			objective.Spec.IndicatorSelector["datadog_numerator_query"] = promptDatadogQuery("numerator", "good events")
			objective.Spec.IndicatorSelector["datadog_denominator_query"] = promptDatadogQuery("denominator", "total events")
		}

		objective.Spec.Window = core.Duration{Duration: getObservationWindow().ToDuration()}

		defaultSloName := generateDefaultSloName(objective)
		name := question.WithStringAnswerV2("What is the name of this SLO?", "", defaultSloName, emptyOptions)
		objective.Metadata.Name = name
		objective.Metadata.Labels["name"] = name

		objective.Metadata.Labels["service"] = serviceName

		objectives = append(objectives, objective)

		fmt.Println(color.Green(fmt.Sprintf("SLO '%s' added to Service '%s'", name, serviceName)))

		fmt.Fprintln(w)
		askForObjective = question.WithBoolAnswer("Do you want to add another SLO?", emptyOptions, question.WithNoAsDefault)
	}

	return objectives, nil
}

func promptForProvider(io *iostreams.IOStreams) (string, entities.Labels, error) {

	//var labels entities.Labels = make(entities.Labels)

	providers := []string{}
	for key := range providersMap {
		providers = append(providers, key)
	}
	sort.Strings(providers) // sorts slice in-place

	providerFullName := question.WithSingleChoiceAnswer("Which cloud provider are you targeting?", emptyOptions, providers...)
	provider := providersMap[providerFullName]
	ps := getProviderSelectors(provider)
	return provider, entities.Labels(ps), nil
}

func getProviderSelectors(provider string) map[string]string {
	var selectors map[string]string = make(map[string]string)

	switch provider {
	case "aws":
		selectors["aws_arn"] = promptAWSArn()
	case "gcp":
		r := buildGCPResourceID()
		if r != nil {
			switch r.ResourceType {
			case "google-cloud-load-balancers":
				selectors["gcp_project_id"] = r.ProjectID
				selectors["gcp_loadbalancer_name"] = r.ResourceName
			}
		}
		/*
			default:
				return question.WithStringAnswer("What is the ID of the resource? This could be the AWS ARN, azure resource ID, etc.", emptyOptions)
		*/
	}

	return selectors
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
