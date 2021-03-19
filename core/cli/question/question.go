package question

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/reliablyhq/cli/core"
)

func WithStringAnswer(scanner *bufio.Scanner, questionText string) string {
	var text string

	for len(text) == 0 {
		fmt.Println(questionText)
		scanner.Scan()
		text = scanner.Text()
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

func WithDurationAnswer(scanner *bufio.Scanner, question string) core.Duration {
	for {
		answer := WithStringAnswer(scanner, question)
		if d, err := time.ParseDuration(answer); err != nil {
			fmt.Println("The value you entered could not be parsed to a duration.")
		} else {
			return core.Duration{Duration: d}
		}
	}
}

func WithBoolAnswer(scanner *bufio.Scanner, question string) bool {
	for {
		answer := WithStringAnswer(scanner, question)
		if b, err := strconv.ParseBool(answer); err == nil {
			return b
		} else {
			// do some noddy-level parsing
			lAnswer := strings.ToLower(answer)
			if lAnswer == "y" || lAnswer == "yes" {
				return true
			} else if lAnswer == "n" || lAnswer == "no" {
				return false
			}

			fmt.Println("the answer you gave could not be parsed to a boolean")
		}
	}
}
