// Package mdproof provides a facade over the internal packages for parsing,
// executing, and reporting on Markdown runbooks.
package mdproof

import (
	"io"
	"time"

	"github.com/runkids/mdproof/internal/assertion"
	"github.com/runkids/mdproof/internal/config"
	"github.com/runkids/mdproof/internal/core"
	"github.com/runkids/mdproof/internal/coverage"
	"github.com/runkids/mdproof/internal/executor"
	"github.com/runkids/mdproof/internal/parser"
	"github.com/runkids/mdproof/internal/report"
	"github.com/runkids/mdproof/internal/runner"
)

// Type aliases — re-export all public types from internal packages.
type (
	Step            = core.Step
	StepResult      = core.StepResult
	AssertionResult = core.AssertionResult
	Report          = core.Report
	Summary         = core.Summary
	Meta            = core.Meta
	Runbook         = core.Runbook
	Config          = config.Config
	RunOptions      = runner.RunOptions
	HookResult      = runner.HookResult
)

// Status constants
const (
	StatusPassed  = core.StatusPassed
	StatusFailed  = core.StatusFailed
	StatusSkipped = core.StatusSkipped
)

// Executor mode constants
const (
	ExecutorAuto   = core.ExecutorAuto
	ExecutorManual = core.ExecutorManual
)

// Assertion type constants
const (
	AssertSubstring = core.AssertSubstring
	AssertExitCode  = core.AssertExitCode
	AssertRegex     = core.AssertRegex
	AssertJQ        = core.AssertJQ
	AssertSnapshot  = core.AssertSnapshot
)

// Default timeouts.
const (
	DefaultStepTimeout    = core.DefaultStepTimeout
	DefaultSessionTimeout = core.DefaultSessionTimeout
)

// ConfigFileName is the conventional name for directory-level runbook config.
const ConfigFileName = config.ConfigFileName

// ErrNotInContainer is returned when execution is attempted outside a container.
var ErrNotInContainer = executor.ErrNotInContainer

// --- Parser ---

// ParseRunbook reads a Markdown runbook and extracts metadata and steps.
func ParseRunbook(r io.Reader) (*Runbook, error) {
	return parser.ParseRunbook(r)
}

// ParseInline scans a Markdown file for inline test blocks.
func ParseInline(r io.Reader, filename string) (*Runbook, error) {
	return parser.ParseInline(r, filename)
}

// Classify determines execution mode for a step.
func Classify(s Step) string {
	return parser.Classify(s)
}

// ClassifyAll batch classifies all steps.
func ClassifyAll(steps []Step) []Step {
	return parser.ClassifyAll(steps)
}

// --- Assertion ---

// RunAssertions checks all expected patterns against a step result.
func RunAssertions(result *StepResult, expected []string) []AssertionResult {
	return assertion.RunAssertions(result, expected)
}

// MatchAssertions checks each expected pattern against the command output.
func MatchAssertions(output string, expected []string) []AssertionResult {
	return assertion.MatchAssertions(output, expected)
}

// AllPassed returns true when every assertion matched successfully.
func AllPassed(results []AssertionResult) bool {
	return assertion.AllPassed(results)
}

// --- Executor ---

// IsContainerEnv returns true if we're running inside a Docker container
// or if the MDPROOF_ALLOW_EXECUTE env var is set.
func IsContainerEnv() bool {
	return executor.IsContainerEnv()
}

// --- Config ---

// LoadConfig reads an mdproof.json from the given directory.
func LoadConfig(dir string) (Config, error) {
	return config.Load(dir)
}

// MergeConfig applies CLI flag overrides on top of file-based config.
func MergeConfig(file Config, cliBuild, cliSetup, cliTeardown string, cliTimeout time.Duration) Config {
	return config.Merge(file, cliBuild, cliSetup, cliTeardown, cliTimeout)
}

// --- Report ---

// WriteJSONReport writes the report as indented JSON.
func WriteJSONReport(w io.Writer, r Report) error {
	return report.WriteJSONReport(w, r)
}

// WriteJSONReports writes multiple reports as a JSON array.
func WriteJSONReports(w io.Writer, reports []Report) error {
	return report.WriteJSONReports(w, reports)
}

// WriteSingleReport prints a single runbook result in plain text.
func WriteSingleReport(w io.Writer, r Report, verbosity int) {
	report.WriteSingleReport(w, r, verbosity)
}

// WritePlainSummary prints a multi-runbook batch summary.
func WritePlainSummary(w io.Writer, reports []Report, verbosity int) {
	report.WritePlainSummary(w, reports, verbosity)
}

// --- Coverage ---

// CoverageEntry pairs a file name with its coverage result.
type CoverageEntry = report.CoverageEntry

// CoverageResult holds coverage analysis for a set of steps.
type CoverageResult = coverage.Result

// AnalyzeCoverage computes coverage metrics for a list of steps.
func AnalyzeCoverage(steps []Step) coverage.Result {
	return coverage.Analyze(steps)
}

// WriteCoverageReport writes a plain-text coverage table.
func WriteCoverageReport(w io.Writer, entries []CoverageEntry) {
	report.WriteCoverageReport(w, entries)
}

// CoverageTotalScore computes the aggregate score across all entries.
func CoverageTotalScore(entries []CoverageEntry) int {
	return report.TotalScore(entries)
}

// --- Runner ---

// Run parses, classifies, executes, and reports a runbook.
func Run(r io.Reader, name string, opts RunOptions) (Report, error) {
	return runner.Run(r, name, opts)
}

// RunBuildHook executes the build command as a simple shell process.
func RunBuildHook(command string) *HookResult {
	return runner.RunBuildHook(command)
}

// ResolveFiles finds runbook/proof files from a path (file or directory).
func ResolveFiles(target string) ([]string, error) {
	return runner.ResolveFiles(target)
}
