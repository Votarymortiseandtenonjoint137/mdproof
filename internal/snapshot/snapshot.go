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

type snapEntry struct {
	content   string
	createdAt string
}

// Store manages snapshot files for runbook assertion comparisons.
type Store struct {
	baseDir    string
	updateMode bool
	cache      map[string]map[string]snapEntry
}

// NewStore creates a new snapshot Store rooted at baseDir.
// If updateMode is true, mismatched snapshots are overwritten instead of failing.
func NewStore(baseDir string, updateMode bool) *Store {
	return &Store{
		baseDir:    baseDir,
		updateMode: updateMode,
		cache:      make(map[string]map[string]snapEntry),
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
	entry, exists := snapshots[name]

	if !exists || s.updateMode {
		now := time.Now().UTC().Format(time.RFC3339)
		createdAt := now
		if exists {
			createdAt = entry.createdAt // preserve original timestamp
		}
		snapshots[name] = snapEntry{content: actual, createdAt: createdAt}
		s.cache[runbookName] = snapshots
		if err := s.saveRunbook(runbookName, snapshots); err != nil {
			result.Detail = fmt.Sprintf("failed to write snapshot: %v", err)
			return result
		}
		result.Matched = true
		if !exists {
			result.Detail = "snapshot created"
		} else if s.updateMode && entry.content != actual {
			result.Detail = "snapshot updated"
		}
		return result
	}

	if actual == entry.content {
		result.Matched = true
		return result
	}

	result.Matched = false
	result.Detail = buildDiff(entry.content, actual)
	return result
}

func (s *Store) snapshotDir() string {
	return filepath.Join(s.baseDir, "__snapshots__")
}

func (s *Store) snapFilePath(runbookName string) string {
	base := strings.TrimSuffix(runbookName, ".md")
	return filepath.Join(s.snapshotDir(), base+".snap")
}

func (s *Store) loadRunbook(runbookName string) map[string]snapEntry {
	if cached, ok := s.cache[runbookName]; ok {
		return cached
	}

	snapshots := make(map[string]snapEntry)
	data, err := os.ReadFile(s.snapFilePath(runbookName))
	if err != nil {
		s.cache[runbookName] = snapshots
		return snapshots
	}

	snapshots = parseSnapFile(string(data))
	s.cache[runbookName] = snapshots
	return snapshots
}

func (s *Store) saveRunbook(runbookName string, snapshots map[string]snapEntry) error {
	dir := s.snapshotDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := formatSnapFile(snapshots)
	return os.WriteFile(s.snapFilePath(runbookName), []byte(content), 0644)
}

func parseSnapFile(data string) map[string]snapEntry {
	result := make(map[string]snapEntry)
	sections := strings.Split(data, "\n---\n")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		name, entry := parseSection(section)
		if name != "" {
			result[name] = entry
		}
	}

	return result
}

func parseSection(section string) (string, snapEntry) {
	lines := strings.Split(section, "\n")
	name := ""
	createdAt := ""
	contentStart := 0

	for i, line := range lines {
		if strings.HasPrefix(line, "// Snapshot: ") {
			name = strings.TrimPrefix(line, "// Snapshot: ")
		}
		if strings.HasPrefix(line, "// Created: ") {
			createdAt = strings.TrimPrefix(line, "// Created: ")
			contentStart = i + 1
			break
		}
	}

	if name == "" || contentStart == 0 || contentStart >= len(lines) {
		return "", snapEntry{}
	}

	if contentStart < len(lines) && lines[contentStart] == "" {
		contentStart++
	}

	content := strings.Join(lines[contentStart:], "\n")
	return name, snapEntry{content: content, createdAt: createdAt}
}

func formatSnapFile(snapshots map[string]snapEntry) string {
	var sb strings.Builder
	first := true

	keys := make([]string, 0, len(snapshots))
	for k := range snapshots {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		entry := snapshots[name]
		if !first {
			sb.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&sb, "// Snapshot: %s\n", name)
		fmt.Fprintf(&sb, "// Created: %s\n", entry.createdAt)
		sb.WriteString("\n")
		sb.WriteString(entry.content)
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
