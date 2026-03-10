package executor

import (
	"testing"
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
		gotIs, gotName := parseSnapshotPattern(tc.pat)
		if gotIs != tc.wantIs || gotName != tc.wantName {
			t.Errorf("parseSnapshotPattern(%q) = (%v, %q), want (%v, %q)",
				tc.pat, gotIs, gotName, tc.wantIs, tc.wantName)
		}
	}
}

func TestSplitSnapshotExpected(t *testing.T) {
	expected := []string{"apple", "snapshot: fruit-snap", "banana", "snapshot:veggie-snap"}
	regular, snapNames := splitSnapshotExpected(expected)

	if len(regular) != 2 || regular[0] != "apple" || regular[1] != "banana" {
		t.Errorf("regular = %v, want [apple banana]", regular)
	}
	if len(snapNames) != 2 || snapNames[0] != "fruit-snap" || snapNames[1] != "veggie-snap" {
		t.Errorf("snapNames = %v, want [fruit-snap veggie-snap]", snapNames)
	}
}

func TestSplitSnapshotExpected_NoSnapshots(t *testing.T) {
	expected := []string{"apple", "banana"}
	regular, snapNames := splitSnapshotExpected(expected)

	if len(regular) != 2 {
		t.Errorf("regular = %v, want [apple banana]", regular)
	}
	if len(snapNames) != 0 {
		t.Errorf("snapNames = %v, want []", snapNames)
	}
}
