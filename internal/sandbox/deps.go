package sandbox

import (
	"regexp"
	"sort"
	"strings"
)

// knownTools maps command names found in runbook code blocks to apt package names.
var knownTools = map[string]string{
	"jq":      "jq",
	"curl":    "curl",
	"python3": "python3",
	"python":  "python3",
	"node":    "nodejs",
	"pip":     "python3-pip",
	"git":     "git",
	"wget":    "wget",
}

// toolPatterns caches compiled regexes for each tool name.
var toolPatterns map[string]*regexp.Regexp

func init() {
	toolPatterns = make(map[string]*regexp.Regexp, len(knownTools))
	for name := range knownTools {
		toolPatterns[name] = regexp.MustCompile(`(?:^|[\s|;&(])` + regexp.QuoteMeta(name) + `(?:$|\s)`)
	}
}

// DetectDeps scans command strings for known tool names and returns
// the sorted, deduplicated list of apt packages to install.
// Always includes ca-certificates.
func DetectDeps(commands []string) []string {
	pkgs := map[string]bool{"ca-certificates": true}

	joined := strings.Join(commands, "\n")
	for name, pattern := range toolPatterns {
		if pattern.MatchString(joined) {
			pkgs[knownTools[name]] = true
		}
	}

	result := make([]string, 0, len(pkgs))
	for pkg := range pkgs {
		result = append(result, pkg)
	}
	sort.Strings(result)
	return result
}
