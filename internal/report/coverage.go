package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/runkids/mdproof/internal/coverage"
)

// CoverageEntry pairs a file name with its coverage result.
type CoverageEntry struct {
	File   string
	Result coverage.Result
}

// WriteCoverageReport writes a plain-text coverage table.
func WriteCoverageReport(w io.Writer, entries []CoverageEntry) {
	fmt.Fprintf(w, "\n mdproof coverage report\n")
	fmt.Fprintf(w, " %s\n", strings.Repeat("\u2500", 65))

	fmt.Fprintf(w, " %-30s %5s  %7s  %10s  %5s\n", "File", "Steps", "Covered", "Assertions", "Score")
	fmt.Fprintf(w, " %s\n", strings.Repeat("\u2500", 65))

	totalCoverable, totalCovered, totalAssertions := 0, 0, 0

	for _, e := range entries {
		fmt.Fprintf(w, " %-30s %5d  %7d  %10d   %3d%%\n",
			e.File, e.Result.CoverableSteps, e.Result.CoveredSteps,
			e.Result.TotalAssertions, e.Result.Score)
		totalCoverable += e.Result.CoverableSteps
		totalCovered += e.Result.CoveredSteps
		totalAssertions += e.Result.TotalAssertions
	}

	totalScore := 100
	if totalCoverable > 0 {
		totalScore = (totalCovered * 100) / totalCoverable
	}

	fmt.Fprintf(w, " %s\n", strings.Repeat("\u2500", 65))
	fmt.Fprintf(w, " %-30s %5d  %7d  %10d   %3d%%\n",
		"Total", totalCoverable, totalCovered, totalAssertions, totalScore)
	fmt.Fprintln(w)

	for _, e := range entries {
		if len(e.Result.UncoveredSteps) > 0 {
			nums := make([]string, len(e.Result.UncoveredSteps))
			for i, n := range e.Result.UncoveredSteps {
				nums[i] = fmt.Sprintf("%d", n)
			}
			fmt.Fprintf(w, " ! %s: Step %s have no assertions\n", e.File, strings.Join(nums, ", "))
		}
		if e.Result.LowDiversity {
			fmt.Fprintf(w, " ! %s: only uses substring assertions\n", e.File)
		}
	}
	fmt.Fprintln(w)
}

// TotalScore computes the aggregate score across all entries.
func TotalScore(entries []CoverageEntry) int {
	totalCoverable, totalCovered := 0, 0
	for _, e := range entries {
		totalCoverable += e.Result.CoverableSteps
		totalCovered += e.Result.CoveredSteps
	}
	if totalCoverable == 0 {
		return 100
	}
	return (totalCovered * 100) / totalCoverable
}
