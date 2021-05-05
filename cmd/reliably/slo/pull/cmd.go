package pull

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/cli/question"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	output       string
	service      string
	emptyOptions []question.AskOpt
)

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull",
		Short: "pull/download SLO manifest from reliably",
		Long:  longCommandDescription(),
		// Example: examples(),
		RunE: runE,
	}

	cmd.Flags().StringVarP(&output, "output", "o", "reliably.yaml", "output path manifest file will be stored")
	cmd.Flags().StringVarP(&service, "service", "s", "", "the name of the specific service you want to pull")
	return &cmd
}

func runE(_ *cobra.Command, args []string) (err error) {
	if _, err := os.Stat(output); err == nil {
		if !question.WithBoolAnswer(fmt.Sprintf("Existing local manifest detected in output path (%s); Do you want to overwrite it?", output), emptyOptions, question.WithNoAsDefault) {
			return nil
		}
	}

	log.Debugf("pulling manifest to: [%s]", output)
	var m *manifest.Manifest
	client := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	if service != "" {
		m, err = api.PullServiceManifest(client, service)
	} else {
		// else pull all
		m, err = api.PullManifest(client)
	}

	if err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		return err
	}
	return nil
}

func longCommandDescription() string {
	return heredoc.Doc(`
	Pull Manifest from reliably API

	A copy of the current manifest is returned. By default the entire
	manifest is retrieved. However, you can specify specific services using
	the flags, "--service/-s"`)
}
