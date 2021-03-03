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
	context *ctx.Context

	violations core.ResultSet

	supportedFormats = Choice{"simple", "json", "yaml", "sarif", "codeclimate"}
	supportedLevels  = Choice(core.Levels)
)

type ScanOptions struct {
	IO *iostreams.IOStreams

	Files []string
	Exec  *api.NewExecution

	BaseDirectory       string
	OutputFormat        string
	OutputFile          string
	LevelFilter         string
	EnableLiveScan      bool
	KubernetesNamespace string
	KubernetesContext   string
	KubeConfigPath      string
}

func NewCmdScan() *cobra.Command {
	opts := &ScanOptions{
		IO: iostreams.System(),
	}

	cmd := &cobra.Command{
		Use:   "scan [path]",
		Short: "Check for Reliably Suggestions",
		Long: `Check your manifests for Reliably Suggestions.

Manifest(s) can be provided in several ways:
- read from standard input, with dash '-' as argument
- path to a single manifest file
- path to a folder that will be scanned recursively for manifests files

By default, the scan command, run without arguments, is scanning
manifests file from the current working directory.

Reliably can also scan for your live kubernetes cluster.`,
		Example: heredoc.Doc(`
			# Scan a single file:
			$ reliably scan manifest.yaml

			# Scan a folder:
			$ reliably scan
			$ reliably scan .
			$ reliably scan ./manifests

			# Scan with reading manifest from stdin:
			$ cat manifest.yaml | reliably scan -

			# Scan with custom format & output to local file
			$ reliably scan --format json --output report.json

			# Scan a live Kubernetes cluster
			$ reliably scan --live
			$ reliably scan --live [--namespace n] [--kubecontext c] [--kubeconfig c]`),
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
			// Check the suggestion level is valid
			if opts.LevelFilter != "" && !supportedLevels.Has(opts.LevelFilter) {
				return fmt.Errorf("Level '%v' is not valid. Use one of the supported levels: %s", opts.LevelFilter, supportedLevels)
			}
			// Check the file for the kubeconfig argument exists
			// Check ONLY if we turn on the live mode - otherwise we don't want to raise an error for people that don't have kubernetes installed -
			if opts.EnableLiveScan && opts.KubeConfigPath != "" && !k8s.FileExists(opts.KubeConfigPath) {
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
				"live":       opts.EnableLiveScan,
				"namespace":  opts.KubernetesNamespace,
				"kubeconfig": opts.KubeConfigPath,
			}).Debug("Run 'scan' command with")

			if !opts.EnableLiveScan {
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

			violationCount, err := scanRun(opts)
			if err != nil {
				return err
			}

			if violationCount > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

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
		&opts.EnableLiveScan, "live", false,
		"Look for weaknesses in a live Kubernetes cluster")

	cmd.Flags().StringVarP(
		&opts.KubernetesNamespace, "namespace", "n", "", "The namespace to use when using a live cluster",
	)

	cmd.Flags().StringVarP(
		&opts.KubernetesContext, "kubecontext", "c", os.Getenv("KUBECONTEXT"), "Specifies the Kubernetes context to evaluate when scanning live cluster",
	)

	configPath, _ := k8s.FindKubeConfigPath()
	cmd.Flags().StringVarP(
		&opts.KubeConfigPath, "kubeconfig", "k", configPath, "Specifies the path and file to use for kubeconfig for live scan")

	return cmd
}

func saveOutput(opts *ScanOptions, suggestions []*core.Suggestion) error {

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

// scanRun execute the workflow for the scan command
func scanRun(opts *ScanOptions) (count int, err error) {

	hostname := core.Hostname()
	apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

	// Sends the context to Reliably prior executing the command
	orgID, err := api.CurrentUserOrganizationID(apiClient, hostname)
	if err != nil {
		return
	}

	context = ctx.NewContext()
	opts.Exec, err = api.SendExecutionContext(apiClient, hostname, orgID, context)
	if err != nil {
		return
	}
	log.WithFields(log.Fields{
		"ID":    opts.Exec.ID,
		"orgID": opts.Exec.OrgID,
		//"ctxID": opts.Exec.ContextID,
		"srcID": opts.Exec.SourceID,
	}).Debug("New execution recorded")

	// Choose to run either live cluster or static manifests scan
	if opts.EnableLiveScan {
		violations, err = liveScan(opts)
	} else {
		violations, err = staticScan(opts)
	}
	if err != nil {
		fmt.Fprintln(opts.IO.ErrOut, err)
		return
	}

	// filter out violations based on the optional level filter
	if opts.LevelFilter != "" {
		violations, _ = filterViolations(violations, opts.LevelFilter)
	}

	// Convert internal OPA violations into output-compliant structure
	suggestions := core.ConvertViolationsToSuggestions(violations, opts.EnableLiveScan)

	// Record raised suggestions into user account, for historical reason
	// for now, we can simple ignore err with API !
	_ = api.RecordSuggestions(apiClient, hostname, orgID, opts.Exec.ID, &suggestions)

	// Create output report & show it to the user (or save it to local file)
	if err = saveOutput(opts, suggestions); err != nil {
		fmt.Fprintln(opts.IO.ErrOut, err)
		return
	}

	count = len(violations)

	if (opts.OutputFormat == "simple" || opts.OutputFormat == "") && count > 0 && opts.OutputFile == "" {
		plural := utils.IfThenElse(count > 1, "s", "")
		fmt.Fprintf(opts.IO.ErrOut, "%v suggestion%s found\n", count, plural)
	}

	return

}

// staticScan runs the scan on static files ie manifests
func staticScan(opts *ScanOptions) (core.ResultSet, error) {

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

			path := fmt.Sprintf("%s/%s", header.APIVersion, header.Kind)
			ppath, err := core.FetchPolicy(workspace, platform, path)
			if err != nil {
				log.Error(fmt.Sprintf(
					"Unable to review resource #%v (%v) in file '%v'", i, kind, fpath))
				continue
			}
			rs := core.Eval(ppath, input)
			newIssues := core.ReportViolations(rs, fpath, platform, kind, startLine, name, uri)
			violations = append(violations, newIssues...)
		}

	}

	return violations, nil
}

// liveScan runs the scan on a live Kubernetes cluster
func liveScan(opts *ScanOptions) (core.ResultSet, error) {

	var violations core.ResultSet = core.ResultSet{} // empty slice

	startLine := 0
	linesCount := 0

	// if flag --cluster is set
	// we will want to get the cluster configuration
	// and search for weaknesses from there

	kubeconfigPath, _ := k8s.FindKubeConfigPath(opts.KubeConfigPath)
	// if the namespace flag is provided use that

	// 1. Connect to the Cluster
	clientSet, err := k8s.GetKubernetesClientSet(kubeconfigPath, opts.KubernetesContext)
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

	var resourceList = k8s.GetResourceList(*clientSet, namespace)

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
		path := fmt.Sprintf("%s/%s", header.APIVersion, header.Kind)
		ppath, err := core.FetchPolicy(workspace, platform, path)
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
