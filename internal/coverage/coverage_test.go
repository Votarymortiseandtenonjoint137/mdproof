package coverage

import (
	"testing"

	"github.com/runkids/mdproof/internal/core"
)

func TestAnalyze_AllCovered(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "Create user", Command: "curl ...", Lang: "bash", Expected: []string{"exit_code: 0", "alice"}},
		{Number: 2, Title: "Verify user", Command: "curl ...", Lang: "bash", Expected: []string{"jq: .id != null"}},
	}

	result := Analyze(steps)

	if result.Score != 100 {
		t.Fatalf("expected score 100, got %d", result.Score)
	}
	if result.CoverableSteps != 2 || result.CoveredSteps != 2 {
		t.Fatalf("expected 2/2 covered, got %d/%d", result.CoveredSteps, result.CoverableSteps)
	}
}

func TestAnalyze_PartialCoverage(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "Create", Command: "curl ...", Lang: "bash", Expected: []string{"ok"}},
		{Number: 2, Title: "Verify", Command: "curl ...", Lang: "bash"},
		{Number: 3, Title: "Delete", Command: "curl ...", Lang: "bash", Expected: []string{"deleted"}},
	}

	result := Analyze(steps)

	if result.Score != 66 {
		t.Fatalf("expected score 66 (2/3), got %d", result.Score)
	}
	if len(result.UncoveredSteps) != 1 || result.UncoveredSteps[0] != 2 {
		t.Fatalf("expected step 2 uncovered, got %v", result.UncoveredSteps)
	}
}

func TestAnalyze_ManualStepsExcluded(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "Auto", Command: "curl ...", Lang: "bash", Expected: []string{"ok"}},
		{Number: 2, Title: "Manual check", Lang: ""},
		{Number: 3, Title: "Auto no assert", Command: "echo hi", Lang: "bash"},
	}

	result := Analyze(steps)

	if result.CoverableSteps != 2 {
		t.Fatalf("expected 2 coverable (manual excluded), got %d", result.CoverableSteps)
	}
	if result.Score != 50 {
		t.Fatalf("expected score 50, got %d", result.Score)
	}
}

func TestAnalyze_TypeDiversityWarning(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Command: "a", Lang: "bash", Expected: []string{"hello"}},
		{Number: 2, Command: "b", Lang: "bash", Expected: []string{"world"}},
		{Number: 3, Command: "c", Lang: "bash", Expected: []string{"foo"}},
	}

	result := Analyze(steps)

	if !result.LowDiversity {
		t.Fatal("3+ steps with only substring assertions should trigger low diversity warning")
	}
}

func TestAnalyze_TypeDiversityOK(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Command: "a", Lang: "bash", Expected: []string{"hello"}},
		{Number: 2, Command: "b", Lang: "bash", Expected: []string{"exit_code: 0"}},
		{Number: 3, Command: "c", Lang: "bash", Expected: []string{"regex: .*"}},
	}

	result := Analyze(steps)

	if result.LowDiversity {
		t.Fatal("mixed assertion types should not trigger low diversity")
	}
}

func TestAnalyze_Empty(t *testing.T) {
	result := Analyze(nil)
	if result.Score != 100 {
		t.Fatalf("empty runbook should have 100%% score, got %d", result.Score)
	}
}
