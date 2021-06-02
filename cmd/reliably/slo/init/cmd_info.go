package init

import "github.com/MakeNowJust/heredoc/v2"

func longCommandDescription() string {
	return heredoc.Doc(`
Initialise the reliably manifest.

The manifest describes the operational contraints of the application,
as well as some metadata about the app that allows users to reach out
and communicate with the maintainer.`)
}

func examples() string {
	return heredoc.Doc(`
$ reliably slo init:
  this method interactively creates a manifest file, asking you questions
  on the command line and adding your answers to the manifest file.

$ reliably slo init -f <path>:
  this method works the same as reliably init, but allows you to specify
  the location of the file. This is useful if you use a multi-repo approach
  to source control.`)
}
