package runner

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/runkids/mdproof/internal/core"
)

func TestMain(m *testing.M) {
	// Tests run on the host — allow execution for test suite.
	os.Setenv("MDPROOF_ALLOW_EXECUTE", "1")
	os.Exit(m.Run())
}

func makeRunbook(steps string) string {
	return "# Test Runbook\n\n## Steps\n\n" + steps
}

func TestRun_SimplePass(t *testing.T) {
	md := makeRunbook(`### Step 1: Echo hello

` + "```bash" + `
echo hello
` + "```" + `

**Expected:**
- hello
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 1 {
		t.Fatalf("expected 1 step, got %d", report.Summary.Total)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Summary.Passed)
	}
	if report.Steps[0].Status != "passed" {
		t.Errorf("expected passed, got %s", report.Steps[0].Status)
	}
}

func TestRun_Failure(t *testing.T) {
	md := makeRunbook(`### Step 1: Fail

` + "```bash" + `
echo "nope" && exit 1
` + "```" + `

**Expected:**
- success
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Summary.Failed)
	}
	if report.Steps[0].Status != "failed" {
		t.Errorf("expected failed, got %s", report.Steps[0].Status)
	}
}

func TestRun_SkipsManual(t *testing.T) {
	md := makeRunbook(`### Step 1: Manual step

` + "```go" + `
fmt.Println("manual")
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Summary.Skipped)
	}
	if report.Steps[0].Status != "skipped" {
		t.Errorf("expected skipped, got %s", report.Steps[0].Status)
	}
}

func TestRun_DryRun(t *testing.T) {
	md := makeRunbook(`### Step 1: Echo

` + "```bash" + `
echo should not run
` + "```" + `

### Step 2: Another

` + "```bash" + `
echo also not run
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 2 {
		t.Fatalf("expected 2 steps, got %d", report.Summary.Total)
	}
	if report.Summary.Skipped != 2 {
		t.Errorf("expected 2 skipped, got %d", report.Summary.Skipped)
	}
	for i, sr := range report.Steps {
		if sr.Status != "skipped" {
			t.Errorf("step %d: expected skipped, got %s", i, sr.Status)
		}
		if sr.Stdout != "" {
			t.Errorf("step %d: expected no stdout in dry run, got %q", i, sr.Stdout)
		}
	}
}

func TestRun_AssertionFailureExitZero(t *testing.T) {
	md := makeRunbook(`### Step 1: Wrong output

` + "```bash" + `
echo "apple orange"
` + "```" + `

**Expected:**
- banana
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Steps[0].Status != "failed" {
		t.Errorf("expected failed due to assertion mismatch, got %s", report.Steps[0].Status)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Summary.Failed)
	}
	if len(report.Steps[0].Assertions) != 1 {
		t.Fatalf("expected 1 assertion, got %d", len(report.Steps[0].Assertions))
	}
	if report.Steps[0].Assertions[0].Matched {
		t.Error("expected assertion to not match")
	}
}

func TestShouldRun_NoFilter(t *testing.T) {
	opts := RunOptions{}
	for _, n := range []int{1, 2, 3, 10} {
		if !opts.shouldRun(n) {
			t.Errorf("shouldRun(%d) = false, want true (no filter)", n)
		}
	}
}

func TestShouldRun_StepsFilter(t *testing.T) {
	opts := RunOptions{Steps: []int{1, 3}}
	cases := []struct {
		step int
		want bool
	}{
		{1, true}, {2, false}, {3, true}, {4, false},
	}
	for _, tc := range cases {
		if got := opts.shouldRun(tc.step); got != tc.want {
			t.Errorf("shouldRun(%d) = %v, want %v", tc.step, got, tc.want)
		}
	}
}

func TestShouldRun_FromFilter(t *testing.T) {
	opts := RunOptions{From: 3}
	cases := []struct {
		step int
		want bool
	}{
		{1, false}, {2, false}, {3, true}, {4, true}, {10, true},
	}
	for _, tc := range cases {
		if got := opts.shouldRun(tc.step); got != tc.want {
			t.Errorf("shouldRun(%d) = %v, want %v", tc.step, got, tc.want)
		}
	}
}

func TestRun_StepsFilter(t *testing.T) {
	md := makeRunbook(`### Step 1: First

` + "```bash" + `
echo one
` + "```" + `

### Step 2: Second

` + "```bash" + `
echo two
` + "```" + `

### Step 3: Third

` + "```bash" + `
echo three
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{Steps: []int{1, 3}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 3 {
		t.Fatalf("expected 3 steps, got %d", report.Summary.Total)
	}
	if report.Summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", report.Summary.Passed)
	}
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Summary.Skipped)
	}
	// Step 2 should be skipped.
	if report.Steps[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped, got %s", report.Steps[1].Status)
	}
}

func TestRun_FromFilter(t *testing.T) {
	md := makeRunbook(`### Step 1: First

` + "```bash" + `
echo one
` + "```" + `

### Step 2: Second

` + "```bash" + `
echo two
` + "```" + `

### Step 3: Third

` + "```bash" + `
echo three
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{From: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 3 {
		t.Fatalf("expected 3 steps, got %d", report.Summary.Total)
	}
	if report.Summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", report.Summary.Passed)
	}
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Summary.Skipped)
	}
	// Step 1 should be skipped.
	if report.Steps[0].Status != core.StatusSkipped {
		t.Errorf("step 1: expected skipped, got %s", report.Steps[0].Status)
	}
}

func TestRun_FailFast(t *testing.T) {
	md := makeRunbook(`### Step 1: Fail early

` + "```bash" + `
echo "step1" && exit 1
` + "```" + `

### Step 2: Should skip

` + "```bash" + `
echo "step2"
` + "```" + `

### Step 3: Should also skip

` + "```bash" + `
echo "step3"
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{FailFast: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 3 {
		t.Fatalf("expected 3 steps, got %d", report.Summary.Total)
	}
	if report.Steps[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed, got %s", report.Steps[0].Status)
	}
	if report.Steps[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped, got %s", report.Steps[1].Status)
	}
	if report.Steps[2].Status != core.StatusSkipped {
		t.Errorf("step 3: expected skipped, got %s", report.Steps[2].Status)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Summary.Failed)
	}
	if report.Summary.Skipped != 2 {
		t.Errorf("expected 2 skipped, got %d", report.Summary.Skipped)
	}
}

func TestRun_FailFastAssertionOnly(t *testing.T) {
	md := makeRunbook(`### Step 1: Exit 0 but assertion fails

` + "```bash" + `
echo "apple"
` + "```" + `

**Expected:**
- banana

### Step 2: Should skip

` + "```bash" + `
echo "step2"
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{FailFast: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Steps[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed (assertion), got %s", report.Steps[0].Status)
	}
	if report.Steps[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped, got %s", report.Steps[1].Status)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Summary.Failed)
	}
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Summary.Skipped)
	}
}

func TestRun_EnvSeeding(t *testing.T) {
	md := makeRunbook(`### Step 1: Check env

` + "```bash" + `
echo "MY_VAR=$MY_VAR"
` + "```" + `

**Expected:**
- MY_VAR=hello
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{
		Env: map[string]string{"MY_VAR": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Steps[0].Status != core.StatusPassed {
		t.Errorf("expected passed, got %s (stdout=%q)", report.Steps[0].Status, report.Steps[0].Stdout)
	}
}

func TestRun_DependsOn(t *testing.T) {
	md := makeRunbook(`### Step 1: Fail

` + "```bash" + `
echo "step1" && exit 1
` + "```" + `

### Step 2: Depends on step 1

<!-- runbook: depends=1 -->

` + "```bash" + `
echo "step2"
` + "```" + `

### Step 3: Independent

` + "```bash" + `
echo "step3"
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Total != 3 {
		t.Fatalf("expected 3 steps, got %d", report.Summary.Total)
	}
	// Step 1: fails.
	if report.Steps[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed, got %s", report.Steps[0].Status)
	}
	// Step 2: skipped because depends on step 1.
	if report.Steps[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped (depends), got %s", report.Steps[1].Status)
	}
	if !strings.Contains(report.Steps[1].Error, "depends") {
		t.Errorf("step 2 error should mention depends, got: %q", report.Steps[1].Error)
	}
	// Step 3: independent, should pass.
	if report.Steps[2].Status != core.StatusPassed {
		t.Errorf("step 3: expected passed (independent), got %s", report.Steps[2].Status)
	}
}

func TestRun_Retry(t *testing.T) {
	// Step uses a counter file to fail on first attempt and pass on second.
	md := makeRunbook(`### Step 1: Retry step

<!-- runbook: retry=2 -->

` + "```bash" + `
F="/tmp/runbook_retry_test_$$"
if [ -f "$F" ]; then
  echo "passed"
  rm -f "$F"
else
  touch "$F"
  echo "failed" && exit 1
fi
` + "```" + `

**Expected:**
- passed
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Steps[0].Status != core.StatusPassed {
		t.Errorf("expected passed after retry, got %s (stdout=%q, stderr=%q)",
			report.Steps[0].Status, report.Steps[0].Stdout, report.Steps[0].Stderr)
	}
}

func TestRun_DependsOnWithFailFast(t *testing.T) {
	// Ensure depends-skip doesn't cascade via fail-fast.
	md := makeRunbook(`### Step 1: Fail

` + "```bash" + `
echo "step1" && exit 1
` + "```" + `

### Step 2: Depends on step 1

<!-- runbook: depends=1 -->

` + "```bash" + `
echo "step2"
` + "```" + `

### Step 3: Independent

` + "```bash" + `
echo "step3"
` + "```" + `
`)

	report, err := Run(strings.NewReader(md), "test", RunOptions{FailFast: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Step 1 fails -> fail-fast triggers.
	if report.Steps[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed, got %s", report.Steps[0].Status)
	}
	// Step 2: skipped (fail-fast, since step 1 already failed).
	if report.Steps[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped, got %s", report.Steps[1].Status)
	}
	// Step 3: also skipped via fail-fast.
	if report.Steps[2].Status != core.StatusSkipped {
		t.Errorf("step 3: expected skipped, got %s", report.Steps[2].Status)
	}
}

func TestRun_JSONOutput(t *testing.T) {
	md := makeRunbook(`### Step 1: Echo

` + "```bash" + `
echo ok
` + "```" + `
`)

	var buf bytes.Buffer
	_, err := Run(strings.NewReader(md), "json-test", RunOptions{JSONOutput: &buf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"runbook": "json-test"`) {
		t.Errorf("JSON output missing runbook name, got: %s", buf.String())
	}
}

// Hook tests

func TestRun_SetupPersistsEnv(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Check var
` + "```bash\necho \"VAR=$MY_VAR\"\n```" + `
**Expected:**
- exit_code: 0
- VAR=hello
`
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		Setup: "export MY_VAR=hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range report.Steps {
		t.Logf("step %d (%s): status=%s stdout=%q stderr=%q error=%q exit=%d",
			s.Step.Number, s.Step.Title, s.Status, s.Stdout, s.Stderr, s.Error, s.ExitCode)
		for _, a := range s.Assertions {
			t.Logf("  assertion: %s matched=%v detail=%q", a.Pattern, a.Matched, a.Detail)
		}
	}
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Summary.Passed)
	}
}

func TestRun_SetupFailSkipsAll(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Should be skipped
` + "```bash\necho ok\n```" + `
**Expected:**
- exit_code: 0
`
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		Setup: "exit 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d (passed=%d failed=%d)",
			report.Summary.Skipped, report.Summary.Passed, report.Summary.Failed)
	}
}

func TestRun_TeardownRuns(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Create marker
` + "```bash\necho ok\n```" + `
**Expected:**
- exit_code: 0
`
	// Teardown runs but its result doesn't affect pass/fail.
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		Teardown: "echo teardown-ran",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Summary.Passed)
	}
	// Teardown is not in steps count.
	if report.Summary.Total != 1 {
		t.Errorf("expected 1 total, got %d", report.Summary.Total)
	}
}

func TestRun_TeardownFailureIgnored(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Passes
` + "```bash\necho ok\n```" + `
**Expected:**
- exit_code: 0
`
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		Teardown: "exit 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Step still passes despite teardown failure.
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Summary.Passed)
	}
	if report.Summary.Failed != 0 {
		t.Errorf("teardown failure should not affect results, got %d failed", report.Summary.Failed)
	}
}

func TestRun_SetupAndTeardown(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Use setup var
` + "```bash\necho \"X=$X\"\n```" + `
**Expected:**
- exit_code: 0
- X=42
`
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		Setup:    "export X=42",
		Teardown: "echo done",
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Summary.Passed)
	}
}

func TestRunBuildHook_Success(t *testing.T) {
	result := RunBuildHook("echo building && echo done")
	if !result.OK {
		t.Errorf("expected OK, got exit %d", result.ExitCode)
	}
	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestRunBuildHook_Failure(t *testing.T) {
	result := RunBuildHook("exit 42")
	if result.OK {
		t.Error("expected failure")
	}
	if result.ExitCode != 42 {
		t.Errorf("exit code = %d, want 42", result.ExitCode)
	}
}

func TestRun_DryRunIgnoresHooks(t *testing.T) {
	md := `# Test
## Steps
### Step 1: Noop
` + "```bash\necho ok\n```" + `
`
	report, err := Run(strings.NewReader(md), "test.md", RunOptions{
		DryRun:   true,
		Setup:    "exit 1", // would fail if executed
		Teardown: "exit 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Dry-run skips everything including hooks.
	if report.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Summary.Skipped)
	}
}

// Integration tests

func TestIntegration_SelfContainedRunbook(t *testing.T) {
	md := `# Self-Contained Integration Test

## Scope

Verify Run end-to-end with simple bash commands.

## Steps

### Step 1: Create temp directory

` + "```bash\nDIR=$(mktemp -d) && echo \"created $DIR\" && [ -d \"$DIR\" ] && echo ok\n```" + `

Expected:

- created
- ok

### Step 2: Echo multiple lines

` + "```bash\necho hello && echo world\n```" + `

Expected:

- hello
- world

### Step 3: Arithmetic check

` + "```bash\necho $((2 + 3))\n```" + `

Expected:

- 5
`

	var jsonBuf bytes.Buffer
	report, err := Run(strings.NewReader(md), "self-contained-test", RunOptions{
		Timeout:    30 * time.Second,
		JSONOutput: &jsonBuf,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	// Verify summary counts.
	if report.Summary.Total != 3 {
		t.Errorf("Total = %d, want 3", report.Summary.Total)
	}
	if report.Summary.Passed != 3 {
		t.Errorf("Passed = %d, want 3", report.Summary.Passed)
	}
	if report.Summary.Failed != 0 {
		t.Errorf("Failed = %d, want 0", report.Summary.Failed)
	}

	// Verify each step passed.
	for i, sr := range report.Steps {
		if sr.Status != "passed" {
			t.Errorf("Step %d (%s): status = %q, want passed; stdout=%q stderr=%q",
				i+1, sr.Step.Title, sr.Status, sr.Stdout, sr.Stderr)
		}
	}

	// Verify JSON output is valid.
	if jsonBuf.Len() == 0 {
		t.Fatal("JSON output is empty")
	}
	var parsed core.Report
	if err := json.Unmarshal(jsonBuf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON output is not valid: %v\nraw: %s", err, jsonBuf.String())
	}
	if parsed.Runbook != "self-contained-test" {
		t.Errorf("JSON runbook = %q, want %q", parsed.Runbook, "self-contained-test")
	}
	if parsed.Summary.Total != 3 {
		t.Errorf("JSON summary total = %d, want 3", parsed.Summary.Total)
	}
	if parsed.Summary.Passed != 3 {
		t.Errorf("JSON summary passed = %d, want 3", parsed.Summary.Passed)
	}

	t.Logf("Report: %d total, %d passed, %d failed, %d skipped, duration=%dms",
		report.Summary.Total, report.Summary.Passed, report.Summary.Failed,
		report.Summary.Skipped, report.DurationMs)
}

func TestIntegration_ParseAllRealRunbooks(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "ai_docs", "tests")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("Skipping: directory %s does not exist: %v", dir, err)
	}

	var runbooks []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_runbook.md") {
			runbooks = append(runbooks, e.Name())
		}
	}

	if len(runbooks) == 0 {
		t.Skip("No runbook files found")
	}

	t.Logf("Found %d runbook files", len(runbooks))

	for _, name := range runbooks {
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer f.Close()

			// Use Run with DryRun to test parse + classify.
			report, err := Run(f, name, RunOptions{
				DryRun:  true,
				Timeout: 30 * time.Second,
			})
			if err != nil {
				t.Fatalf("Run error: %v", err)
			}

			if report.Summary.Total == 0 {
				t.Fatal("Should have at least 1 step")
			}

			t.Logf("  %s: %d steps", name, report.Summary.Total)
		})
	}
}

func TestIntegration_DryRunAllRunbooks(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "ai_docs", "tests")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("Skipping: directory %s does not exist: %v", dir, err)
	}

	var runbooks []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_runbook.md") {
			runbooks = append(runbooks, e.Name())
		}
	}

	if len(runbooks) == 0 {
		t.Skip("No runbook files found")
	}

	t.Logf("Found %d runbook files for dry-run", len(runbooks))

	for _, name := range runbooks {
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer f.Close()

			report, err := Run(f, name, RunOptions{
				DryRun:  true,
				Timeout: 30 * time.Second,
			})
			if err != nil {
				t.Fatalf("Run dry-run error: %v", err)
			}

			// All steps should be skipped in dry-run mode.
			for i, sr := range report.Steps {
				if sr.Status != "skipped" {
					t.Errorf("Step %d (%s): status = %q, want skipped",
						i, sr.Step.Title, sr.Status)
				}
			}

			if report.Summary.Skipped != report.Summary.Total {
				t.Errorf("Skipped = %d, Total = %d — all should be skipped",
					report.Summary.Skipped, report.Summary.Total)
			}

			if report.Summary.Failed != 0 {
				t.Errorf("Failed = %d, want 0 in dry-run", report.Summary.Failed)
			}

			t.Logf("  %s: %d steps, all skipped", name, report.Summary.Total)
		})
	}
}

func TestRun_StepSetup(t *testing.T) {
	md := makeRunbook("### Step 1: First\n\n" + "```bash" + "\necho step1\n" + "```" + "\n\n### Step 2: Second\n\n" + "```bash" + "\necho step2\n" + "```" + "\n")
	report, err := Run(strings.NewReader(md), "test", RunOptions{
		StepSetup: "echo setup",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", report.Summary.Passed)
	}
	for i, s := range report.Steps {
		if s.StepSetup == nil {
			t.Errorf("step %d: expected StepSetup result", i+1)
		}
	}
}

func TestRun_StepSetupFailSkips(t *testing.T) {
	md := makeRunbook("### Step 1: Should fail\n\n" + "```bash" + "\necho should-not-run\n" + "```" + "\n")
	report, err := Run(strings.NewReader(md), "test", RunOptions{
		StepSetup: "exit 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Summary.Failed)
	}
	if !strings.Contains(report.Steps[0].Error, "step-setup") {
		t.Errorf("expected step-setup error, got %q", report.Steps[0].Error)
	}
}
