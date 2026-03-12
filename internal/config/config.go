package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ConfigFileName is the conventional name for directory-level runbook config.
const ConfigFileName = "mdproof.json"

// SandboxConfig holds settings for the sandbox subcommand.
type SandboxConfig struct {
	Image string `json:"image,omitempty"` // container image (default: debian:bookworm-slim)
	Keep  bool   `json:"keep,omitempty"`  // don't auto-remove container
	RO    bool   `json:"ro,omitempty"`    // mount workspace read-only
}

// Config holds lifecycle hooks and defaults for runbook execution.
type Config struct {
	Build    string            `json:"build,omitempty"`    // command to run once before all runbooks
	Setup    string            `json:"setup,omitempty"`    // command to run before each runbook
	Teardown      string            `json:"teardown,omitempty"`       // command to run after each runbook
	StepSetup    string            `json:"step_setup,omitempty"`    // command to run before each step
	StepTeardown string            `json:"step_teardown,omitempty"` // command to run after each step
	Timeout      string            `json:"timeout,omitempty"`       // per-step timeout (e.g., "5m")
	Env      map[string]string `json:"env,omitempty"`      // environment variables seeded into all steps
	Strict   *bool             `json:"strict,omitempty"`   // container-only execution (default: true)
	Sandbox  *SandboxConfig    `json:"sandbox,omitempty"`  // sandbox subcommand settings
}

// TimeoutDuration parses the timeout string into a time.Duration.
// Returns zero if empty or invalid.
func (c Config) TimeoutDuration() time.Duration {
	if c.Timeout == "" {
		return 0
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 0
	}
	return d
}

// Load reads an mdproof.json from the given directory.
// Returns an empty config (no error) if the file doesn't exist.
func Load(dir string) (Config, error) {
	path := filepath.Join(dir, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Merge applies CLI flag overrides on top of file-based config.
// CLI flags take precedence when non-empty. strictExplicit indicates
// whether --strict was explicitly passed on the command line.
func Merge(file Config, cliBuild, cliSetup, cliTeardown, cliStepSetup, cliStepTeardown string, cliTimeout time.Duration, cliStrict bool, strictExplicit bool) Config {
	merged := file
	if cliBuild != "" {
		merged.Build = cliBuild
	}
	if cliSetup != "" {
		merged.Setup = cliSetup
	}
	if cliTeardown != "" {
		merged.Teardown = cliTeardown
	}
	if cliStepSetup != "" {
		merged.StepSetup = cliStepSetup
	}
	if cliStepTeardown != "" {
		merged.StepTeardown = cliStepTeardown
	}
	if cliTimeout != 0 {
		merged.Timeout = cliTimeout.String()
	}
	if strictExplicit {
		merged.Strict = &cliStrict
	}
	return merged
}

// IsStrict returns the effective strict mode value.
// Default is true if not set in config or CLI.
func (c Config) IsStrict() bool {
	if c.Strict == nil {
		return true
	}
	return *c.Strict
}
