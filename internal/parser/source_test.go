package parser

import (
	"strings"
	"testing"
)

func TestParseRunbookFile_PreservesSourceRanges(t *testing.T) {
	input := `# Source Demo

## Steps

### Step 1: First

` + "```bash" + `
echo first
` + "```" + `

Expected:

- first

### Step 2: Second

` + "```bash" + `
echo second
` + "```" + `
` + "```bash" + `
echo third
` + "```" + `

Expected:

- second
- third
`

	rb, err := ParseRunbookFile(strings.NewReader(input), "docs/source.md")
	if err != nil {
		t.Fatalf("ParseRunbookFile error: %v", err)
	}
	if len(rb.Steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(rb.Steps))
	}

	step1 := rb.Steps[0]
	if step1.File != "docs/source.md" {
		t.Fatalf("step1.File = %q, want docs/source.md", step1.File)
	}
	if step1.HeadingSource.Start.Line != 5 || step1.HeadingSource.End.Line != 5 {
		t.Fatalf("step1 heading source = %+v, want line 5", step1.HeadingSource)
	}
	if len(step1.CodeSources) != 1 {
		t.Fatalf("len(step1.CodeSources) = %d, want 1", len(step1.CodeSources))
	}
	if step1.CodeSources[0].Start.Line != 7 || step1.CodeSources[0].End.Line != 9 {
		t.Fatalf("step1 code source = %+v, want 7-9", step1.CodeSources[0])
	}
	if len(step1.Expected) != 1 {
		t.Fatalf("len(step1.Expected) = %d, want 1", len(step1.Expected))
	}
	if step1.Expected[0].Text != "first" {
		t.Fatalf("step1.Expected[0].Text = %q, want first", step1.Expected[0].Text)
	}
	if step1.Expected[0].Source.Start.Line != 13 || step1.Expected[0].Source.End.Line != 13 {
		t.Fatalf("step1.Expected[0].Source = %+v, want line 13", step1.Expected[0].Source)
	}

	step2 := rb.Steps[1]
	if len(step2.CodeSources) != 2 {
		t.Fatalf("len(step2.CodeSources) = %d, want 2", len(step2.CodeSources))
	}
	if step2.CodeSources[0].Start.Line != 17 || step2.CodeSources[0].End.Line != 19 {
		t.Fatalf("step2 first code source = %+v, want 17-19", step2.CodeSources[0])
	}
	if step2.CodeSources[1].Start.Line != 20 || step2.CodeSources[1].End.Line != 22 {
		t.Fatalf("step2 second code source = %+v, want 20-22", step2.CodeSources[1])
	}
	if step2.Expected[1].Source.Start.Line != 27 {
		t.Fatalf("step2.Expected[1].Source = %+v, want line 27", step2.Expected[1].Source)
	}
}

func TestParseRunbookFile_UnclosedFenceIncludesFileAndLine(t *testing.T) {
	input := `# Broken

## Steps

### Step 1: Broken

` + "```bash" + `
echo nope
`

	_, err := ParseRunbookFile(strings.NewReader(input), "docs/broken.md")
	if err == nil {
		t.Fatal("expected error for unclosed fence")
	}
	if !strings.Contains(err.Error(), "docs/broken.md:") {
		t.Fatalf("error %q should include filename", err)
	}
	if !strings.Contains(err.Error(), "unclosed code fence") {
		t.Fatalf("error %q should mention unclosed code fence", err)
	}
}

func TestParseRunbookFile_DirectiveOutsideStepIncludesFileAndLine(t *testing.T) {
	input := `# Broken

<!-- runbook: retry=2 -->

## Steps
`

	_, err := ParseRunbookFile(strings.NewReader(input), "docs/directive.md")
	if err == nil {
		t.Fatal("expected error for misplaced directive")
	}
	if !strings.Contains(err.Error(), "docs/directive.md:3") {
		t.Fatalf("error %q should include docs/directive.md:3", err)
	}
}

func TestParseInline_PreservesNearestHeadingAndSourceRanges(t *testing.T) {
	input := `# README

## Install

<!-- mdproof:start -->
` + "```bash" + `
echo install
` + "```" + `

Expected:

- install
<!-- mdproof:end -->

## Verify

<!-- mdproof:start -->
` + "```bash" + `
echo verify
` + "```" + `

Expected:

- verify
<!-- mdproof:end -->
`

	rb, err := ParseInline(strings.NewReader(input), "README.md")
	if err != nil {
		t.Fatalf("ParseInline error: %v", err)
	}
	if len(rb.Steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(rb.Steps))
	}
	if rb.Steps[0].Title != "Install" {
		t.Fatalf("step1 title = %q, want Install", rb.Steps[0].Title)
	}
	if rb.Steps[0].HeadingSource.Start.Line != 3 {
		t.Fatalf("step1 heading source = %+v, want line 3", rb.Steps[0].HeadingSource)
	}
	if rb.Steps[0].CodeSources[0].Start.Line != 6 || rb.Steps[0].CodeSources[0].End.Line != 8 {
		t.Fatalf("step1 code source = %+v, want 6-8", rb.Steps[0].CodeSources[0])
	}
	if rb.Steps[0].Expected[0].Source.Start.Line != 12 {
		t.Fatalf("step1 expected source = %+v, want line 12", rb.Steps[0].Expected[0].Source)
	}
	if rb.Steps[1].Title != "Verify" {
		t.Fatalf("step2 title = %q, want Verify", rb.Steps[1].Title)
	}
	if rb.Steps[1].HeadingSource.Start.Line != 15 {
		t.Fatalf("step2 heading source = %+v, want line 15", rb.Steps[1].HeadingSource)
	}
}
