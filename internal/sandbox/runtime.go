package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RunOpts holds options for a container run invocation.
type RunOpts struct {
	Image      string            // container image
	WorkDir    string            // host directory to mount at /workspace
	MountRO    bool              // mount read-only
	Keep       bool              // don't auto-remove container
	BinaryPath string            // path to cross-compiled mdproof binary
	Deps       []string          // apt packages to install
	Args       []string          // mdproof CLI args (flags + files) to pass through
	Env        map[string]string // extra env vars
}

// Runtime abstracts container execution.
type Runtime interface {
	Available() bool
	Run(opts RunOpts) (exitCode int, err error)
	// TargetPlatform returns GOOS and GOARCH for cross-compilation.
	TargetPlatform() (string, string)
}

// DockerRuntime executes containers via the docker CLI.
type DockerRuntime struct{}

// Available checks if the docker CLI is installed and responsive.
func (d *DockerRuntime) Available() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// TargetPlatform returns linux/amd64 for Docker containers.
func (d *DockerRuntime) TargetPlatform() (string, string) {
	return "linux", "amd64"
}

// buildArgs constructs the docker run argument list.
func (d *DockerRuntime) buildArgs(opts RunOpts) []string {
	args := []string{"run"}

	if !opts.Keep {
		args = append(args, "--rm")
	}

	// Mount workspace.
	mount := opts.WorkDir + ":/workspace"
	if opts.MountRO {
		mount += ":ro"
	}
	args = append(args, "-v", mount)

	// Mount mdproof binary.
	args = append(args, "-v", opts.BinaryPath+":/usr/local/bin/mdproof:ro")

	// Working directory.
	args = append(args, "-w", "/workspace")

	// Environment: always set MDPROOF_ALLOW_EXECUTE.
	args = append(args, "-e", "MDPROOF_ALLOW_EXECUTE=1")
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}

	// Image.
	args = append(args, opts.Image)

	// Command: install deps then run mdproof.
	shellCmd := buildContainerCommand(opts.Deps, opts.Args)
	args = append(args, "bash", "-c", shellCmd)

	return args
}

// Run executes mdproof inside a Docker container.
func (d *DockerRuntime) Run(opts RunOpts) (int, error) {
	args := d.buildArgs(opts)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("docker run: %w", err)
	}
	return 0, nil
}

// DetectRuntime returns the best available container runtime.
// Override with MDPROOF_RUNTIME=docker|apple.
func DetectRuntime(override string) (Runtime, error) {
	switch strings.ToLower(override) {
	case "docker":
		d := &DockerRuntime{}
		if !d.Available() {
			return nil, fmt.Errorf("docker not found in PATH")
		}
		return d, nil
	case "apple":
		a := &AppleRuntime{}
		if !a.Available() {
			return nil, fmt.Errorf("Apple container CLI not found")
		}
		return a, nil
	case "":
		// Auto-detect.
	default:
		return nil, fmt.Errorf("unknown runtime %q (use docker or apple)", override)
	}

	// Auto-detect: prefer Apple containers on macOS arm64.
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		a := &AppleRuntime{}
		if a.Available() {
			return a, nil
		}
	}

	d := &DockerRuntime{}
	if d.Available() {
		return d, nil
	}

	return nil, fmt.Errorf("no container runtime found\n  Install Docker: https://docs.docker.com/get-docker/")
}

// buildContainerCommand generates the bash -c command string for inside the container.
func buildContainerCommand(deps []string, mdproofArgs []string) string {
	var parts []string

	// Install dependencies if any.
	if len(deps) > 0 {
		pkgList := strings.Join(deps, " ")
		parts = append(parts,
			"apt-get update -qq >/dev/null 2>&1",
			fmt.Sprintf("apt-get install -y -qq %s >/dev/null 2>&1", pkgList),
		)
	}

	// Run mdproof with passthrough args.
	mdproofCmd := "mdproof " + strings.Join(mdproofArgs, " ")
	parts = append(parts, mdproofCmd)

	return strings.Join(parts, " && ")
}
