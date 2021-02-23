package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/icza/dyno"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	// "k8s.io/client-go/kubernetes"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	ctx "github.com/reliablyhq/cli/core/context"
	finder "github.com/reliablyhq/cli/core/find"
	"github.com/reliablyhq/cli/core/iostreams"
	k8s "github.com/reliablyhq/cli/core/kubernetes"
	output "github.com/reliablyhq/cli/core/output"
	"github.com/reliablyhq/cli/utils"
)

const (
	platform = "Kubernetes"
)

var (
	context   *ctx.Context
	contextID string

	violations core.ResultSet

	supportedFormats = Choice{"simple", "json", "yaml", "sarif", "codeclimate"}
	supportedLevels  = Choice(core.Levels)
)

type DiscoveryOptions struct {
	IO *iostreams.IOStreams

	Files []string

	BaseDirectory       string
	OutputFormat        string
	OutputFile          string
	LevelFilter         string
	EnableLiveDiscovery bool
	KubernetesNamespace string
	KubeConfigPath      string
}

func NewCmdDiscover() *cobra.Command {
	opts := &DiscoveryOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "discover [path]",
		Short: "Check for Reliably Suggestions",
		Long: `Check your manifests for Reliably Suggestions.

Manifest(s) can be provided in several ways:
- read from standard input, with dash '-' as argument
- path to a single manifest file
- path to a folder that will be scanned recursively for manifests files

By default, the discover command, run without arguments, is scanning
manifests file from the current working directory.`,
		Example: heredoc.Doc(`
			# Discover with a single file:
			$ reliably discover manifest.yaml

			# Discover with a folder:
			$ reliably discover
			$ reliably discover .
			$ reliably discover ./manifests

			# Discover with reading manifest from stdin:
			$ cat manifest.yaml | reliably discover -

			# Discover with custom format & output to local file
			$ reliably discover --format json --output report.json`),
		ValidArgs: []string{"-"},
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}

			// no argument is given, make it default with directory scanning
			if len(args) == 0 {
				return nil
			}

			// When dash is used, we capture from stdin
			if args[0] == "-" {
				return nil
			}

			// or consider a valid path to file or folder
			f := args[0]
			if f != "" {
				if _, err := os.Stat(f); os.IsNotExist(err) {
					return fmt.Errorf("Invalid argument '%s': does not exist", f)
				}
			}

			return nil
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if !cmdutil.CheckAuth() {
				cmdutil.PrintRequireAuthMsg()
				os.Exit(1)
			}
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate command options
			if opts.OutputFormat != "" && !supportedFormats.Has(opts.OutputFormat) {
				return fmt.Errorf("Format '%v' is not valid. Use one of the supported formats: %v", opts.OutputFormat, supportedFormats)
			}
			if opts.OutputFormat == "sarif" && opts.OutputFile != "" {
				// The file name of a SARIF log file SHOULD end with the extension ".sarif".
				// Example 1: output.sarif
				// The file name MAY end with the additional extension ".json".
				// Example 2: output.sarif.json
				//see: https://doclientSet.oasis-open.org/sarif/sarif/v2.1.0/os/sarif-v2.1.0-os.html#_Toc34317421
				if !strings.HasSuffix(opts.OutputFile, ".sarif") && !strings.HasSuffix(opts.OutputFile, ".sarif.json") {
					return fmt.Errorf("The output file name for a SARIF report should end with the extension '.sarif' or '.sarif.json'")
				}
			}
			// Add a warning to user for deprecated --dir flag
			if opts.BaseDirectory != "" {
				log.Warning(
					"--dir flag is deprecated and shall not be used anymore. ",
					fmt.Sprintf("Please run `reliably discover %s` instead.", opts.BaseDirectory),
				)
			}
			// Check the suggestion level is valid
			if opts.LevelFilter != "" && !supportedLevels.Has(opts.LevelFilter) {
				return fmt.Errorf("Level '%v' is not valid. Use one of the supported levels: %s", opts.LevelFilter, supportedLevels)
			}
			// Check the file for the kubeconfig argument exists
			if opts.KubeConfigPath != "" && !k8s.FileExists(opts.KubeConfigPath) {
				return fmt.Errorf("The kubeconfig argument %v is not a path to a file", opts.KubeConfigPath)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				argStr string
				err    error
			)

			if len(args) > 0 {
				argStr = args[0]
			}

			log.WithFields(log.Fields{
				"arg":        argStr,
				"directory":  opts.BaseDirectory,
				"format":     opts.OutputFormat,
				"output":     opts.OutputFile,
				"live":       opts.EnableLiveDiscovery,
				"namespace":  opts.KubernetesNamespace,
				"kubeconfig": opts.KubeConfigPath,
			}).Debug("Run 'discover' command with")

			if !opts.EnableLiveDiscovery {
				if len(args) > 0 {
					fpath := args[0]
					if fpath == "-" {
						opts.Files = append(opts.Files, fpath)
					} else {
						info, _ := os.Stat(fpath)
						if info.IsDir() {
							opts.Files = finder.GetKubernetesFiles(fpath)
						} else {
							// single file
							opts.Files = append(opts.Files, fpath)
						}
					}
				} else {
					// when no arg is provided, set the current working dir as default
					if opts.BaseDirectory == "" {
						opts.BaseDirectory = "."
					}
					opts.Files = finder.GetKubernetesFiles(opts.BaseDirectory)
				}
				log.Debug(fmt.Sprintf("Kubernetes files found: %v", opts.Files))
			}

			violationCount, err := discoverRun(opts)
			if err != nil {
				return err
			}

			if violationCount > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	// This flag is deprecated and soon to be removed;
	// But it's kept as hidden flag for backward compatibility
	cmd.Flags().StringVar(
		&opts.BaseDirectory, "dir", "", "Base directory to look for candidates",
	)
	// Does not make it visible to users in help anymore as deprecated
	_ = cmd.Flags().MarkHidden("dir")

	cmd.Flags().StringVarP(
		&opts.OutputFormat, "format", "f", "",
		fmt.Sprintf("Specify the output format: %v", supportedFormats))
	//reviewCmd.Flags().Lookup("format").NoOptDefVal = "default"

	cmd.Flags().StringVarP(
		&opts.OutputFile, "output", "o", "", "Write results to a file instead of standard output")

	cmd.Flags().StringVarP(
		&opts.LevelFilter, "level", "l", "", "Display suggestions only for level and higher")

	// Declare and hide herited options from kubectl
	cmd.Flags().BoolVar(
		&opts.EnableLiveDiscovery, "live", false,
		"Look for weaknesses in a live Kubernetes cluster")

	cmd.Flags().StringVarP(
		&opts.KubernetesNamespace, "namespace", "n", "", "The namespace to use when using a live cluster",
	)

	cmd.Flags().StringVarP(
		&opts.KubeConfigPath, "kubeconfig", "k", "", "Specifiies the path and file to use for kubeconfig for live discovery")
	configPath, _ := k8s.FindKubeConfigPath("")
	cmd.Flags().Lookup("kubeconfig").NoOptDefVal = configPath

	return cmd
}

func saveOutput(opts *DiscoveryOptions, suggestions []*core.Suggestion) error {

	log.WithFields(log.Fields{
		"filename":    opts.OutputFile,
		"format":      opts.OutputFormat,
		"baseDir":     opts.BaseDirectory,
		"suggestions": fmt.Sprintf("%v", len(suggestions)),
	}).Debug("saveOutput")

	if opts.OutputFile != "" {
		outfile, err := os.Create(opts.OutputFile) // creates or truncates with O_RDWR mode
		if err != nil {
			log.Error("error creating output file")
			log.Error(opts.OutputFile)
			log.Error(err)
			return err
		}
		defer outfile.Close()
		err = output.CreateReport(outfile, opts.OutputFormat, opts.BaseDirectory, suggestions)
		if err != nil {
			log.Error("Error creating the report")
			log.Error(err)
			return err
		}
	} else {
		// os.Stdout is an opened file to stdout
		err := output.CreateReport(os.Stdout, opts.OutputFormat, opts.BaseDirectory, suggestions)
		if err != nil {
			log.Error("Error creating the report")
			log.Error(err)
			return err
		}
	}

	return nil
}

// discoverRun execute the discover workflow for the command
func discoverRun(opts *DiscoveryOptions) (count int, err error) {

	hostname := core.Hostname()
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	// Sends the context to Reliably prior executing the command
	orgID, err := api.CurrentUserOrganizationID(apiClient, hostname)
	if err != nil {
		return
	}

	context = ctx.NewContext()
	contextID, err = api.SendExecutionContext(apiClient, hostname, orgID, context)
	if err != nil {
		return
	}

	// Choose to run either live cluster or static manifests discovery
	if opts.EnableLiveDiscovery {
		violations, err = liveDiscover(opts)
	} else {
		violations, err = staticDiscover(opts)
	}
	if err != nil {
		fmt.Fprintln(opts.IO.ErrOut, err)
		return
	}

	// filter out violations based on the optional level filter
	if opts.LevelFilter != "" {
		violations, _ = filterViolations(violations, opts.LevelFilter)
	}

	// Create output report
	suggestions := core.ConvertViolationsToSuggestions(violations)
	if err = saveOutput(opts, suggestions); err != nil {
		fmt.Fprintln(opts.IO.ErrOut, err)
		return
	}

	return len(violations), nil

}

// staticDiscover runs the discovery on static files
func staticDiscover(opts *DiscoveryOptions) (core.ResultSet, error) {

	var violations core.ResultSet = core.ResultSet{} // empty slice

	for _, fpath := range opts.Files {
		log.Debug(fmt.Sprintf("Processing file %v", fpath))

		startLine := 0
		linesCount := 0

		resources := finder.ReadAndSplitKubernetesFile(fpath)
		for i, resource := range resources {

			// for a new resource, append the previous resource length to
			// global file lines counter (beginning of the current resource);
			// then calculate the current resource length for next iteration
			startLine += linesCount
			linesCount = strings.Count(resource, "\n") + 1 // Read & split function removes 1 line return

			header, err := finder.GetYamlInfo(resource)
			if err != nil {
				// Unable to identify Yaml as K8s resource
				continue
			}
			kind := header.Kind
			name := header.Metadata.Name
			uri := header.URI()

			// unmarshall the yaml content into standard map for OPA
			var input interface{}
			if err := yaml.Unmarshal([]byte(resource), &input); err != nil {
				// Unable to load YAML - shall not happen as already parsed once
				// by the GetYamlInfo method
				continue
			}

			log.Debug(fmt.Sprintf("Processing resource #%v: %v", i, kind))

			// !! m cannot be marshaled using encoding/json !!
			// -> unmarshalled YAML struct is not compliant with JSON
			// due to interface{} as map keys, while JSON supports
			// only strings as keys
			// --> underneath, it converts map[interface{}]interface{}
			// to map[string]interface{} (recursively)
			input = dyno.ConvertMapI2MapS(input)

			ppath, err := core.FetchPolicy(workspace, platform, kind)
			if err != nil {
				log.Error(fmt.Sprintf(
					"Unable to review resource #%v (%v) in file '%v'", i, kind, fpath))
				continue
			}
			// todo remove the next line
			fmt.Printf("input  %v", input)
			rs := core.Eval(ppath, input)
			newIssues := core.ReportViolations(rs, fpath, platform, kind, startLine, name, uri)
			violations = append(violations, newIssues...)
		}

	}

	return violations, nil
}

// liveDiscover runs the discovery on a live Kubernetes cluster
func liveDiscover(opts *DiscoveryOptions) (core.ResultSet, error) {

	var violations core.ResultSet = core.ResultSet{} // empty slice

	startLine := 0
	linesCount := 0

	// if flag --cluster is set
	// we will want to get the cluster configuration
	// and search for weaknesses from there

	kubeconfigPath, _ := k8s.FindKubeConfigPath(opts.KubeConfigPath)
	// if the namespace flag is provided use that

	// 1. Connect to the Cluster
	clientSet, err := k8s.GetKubernetesClientSet(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	// 2. Scan the API for "configuration"

	namespace := "default"
	// if the namespace flag is provided use that
	if opts.KubernetesNamespace != "" {
		namespace = opts.KubernetesNamespace
	}

	// log.Debugf("Get pods for namespace %v", namespace)

	var resourceList []string = make([]string, 0, 0)

	podList, _ := k8s.GetPodSpec(*clientSet, namespace)
	deploymentList, _ := k8s.GetDeploymentSpec(*clientSet, namespace)
	clusterRoleBindingList, _ := k8s.GetClusterRoleBindingSpec(*clientSet)
	ingressList, _ := k8s.GetIngressSpec(*clientSet, namespace)
	podSecurityPolicyList, _ := k8s.GetPodSecurityPolicySpec(*clientSet)

	lists := [][]string{
		podList,
		deploymentList,
		clusterRoleBindingList,
		ingressList,
		podSecurityPolicyList,
	}

	for _, l := range lists {
		resourceList = append(resourceList, l...)
	}

	for _, r := range resourceList {

		startLine += linesCount
		linesCount = strings.Count(r, "\n")

		header, err := k8s.GetHeaderInfo(r)
		if err != nil {
			log.Debugf("Error from k8s.GetHeaderInfo: %v", err)

			continue
		}
		log.Debug(fmt.Sprintf("Processing Pod #%v: in namespace: %v",
			header.Metadata.Name, namespace))

		kind := header.Kind
		name := header.Metadata.Name
		uri := header.URI()

		var input interface{}
		if err := json.Unmarshal([]byte(r), &input); err != nil {
			// Unable to load JSON - shall not happen as already parsed once
			// by the GetHeaderInfo method
			// continue
			log.Debugf("Error Unmarshalling Pod: %v", err)
			continue
		}
		// todo consider removing this, I dont think its required
		input = dyno.ConvertMapI2MapS(input)

		// fetch the policies
		ppath, err := core.FetchPolicy(workspace, platform, kind)
		if err != nil {
			log.Error(fmt.Sprintf(
				"Unable to review resource #%v (%v) in file '%v'", 0, kind, "live"))
			// continue
		}
		log.Debugf("policy path %v", ppath)

		// evaluate the input with the policies
		rs := core.Eval(ppath, input)

		newIssues := core.ReportViolations(rs, "fpath", platform, kind, startLine, name, uri)
		violations = append(violations, newIssues...)

	}

	return violations, nil
}

// filterViolations allows to filter out violations with a level lower than
// the requested level
func filterViolations(violations core.ResultSet, l string) (core.ResultSet, error) {

	level, err := core.NewLevel(l)
	if err != nil {
		return violations, err // unknown level
	}

	filtered := utils.Filter(violations, func(val interface{}) bool {
		return val.(core.Result).Rule.Level >= level
	})

	return filtered.(core.ResultSet), nil
}
