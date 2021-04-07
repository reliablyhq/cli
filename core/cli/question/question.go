package question

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/reliablyhq/cli/core"
)

func WithStringAnswer(questionText string) string {
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

func WithFloat64Answer(question string, min, max float64) float64 {
	for {
		answer := WithStringAnswer(question)
		if f, err := strconv.ParseFloat(answer, 64); err != nil {
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

func WithInt64Answer(question string) int64 {
	for {
		answer := WithStringAnswer(question)
		if i, err := strconv.ParseInt(answer, 10, 64); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			return i
		}
	}
}

func WithDurationAnswer(question string) core.Duration {
	for {
		answer := WithInt64Answer(question)
		ms := answer * 1000000
		return core.Duration{Duration: time.Duration(ms)}
	}
}

func WithBoolAnswer(question string) bool {
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

func WithSingleChoiceAnswer(question string, choices ...string) string {
	var answer string
	prompt := survey.Select{
		Options: choices,
		Message: question,
	}

	if err := survey.AskOne(&prompt, &answer); err == terminal.InterruptErr {
		os.Exit(0)
	}

	return answer
}
