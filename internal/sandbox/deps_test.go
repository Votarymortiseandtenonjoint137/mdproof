package sandbox

import (
	"testing"
)

func TestDetectDeps(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		want     []string
	}{
		{
			name:     "jq and curl detected",
			commands: []string{"curl -s http://localhost | jq .status"},
			want:     []string{"ca-certificates", "curl", "jq"},
		},
		{
			name:     "python detected",
			commands: []string{"python3 -c 'print(1)'"},
			want:     []string{"ca-certificates", "python3"},
		},
		{
			name:     "no tools detected",
			commands: []string{"echo hello", "ls -la"},
			want:     []string{"ca-certificates"},
		},
		{
			name:     "deduplication",
			commands: []string{"curl http://a", "curl http://b"},
			want:     []string{"ca-certificates", "curl"},
		},
		{
			name:     "python alias maps to python3",
			commands: []string{"python script.py"},
			want:     []string{"ca-certificates", "python3"},
		},
		{
			name:     "word boundary — no false positive on downloading",
			commands: []string{"echo 'downloading...'"},
			want:     []string{"ca-certificates"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDeps(tt.commands)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
