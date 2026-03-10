package watcher

import (
	"os"
	"time"
)

// Watcher monitors files for changes using os.Stat polling.
type Watcher struct {
	files    []string
	modTimes map[string]time.Time
}

// New creates a Watcher for the given files.
func New(files []string) *Watcher {
	return &Watcher{
		files:    files,
		modTimes: make(map[string]time.Time),
	}
}

// Snapshot records the current modification times of all watched files.
func (w *Watcher) Snapshot() {
	for _, f := range w.files {
		info, err := os.Stat(f)
		if err == nil {
			w.modTimes[f] = info.ModTime()
		}
	}
}

// DetectChanges compares current file state against the last snapshot.
// Returns list of changed files and updates the snapshot.
func (w *Watcher) DetectChanges() []string {
	var changed []string

	for _, f := range w.files {
		info, err := os.Stat(f)
		if err != nil {
			// File was deleted or inaccessible
			if _, existed := w.modTimes[f]; existed {
				changed = append(changed, f)
				delete(w.modTimes, f)
			}
			continue
		}

		prev, existed := w.modTimes[f]
		if !existed || info.ModTime().After(prev) {
			changed = append(changed, f)
			w.modTimes[f] = info.ModTime()
		}
	}

	return changed
}

// SetFiles updates the list of watched files (for directory re-scanning).
// Stale entries no longer in the new list are pruned from modTimes.
func (w *Watcher) SetFiles(files []string) {
	w.files = files
	current := make(map[string]bool, len(files))
	for _, f := range files {
		current[f] = true
	}
	for k := range w.modTimes {
		if !current[k] {
			delete(w.modTimes, k)
		}
	}
}

// Dedup removes duplicate file paths from a list.
func Dedup(files []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	return result
}
