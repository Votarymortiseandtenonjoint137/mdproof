package coverage

import (
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

// Result holds coverage analysis for a set of steps.
type Result struct {
	CoverableSteps  int
	CoveredSteps    int
	TotalAssertions int
	Score           int
	UncoveredSteps  []int
	LowDiversity    bool
}

// Analyze computes coverage metrics for a list of steps.
func Analyze(steps []core.Step) Result {
	var r Result

	hasNonSubstring := false
	coveredCount := 0

	for _, s := range steps {
		if !isCoverable(s) {
			continue
		}

		r.CoverableSteps++
		r.TotalAssertions += len(s.Expected)

		if len(s.Expected) > 0 {
			r.CoveredSteps++
			coveredCount++

			for _, exp := range s.Expected {
				if isTypedAssertion(exp) {
					hasNonSubstring = true
				}
			}
		} else {
			r.UncoveredSteps = append(r.UncoveredSteps, s.Number)
		}
	}

	if r.CoverableSteps == 0 {
		r.Score = 100
	} else {
		r.Score = (r.CoveredSteps * 100) / r.CoverableSteps
	}

	if r.CoverableSteps >= 3 && coveredCount > 0 && !hasNonSubstring {
		r.LowDiversity = true
	}

	return r
}

func isCoverable(s core.Step) bool {
	if s.Command == "" {
		return false
	}
	switch s.Lang {
	case "bash", "sh", "":
		return true
	default:
		return false
	}
}

func isTypedAssertion(pat string) bool {
	prefixes := []string{"exit_code:", "regex:", "jq:", "snapshot:"}
	for _, p := range prefixes {
		if strings.HasPrefix(pat, p) {
			return true
		}
	}
	return false
}
