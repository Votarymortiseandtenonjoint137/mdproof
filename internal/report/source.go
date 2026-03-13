package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

func reportPath(path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

func firstFailingAssertion(step core.StepResult) *core.AssertionResult {
	for i := range step.Assertions {
		if !step.Assertions[i].Matched {
			return &step.Assertions[i]
		}
	}
	return nil
}

func firstCodeRange(step core.StepResult) *core.SourceRange {
	if len(step.Step.CodeSources) == 0 {
		return nil
	}
	return &step.Step.CodeSources[0]
}

func headingRange(step core.StepResult) *core.SourceRange {
	if step.Step.HeadingSource.IsZero() {
		return nil
	}
	return &step.Step.HeadingSource
}

func failureHeaderLocation(step core.StepResult) (string, int) {
	path := reportPath(step.Step.File)
	if path == "" {
		return "", 0
	}
	if a := firstFailingAssertion(step); a != nil && a.Source != nil {
		return path, a.Source.Start.Line
	}
	if heading := headingRange(step); heading != nil {
		return path, heading.Start.Line
	}
	if code := firstCodeRange(step); code != nil {
		return path, code.Start.Line
	}
	return path, 0
}

func commandFailureLocation(step core.StepResult) string {
	path := reportPath(step.Step.File)
	code := firstCodeRange(step)
	if path == "" || code == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d-%d", path, code.Start.Line, code.End.Line)
}

func assertionFailureLocation(step core.StepResult) (string, string) {
	path := reportPath(step.Step.File)
	a := firstFailingAssertion(step)
	if path == "" || a == nil || a.Source == nil {
		return "", ""
	}
	return fmt.Sprintf("%s:%d", path, a.Source.Start.Line), a.Pattern
}

func junitLocationLine(step core.StepResult) string {
	if path, line := failureHeaderLocation(step); path != "" && line > 0 {
		return fmt.Sprintf("Location: %s:%d", path, line)
	}
	if loc := commandFailureLocation(step); loc != "" {
		return "Location: " + loc
	}
	return ""
}
