package parser

import (
	"strings"
	"testing"
)

func TestParseInline_SingleBlock(t *testing.T) {
	input := `# My README

Some documentation here.

<!-- mdproof:start -->
` + "```bash" + `
echo hello
` + "```" + `

Expected:

- hello
<!-- mdproof:end -->

More docs.
`

	rb, err := ParseInline(strings.NewReader(input), "README.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rb.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(rb.Steps))
	}
	if rb.Steps[0].Command != "echo hello" {
		t.Fatalf("unexpected command: %q", rb.Steps[0].Command)
	}
	if len(rb.Steps[0].Expected) != 1 || rb.Steps[0].Expected[0].Text != "hello" {
		t.Fatalf("unexpected expected: %v", rb.Steps[0].Expected)
	}
	if rb.Steps[0].Number != 1 {
		t.Fatalf("expected step number 1, got %d", rb.Steps[0].Number)
	}
}

func TestParseInline_MultipleBlocks(t *testing.T) {
	input := `# API Docs

<!-- mdproof:start -->
` + "```bash" + `
echo first
` + "```" + `

Expected:

- first
<!-- mdproof:end -->

Some text between.

<!-- mdproof:start -->
` + "```bash" + `
echo second
` + "```" + `

Expected:

- second
<!-- mdproof:end -->
`

	rb, err := ParseInline(strings.NewReader(input), "api-docs.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rb.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(rb.Steps))
	}
	if rb.Steps[0].Number != 1 || rb.Steps[1].Number != 2 {
		t.Fatal("steps should be auto-numbered 1, 2")
	}
	if rb.Meta.Title != "API Docs" {
		t.Fatalf("expected title from # heading, got %q", rb.Meta.Title)
	}
}

func TestParseInline_NoMarkers(t *testing.T) {
	input := "# README\n\nJust docs, no tests.\n"
	rb, err := ParseInline(strings.NewReader(input), "README.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rb.Steps) != 0 {
		t.Fatalf("expected 0 steps, got %d", len(rb.Steps))
	}
}

func TestParseInline_UnclosedMarker(t *testing.T) {
	input := `# README

<!-- mdproof:start -->
` + "```bash" + `
echo hello
` + "```" + `
`
	_, err := ParseInline(strings.NewReader(input), "README.md")
	if err == nil {
		t.Fatal("expected error for unclosed marker")
	}
}

func TestParseInline_NestedMarkers(t *testing.T) {
	input := `
<!-- mdproof:start -->
<!-- mdproof:start -->
` + "```bash" + `
echo hello
` + "```" + `
<!-- mdproof:end -->
<!-- mdproof:end -->
`
	_, err := ParseInline(strings.NewReader(input), "README.md")
	if err == nil {
		t.Fatal("expected error for nested markers")
	}
}

func TestParseInline_WithSnapshotAssertion(t *testing.T) {
	input := `
<!-- mdproof:start -->
` + "```bash" + `
curl -s http://localhost/api
` + "```" + `

Expected:

- snapshot: api-response
<!-- mdproof:end -->
`

	rb, err := ParseInline(strings.NewReader(input), "docs.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rb.Steps[0].Expected) != 1 || rb.Steps[0].Expected[0].Text != "snapshot: api-response" {
		t.Fatalf("snapshot assertion not preserved: %v", rb.Steps[0].Expected)
	}
}
