package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_WithBuild(t *testing.T) {
	dir := t.TempDir()
	data := `{"build": "make build", "setup": "echo setup", "timeout": "5m"}`
	if err := os.WriteFile(filepath.Join(dir, "mdproof.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Build != "make build" {
		t.Errorf("build = %q, want %q", cfg.Build, "make build")
	}
}

func TestMerge_CLIBuildOverrides(t *testing.T) {
	file := Config{Build: "file-build"}
	merged := Merge(file, "cli-build", "", "", 0)
	if merged.Build != "cli-build" {
		t.Errorf("build = %q, want %q", merged.Build, "cli-build")
	}
}

func TestLoad_FileExists(t *testing.T) {
	dir := t.TempDir()
	data := `{"setup": "echo setup", "teardown": "echo teardown", "timeout": "5m"}`
	if err := os.WriteFile(filepath.Join(dir, "mdproof.json"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Setup != "echo setup" {
		t.Errorf("setup = %q, want %q", cfg.Setup, "echo setup")
	}
	if cfg.Teardown != "echo teardown" {
		t.Errorf("teardown = %q, want %q", cfg.Teardown, "echo teardown")
	}
	if cfg.Timeout != "5m" {
		t.Errorf("timeout = %q, want %q", cfg.Timeout, "5m")
	}
}

func TestLoad_NoFile(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Setup != "" || cfg.Teardown != "" || cfg.Timeout != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mdproof.json"), []byte("{bad"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMerge_CLIOverrides(t *testing.T) {
	file := Config{
		Setup:    "file-setup",
		Teardown: "file-teardown",
		Timeout:  "1m",
	}

	merged := Merge(file, "", "cli-setup", "", 0)
	if merged.Setup != "cli-setup" {
		t.Errorf("setup = %q, want %q", merged.Setup, "cli-setup")
	}
	if merged.Teardown != "file-teardown" {
		t.Errorf("teardown should keep file value, got %q", merged.Teardown)
	}
}

func TestMerge_CLITimeoutOverrides(t *testing.T) {
	file := Config{Timeout: "1m"}
	merged := Merge(file, "", "", "", 5*time.Minute)
	if merged.Timeout != "5m0s" {
		t.Errorf("timeout = %q, want %q", merged.Timeout, "5m0s")
	}
}

func TestTimeoutDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"30s", 30 * time.Second},
		{"", 0},
		{"invalid", 0},
	}
	for _, tt := range tests {
		cfg := Config{Timeout: tt.input}
		got := cfg.TimeoutDuration()
		if got != tt.want {
			t.Errorf("TimeoutDuration(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
