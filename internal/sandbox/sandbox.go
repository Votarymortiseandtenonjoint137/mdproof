package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// 4. Cross-compile (or use cache).
	targetOS, targetArch := rt.TargetPlatform()
	binaryPath, fromCache, err := CachedBuild(version, targetOS, targetArch)
	if err != nil {
		// Fallback: try downloading from releases.
		binaryPath, err = DownloadBinary(version, targetOS, targetArch)
		if err != nil {
			return 1, fmt.Errorf("build binary: %w", err)
		}
		fromCache = false
	}
	if fromCache {
		fmt.Fprintf(os.Stderr, "Using cached binary (%s)\n", version)
	} else {
		fmt.Fprintf(os.Stderr, "Building mdproof for %s/%s...\n", targetOS, targetArch)
		defer os.Remove(binaryPath)
	}

	// 5. Detect dependencies from passthrough args (file paths).
	deps := detectDepsFromFiles(passthrough)

	// 6. Get working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return 1, fmt.Errorf("get working directory: %w", err)
	}

	// 7. Resolve file paths relative to CWD for container mount.
	passthrough, err = resolvePassthroughPaths(passthrough, cwd)
	if err != nil {
		return 1, err
	}

	// 8. Run in container.
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
	lines := strings.Split(content, "\n")
	inBlock := false
	var current []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock {
			if trimmed == "```bash" || trimmed == "```sh" {
				inBlock = true
				current = nil
			}
		} else {
			if trimmed == "```" {
				blocks = append(blocks, strings.Join(current, "\n"))
				inBlock = false
			} else {
				current = append(current, line)
			}
		}
	}
	return blocks
}

// resolvePassthroughPaths converts absolute file/directory paths in passthrough
// args to relative paths from CWD. This ensures they resolve correctly inside
// the container where CWD is mounted at /workspace.
// Returns an error if a path resolves outside CWD (not mounted in container).
func resolvePassthroughPaths(args []string, cwd string) ([]string, error) {
	resolved := make([]string, len(args))
	for i, arg := range args {
		// Skip flags.
		if strings.HasPrefix(arg, "-") {
			resolved[i] = arg
			continue
		}

		// Only process paths that exist on the host.
		if _, err := os.Stat(arg); err != nil {
			resolved[i] = arg
			continue
		}

		abs, err := filepath.Abs(arg)
		if err != nil {
			resolved[i] = arg
			continue
		}

		rel, err := filepath.Rel(cwd, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("path %q is outside the working directory %q\n  sandbox mounts only the working directory into the container\n  move the file into %s or cd to a parent directory", arg, cwd, cwd)
		}

		resolved[i] = rel
	}
	return resolved, nil
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
