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
func RunAssertions(result *core.StepResult, expected []string) []core.AssertionResult {
	if len(expected) == 0 {
		return nil
	}

	combined := result.Stdout + "\n" + result.Stderr
	results := make([]core.AssertionResult, 0, len(expected))

	for _, pat := range expected {
		r := dispatchAssertion(pat, combined, result.Stdout, result.ExitCode)
		results = append(results, r)
	}

	return results
}

// dispatchAssertion detects the assertion type and runs the appropriate check.
func dispatchAssertion(pat, combined, stdout string, exitCode int) core.AssertionResult {
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
	return checkSubstring(pat, combined)
}

// checkSubstring performs case-insensitive substring matching with negation.
func checkSubstring(pat, output string) core.AssertionResult {
	r := core.AssertionResult{Pattern: pat, Type: core.AssertSubstring}
	lower := strings.ToLower(output)

	inner := pat
	for _, prefix := range negationPrefixes {
		if strings.HasPrefix(pat, prefix) {
			r.Negated = true
			inner = pat[len(prefix):]
			break
		}
	}

	found := strings.Contains(lower, strings.ToLower(inner))
	if r.Negated {
		r.Matched = !found
	} else {
		r.Matched = found
	}

	return r
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

	cmd := exec.Command("jq", "-e", expr)
	cmd.Stdin = bytes.NewBufferString(output)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
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
		r := core.AssertionResult{Pattern: pat, Type: core.AssertSubstring}

		inner := pat
		for _, prefix := range negationPrefixes {
			if strings.HasPrefix(pat, prefix) {
				r.Negated = true
				inner = pat[len(prefix):]
				break
			}
		}

		found := strings.Contains(lower, strings.ToLower(inner))
		if r.Negated {
			r.Matched = !found
		} else {
			r.Matched = found
		}

		results = append(results, r)
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
