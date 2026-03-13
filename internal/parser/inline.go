package parser

import (
	"bufio"
	"io"
	"path/filepath"
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

const (
	markerStart = "<!-- mdproof:start -->"
	markerEnd   = "<!-- mdproof:end -->"
)

// inlineBlock captures a single <!-- mdproof:start/end --> block with its
// surrounding context so parseInlineBlock can receive it as one value.
type inlineBlock struct {
	content       string
	startLine     int
	filename      string
	headingTitle  string
	headingSource core.SourceRange
}

// ParseInline scans an arbitrary Markdown file for <!-- mdproof:start/end --> blocks.
// Each block is parsed as a runbook step. Steps are auto-numbered 1, 2, 3...
func ParseInline(r io.Reader, filename string) (*core.Runbook, error) {
	scanner := bufio.NewScanner(r)
	rb := &core.Runbook{}

	var blocks []inlineBlock
	var currentBlock strings.Builder
	inBlock := false
	lineNum := 0
	startLine := 0
	currentHeading := filepath.Base(filename)
	var currentHeadingSource core.SourceRange

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		trimmed := strings.TrimSpace(line)

		// Extract title from # heading (top-level only)
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") && rb.Meta.Title == "" {
			rb.Meta.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			rb.Meta.Title = stripInlineMarkdown(rb.Meta.Title)
		}
		if strings.HasPrefix(trimmed, "#") {
			if title := parseMarkdownHeading(trimmed); title != "" {
				currentHeading = title
				currentHeadingSource = singleLineRange(lineNum)
			}
		}

		if trimmed == markerStart {
			if inBlock {
				return nil, sourceError(filename, lineNum, "nested <!-- mdproof:start --> markers not allowed")
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
			blocks = append(blocks, inlineBlock{
				content:       currentBlock.String(),
				startLine:     startLine + 1,
				filename:      filename,
				headingTitle:  currentHeading,
				headingSource: currentHeadingSource,
			})
			continue
		}

		if inBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, sourceError(filename, lineNum, "scanner error: %v", err)
	}

	if inBlock {
		return nil, sourceError(filename, startLine, "unclosed <!-- mdproof:start --> marker")
	}

	for i, block := range blocks {
		step, err := parseInlineBlock(block, i+1)
		if err != nil {
			return nil, err
		}
		rb.Steps = append(rb.Steps, step)
	}

	return rb, nil
}

func parseInlineBlock(block inlineBlock, stepNum int) (core.Step, error) {
	title := block.headingTitle
	if title == "" {
		title = filepath.Base(block.filename)
	}
	step := core.Step{
		Number:        stepNum,
		Title:         title,
		File:          block.filename,
		HeadingSource: block.headingSource,
	}

	lines := strings.Split(block.content, "\n")
	inCode := false
	inExpected := false
	var codeLines []string
	var codeLang string
	codeStartLine := 0

	for idx, line := range lines {
		lineNum := block.startLine + idx
		if inCode {
			if codeFenceCloseRe.MatchString(line) {
				inCode = false
				step.Command = strings.TrimRight(strings.Join(codeLines, "\n"), "\n")
				step.Lang = codeLang
				step.CodeSources = append(step.CodeSources, core.SourceRange{
					Start: core.SourcePos{Line: codeStartLine},
					End:   core.SourcePos{Line: lineNum},
				})
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
			codeStartLine = lineNum
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
				step.Expected = append(step.Expected, core.Expectation{
					Text:   item,
					Source: singleLineRange(lineNum),
				})
				continue
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			inExpected = false
		}
	}

	if inCode {
		return core.Step{}, sourceError(block.filename, codeStartLine, "unclosed code fence")
	}

	return step, nil
}

func parseMarkdownHeading(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return ""
	}
	trimmed = strings.TrimLeft(trimmed, "#")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return ""
	}
	return stripInlineMarkdown(trimmed)
}
