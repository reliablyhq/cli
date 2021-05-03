package edit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/reliablyhq/cli/api"
	"github.com/reliablyhq/cli/core"
	"github.com/reliablyhq/cli/core/manifest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var editor string

func longCommandDescription() string {
	return `
The edit command pulls a copy of the organization service manifest
and opens the default text editor. Once the file is save and the
editor is closed. The resulting file is applied to the organization

NOTE: This feature only supports terminal based text editors
`
}

func NewCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "edit",
		Short: "Edit slo manifest file on server",
		Long:  longCommandDescription(),
		// Example: examples(),
		RunE: runE,
	}

	cmd.Flags().StringVarP(&editor, "editor", "e", os.Getenv("EDITOR"), "path to text editor binary/app")
	return &cmd
}

func runE(_ *cobra.Command, args []string) error {
	tmpfilePath := fmt.Sprintf(".manifest-edit-%d.yaml", time.Now().Unix())
	defer os.Remove(tmpfilePath)
	client := api.NewClientFromHTTP(api.AuthHTTPClient(core.Hostname()))
	m, err := api.PullManifest(client)
	if err != nil {
		return err
	}

	if m == nil {
		return errors.New("no remote manifest detected, trying running `reliably slo init`")
	}

	f, err := os.Create(tmpfilePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %s", err)
	}

	if err := yaml.NewEncoder(f).Encode(&m); err != nil {
		return err
	}

	f.Close()

	cmd := exec.Command(editor, tmpfilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	log.Debugf("executing: %s %s", editor, tmpfilePath)
	if cmd.Run() != nil {
		return fmt.Errorf("error running text editor: %v", err)
	}

	m, err = manifest.Load(tmpfilePath)
	if err != nil {
		return err
	}

	// TODO: implement hash check so that
	// file is only push to API if changes
	// are made

	// finally push to api
	return api.PushManifest(client, m)
}
