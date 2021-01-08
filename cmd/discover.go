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

	"github.com/reliablyhq/cli/core"
	finder "github.com/reliablyhq/cli/core/find"
	output "github.com/reliablyhq/cli/core/output"
	"github.com/reliablyhq/cli/utils"
)

// Choice is a list of string
type Choice []string

// Has indicates whether the string slice contains the value
func (list Choice) Has(a string) bool {
	return utils.StringInArray(a, list)
}

const (
	platform = "Kubernetes"
)

var (
	baseDirectory string
	outputFormat  string
	outputFile    string

	violations core.ResultSet

	supportedFormats = Choice{"simple", "json", "yaml", "sarif", "codeclimate"}

	reviewCmd = &cobra.Command{
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

		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate command options
			if outputFormat != "" && !supportedFormats.Has(outputFormat) {
				return fmt.Errorf("Format '%v' is not valid. Use one of the supported formats: %v", outputFormat, supportedFormats)
			}
			if outputFormat == "sarif" && outputFile != "" {
				// The file name of a SARIF log file SHOULD end with the extension ".sarif".
				// Example 1: output.sarif
				// The file name MAY end with the additional extension ".json".
				// Example 2: output.sarif.json
				//see: https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/sarif-v2.1.0-os.html#_Toc34317421
				if !strings.HasSuffix(outputFile, ".sarif") && !strings.HasSuffix(outputFile, ".sarif.json") {
					return fmt.Errorf("The output file name for a SARIF report should end with the extension '.sarif' or '.sarif.json'")
				}
			}
			// Add a warning to user for deprecated --dir flag
			if baseDirectory != "" {
				log.Warning(
					"--dir flag is deprecated and shall not be used anymore. ",
					fmt.Sprintf("Please run `reliably discover %s` instead.", baseDirectory),
				)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {

			var argStr string

			if len(args) > 0 {
				argStr = args[0]
			}

			log.WithFields(log.Fields{
				"arg":       argStr,
				"directory": baseDirectory,
				"format":    outputFormat,
				"output":    outputFile,
			}).Debug("Run 'discover' command with")

			// Run the command
			violationCount := 0 // initializes the global number of violations

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
				if baseDirectory == "" {
					baseDirectory = "."
				}
				files = finder.GetKubernetesFiles(baseDirectory)
			}
			log.Debug(fmt.Sprintf("Kubernetes files found: %v", files))

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

					violationCount += core.CountViolations(rs, platform, kind)
				}

			}

			// Create output report
			suggestions := core.ConvertViolationsToSuggestions(violations)
			if err := saveOutput(outputFile, outputFormat, baseDirectory, suggestions); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if violationCount > 0 {
				os.Exit(1)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(reviewCmd)

	// This flag is deprecated and soon to be removed;
	// But it's kept as hidden flag for backward compatibility
	reviewCmd.Flags().StringVar(
		&baseDirectory, "dir", "", "Base directory to look for candidates",
	)
	// Does not make it visible to users in help anymore as deprecated
	reviewCmd.Flags().MarkHidden("dir")

	reviewCmd.Flags().StringVarP(
		&outputFormat, "format", "f", "",
		fmt.Sprintf("Specify the output format: %v", supportedFormats))
	//reviewCmd.Flags().Lookup("format").NoOptDefVal = "default"

	reviewCmd.Flags().StringVarP(
		&outputFile, "output", "o", "", "Write results to a file instead of standard output")

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
