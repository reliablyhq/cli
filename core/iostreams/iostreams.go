package iostreams

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/reliablyhq/cli/core/color"

	"github.com/mattn/go-isatty"
)

// SuccessIcon returns the success icon as colored string
func SuccessIcon() string {
	return color.Green("✓")
}

// FailureIcon returns the failure icon as colored string
func FailureIcon() string {
	return color.Red("✕")
}

func UnknownIcon() string {
	return color.Grey("?")
}

func WarningIcon() string {
	return color.Yellow("!")
}

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	stdinTTYOverride  bool
	stdinIsTTY        bool
	stdoutTTYOverride bool
	stdoutIsTTY       bool
	stderrTTYOverride bool
	stderrIsTTY       bool

	neverPrompt bool

	progressIndicatorEnabled bool
	progressIndicator        *spinner.Spinner
}

func (s *IOStreams) SetStdinTTY(isTTY bool) {
	s.stdinTTYOverride = true
	s.stdinIsTTY = isTTY
}

func (s *IOStreams) IsStdinTTY() bool {
	if s.stdinTTYOverride {
		return s.stdinIsTTY
	}
	if stdin, ok := s.In.(*os.File); ok {
		return isTerminal(stdin)
	}
	return false
}

func (s *IOStreams) SetStdoutTTY(isTTY bool) {
	s.stdoutTTYOverride = true
	s.stdoutIsTTY = isTTY
}

func (s *IOStreams) IsStdoutTTY() bool {
	if s.stdoutTTYOverride {
		return s.stdoutIsTTY
	}
	if stdout, ok := s.Out.(*os.File); ok {
		return isTerminal(stdout)
	}
	return false
}

func (s *IOStreams) SetStderrTTY(isTTY bool) {
	s.stderrTTYOverride = true
	s.stderrIsTTY = isTTY
}

func (s *IOStreams) IsStderrTTY() bool {
	if s.stderrTTYOverride {
		return s.stderrIsTTY
	}
	if stderr, ok := s.ErrOut.(*os.File); ok {
		return isTerminal(stderr)
	}
	return false
}

func (s *IOStreams) CanPrompt() bool {
	if s.neverPrompt {
		return false
	}

	return s.IsStdinTTY() && s.IsStdoutTTY()
}

func (s *IOStreams) SetNeverPrompt(v bool) {
	s.neverPrompt = v
}

func System() *IOStreams {
	stdoutIsTTY := isTerminal(os.Stdout)
	stderrIsTTY := isTerminal(os.Stderr)

	io := &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	if stdoutIsTTY && stderrIsTTY {
		io.progressIndicatorEnabled = true
	}

	// prevent duplicate isTerminal queries now that we know the answer
	io.SetStdoutTTY(stdoutIsTTY)
	io.SetStderrTTY(stderrIsTTY)
	return io
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func isCygwinTerminal(w io.Writer) bool {
	if f, isFile := w.(*os.File); isFile {
		return isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

func (s *IOStreams) StartProgressIndicator() {
	if !s.progressIndicatorEnabled {
		return
	}
	sp := spinner.New(spinner.CharSets[11], 200*time.Millisecond, spinner.WithWriter(s.ErrOut))
	sp.Reverse() // make it spin clockwise
	sp.Start()
	s.progressIndicator = sp
}

func (s *IOStreams) StopProgressIndicator() {
	if s.progressIndicator == nil {
		return
	}
	s.progressIndicator.Stop()
	s.progressIndicator = nil
}

func (s *IOStreams) SetProgessMessage(msg string) {
	if s.progressIndicator == nil {
		return
	}
	s.progressIndicator.Suffix = msg
}
