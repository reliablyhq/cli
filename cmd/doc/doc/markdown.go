package doc

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewMarkdownCommand(rootCmd *cobra.Command) *cobra.Command {
	var outputDirectory string = "."

	const baseURLPath = "/docs/reference/cli/"

	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return baseURLPath + strings.ToLower(strings.Replace(base, "_", "-", -1)) + "/"
	}

	const fmTemplate = `---
title: %s
excerpt: Documentation for the %s command in the Reliably CLI
categories: ["reference", "cli"]
status: published
type: doc
---
`

	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		return fmt.Sprintf(fmTemplate, strings.Replace(base, "_", " ", -1), filename)
	}

	cmd := &cobra.Command{
		Use:   "markdown",
		Args:  cobra.ExactArgs(0),
		Short: "Generates Markdown pages for the reliably CLI.",
		Long:  "Generate one Markdown document per command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTreeCustom(rootCmd, outputDirectory, filePrepender, linkHandler)
		},
	}

	cmd.Flags().StringVar(
		&outputDirectory,
		"output-dir",
		"",
		"Output directory of the generate Markdown documents",
	)

	return cmd
}
