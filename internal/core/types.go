package core

import (
	"fmt"
	"sort"
	"time"
)

// Status constants
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Executor mode constants
const (
	ExecutorAuto   = "auto"
	ExecutorManual = "manual"
)

// Assertion type constants.
const (
	AssertSubstring = "substring"
	AssertExitCode  = "exit_code"
	AssertRegex     = "regex"
	AssertJQ        = "jq"
	AssertSnapshot  = "snapshot"
)

// Default timeouts.
const (
	DefaultStepTimeout    = 2 * time.Minute
	DefaultSessionTimeout = 10 * time.Minute
)

// Sentinel exit codes used by the session executor.
const (
	ExitCodeFailFastSkipped = -1 // step skipped due to --fail-fast
	ExitCodeDependsSkipped  = -2 // step skipped due to depends directive
)

// Step represents a single test step in a runbook.
type Step struct {
	Number      int           `json:"number"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Command     string        `json:"command,omitempty"`
	Lang        string        `json:"lang,omitempty"`
	Expected    []string      `json:"expected,omitempty"`
	Executor    string        `json:"executor,omitempty"` // "auto", "ai-delegate", "manual"
	Timeout     time.Duration `json:"timeout,omitempty"`  // per-step timeout override (0 = use global)
	Retry       int           `json:"retry,omitempty"`    // retry count on failure (0 = no retry)
	RetryDelay  time.Duration `json:"retry_delay,omitempty"`
	DependsOn   int           `json:"depends_on,omitempty"` // skip if this step number failed (0 = none)
}

// HookExecResult holds the outcome of a step-setup or step-teardown execution.
type HookExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

// SubCommandResult holds the outcome of a single sub-command within a step.
type SubCommandResult struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
}

// StepResult represents the execution result of a single step.
type StepResult struct {
	Step       Step              `json:"step"`
	Status     string            `json:"status"` // "passed", "failed", "skipped", "running"
	DurationMs int64             `json:"duration_ms"`
	Stdout     string            `json:"stdout,omitempty"`
	Stderr     string            `json:"stderr,omitempty"`
	ExitCode   int               `json:"exit_code"`
	Assertions []AssertionResult `json:"assertions,omitempty"`
	Error        string            `json:"error,omitempty"`
	StepSetup    *HookExecResult   `json:"step_setup,omitempty"`
	StepTeardown *HookExecResult   `json:"step_teardown,omitempty"`
	SubCommands  []SubCommandResult `json:"sub_commands,omitempty"`
}

// AssertionResult represents the result of a single assertion check.
type AssertionResult struct {
	Pattern string `json:"pattern"`
	Type    string `json:"type,omitempty"` // "substring", "exit_code", "regex", "jq"
	Matched bool   `json:"matched"`
	Negated bool   `json:"negated,omitempty"`
	Detail  string `json:"detail,omitempty"` // extra info on failure (e.g., "got exit_code=1")
}

// Report represents the full execution report for a runbook.
type Report struct {
	Version     string            `json:"version"`
	Runbook     string            `json:"runbook"`
	Environment map[string]string `json:"environment,omitempty"`
	Hooks       map[string]string `json:"hooks,omitempty"` // setup/teardown status
	DurationMs  int64             `json:"duration_ms"`
	Summary     Summary           `json:"summary"`
	Steps       []StepResult      `json:"steps"`
}

// Summary represents execution summary counts.
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// Meta represents metadata extracted from runbook headings.
type Meta struct {
	Title string
	Scope string
	Env   string
}

// Runbook is the fully parsed representation of an E2E runbook Markdown file.
type Runbook struct {
	Meta  Meta
	Steps []Step
}

// SortedKeys returns map keys in sorted order for deterministic output.
func SortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TruncateText shortens s to max characters, adding ellipsis if needed.
func TruncateText(s string, max int) string {
	if len(s) <= max {
		return s // fast path: byte len fits means rune len also fits
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return "\u2026"
	}
	return string(runes[:max-1]) + "\u2026"
}

// FormatDurationMs formats milliseconds into a human-readable string.
func FormatDurationMs(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}

// StepFailReason extracts a concise failure reason from a StepResult.
func StepFailReason(r StepResult) string {
	for _, a := range r.Assertions {
		if !a.Matched {
			if a.Negated {
				return fmt.Sprintf("unexpected match: %s", a.Pattern)
			}
			return fmt.Sprintf("expected: %s", a.Pattern)
		}
	}
	if r.Error != "" {
		return r.Error
	}
	if r.ExitCode != 0 {
		return fmt.Sprintf("exit code %d", r.ExitCode)
	}
	return ""
}
