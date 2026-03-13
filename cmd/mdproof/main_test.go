package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/runkids/mdproof"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("MDPROOF_ALLOW_EXECUTE", "1")
	os.Exit(m.Run())
}

func TestRunAllAndReport_PerRunbookIsolationUsesFreshHomeAndTmpdir(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		writeEnvRunbook(t, dir, "first-proof.md"),
		writeEnvRunbook(t, dir, "second-proof.md"),
	}

	var (
		reports []mdproof.Report
		errs    int
	)
	captureStdout(t, func() {
		reports, errs = runAllAndReport(files, false, 5*time.Second, mdproof.Config{
			Isolation: mdproof.IsolationPerRunbook,
			Env: map[string]string{
				"BASE": "present",
			},
		}, mdproof.RunOptions{}, "", 0, false)
	})

	if errs != 0 {
		t.Fatalf("errs = %d, want 0", errs)
	}
	if len(reports) != 2 {
		t.Fatalf("len(reports) = %d, want 2", len(reports))
	}

	firstHome := envValue(t, reports[0].Steps[0].Stdout, "HOME")
	firstTmp := envValue(t, reports[0].Steps[0].Stdout, "TMPDIR")
	firstBase := envValue(t, reports[0].Steps[0].Stdout, "BASE")
	secondHome := envValue(t, reports[1].Steps[0].Stdout, "HOME")
	secondTmp := envValue(t, reports[1].Steps[0].Stdout, "TMPDIR")
	secondBase := envValue(t, reports[1].Steps[0].Stdout, "BASE")

	if firstHome == secondHome {
		t.Fatalf("HOME reused across runbooks: %q", firstHome)
	}
	if firstTmp == secondTmp {
		t.Fatalf("TMPDIR reused across runbooks: %q", firstTmp)
	}
	if got, want := firstBase, "present"; got != want {
		t.Fatalf("first BASE = %q, want %q", got, want)
	}
	if got, want := secondBase, "present"; got != want {
		t.Fatalf("second BASE = %q, want %q", got, want)
	}
	if got, want := firstTmp, filepath.Join(firstHome, "tmp"); got != want {
		t.Fatalf("first TMPDIR = %q, want %q", got, want)
	}
	if got, want := secondTmp, filepath.Join(secondHome, "tmp"); got != want {
		t.Fatalf("second TMPDIR = %q, want %q", got, want)
	}
}

func TestRunAllAndReport_PerRunbookIsolationCleansUpTempHomes(t *testing.T) {
	dir := t.TempDir()
	file := writeEnvRunbook(t, dir, "cleanup-proof.md")

	var (
		reports []mdproof.Report
		errs    int
	)
	captureStdout(t, func() {
		reports, errs = runAllAndReport([]string{file}, false, 5*time.Second, mdproof.Config{
			Isolation: mdproof.IsolationPerRunbook,
			Env: map[string]string{
				"BASE": "present",
			},
		}, mdproof.RunOptions{}, "", 0, false)
	})

	if errs != 0 {
		t.Fatalf("errs = %d, want 0", errs)
	}
	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d, want 1", len(reports))
	}

	home := envValue(t, reports[0].Steps[0].Stdout, "HOME")
	tmp := envValue(t, reports[0].Steps[0].Stdout, "TMPDIR")

	if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("HOME dir still exists or stat failed unexpectedly: path=%q err=%v", home, err)
	}
	if _, err := os.Stat(tmp); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("TMPDIR still exists or stat failed unexpectedly: path=%q err=%v", tmp, err)
	}
}

func writeEnvRunbook(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	content := `# Env Isolation

## Steps

### Step 1: Print environment

` + "```bash\n" + `printf 'HOME=%s\nTMPDIR=%s\nBASE=%s\n' "$HOME" "$TMPDIR" "$BASE"
` + "```" + `

Expected:

- regex: HOME=/
- regex: TMPDIR=/
- BASE=present
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write runbook: %v", err)
	}
	return path
}

func envValue(t *testing.T, stdout, key string) string {
	t.Helper()

	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `=(.+)$`)
	match := re.FindStringSubmatch(stdout)
	if len(match) != 2 {
		t.Fatalf("missing %s in stdout %q", key, stdout)
	}
	return match[1]
}

func captureStdout(t *testing.T, fn func()) {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = orig
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	_, _ = io.Copy(io.Discard, r)
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
}
