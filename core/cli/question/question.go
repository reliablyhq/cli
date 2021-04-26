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

type AskOpt = survey.AskOpt

var (
	Required    = survey.WithValidator(survey.Required)
	Cursor      = survey.WithShowCursor(true)
	Subquestion = survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Text = "|"
		icons.Question.Format = "green+d"
	})
)

func WithStringAnswer(questionText string, opts []survey.AskOpt) string {
	var text string
	opts = append(opts, Required, Cursor)

	err := survey.AskOne(&survey.Input{
		Message: questionText,
	}, &text, opts...)
	if err == terminal.InterruptErr {
		os.Exit(0)
	}

	return text
}

func WithFloat64Answer(question string, opts []survey.AskOpt, min, max float64) float64 {
	for {
		answer := WithStringAnswer(question, opts)
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

func WithInt64Answer(question string, opts []survey.AskOpt) int64 {
	for {
		answer := WithStringAnswer(question, opts)
		if i, err := strconv.ParseInt(answer, 10, 64); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			return i
		}
	}
}

func WithDurationAnswer(question string, opts []survey.AskOpt) core.Duration {
	answer := WithInt64Answer(question, opts)
	ms := answer * 1000000
	return core.Duration{Duration: time.Duration(ms)}
}

type BoolAnswer = bool

const (
	WithYesAsDefault bool = true
	WithNoAsDefault  bool = false
)

func WithBoolAnswer(question string, opts []survey.AskOpt, yesno ...BoolAnswer) bool {
	var b bool
	var defaultAnwser bool = true

	// use variadic argument as single optional param
	if len(yesno) > 0 {
		defaultAnwser = yesno[0]
	}

	err := survey.AskOne(&survey.Confirm{
		Message: question,
		Default: defaultAnwser,
	}, &b, opts...)
	if err == terminal.InterruptErr {
		os.Exit(0)
	}
	return b
}

func WithSingleChoiceAnswer(question string, opts []survey.AskOpt, choices ...string) string {
	var answer string
	prompt := survey.Select{
		Options: choices,
		Message: question,
	}

	if err := survey.AskOne(&prompt, &answer, opts...); err == terminal.InterruptErr {
		os.Exit(0)
	}

	return answer
}
