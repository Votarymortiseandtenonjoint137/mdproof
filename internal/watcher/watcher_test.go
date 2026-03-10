package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectChanges_ModifiedFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test-proof.md")
	os.WriteFile(f, []byte("# test"), 0644)

	w := New([]string{f})
	w.Snapshot() // take initial snapshot

	// Modify file
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(f, []byte("# modified"), 0644)

	changed := w.DetectChanges()
	if len(changed) != 1 || changed[0] != f {
		t.Fatalf("expected 1 changed file, got %v", changed)
	}
}

func TestDetectChanges_NoChange(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test-proof.md")
	os.WriteFile(f, []byte("# test"), 0644)

	w := New([]string{f})
	w.Snapshot()

	changed := w.DetectChanges()
	if len(changed) != 0 {
		t.Fatalf("expected no changes, got %v", changed)
	}
}

func TestDetectChanges_DeletedFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test-proof.md")
	os.WriteFile(f, []byte("# test"), 0644)

	w := New([]string{f})
	w.Snapshot()

	os.Remove(f)

	changed := w.DetectChanges()
	if len(changed) != 1 {
		t.Fatalf("expected 1 changed (deleted) file, got %v", changed)
	}
}

func TestDedup(t *testing.T) {
	events := []string{"a.md", "b.md", "a.md", "c.md", "a.md"}
	result := Dedup(events)
	if len(result) != 3 {
		t.Fatalf("expected 3 unique files, got %d: %v", len(result), result)
	}
}
