package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/runkids/mdproof/internal/core"
)

func TestWriteSingleReport_FailedAssertionShowsSourceLocation(t *testing.T) {
	report := core.Report{
		Runbook: "example.md",
		Summary: core.Summary{Total: 1, Failed: 1},
		Steps: []core.StepResult{
			{
				Step: core.Step{
					Number: 1,
					Title:  "Check hello",
					File:   "docs/example.md",
					HeadingSource: core.SourceRange{
						Start: core.SourcePos{Line: 5},
						End:   core.SourcePos{Line: 5},
					},
					CodeSources: []core.SourceRange{{
						Start: core.SourcePos{Line: 7},
						End:   core.SourcePos{Line: 9},
					}},
				},
				Source: &core.StepSource{
					Heading: &core.SourceRange{
						Start: core.SourcePos{Line: 5},
						End:   core.SourcePos{Line: 5},
					},
					CodeBlocks: []core.SourceRange{{
						Start: core.SourcePos{Line: 7},
						End:   core.SourcePos{Line: 9},
					}},
				},
				Status: core.StatusFailed,
				Assertions: []core.AssertionResult{{
					Pattern: "hello",
					Type:    core.AssertSubstring,
					Matched: false,
					Source: &core.SourceRange{
						Start: core.SourcePos{Line: 12},
						End:   core.SourcePos{Line: 12},
					},
				}},
			},
		},
	}

	var buf bytes.Buffer
	WriteSingleReport(&buf, report, 0)

	out := buf.String()
	if !strings.Contains(out, "FAIL docs/example.md:12 Step 1: Check hello") {
		t.Fatalf("plain report missing failure header with source:\n%s", out)
	}
	if !strings.Contains(out, "Assertion docs/example.md:12 hello") {
		t.Fatalf("plain report missing assertion source line:\n%s", out)
	}
}

func TestWriteSingleReport_CommandFailureShowsCodeBlockRange(t *testing.T) {
	report := core.Report{
		Runbook: "example.md",
		Summary: core.Summary{Total: 1, Failed: 1},
		Steps: []core.StepResult{
			{
				Step: core.Step{
					Number: 2,
					Title:  "Run command",
					File:   "docs/example.md",
					CodeSources: []core.SourceRange{{
						Start: core.SourcePos{Line: 20},
						End:   core.SourcePos{Line: 24},
					}},
				},
				Source: &core.StepSource{
					CodeBlocks: []core.SourceRange{{
						Start: core.SourcePos{Line: 20},
						End:   core.SourcePos{Line: 24},
					}},
				},
				Status:   core.StatusFailed,
				ExitCode: 2,
			},
		},
	}

	var buf bytes.Buffer
	WriteSingleReport(&buf, report, 0)

	out := buf.String()
	if !strings.Contains(out, "FAIL docs/example.md:20 Step 2: Run command") {
		t.Fatalf("plain report missing command failure header:\n%s", out)
	}
	if !strings.Contains(out, "Command docs/example.md:20-24") {
		t.Fatalf("plain report missing command source range:\n%s", out)
	}
}
