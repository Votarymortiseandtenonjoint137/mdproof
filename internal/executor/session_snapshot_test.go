package executor

import (
	"testing"

	"github.com/runkids/mdproof/internal/core"
)

func TestIsSnapshotPattern(t *testing.T) {
	tests := []struct {
		pat      string
		wantIs   bool
		wantName string
	}{
		{"snapshot: my-snap", true, "my-snap"},
		{"snapshot:my-snap", true, "my-snap"},
		{"substring match", false, ""},
		{"exit_code: 0", false, ""},
		{"", false, ""},
	}
	for _, tc := range tests {
		gotIs, gotName := ParseSnapshotPattern(tc.pat)
		if gotIs != tc.wantIs || gotName != tc.wantName {
			t.Errorf("ParseSnapshotPattern(%q) = (%v, %q), want (%v, %q)",
				tc.pat, gotIs, gotName, tc.wantIs, tc.wantName)
		}
	}
}

func TestSplitSnapshotExpected(t *testing.T) {
	expected := core.Expectations("apple", "snapshot: fruit-snap", "banana", "snapshot:veggie-snap")
	regular, snapNames := splitSnapshotExpected(expected)

	if len(regular) != 2 || regular[0].Text != "apple" || regular[1].Text != "banana" {
		t.Errorf("regular = %v, want [apple banana]", regular)
	}
	if len(snapNames) != 2 || snapNames[0].Text != "snapshot: fruit-snap" || snapNames[1].Text != "snapshot:veggie-snap" {
		t.Errorf("snapNames = %v, want snapshot expectations preserved", snapNames)
	}
}

func TestSplitSnapshotExpected_NoSnapshots(t *testing.T) {
	expected := core.Expectations("apple", "banana")
	regular, snapNames := splitSnapshotExpected(expected)

	if len(regular) != 2 {
		t.Errorf("regular = %v, want [apple banana]", regular)
	}
	if len(snapNames) != 0 {
		t.Errorf("snapNames = %v, want []", snapNames)
	}
}
