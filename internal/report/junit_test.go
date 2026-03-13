package report

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/runkids/mdproof/internal/core"
)

func TestJUnit_XMLDeclaration(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{newTestReport()}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}
	if !strings.HasPrefix(buf.String(), "<?xml") {
		t.Error("output should start with XML declaration")
	}
}

func TestJUnit_SingleReport_RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{newTestReport()}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	if len(got.Testsuites) != 1 {
		t.Fatalf("len(Testsuites) = %d, want 1", len(got.Testsuites))
	}
	suite := got.Testsuites[0]
	if suite.Name != "test-runbook.md" {
		t.Errorf("suite.Name = %q, want %q", suite.Name, "test-runbook.md")
	}
	if suite.Tests != 3 {
		t.Errorf("suite.Tests = %d, want 3", suite.Tests)
	}
	if len(suite.Testcases) != 3 {
		t.Errorf("len(Testcases) = %d, want 3", len(suite.Testcases))
	}
}

func TestJUnit_MultipleReports(t *testing.T) {
	r1 := newTestReport()
	r2 := newTestReport()
	r2.Runbook = "second.md"
	r2.Summary.Total = 2
	r2.Summary.Passed = 1
	r2.Summary.Failed = 1
	r2.Summary.Skipped = 0

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r1, r2}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	if len(got.Testsuites) != 2 {
		t.Fatalf("len(Testsuites) = %d, want 2", len(got.Testsuites))
	}
	if got.Tests != 5 {
		t.Errorf("root Tests = %d, want 5", got.Tests)
	}
	if got.Failures != 1 {
		t.Errorf("root Failures = %d, want 1", got.Failures)
	}
}

func TestJUnit_FailedStep(t *testing.T) {
	r := newTestReport()
	r.Steps[0].Status = core.StatusFailed
	r.Steps[0].Step.File = "docs/test-runbook.md"
	r.Steps[0].Step.HeadingSource = core.SourceRange{
		Start: core.SourcePos{Line: 7},
		End:   core.SourcePos{Line: 7},
	}
	r.Steps[0].Assertions = []core.AssertionResult{
		{
			Pattern: "expected output",
			Type:    "substring",
			Matched: false,
			Detail:  "not found in stdout",
			Source: &core.SourceRange{
				Start: core.SourcePos{Line: 12},
				End:   core.SourcePos{Line: 12},
			},
		},
	}
	r.Steps[0].Source = core.StepSourceFromStep(r.Steps[0].Step)
	r.Summary.Passed = 1
	r.Summary.Failed = 1

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	tc := got.Testsuites[0].Testcases[0]
	if tc.Failure == nil {
		t.Fatal("expected <failure> element for failed step")
	}
	if tc.Failure.Message != "expected output" {
		t.Errorf("failure.Message = %q, want %q", tc.Failure.Message, "expected output")
	}
	if tc.Failure.Type != "substring" {
		t.Errorf("failure.Type = %q, want %q", tc.Failure.Type, "substring")
	}
	if !strings.Contains(tc.Failure.Body, "not found in stdout") {
		t.Errorf("failure.Body should contain detail, got %q", tc.Failure.Body)
	}
	if !strings.HasPrefix(tc.Failure.Body, "Location: docs/test-runbook.md:12") {
		t.Errorf("failure.Body should start with location, got %q", tc.Failure.Body)
	}
}

func TestJUnit_SkippedStep(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{newTestReport()}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	// Step 3 in newTestReport is "skipped"
	tc := got.Testsuites[0].Testcases[2]
	if tc.Skipped == nil {
		t.Fatal("expected <skipped> element for skipped step")
	}
}

func TestJUnit_SystemOut(t *testing.T) {
	r := newTestReport()
	r.Steps[0].Stdout = "hello world\n"

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	tc := got.Testsuites[0].Testcases[0]
	if tc.SystemOut != "hello world\n" {
		t.Errorf("system-out = %q, want %q", tc.SystemOut, "hello world\n")
	}
}

func TestJUnit_TimeFormat(t *testing.T) {
	r := newTestReport()
	r.DurationMs = 1500
	r.Steps[0].DurationMs = 1500

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	if got.Testsuites[0].Time != "1.500" {
		t.Errorf("suite time = %q, want %q", got.Testsuites[0].Time, "1.500")
	}
	if got.Testsuites[0].Testcases[0].Time != "1.500" {
		t.Errorf("testcase time = %q, want %q", got.Testsuites[0].Testcases[0].Time, "1.500")
	}
}

func TestJUnit_MultipleFailedAssertions(t *testing.T) {
	r := newTestReport()
	r.Steps[0].Status = core.StatusFailed
	r.Steps[0].Assertions = []core.AssertionResult{
		{Pattern: "first pattern", Type: "substring", Matched: false, Detail: "not found"},
		{Pattern: "second pattern", Type: "regex", Matched: false, Detail: "no match"},
		{Pattern: "ok pattern", Type: "substring", Matched: true},
	}
	r.Summary.Passed = 1
	r.Summary.Failed = 1

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	tc := got.Testsuites[0].Testcases[0]
	if tc.Failure == nil {
		t.Fatal("expected <failure> element")
	}
	// Message/Type come from the first failed assertion
	if tc.Failure.Message != "first pattern" {
		t.Errorf("failure.Message = %q, want %q", tc.Failure.Message, "first pattern")
	}
	if tc.Failure.Type != "substring" {
		t.Errorf("failure.Type = %q, want %q", tc.Failure.Type, "substring")
	}
	// Body should contain both failed assertions
	if !strings.Contains(tc.Failure.Body, "first pattern") {
		t.Error("failure.Body missing first pattern")
	}
	if !strings.Contains(tc.Failure.Body, "second pattern") {
		t.Error("failure.Body missing second pattern")
	}
}

func TestJUnit_FailedByExitCodeOnly(t *testing.T) {
	r := newTestReport()
	r.Steps[0].Status = core.StatusFailed
	r.Steps[0].ExitCode = 2
	r.Steps[0].Assertions = nil // no assertions — failure is purely from exit code
	r.Summary.Passed = 1
	r.Summary.Failed = 1

	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{r}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	tc := got.Testsuites[0].Testcases[0]
	if tc.Failure == nil {
		t.Fatal("expected <failure> element for exit code failure")
	}
	if !strings.Contains(tc.Failure.Message, "exit code 2") {
		t.Errorf("failure.Message = %q, want it to contain 'exit code 2'", tc.Failure.Message)
	}
}

func TestJUnit_EmptyReports(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJUnitReport(&buf, []core.Report{}); err != nil {
		t.Fatalf("WriteJUnitReport: %v", err)
	}

	var got junitTestsuites
	if err := xml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("xml.Unmarshal: %v", err)
	}

	if got.Tests != 0 {
		t.Errorf("Tests = %d, want 0", got.Tests)
	}
	if len(got.Testsuites) != 0 {
		t.Errorf("len(Testsuites) = %d, want 0", len(got.Testsuites))
	}
}
