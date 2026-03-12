package executor

import (
	"context"
	"os"
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

func TestIsContainerEnv_EnvOverride(t *testing.T) {
	// Already set by TestMain, so should return true.
	if !IsContainerEnv() {
		t.Fatal("expected IsContainerEnv=true with MDPROOF_ALLOW_EXECUTE set")
	}
}

func TestIsContainerEnv_NoOverride(t *testing.T) {
	old := os.Getenv("MDPROOF_ALLOW_EXECUTE")
	os.Unsetenv("MDPROOF_ALLOW_EXECUTE")
	defer os.Setenv("MDPROOF_ALLOW_EXECUTE", old)

	// On host (no /.dockerenv), should return false.
	// Inside a real container, this test still passes (/.dockerenv exists).
	got := IsContainerEnv()
	_, hasDocker := os.Stat("/.dockerenv")
	_, hasPodman := os.Stat("/run/.containerenv")
	inContainer := hasDocker == nil || hasPodman == nil

	if got != inContainer {
		t.Fatalf("IsContainerEnv=%v, expected %v (inContainer=%v)", got, inContainer, inContainer)
	}
}

func TestExecuteSession_SimpleEcho(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "echo", Command: "echo hello", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("expected passed, got %s (err=%s stderr=%q)", results[0].Status, results[0].Error, results[0].Stderr)
	}
	if got := strings.TrimSpace(results[0].Stdout); got != "hello" {
		t.Fatalf("expected stdout 'hello', got %q", got)
	}
}

func TestExecuteSession_VariablePersistence(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "set var", Command: "MY_VAR=fromstep1\necho \"set MY_VAR=$MY_VAR\"", Executor: core.ExecutorAuto},
		{Number: 2, Title: "read var", Command: "echo \"got MY_VAR=$MY_VAR\"", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("step 1: expected passed, got %s (err=%s stderr=%q)", results[0].Status, results[0].Error, results[0].Stderr)
	}
	if results[1].Status != core.StatusPassed {
		t.Fatalf("step 2: expected passed, got %s (err=%s stderr=%q)", results[1].Status, results[1].Error, results[1].Stderr)
	}
	if !strings.Contains(results[1].Stdout, "got MY_VAR=fromstep1") {
		t.Fatalf("step 2: expected variable from step 1, got stdout=%q", results[1].Stdout)
	}
}

func TestExecuteSession_StepFailureContinues(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "fail", Command: "echo before_fail && exit 1", Executor: core.ExecutorAuto},
		{Number: 2, Title: "still runs", Command: "echo after_fail", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusFailed {
		t.Fatalf("step 1: expected failed, got %s", results[0].Status)
	}
	if results[0].ExitCode != 1 {
		t.Fatalf("step 1: expected exit code 1, got %d", results[0].ExitCode)
	}
	if results[1].Status != core.StatusPassed {
		t.Fatalf("step 2: expected passed, got %s (err=%s stderr=%q)", results[1].Status, results[1].Error, results[1].Stderr)
	}
	if !strings.Contains(results[1].Stdout, "after_fail") {
		t.Fatalf("step 2: expected 'after_fail', got %q", results[1].Stdout)
	}
}

func TestExecuteSession_SkipsManual(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "auto", Command: "echo ok", Executor: core.ExecutorAuto},
		{Number: 2, Title: "manual", Command: "echo skip", Executor: core.ExecutorManual},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("step 1: expected passed, got %s", results[0].Status)
	}
	if results[1].Status != core.StatusSkipped {
		t.Fatalf("step 2: expected skipped, got %s", results[1].Status)
	}
}

func TestExecuteSession_CapturesStderr(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "stderr", Command: "echo out && echo err >&2", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("expected passed, got %s (err=%s)", results[0].Status, results[0].Error)
	}
	if !strings.Contains(results[0].Stdout, "out") {
		t.Errorf("expected stdout to contain 'out', got %q", results[0].Stdout)
	}
	if !strings.Contains(results[0].Stderr, "err") {
		t.Errorf("expected stderr to contain 'err', got %q", results[0].Stderr)
	}
}

func TestExecuteSession_VariableSurvivedFailedStep(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "set and fail", Command: "SURV=yes\necho set_surv\nfalse", Executor: core.ExecutorAuto},
		{Number: 2, Title: "check surv", Command: "echo \"SURV=$SURV\"", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusFailed {
		t.Fatalf("step 1: expected failed, got %s", results[0].Status)
	}
	// EXIT trap should have saved SURV even though step failed.
	if results[1].Status != core.StatusPassed {
		t.Fatalf("step 2: expected passed, got %s (err=%s stderr=%q)", results[1].Status, results[1].Error, results[1].Stderr)
	}
	if !strings.Contains(results[1].Stdout, "SURV=yes") {
		t.Fatalf("step 2: expected SURV=yes, got %q", results[1].Stdout)
	}
}

func TestExecuteSession_Assertions(t *testing.T) {
	steps := []core.Step{
		{
			Number:   1,
			Title:    "with expected",
			Command:  "echo apple banana",
			Expected: []string{"apple", "cherry"},
			Executor: core.ExecutorAuto,
		},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	// Command succeeds but assertion for "cherry" fails.
	if results[0].Status != core.StatusFailed {
		t.Fatalf("expected failed due to assertion, got %s", results[0].Status)
	}
	if len(results[0].Assertions) != 2 {
		t.Fatalf("expected 2 assertions, got %d", len(results[0].Assertions))
	}
	if !results[0].Assertions[0].Matched {
		t.Error("'apple' assertion should have matched")
	}
	if results[0].Assertions[1].Matched {
		t.Error("'cherry' assertion should NOT have matched")
	}
}

func TestExecuteSession_EmptySteps(t *testing.T) {
	results := ExecuteSession(context.Background(), nil, SessionOptions{Timeout: 30 * time.Second})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestExecuteSession_MergedCodeBlocks(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "merged", Command: "echo first\n---\necho second", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("expected passed, got %s (err=%s stderr=%q)", results[0].Status, results[0].Error, results[0].Stderr)
	}
	if !strings.Contains(results[0].Stdout, "first") || !strings.Contains(results[0].Stdout, "second") {
		t.Fatalf("expected both 'first' and 'second', got %q", results[0].Stdout)
	}
}

func TestExecuteSession_FailFast(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "fail", Command: "echo step1 && exit 1", Executor: core.ExecutorAuto},
		{Number: 2, Title: "skip", Command: "echo step2", Executor: core.ExecutorAuto},
		{Number: 3, Title: "skip", Command: "echo step3", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second, FailFast: true})

	if results[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed, got %s", results[0].Status)
	}
	if results[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped, got %s", results[1].Status)
	}
	if results[2].Status != core.StatusSkipped {
		t.Errorf("step 3: expected skipped, got %s", results[2].Status)
	}
}

func TestExecuteSession_EnvSeeding(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "check env", Command: "echo MY_VAR=$MY_VAR", Executor: core.ExecutorAuto},
	}
	env := map[string]string{"MY_VAR": "seeded"}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second, EnvVars: env})

	if results[0].Status != core.StatusPassed {
		t.Fatalf("expected passed, got %s (err=%s)", results[0].Status, results[0].Error)
	}
	if !strings.Contains(results[0].Stdout, "MY_VAR=seeded") {
		t.Errorf("expected MY_VAR=seeded, got %q", results[0].Stdout)
	}
}

func TestExecuteSession_Retry(t *testing.T) {
	// Uses a temp file as counter: first attempt creates file + fails,
	// second attempt sees file + passes.
	steps := []core.Step{
		{
			Number:   1,
			Title:    "retry step",
			Command:  "F=/tmp/rb_test_retry_$$\nif [ -f \"$F\" ]; then\n  echo passed\n  rm -f \"$F\"\nelse\n  touch \"$F\"\n  echo failed && exit 1\nfi",
			Retry:    1,
			Executor: core.ExecutorAuto,
		},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Errorf("expected passed after retry, got %s (stdout=%q stderr=%q)",
			results[0].Status, results[0].Stdout, results[0].Stderr)
	}
}

func TestExecuteSession_DependsOn(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "fail", Command: "exit 1", Executor: core.ExecutorAuto},
		{Number: 2, Title: "depends", Command: "echo should_not_run", DependsOn: 1, Executor: core.ExecutorAuto},
		{Number: 3, Title: "independent", Command: "echo ok", Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusFailed {
		t.Errorf("step 1: expected failed, got %s", results[0].Status)
	}
	if results[1].Status != core.StatusSkipped {
		t.Errorf("step 2: expected skipped (depends), got %s", results[1].Status)
	}
	if !strings.Contains(results[1].Error, "depends") {
		t.Errorf("step 2 error should mention depends, got: %q", results[1].Error)
	}
	if results[2].Status != core.StatusPassed {
		t.Errorf("step 3: expected passed, got %s (err=%s)", results[2].Status, results[2].Error)
	}
}

func TestExecuteSession_DependsOnPasses(t *testing.T) {
	steps := []core.Step{
		{Number: 1, Title: "pass", Command: "echo ok", Executor: core.ExecutorAuto},
		{Number: 2, Title: "depends", Command: "echo depends_ok", DependsOn: 1, Executor: core.ExecutorAuto},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusPassed {
		t.Errorf("step 1: expected passed, got %s", results[0].Status)
	}
	if results[1].Status != core.StatusPassed {
		t.Errorf("step 2: expected passed (depends satisfied), got %s (err=%s)", results[1].Status, results[1].Error)
	}
	if !strings.Contains(results[1].Stdout, "depends_ok") {
		t.Errorf("step 2: expected 'depends_ok', got %q", results[1].Stdout)
	}
}

func TestExecuteSession_PerStepTimeout(t *testing.T) {
	steps := []core.Step{
		{
			Number:   1,
			Title:    "slow step",
			Command:  "sleep 10 && echo done",
			Timeout:  1 * time.Second,
			Executor: core.ExecutorAuto,
		},
	}
	results := ExecuteSession(context.Background(), steps, SessionOptions{Timeout: 30 * time.Second})

	if results[0].Status != core.StatusFailed {
		t.Errorf("expected failed (timeout), got %s", results[0].Status)
	}
	if results[0].ExitCode == 0 {
		t.Errorf("expected non-zero exit code from timeout")
	}
}
