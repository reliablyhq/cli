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

func WithStringAnswer(questionText string, subquestion bool) string {
	var text string

	for len(text) == 0 {
		err := survey.AskOne(&survey.Input{
			Message: questionText,
		}, &text, SetIcon(subquestion), survey.WithValidator(survey.Required), survey.WithShowCursor(true))
		if err == terminal.InterruptErr {
			os.Exit(0)
		}
	}

	return text
}

func WithFloat64Answer(question string, subquestion bool, min, max float64) float64 {
	for {
		answer := WithStringAnswer(question, subquestion)
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

func WithInt64Answer(question string, subquestion bool) int64 {
	for {
		answer := WithStringAnswer(question, subquestion)
		if i, err := strconv.ParseInt(answer, 10, 64); err != nil {
			fmt.Println("Please make sure you type a number")
		} else {
			return i
		}
	}
}

func WithDurationAnswer(question string, subquestion bool) core.Duration {
	for {
		answer := WithInt64Answer(question, subquestion)
		ms := answer * 1000000
		return core.Duration{Duration: time.Duration(ms)}
	}
}

type BoolAnswer = bool

const (
	WithYesAsDefault bool = true
	WithNoAsDefault  bool = false
)

func WithBoolAnswer(question string, subquestion bool, yesno ...BoolAnswer) bool {
	var b bool
	var defaultAnwser bool = true

	// use variadic argument as single optional param
	if len(yesno) > 0 {
		defaultAnwser = yesno[0]
	}

	err := survey.AskOne(&survey.Confirm{
		Message: question,
		Default: defaultAnwser,
	}, &b, SetIcon(subquestion)survey.WithShowCursor(true))
	if err == terminal.InterruptErr {
		os.Exit(0)
	}
	return b
}

func WithSingleChoiceAnswer(question string, subquestion bool, choices ...string) string {
	var answer string
	prompt := survey.Select{
		Options: choices,
		Message: question,
	}

	if err := survey.AskOne(&prompt, &answer, SetIcon(subquestion)); err == terminal.InterruptErr {
		os.Exit(0)
	}

	return answer
}

func SetIcon(subquestion bool) survey.AskOpt {
	var askOpt survey.AskOpt
	if subquestion {
		askOpt = survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "|"
			icons.Question.Format = "green+d"
		})
	} else {
		askOpt = survey.WithIcons(func(icons *survey.IconSet) {
			icons.Question.Text = "?"
			icons.Question.Format = "green+hb"
		})
	}
	return askOpt
}
