package sandbox

import (
	"fmt"
	"os/exec"
)

// AppleRuntime executes containers via Apple's container CLI (macOS 26+).
type AppleRuntime struct{}

// Available checks if the Apple container CLI is installed.
func (a *AppleRuntime) Available() bool {
	_, err := exec.LookPath("container")
	return err == nil
}

// TargetPlatform returns linux/arm64 for Apple containers on Apple Silicon.
func (a *AppleRuntime) TargetPlatform() (string, string) {
	return "linux", "arm64"
}

// buildArgs constructs the container run argument list for Apple containers.
func (a *AppleRuntime) buildArgs(opts RunOpts) []string {
	args := []string{"run"}

	if !opts.Keep {
		args = append(args, "--rm")
	}

	// Apple containers use --mount syntax.
	mountOpt := fmt.Sprintf("type=bind,src=%s,dst=%s", opts.WorkDir, containerWorkDir)
	if opts.MountRO {
		mountOpt += ",readonly"
	}
	args = append(args, "--mount", mountOpt)

	// Mount mdproof binary.
	args = append(args, "--mount",
		fmt.Sprintf("type=bind,src=%s,dst=/usr/local/bin/mdproof,readonly", opts.BinaryPath))

	// Working directory.
	args = append(args, "-w", containerWorkDir)

	// Environment.
	args = append(args, "-e", allowExecuteEnv)
	for k, v := range opts.Env {
		args = append(args, "-e", k+"="+v)
	}

	// Image.
	args = append(args, opts.Image)

	// Command: reuse shared buildContainerCommand from runtime.go.
	shellCmd := buildContainerCommand(opts.Deps, opts.Args)
	args = append(args, "bash", "-c", shellCmd)

	return args
}

// Run executes mdproof inside an Apple container.
func (a *AppleRuntime) Run(opts RunOpts) (int, error) {
	return runContainer("container", a.buildArgs(opts))
}
