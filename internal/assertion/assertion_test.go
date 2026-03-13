package assertion

import (
	"strings"
	"testing"

	"github.com/runkids/mdproof/internal/core"
)

func TestMatchAssertions_SubstringMatch(t *testing.T) {
	results := MatchAssertions("config_created=yes\nstatus=ok", []string{"config_created=yes"})
	if len(results) != 1 || !results[0].Matched {
		t.Fatal("expected substring match")
	}
}

func TestMatchAssertions_NegatedPasses(t *testing.T) {
	results := MatchAssertions("all good", []string{"No error found"})
	if len(results) != 1 || !results[0].Matched || !results[0].Negated {
		t.Fatal("negated pattern should pass when text is absent")
	}
}

func TestMatchAssertions_NegatedFails(t *testing.T) {
	results := MatchAssertions("error found in log", []string{"No error found"})
	if len(results) != 1 || results[0].Matched {
		t.Fatal("negated pattern should fail when text is present")
	}
	if !results[0].Negated {
		t.Fatal("expected Negated flag")
	}
	if results[0].Detail == "" {
		t.Fatal("negated failure should include detail showing matched line")
	}
	if !strings.Contains(results[0].Detail, "error found") {
		t.Fatalf("detail should show matched context, got: %s", results[0].Detail)
	}
}

func TestMatchAssertions_NegatedWordBoundary(t *testing.T) {
	// "Not FAIL" should NOT trigger on "0 failed" — word boundary prevents it.
	results := MatchAssertions("Uninstall complete: 2 removed, 0 failed (0.0s)", []string{"Not FAIL"})
	if !results[0].Matched {
		t.Fatal("Not FAIL should pass because 'failed' is not the word 'FAIL'")
	}
}

func TestMatchAssertions_NegatedWordBoundary_StillCatches(t *testing.T) {
	// "Not FAIL" SHOULD trigger on "FAIL: something" — exact word match.
	results := MatchAssertions("FAIL: config has skills", []string{"Not FAIL"})
	if results[0].Matched {
		t.Fatal("Not FAIL should fail because 'FAIL' appears as a word")
	}
	if results[0].Detail == "" {
		t.Fatal("should include detail on failure")
	}
}

func TestMatchAssertions_NegatedWordBoundary_Standalone(t *testing.T) {
	// "Not FAIL" triggers on standalone "FAIL" at end of line.
	results := MatchAssertions("test result: FAIL", []string{"Not FAIL"})
	if results[0].Matched {
		t.Fatal("Not FAIL should fail on standalone FAIL")
	}
}

func TestMatchAssertions_EqualsStyle(t *testing.T) {
	results := MatchAssertions("claude_ok=yes\nother=no", []string{"claude_ok=yes"})
	if len(results) != 1 || !results[0].Matched {
		t.Fatal("equals-style pattern should match")
	}
}

func TestMatchAssertions_CaseInsensitive(t *testing.T) {
	results := MatchAssertions("Config_Created=YES", []string{"config_created=yes"})
	if len(results) != 1 || !results[0].Matched {
		t.Fatal("matching should be case-insensitive")
	}
}

func TestMatchAssertions_PatternNotFound(t *testing.T) {
	results := MatchAssertions("nothing here", []string{"missing_key=true"})
	if len(results) != 1 || results[0].Matched {
		t.Fatal("pattern not found should yield matched=false")
	}
}

func TestAllPassed_AllTrue(t *testing.T) {
	results := []core.AssertionResult{
		{Pattern: "a", Matched: true},
		{Pattern: "b", Matched: true},
	}
	if !AllPassed(results) {
		t.Fatal("AllPassed should return true when all matched")
	}
}

func TestAllPassed_OneFalse(t *testing.T) {
	results := []core.AssertionResult{
		{Pattern: "a", Matched: true},
		{Pattern: "b", Matched: false},
	}
	if AllPassed(results) {
		t.Fatal("AllPassed should return false when one is unmatched")
	}
}

func TestMatchAssertions_Multiple(t *testing.T) {
	output := "status=ok\ncount=3\nmode=merge"
	expected := []string{"status=ok", "count=3", "Not missing_key", "mode=merge"}

	results := MatchAssertions(output, expected)
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for i, r := range results {
		if !r.Matched {
			t.Errorf("result[%d] (%s) should have matched", i, r.Pattern)
		}
	}
	if !results[2].Negated {
		t.Error("result[2] should be negated")
	}
}

func TestMatchAssertions_EmptyExpected(t *testing.T) {
	results := MatchAssertions("some output", nil)
	if len(results) != 0 {
		t.Fatalf("empty expected should return empty results, got %d", len(results))
	}
}

func TestRunAssertions_RegexMultilineDefault(t *testing.T) {
	// ^pattern$ should match individual lines, not just the whole string.
	result := &core.StepResult{
		Stdout:   "line1\n42\nline3",
		Stderr:   "",
		ExitCode: 0,
	}
	results := RunAssertions(result, core.Expectations(`regex: ^42$`))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Matched {
		t.Fatal("regex ^42$ should match line '42' in multi-line output with auto (?m)")
	}
}

func TestRunAssertions_RegexExplicitFlags(t *testing.T) {
	// When pattern already has (? flags, don't prepend (?m).
	result := &core.StepResult{
		Stdout:   "ABC\ndef",
		Stderr:   "",
		ExitCode: 0,
	}
	results := RunAssertions(result, core.Expectations(`regex: (?i)abc`))
	if len(results) != 1 || !results[0].Matched {
		t.Fatal("regex with explicit (?i) flag should still work")
	}
}

func TestRunAssertions_RegexWithoutMultiline(t *testing.T) {
	// Verify ^ matches line boundary (not just string start).
	result := &core.StepResult{
		Stdout:   "header\nfoo=bar\ntrailer",
		Stderr:   "",
		ExitCode: 0,
	}
	results := RunAssertions(result, core.Expectations(`regex: ^foo=bar$`))
	if len(results) != 1 || !results[0].Matched {
		t.Fatal("regex ^foo=bar$ should match middle line with auto (?m)")
	}
}

// Typed assertion tests

func TestRunAssertions_Substring(t *testing.T) {
	r := &core.StepResult{Stdout: "hello world", ExitCode: 0}
	results := RunAssertions(r, core.Expectations("hello", "Not missing"))

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Matched || results[0].Type != core.AssertSubstring {
		t.Errorf("'hello' should match as substring")
	}
	if !results[1].Matched || !results[1].Negated {
		t.Errorf("'Not missing' should match (negated)")
	}
}

func TestRunAssertions_PreservesExpectationSource(t *testing.T) {
	r := &core.StepResult{Stdout: "hello world", ExitCode: 0}
	results := RunAssertions(r, []core.Expectation{{
		Text: "hello",
		Source: core.SourceRange{
			Start: core.SourcePos{Line: 18},
			End:   core.SourcePos{Line: 18},
		},
	}})

	if len(results) != 1 || results[0].Source == nil {
		t.Fatalf("expected assertion source to be preserved, got %+v", results)
	}
	if results[0].Source.Start.Line != 18 {
		t.Fatalf("source line = %d, want 18", results[0].Source.Start.Line)
	}
}

func TestRunAssertions_ExitCode_Match(t *testing.T) {
	r := &core.StepResult{Stdout: "", ExitCode: 0}
	results := RunAssertions(r, core.Expectations("exit_code: 0"))

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Matched {
		t.Errorf("exit_code: 0 should match when exit code is 0")
	}
	if results[0].Type != core.AssertExitCode {
		t.Errorf("type should be exit_code, got %s", results[0].Type)
	}
}

func TestRunAssertions_ExitCode_Mismatch(t *testing.T) {
	r := &core.StepResult{Stdout: "", ExitCode: 1}
	results := RunAssertions(r, core.Expectations("exit_code: 0"))

	if results[0].Matched {
		t.Errorf("exit_code: 0 should NOT match when exit code is 1")
	}
	if results[0].Detail == "" {
		t.Errorf("should have detail on mismatch")
	}
}

func TestRunAssertions_ExitCode_Negated(t *testing.T) {
	r := &core.StepResult{Stdout: "", ExitCode: 1}
	results := RunAssertions(r, core.Expectations("exit_code: !0"))

	if !results[0].Matched {
		t.Errorf("exit_code: !0 should match when exit code is 1")
	}
	if !results[0].Negated {
		t.Errorf("should be flagged as negated")
	}
}

func TestRunAssertions_ExitCode_NegatedFail(t *testing.T) {
	r := &core.StepResult{Stdout: "", ExitCode: 0}
	results := RunAssertions(r, core.Expectations("exit_code: !0"))

	if results[0].Matched {
		t.Errorf("exit_code: !0 should NOT match when exit code is 0")
	}
}

func TestRunAssertions_Regex_Match(t *testing.T) {
	r := &core.StepResult{Stdout: "synced 42 skills in 1.2s", ExitCode: 0}
	results := RunAssertions(r, core.Expectations(`regex: \d+ skills`))

	if !results[0].Matched {
		t.Errorf("regex should match")
	}
	if results[0].Type != core.AssertRegex {
		t.Errorf("type should be regex, got %s", results[0].Type)
	}
}

func TestRunAssertions_Regex_NoMatch(t *testing.T) {
	r := &core.StepResult{Stdout: "no numbers here", ExitCode: 0}
	results := RunAssertions(r, core.Expectations(`regex: \d+ skills`))

	if results[0].Matched {
		t.Errorf("regex should not match")
	}
}

func TestRunAssertions_Regex_Invalid(t *testing.T) {
	r := &core.StepResult{Stdout: "test", ExitCode: 0}
	results := RunAssertions(r, core.Expectations(`regex: [invalid`))

	if results[0].Matched {
		t.Errorf("invalid regex should not match")
	}
	if results[0].Detail == "" {
		t.Errorf("should have error detail")
	}
}

func TestRunAssertions_JQ_Match(t *testing.T) {
	r := &core.StepResult{Stdout: `{"count": 5, "status": "ok"}`, ExitCode: 0}
	results := RunAssertions(r, core.Expectations("jq: .count > 0"))

	if !results[0].Matched {
		t.Errorf("jq should match: %s", results[0].Detail)
	}
	if results[0].Type != core.AssertJQ {
		t.Errorf("type should be jq, got %s", results[0].Type)
	}
}

func TestRunAssertions_JQ_NoMatch(t *testing.T) {
	r := &core.StepResult{Stdout: `{"count": 0}`, ExitCode: 0}
	results := RunAssertions(r, core.Expectations("jq: .count > 0"))

	if results[0].Matched {
		t.Errorf("jq should not match when count is 0")
	}
}

func TestRunAssertions_JQ_InvalidJSON(t *testing.T) {
	r := &core.StepResult{Stdout: "not json", ExitCode: 0}
	results := RunAssertions(r, core.Expectations("jq: .count"))

	if results[0].Matched {
		t.Errorf("jq should fail on non-JSON input")
	}
	if results[0].Detail == "" {
		t.Errorf("should have error detail")
	}
}

func TestRunAssertions_Mixed(t *testing.T) {
	r := &core.StepResult{
		Stdout:   `{"installed": true, "name": "my-skill"}`,
		Stderr:   "Installed: my-skill",
		ExitCode: 0,
	}
	results := RunAssertions(r, core.Expectations(
		"Installed",      // substring on combined
		"exit_code: 0",   // exit code
		"jq: .installed", // jq on stdout only
		"Not error",      // negated substring
	))

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for i, res := range results {
		if !res.Matched {
			t.Errorf("result[%d] (%s) should have matched: %s", i, res.Pattern, res.Detail)
		}
	}
}

func TestCheckStep_NonZeroExitWithExitCodeAssertion(t *testing.T) {
	result := core.StepResult{
		Step:     core.Step{Expected: core.Expectations("exit_code: 1", "error")},
		Status:   core.StatusFailed,
		ExitCode: 1,
		Stdout:   "error: something went wrong",
	}

	CheckStep(&result, result.Step)

	// With exit_code: 1 assertion, the step should PASS even though exit code is non-zero.
	if result.Status != core.StatusPassed {
		t.Errorf("expected passed (exit_code: 1 matches), got %s", result.Status)
	}
	if len(result.Assertions) != 2 {
		t.Fatalf("expected 2 assertions, got %d", len(result.Assertions))
	}
}

func TestCheckJQ_MeaningfulError(t *testing.T) {
	r := &core.StepResult{Stdout: `{"a":1}`, ExitCode: 0}
	results := RunAssertions(r, core.Expectations("jq: .a == 1"))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Either jq is installed and the assertion matches, or it's not installed
	// and we get a helpful error message — both are valid outcomes.
	if !results[0].Matched && !strings.Contains(results[0].Detail, "not installed") {
		t.Errorf("expected match or 'not installed' message, got: %s", results[0].Detail)
	}
}

func TestCheckStep_NoAssertions_PreservesStatus(t *testing.T) {
	result := core.StepResult{
		Step:     core.Step{Expected: nil},
		Status:   core.StatusFailed,
		ExitCode: 1,
	}

	CheckStep(&result, result.Step)

	// No assertions -> status unchanged.
	if result.Status != core.StatusFailed {
		t.Errorf("expected status unchanged (failed), got %s", result.Status)
	}
}
