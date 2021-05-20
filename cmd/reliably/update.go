package cmd

import (
	"context"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v33/github"
	"github.com/inconshreveable/go-update"
	"github.com/machinebox/progress"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/reliablyhq/cli/cmd/reliably/cmdutil"
	"github.com/reliablyhq/cli/core/color"
	"github.com/reliablyhq/cli/core/iostreams"
	up "github.com/reliablyhq/cli/core/update"
	v "github.com/reliablyhq/cli/version"
)

type UpdateOptions struct {
	IO          *iostreams.IOStreams
	UpdaterRepo string

	Version   string
	NoConfirm bool
}

func NewCmdUpdate() *cobra.Command {

	opts := &UpdateOptions{
		IO:          iostreams.System(),
		UpdaterRepo: updaterRepo,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates Reliably CLI",
		Long: `Updates Reliably CLI to the latest version.

It is also possible to update the CLI to a specific version.
Please note that downgrade is also supported by setting the version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Version, "version", "", "Update to a specific version")
	cmd.Flags().BoolVarP(&opts.NoConfirm, "yes", "y", false, "Don't ask for update confirmation")

	cmdutil.DisableAuthCheck(cmd)

	return cmd
}

func runUpdate(opts *UpdateOptions) (err error) {

	if v.IsDevVersion() {
		return errors.New("This command cannot be run in dev mode")
	}

	current := strings.TrimPrefix(v.Version, "v")
	downgrade := strings.TrimPrefix(v.Version, "v")
	upgrade := strings.TrimPrefix(opts.Version, "v")

	var rel *github.RepositoryRelease

	if opts.Version == "" {
		// check if a new version is available, only when upgrading to latest
		rel, err = up.GetLatestRelease(nil, opts.UpdaterRepo)
		if err == nil {
			upgrade = strings.TrimPrefix(*rel.TagName, "v")
			if !up.VersionGreaterThan(upgrade, current) {
				fmt.Fprintln(opts.IO.ErrOut, "You are already using the latest release.")
				return nil
			}
		}

	} else {
		// get the release related to version specified by user
		// makes sure the version tag starts with 'v' prefix
		tag := fmt.Sprintf("v%s", strings.TrimPrefix(opts.Version, "v"))
		rel, err = up.GetRelease(nil, opts.UpdaterRepo, tag)
		if err != nil {
			log.Debug(err)
			return fmt.Errorf("No release was found matching your version '%s'", opts.Version)
		}
	}

	if !opts.NoConfirm {
		fmt.Println("Your current CLI version is:", color.Cyan(current))
		fmt.Println("You will be upgraded to version:", color.Cyan(upgrade))
		fmt.Println()

		// prompt user for update confirmation
		var keepGoing bool
		err := survey.AskOne(&survey.Confirm{
			Message: "Do you want to continue?",
			Default: true,
		}, &keepGoing)
		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		if !keepGoing {
			return nil // exit properly when user did not confirm
		}
	}

	fmt.Fprintln(opts.IO.ErrOut, color.Grey("Please wait while we download and install the new version..."))

	opts.IO.StartProgressIndicator() // Use
	rc, checksum, size, err := up.DownloadReleaseAsset(nil, updaterRepo, runtime.GOOS, *rel.TagName)
	if err != nil {
		return err
	}
	defer rc.Close()

	ctx := context.Background()
	r := progress.NewReader(rc) // we tunnel the http response body reader into the progress tracker
	go func() {                 // Async we show the download progress
		progressChan := progress.NewTicker(ctx, r, int64(size), time.Second)
		for p := range progressChan {
			msg := fmt.Sprintf(" %.0f%% Complete - %v remaining", p.Percent(), p.Remaining().Round(time.Second))
			opts.IO.SetProgessMessage(msg)
			//fmt.Printf("\r%.0f%% Complete - %v remaining...", p.Percent(), p.Remaining().Round(time.Second))
		}
		//fmt.Println("\rdownload is complete") // this does not get printed
	}()

	updateOpts := update.Options{}

	if checksum != "" {
		updateOpts.Hash = crypto.MD5
		// ! we need to decode the checksum into bytes !
		check, _ := hex.DecodeString(checksum)
		updateOpts.Checksum = check
	}

	err = update.Apply(r, updateOpts) // This will trigger the read on the response body - forwarded by the progress -
	if err != nil {
		return err
	}

	opts.IO.StopProgressIndicator()
	fmt.Fprint(opts.IO.ErrOut, ClearPreviousLine) // we also clear the "please wait..." message

	fmt.Fprintln(opts.IO.ErrOut)
	if opts.Version == "" {
		fmt.Fprintln(opts.IO.ErrOut, "You're now up-to-date!")
	} else {
		fmt.Fprintf(opts.IO.ErrOut, "You're now using version '%s'\n", upgrade)
	}

	fmt.Fprintln(opts.IO.ErrOut)
	fmt.Fprintln(opts.IO.ErrOut, "To revert your CLI to the previously installed version, you may run:")
	fmt.Fprintf(opts.IO.ErrOut, "$ reliably update --version %s\n", downgrade)
	fmt.Fprintln(opts.IO.ErrOut)

	return
}
