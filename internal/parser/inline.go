package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

const (
	markerStart = "<!-- mdproof:start -->"
	markerEnd   = "<!-- mdproof:end -->"
)

// ParseInline scans an arbitrary Markdown file for <!-- mdproof:start/end --> blocks.
// Each block is parsed as a runbook step. Steps are auto-numbered 1, 2, 3...
func ParseInline(r io.Reader, filename string) (*core.Runbook, error) {
	scanner := bufio.NewScanner(r)
	rb := &core.Runbook{}

	var blocks []string
	var currentBlock strings.Builder
	inBlock := false
	lineNum := 0
	startLine := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		trimmed := strings.TrimSpace(line)

		// Extract title from # heading (top-level only)
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") && rb.Meta.Title == "" {
			rb.Meta.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			rb.Meta.Title = stripInlineMarkdown(rb.Meta.Title)
		}

		if trimmed == markerStart {
			if inBlock {
				return nil, fmt.Errorf("line %d: nested <!-- mdproof:start --> markers not allowed", lineNum)
			}
			inBlock = true
			startLine = lineNum
			currentBlock.Reset()
			continue
		}

		if trimmed == markerEnd {
			if !inBlock {
				continue
			}
			inBlock = false
			blocks = append(blocks, currentBlock.String())
			continue
		}

		if inBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	if inBlock {
		return nil, fmt.Errorf("line %d: unclosed <!-- mdproof:start --> marker", startLine)
	}

	for i, block := range blocks {
		step, err := parseInlineBlock(block, i+1)
		if err != nil {
			return nil, fmt.Errorf("inline block %d: %w", i+1, err)
		}
		rb.Steps = append(rb.Steps, step)
	}

	return rb, nil
}

func parseInlineBlock(content string, stepNum int) (core.Step, error) {
	step := core.Step{
		Number: stepNum,
		Title:  fmt.Sprintf("inline %d", stepNum),
	}

	lines := strings.Split(content, "\n")
	inCode := false
	inExpected := false
	var codeLines []string
	var codeLang string

	for _, line := range lines {
		if inCode {
			if codeFenceCloseRe.MatchString(line) {
				inCode = false
				step.Command = strings.TrimRight(strings.Join(codeLines, "\n"), "\n")
				step.Lang = codeLang
			} else {
				codeLines = append(codeLines, line)
			}
			continue
		}

		if cm := codeFenceOpenRe.FindStringSubmatch(line); cm != nil {
			codeLang = cm[1]
			if codeLang == "" {
				codeLang = "bash"
			}
			codeLines = nil
			inCode = true
			continue
		}

		if isExpectedLabel(line) {
			inExpected = true
			continue
		}

		if inExpected {
			if bm := bulletRe.FindStringSubmatch(line); bm != nil {
				item := stripInlineMarkdown(strings.TrimSpace(bm[1]))
				step.Expected = append(step.Expected, item)
				continue
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			inExpected = false
		}
	}

	return step, nil
}
