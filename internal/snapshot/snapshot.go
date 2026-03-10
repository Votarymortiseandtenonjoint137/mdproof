package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/runkids/mdproof/internal/core"
)

// Store manages snapshot files for runbook assertion comparisons.
type Store struct {
	baseDir    string
	updateMode bool
	cache      map[string]map[string]string
}

// NewStore creates a new snapshot Store rooted at baseDir.
// If updateMode is true, mismatched snapshots are overwritten instead of failing.
func NewStore(baseDir string, updateMode bool) *Store {
	return &Store{
		baseDir:    baseDir,
		updateMode: updateMode,
		cache:      make(map[string]map[string]string),
	}
}

// Check compares actual output against the stored snapshot for the given name.
// On first run (no existing snapshot), it creates the snapshot and returns matched.
// In update mode, mismatches overwrite the stored snapshot.
func (s *Store) Check(name, actual, runbookName string) core.AssertionResult {
	result := core.AssertionResult{
		Pattern: "snapshot: " + name,
		Type:    core.AssertSnapshot,
	}

	snapshots := s.loadRunbook(runbookName)
	stored, exists := snapshots[name]

	if !exists || s.updateMode {
		snapshots[name] = actual
		s.cache[runbookName] = snapshots
		if err := s.saveRunbook(runbookName, snapshots); err != nil {
			result.Detail = fmt.Sprintf("failed to write snapshot: %v", err)
			return result
		}
		result.Matched = true
		if !exists {
			result.Detail = "snapshot created"
		} else if s.updateMode && stored != actual {
			result.Detail = "snapshot updated"
		}
		return result
	}

	if actual == stored {
		result.Matched = true
		return result
	}

	result.Matched = false
	result.Detail = buildDiff(stored, actual)
	return result
}

func (s *Store) snapshotDir() string {
	return filepath.Join(s.baseDir, "__snapshots__")
}

func (s *Store) snapFilePath(runbookName string) string {
	base := strings.TrimSuffix(runbookName, ".md")
	return filepath.Join(s.snapshotDir(), base+".snap")
}

func (s *Store) loadRunbook(runbookName string) map[string]string {
	if cached, ok := s.cache[runbookName]; ok {
		return cached
	}

	snapshots := make(map[string]string)
	data, err := os.ReadFile(s.snapFilePath(runbookName))
	if err != nil {
		s.cache[runbookName] = snapshots
		return snapshots
	}

	snapshots = parseSnapFile(string(data))
	s.cache[runbookName] = snapshots
	return snapshots
}

func (s *Store) saveRunbook(runbookName string, snapshots map[string]string) error {
	dir := s.snapshotDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := formatSnapFile(snapshots)
	return os.WriteFile(s.snapFilePath(runbookName), []byte(content), 0644)
}

func parseSnapFile(data string) map[string]string {
	result := make(map[string]string)
	sections := strings.Split(data, "\n---\n")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		name, content := parseSection(section)
		if name != "" {
			result[name] = content
		}
	}

	return result
}

func parseSection(section string) (string, string) {
	lines := strings.SplitN(section, "\n", -1)
	name := ""
	contentStart := 0

	for i, line := range lines {
		if strings.HasPrefix(line, "// Snapshot: ") {
			name = strings.TrimPrefix(line, "// Snapshot: ")
		}
		if strings.HasPrefix(line, "// Created: ") {
			contentStart = i + 1
			break
		}
	}

	if name == "" || contentStart >= len(lines) {
		return "", ""
	}

	if contentStart < len(lines) && lines[contentStart] == "" {
		contentStart++
	}

	content := strings.Join(lines[contentStart:], "\n")
	return name, content
}

func formatSnapFile(snapshots map[string]string) string {
	var sb strings.Builder
	first := true

	keys := sortedKeys(snapshots)

	for _, name := range keys {
		content := snapshots[name]
		if !first {
			sb.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&sb, "// Snapshot: %s\n", name)
		fmt.Fprintf(&sb, "// Created: %s\n", time.Now().UTC().Format(time.RFC3339))
		sb.WriteString("\n")
		sb.WriteString(content)
		sb.WriteString("\n")
		first = false
	}

	return sb.String()
}

func buildDiff(expected, actual string) string {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")

	var diff strings.Builder
	diff.WriteString("snapshot mismatch:\n")

	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	for i := 0; i < maxLines; i++ {
		var exp, act string
		if i < len(expLines) {
			exp = expLines[i]
		}
		if i < len(actLines) {
			act = actLines[i]
		}
		if exp != act {
			if i < len(expLines) {
				fmt.Fprintf(&diff, "  - %s\n", exp)
			}
			if i < len(actLines) {
				fmt.Fprintf(&diff, "  + %s\n", act)
			}
		}
	}

	return strings.TrimRight(diff.String(), "\n")
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
