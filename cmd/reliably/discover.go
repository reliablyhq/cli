package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/icza/dyno"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core"
	ctx "github.com/reliablyhq/cli/core/context"
	finder "github.com/reliablyhq/cli/core/find"
	"github.com/reliablyhq/cli/core/iostreams"
	output "github.com/reliablyhq/cli/core/output"
)

const (
	platform = "Kubernetes"
)

var (
	/*
		baseDirectory string
		outputFormat  string
		outputFile    string
	*/

	context   *ctx.Context
	contextID string

	violations core.ResultSet

	supportedFormats = Choice{"simple", "json", "yaml", "sarif", "codeclimate"}
)

type DiscoveryOptions struct {
	IO *iostreams.IOStreams

	BaseDirectory string
	OutputFormat  string
	OutputFile    string
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
				//see: https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/sarif-v2.1.0-os.html#_Toc34317421
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
				"arg":       argStr,
				"directory": opts.BaseDirectory,
				"format":    opts.OutputFormat,
				"output":    opts.OutputFile,
			}).Debug("Run 'discover' command with")

			hostname := core.Hostname()
			apiClient := api.NewClientFromHTTP(api.AuthHTTPClient(hostname))

			// Sends the context to Reliably prior executing the command
			orgID, err := api.CurrentUserOrganizationID(apiClient, hostname)
			if err != nil {
				return err
			}

			context = ctx.NewContext()
			contextID, err = api.SendExecutionContext(apiClient, hostname, orgID, context)
			if err != nil {
				return err
			}

			// Run the command
			var files []string

			if len(args) > 0 {
				fpath := args[0]
				if fpath == "-" {
					files = append(files, fpath)
				} else {
					info, _ := os.Stat(fpath)
					if info.IsDir() {
						files = finder.GetKubernetesFiles(fpath)
					} else {
						// single file
						files = append(files, fpath)
					}
				}
			} else {
				// when no arg is provided, set the current working dir as default
				if opts.BaseDirectory == "" {
					opts.BaseDirectory = "."
				}
				files = finder.GetKubernetesFiles(opts.BaseDirectory)
			}
			log.Debug(fmt.Sprintf("Kubernetes files found: %v", files))

			violations, err := discoverRun(opts, files)
			if err != nil {
				fmt.Fprintln(opts.IO.ErrOut, err)
				os.Exit(1)
			}

			// Create output report
			suggestions := core.ConvertViolationsToSuggestions(violations)
			if err := saveOutput(opts.OutputFile, opts.OutputFormat, opts.BaseDirectory, suggestions); err != nil {
				fmt.Fprintln(opts.IO.ErrOut, err)
				os.Exit(1)
			}

			if len(violations) > 0 {
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

	return cmd
}

func saveOutput(filename string, format string, baseDir string, suggestions []*core.Suggestion) error {

	log.WithFields(log.Fields{
		"filename":    filename,
		"format":      format,
		"baseDir":     baseDir,
		"suggestions": fmt.Sprintf("%v", len(suggestions)),
	}).Debug("saveOutput")

	if filename != "" {
		outfile, err := os.Create(filename) // creates or truncates with O_RDWR mode
		if err != nil {
			log.Error("error creating output file")
			log.Error(filename)
			log.Error(err)
			return err
		}
		defer outfile.Close()
		err = output.CreateReport(outfile, format, baseDir, suggestions)
		if err != nil {
			log.Error("Error creating the report")
			log.Error(err)
			return err
		}
	} else {
		// os.Stdout is an opened file to stdout
		err := output.CreateReport(os.Stdout, format, baseDir, suggestions)
		if err != nil {
			log.Error("Error creating the report")
			log.Error(err)
			return err
		}
	}

	return nil
}

func discoverRun(opts *DiscoveryOptions, files []string) (core.ResultSet, error) {

	var violations core.ResultSet = core.ResultSet{} // empty slice

	for _, fpath := range files {
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

			rs := core.Eval(ppath, input)
			newIssues := core.ReportViolations(rs, fpath, platform, kind, startLine, name, uri)
			violations = append(violations, newIssues...)
		}

	}

	return violations, nil
}
