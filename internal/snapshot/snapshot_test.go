package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_Check_FirstRun_CreatesSnapshot(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, false)

	result := s.Check("greeting", "hello world", "my-runbook")
	if !result.Matched {
		t.Fatal("first run should pass (create snapshot)")
	}

	// Verify file was created
	snapFile := filepath.Join(dir, "__snapshots__", "my-runbook.snap")
	data, err := os.ReadFile(snapFile)
	if err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}
	if !strings.Contains(string(data), "hello world") {
		t.Fatalf("snapshot file missing output, got: %s", data)
	}
}

func TestStore_Check_Match(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, false)

	s.Check("greeting", "hello world", "my-runbook")

	result := s.Check("greeting", "hello world", "my-runbook")
	if !result.Matched {
		t.Fatal("same output should match snapshot")
	}
}

func TestStore_Check_Mismatch(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, false)

	s.Check("greeting", "hello world", "my-runbook")

	result := s.Check("greeting", "goodbye world", "my-runbook")
	if result.Matched {
		t.Fatal("different output should not match")
	}
	if result.Detail == "" {
		t.Fatal("mismatch should include diff detail")
	}
}

func TestStore_Check_UpdateMode(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, true) // update mode

	s.Check("greeting", "hello world", "my-runbook")
	result := s.Check("greeting", "goodbye world", "my-runbook")

	if !result.Matched {
		t.Fatal("update mode should always pass")
	}

	s2 := NewStore(dir, false)
	result2 := s2.Check("greeting", "goodbye world", "my-runbook")
	if !result2.Matched {
		t.Fatal("updated snapshot should match new output")
	}
}

func TestStore_Check_DuplicateName_SameRunbook(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, false)

	s.Check("greeting", "hello", "my-runbook")
	s.Check("greeting", "hello", "my-runbook")

	result := s.Check("greeting", "different", "my-runbook")
	if result.Matched {
		t.Fatal("should detect mismatch, not create duplicate")
	}
}

func TestStore_MultipleSnapshots_SameRunbook(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir, false)

	r1 := s.Check("users-list", `[{"id":1}]`, "api-proof")
	r2 := s.Check("user-detail", `{"id":1,"name":"alice"}`, "api-proof")

	if !r1.Matched || !r2.Matched {
		t.Fatal("both snapshots should pass on first run")
	}

	data, _ := os.ReadFile(filepath.Join(dir, "__snapshots__", "api-proof.snap"))
	content := string(data)
	if !strings.Contains(content, "users-list") || !strings.Contains(content, "user-detail") {
		t.Fatal("both snapshots should be in same file")
	}
}
