package init

import (
	"fmt"

	"github.com/reliablyhq/cli/core/cli/question"
)

func promptDatadogQuery(name string, help string) string {
	q := fmt.Sprintf("Paste your '%s' (%s) datadog query:", name, help)
	return question.WithStringAnswer(q, []question.AskOpt{question.Subquestion})
}
