package assertion

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/runkids/mdproof/internal/core"
)

// negationPrefixes lists recognized negation prefixes, longest first
// to ensure greedy matching.
var negationPrefixes = []string{
	"Should NOT ",
	"should not ",
	"Must NOT ",
	"must not ",
	"Does not ",
	"does not ",
	"NOT ",
	"Not ",
	"not ",
	"No ",
	"no ",
}

// typedPrefixes maps assertion prefixes to their type.
// Order matters: longest prefix first.
var typedPrefixes = []struct {
	prefix string
	typ    string
}{
	{"exit_code: ", core.AssertExitCode},
	{"exit_code:", core.AssertExitCode},
	{"regex: ", core.AssertRegex},
	{"regex:", core.AssertRegex},
	{"jq: ", core.AssertJQ},
	{"jq:", core.AssertJQ},
}

// RunAssertions checks all expected patterns against a step result.
// It dispatches to the appropriate checker based on assertion type prefix.
// Substring/regex assertions match against stdout+stderr combined.
// JQ assertions match against stdout only (expects JSON).
func RunAssertions(result *core.StepResult, expected []core.Expectation) []core.AssertionResult {
	if len(expected) == 0 {
		return nil
	}

	combined := result.Stdout + "\n" + result.Stderr
	combinedLower := strings.ToLower(combined)
	results := make([]core.AssertionResult, 0, len(expected))

	for _, exp := range expected {
		r := dispatchAssertion(exp.Text, combined, combinedLower, result.Stdout, result.ExitCode)
		if !exp.Source.IsZero() {
			source := exp.Source
			r.Source = &source
		}
		results = append(results, r)
	}

	return results
}

// dispatchAssertion detects the assertion type and runs the appropriate check.
func dispatchAssertion(pat, combined, combinedLower, stdout string, exitCode int) core.AssertionResult {
	// Check for typed prefixes first.
	for _, tp := range typedPrefixes {
		if strings.HasPrefix(pat, tp.prefix) {
			arg := strings.TrimSpace(pat[len(tp.prefix):])
			switch tp.typ {
			case core.AssertExitCode:
				return checkExitCode(pat, arg, exitCode)
			case core.AssertRegex:
				return checkRegex(pat, arg, combined)
			case core.AssertJQ:
				return checkJQ(pat, arg, stdout)
			}
		}
	}

	// Default: substring match with negation support.
	return checkSubstring(pat, combinedLower)
}

// checkSubstring performs case-insensitive substring matching with negation.
// Negated assertions use word boundary matching (\b) to avoid false positives
// (e.g. "Not FAIL" should not trigger on "0 failed").
func checkSubstring(pat, combinedLower string) core.AssertionResult {
	r := core.AssertionResult{Pattern: pat, Type: core.AssertSubstring}

	inner := pat
	for _, prefix := range negationPrefixes {
		if strings.HasPrefix(pat, prefix) {
			r.Negated = true
			inner = pat[len(prefix):]
			break
		}
	}

	needle := strings.ToLower(inner)
	if r.Negated {
		// Use word boundary matching for negated assertions.
		found := negatedContains(combinedLower, needle)
		r.Matched = !found
		if !r.Matched {
			r.Detail = fmt.Sprintf("negated pattern %q was found in: %s",
				inner, findMatchLine(combinedLower, needle))
		}
	} else {
		r.Matched = strings.Contains(combinedLower, needle)
	}

	return r
}

// negatedContains checks if inner appears as a whole word (word boundary match)
// in text, case-insensitively. Falls back to substring if regex compilation fails.
func negatedContains(text, inner string) bool {
	pattern := `(?i)\b` + regexp.QuoteMeta(inner) + `\b`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return strings.Contains(strings.ToLower(text), strings.ToLower(inner))
	}
	return re.MatchString(text)
}

// findMatchLine returns the trimmed line containing the first occurrence of needle.
func findMatchLine(text, needle string) string {
	idx := strings.Index(text, needle)
	if idx < 0 {
		return ""
	}
	// Find line boundaries around the match.
	start := strings.LastIndex(text[:idx], "\n") + 1
	end := strings.Index(text[idx:], "\n")
	if end < 0 {
		end = len(text)
	} else {
		end += idx
	}
	line := strings.TrimSpace(text[start:end])
	if len(line) > 120 {
		line = line[:120] + "..."
	}
	return fmt.Sprintf("%q", line)
}

// checkExitCode verifies the step's exit code.
// Supports: "0", "1", "!0" (not zero), "!1" (not one).
func checkExitCode(pat, arg string, exitCode int) core.AssertionResult {
	r := core.AssertionResult{Pattern: pat, Type: core.AssertExitCode}

	if strings.HasPrefix(arg, "!") {
		// Negated: exit code must NOT equal N.
		r.Negated = true
		n, err := strconv.Atoi(strings.TrimSpace(arg[1:]))
		if err != nil {
			r.Detail = fmt.Sprintf("invalid exit_code value: %s", arg)
			return r
		}
		r.Matched = exitCode != n
		if !r.Matched {
			r.Detail = fmt.Sprintf("got exit_code=%d, expected !%d", exitCode, n)
		}
	} else {
		n, err := strconv.Atoi(strings.TrimSpace(arg))
		if err != nil {
			r.Detail = fmt.Sprintf("invalid exit_code value: %s", arg)
			return r
		}
		r.Matched = exitCode == n
		if !r.Matched {
			r.Detail = fmt.Sprintf("got exit_code=%d, expected %d", exitCode, n)
		}
	}

	return r
}

// checkRegex matches a Go regex pattern against output.
// Automatically prepends (?m) for multi-line mode so ^ and $ match line
// boundaries instead of string boundaries — unless the pattern already
// contains explicit flags.
func checkRegex(pat, pattern, output string) core.AssertionResult {
	r := core.AssertionResult{Pattern: pat, Type: core.AssertRegex}

	p := pattern
	if !strings.HasPrefix(p, "(?") {
		p = "(?m)" + p
	}

	re, err := regexp.Compile(p)
	if err != nil {
		r.Detail = fmt.Sprintf("invalid regex: %v", err)
		return r
	}

	r.Matched = re.MatchString(output)
	if !r.Matched {
		r.Detail = fmt.Sprintf("regex %q did not match", pattern)
	}

	return r
}

// checkJQ runs a jq expression against stdout. Passes if jq -e exits 0.
func checkJQ(pat, expr, output string) core.AssertionResult {
	r := core.AssertionResult{Pattern: pat, Type: core.AssertJQ}

	jqBin, err := exec.LookPath("jq")
	if err != nil {
		r.Detail = "jq is not installed (required for jq: assertions)\n" +
			"  brew install jq  OR  apt-get install jq\n" +
			"  or use: mdproof sandbox (auto-installs dependencies)"
		return r
	}

	cmd := exec.Command(jqBin, "-e", expr)
	cmd.Stdin = bytes.NewBufferString(output)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		r.Detail = fmt.Sprintf("jq failed: %s", strings.TrimSpace(stderr.String()))
		return r
	}

	r.Matched = true
	return r
}

// MatchAssertions checks each expected pattern against the command output.
// This is the legacy API — kept for backward compatibility with existing tests.
// New code should use RunAssertions.
func MatchAssertions(output string, expected []string) []core.AssertionResult {
	results := make([]core.AssertionResult, 0, len(expected))
	lower := strings.ToLower(output)
	for _, pat := range expected {
		results = append(results, checkSubstring(pat, lower))
	}
	return results
}

// AllPassed returns true when every assertion matched successfully.
func AllPassed(results []core.AssertionResult) bool {
	for _, r := range results {
		if !r.Matched {
			return false
		}
	}
	return true
}

// CheckStep runs assertion matching on a step result.
// If assertions are defined, they always run (regardless of exit code)
// and determine the final pass/fail status. If no assertions are defined,
// exit code alone determines the result (0=pass, non-zero=fail).
func CheckStep(result *core.StepResult, step core.Step) {
	if len(step.Expected) == 0 {
		return
	}

	result.Assertions = RunAssertions(result, step.Expected)
	if AllPassed(result.Assertions) {
		result.Status = core.StatusPassed
	} else {
		result.Status = core.StatusFailed
	}
}
