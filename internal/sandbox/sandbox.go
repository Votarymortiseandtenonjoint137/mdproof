package sandbox

import (
	"fmt"
	"os"

	"github.com/runkids/mdproof/internal/config"
)

// DefaultImage is the default container image for sandbox mode.
const DefaultImage = "debian:bookworm-slim"

// Opts holds sandbox execution options.
type Opts struct {
	Image string
	Keep  bool
	RO    bool
}

// DefaultOpts returns sandbox options with default values.
func DefaultOpts() Opts {
	return Opts{
		Image: DefaultImage,
	}
}

// MergeOpts applies config file settings onto defaults.
func MergeOpts(base Opts, cfg *config.SandboxConfig) Opts {
	if cfg == nil {
		return base
	}
	if cfg.Image != "" {
		base.Image = cfg.Image
	}
	if cfg.Keep {
		base.Keep = true
	}
	if cfg.RO {
		base.RO = true
	}
	return base
}

// ParseSandboxArgs separates sandbox-specific flags from mdproof passthrough args.
// Returns sandbox options and the remaining args to pass to mdproof inside the container.
func ParseSandboxArgs(args []string) (Opts, []string) {
	opts := DefaultOpts()
	var passthrough []string

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--image":
			if i+1 < len(args) {
				opts.Image = args[i+1]
				i += 2
			} else {
				i++
			}
		case "--keep":
			opts.Keep = true
			i++
		case "--ro":
			opts.RO = true
			i++
		default:
			passthrough = append(passthrough, args[i])
			i++
		}
	}

	return opts, passthrough
}

// Run is the main entry point for the sandbox subcommand.
// It cross-compiles mdproof, detects dependencies, provisions a container,
// and executes mdproof inside it.
func Run(args []string, fileCfg config.Config, version string) (int, error) {
	// 1. Parse sandbox flags + collect passthrough.
	sOpts, passthrough := ParseSandboxArgs(args)

	// 2. Merge config file settings.
	sOpts = MergeOpts(sOpts, fileCfg.Sandbox)

	if len(passthrough) == 0 {
		return 1, fmt.Errorf("usage: mdproof sandbox [--image IMAGE] [--keep] [--ro] <file.md|directory> [mdproof flags...]")
	}

	// 3. Detect container runtime.
	runtimeOverride := os.Getenv("MDPROOF_RUNTIME")
	rt, err := DetectRuntime(runtimeOverride)
	if err != nil {
		return 1, err
	}

	// 4. Cross-compile mdproof for target platform.
	targetOS, targetArch := rt.TargetPlatform()
	fmt.Fprintf(os.Stderr, "Building mdproof for %s/%s...\n", targetOS, targetArch)

	binaryPath, err := BuildBinary(targetOS, targetArch)
	if err != nil {
		// Fallback: try downloading from releases.
		binaryPath, err = DownloadBinary(version, targetOS, targetArch)
		if err != nil {
			return 1, fmt.Errorf("build binary: %w", err)
		}
	}
	defer os.Remove(binaryPath)

	// 5. Detect dependencies from passthrough args (file paths).
	deps := detectDepsFromFiles(passthrough)

	// 6. Get working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return 1, fmt.Errorf("get working directory: %w", err)
	}

	// 7. Run in container.
	fmt.Fprintf(os.Stderr, "Running in %s (image: %s)...\n", runtimeName(rt), sOpts.Image)

	exitCode, err := rt.Run(RunOpts{
		Image:      sOpts.Image,
		WorkDir:    cwd,
		MountRO:    sOpts.RO,
		Keep:       sOpts.Keep,
		BinaryPath: binaryPath,
		Deps:       deps,
		Args:       passthrough,
	})

	return exitCode, err
}

// detectDepsFromFiles extracts file paths from passthrough args, parses them,
// and detects tool dependencies from code blocks.
func detectDepsFromFiles(args []string) []string {
	var commands []string
	for _, arg := range args {
		// Skip flags (start with -).
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}
		// Try to read the file and extract commands.
		data, err := os.ReadFile(arg)
		if err != nil {
			continue // might be a directory or non-existent — skip
		}
		commands = append(commands, extractCodeBlocks(string(data))...)
	}
	return DetectDeps(commands)
}

// extractCodeBlocks extracts command text from fenced bash code blocks.
func extractCodeBlocks(content string) []string {
	var blocks []string
	lines := splitLines(content)
	inBlock := false
	var current []string

	for _, line := range lines {
		trimmed := trimString(line)
		if !inBlock {
			if trimmed == "```bash" || trimmed == "```sh" {
				inBlock = true
				current = nil
			}
		} else {
			if trimmed == "```" {
				blocks = append(blocks, joinLines(current))
				inBlock = false
			} else {
				current = append(current, line)
			}
		}
	}
	return blocks
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimString(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func joinLines(lines []string) string {
	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

// runtimeName returns a human-readable name for the runtime.
func runtimeName(rt Runtime) string {
	switch rt.(type) {
	case *DockerRuntime:
		return "Docker"
	case *AppleRuntime:
		return "Apple container"
	default:
		return "container"
	}
}
