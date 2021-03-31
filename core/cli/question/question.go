package question

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/reliablyhq/cli/core"
)

func WithStringAnswer(scanner *bufio.Scanner, questionText string) string {
	var text string

	for len(text) == 0 {
		err := survey.AskOne(&survey.Input{
			Message: questionText,
		}, &text, survey.WithValidator(survey.Required))
		if err == terminal.InterruptErr {
			os.Exit(0)
		}
	}

	return text
}

func WithFloat64Answer(scanner *bufio.Scanner, question string, min, max float64) float64 {
	for {
		answer := WithStringAnswer(scanner, question)
		if f, err := strconv.ParseFloat(answer, 32); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			if f < min || f > max {
				fmt.Printf("the value must be between %.2f and %.2f\n", min, max)
			} else {
				return f
			}
		}
	}
}

func WithInt64Answer(scanner *bufio.Scanner, question string) int64 {
	for {
		answer := WithStringAnswer(scanner, question)
		if i, err := strconv.ParseInt(answer, 10, 64); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			return i
		}
	}
}

func WithDurationAnswer(scanner *bufio.Scanner, question string) core.Duration {
	for {
		answer := WithInt64Answer(scanner, question)
		ms := answer * 1000000
		return core.Duration{Duration: time.Duration(ms)}
	}
}

func WithBoolAnswer(scanner *bufio.Scanner, question string) bool {
	var b bool
	err := survey.AskOne(&survey.Confirm{
		Message: question,
		Default: true,
	}, &b)
	if err == terminal.InterruptErr {
		os.Exit(0)
	}
	return b
}
