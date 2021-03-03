package workflows

import (
	"github.com/MakeNowJust/heredoc/v2"
)

const github_Path string = ".github/workflows/reliably.yaml"

var github_Template string = heredoc.Doc(`
name: Reliably workflow

on: push

env:
  RELIABLY_TOKEN: ${{ secrets.RELIABLY_TOKEN }}

jobs:
  reliably:
    runs-on: ubuntu-latest
    steps:
      - name: 'Checkout source code'
        uses: actions/checkout@v2
      - name: 'Run Reliably'
        uses: reliablyhq/gh-action@v1
        continue-on-error: true
        with:
          format: "sarif"
          output: "reliably.sarif"
      - name: Upload result to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v1
        with:
          sarif_file: reliably.sarif
`)

var github_AccessTokenSecretHelp string = `
You must define %s as a Secret in your repository settings:
https://github.com/%s/%s/settings/secrets/actions/new
`
