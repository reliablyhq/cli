package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/cli/safeexec"
)

var GitCommand = func(args ...string) (*exec.Cmd, error) {
	gitExe, err := safeexec.LookPath("git")
	if err != nil {
		programName := "git"
		if runtime.GOOS == "windows" {
			programName = "Git for Windows"
		}
		return nil, fmt.Errorf("unable to find git executable in PATH; please install %s before retrying", programName)
	}
	return exec.Command(gitExe, args...), nil
}

// ToplevelDir returns the top-level directory path of the current repository
func ToplevelDir() (string, error) {
	showCmd, err := GitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	output, err := showCmd.Output()
	return firstLine(output), err
}

// GitDir returns the directory path of the current repository
func GitDir() (string, error) {
	showCmd, err := GitCommand("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	output, err := showCmd.Output()
	return firstLine(output), err
}

// IsGitRepo indicates whether the current working directory is a git repository
func IsGitRepo() bool {
	d, err := GitDir()
	if err != nil {
		return false
	}

	return d != ""
}

func firstLine(output []byte) string {
	if i := bytes.IndexAny(output, "\n"); i >= 0 {
		return string(output)[0:i]
	}
	return string(output)
}
