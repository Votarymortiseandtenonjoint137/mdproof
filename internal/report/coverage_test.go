package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/runkids/mdproof/internal/coverage"
)

func TestWriteCoverageReport_SingleFile(t *testing.T) {
	var buf bytes.Buffer
	results := []CoverageEntry{
		{
			File:   "deploy-proof.md",
			Result: coverage.Result{CoverableSteps: 5, CoveredSteps: 5, TotalAssertions: 12, Score: 100},
		},
	}

	WriteCoverageReport(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "deploy-proof.md") {
		t.Fatal("should contain filename")
	}
	if !strings.Contains(out, "100%") {
		t.Fatal("should contain 100%")
	}
}

func TestWriteCoverageReport_WithWarnings(t *testing.T) {
	var buf bytes.Buffer
	results := []CoverageEntry{
		{
			File: "api-proof.md",
			Result: coverage.Result{
				CoverableSteps: 4, CoveredSteps: 2, TotalAssertions: 2,
				Score: 50, UncoveredSteps: []int{3, 4}, LowDiversity: true,
			},
		},
	}

	WriteCoverageReport(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "50%") {
		t.Fatal("should contain 50%")
	}
	if !strings.Contains(out, "Step 3, 4") {
		t.Fatalf("should warn about uncovered steps, got: %s", out)
	}
	if !strings.Contains(out, "only uses substring") {
		t.Fatal("should warn about low diversity")
	}
}
