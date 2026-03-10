package sandbox

import (
	"os"
	"testing"

	"github.com/runkids/mdproof/internal/config"
)

func TestDefaultOpts(t *testing.T) {
	opts := DefaultOpts()
	if opts.Image != DefaultImage {
		t.Errorf("Image = %q, want %q", opts.Image, DefaultImage)
	}
	if opts.Keep {
		t.Error("Keep should default to false")
	}
	if opts.RO {
		t.Error("RO should default to false")
	}
}

func TestMergeOptsFromConfig(t *testing.T) {
	opts := DefaultOpts()
	cfg := &config.SandboxConfig{
		Image: "node:20",
		Keep:  true,
		RO:    true,
	}

	merged := MergeOpts(opts, cfg)
	if merged.Image != "node:20" {
		t.Errorf("Image = %q, want %q", merged.Image, "node:20")
	}
	if !merged.Keep {
		t.Error("Keep should be true from config")
	}
	if !merged.RO {
		t.Error("RO should be true from config")
	}
}

func TestMergeOptsNilConfig(t *testing.T) {
	opts := DefaultOpts()
	merged := MergeOpts(opts, nil)
	if merged.Image != DefaultImage {
		t.Errorf("Image = %q, want %q", merged.Image, DefaultImage)
	}
}

func TestParseSandboxArgs(t *testing.T) {
	args := []string{"--image", "node:20", "--keep", "--ro", "--report", "json", "tests/"}
	sOpts, passthrough := ParseSandboxArgs(args)

	if sOpts.Image != "node:20" {
		t.Errorf("Image = %q, want %q", sOpts.Image, "node:20")
	}
	if !sOpts.Keep {
		t.Error("Keep should be true")
	}
	if !sOpts.RO {
		t.Error("RO should be true")
	}

	if len(passthrough) != 3 {
		t.Fatalf("passthrough = %v, want 3 items", passthrough)
	}
	if passthrough[0] != "--report" || passthrough[1] != "json" || passthrough[2] != "tests/" {
		t.Errorf("passthrough = %v", passthrough)
	}
}

func TestParseSandboxArgsDefaults(t *testing.T) {
	args := []string{"tests/hello-proof.md"}
	sOpts, passthrough := ParseSandboxArgs(args)

	if sOpts.Image != DefaultImage {
		t.Errorf("Image = %q, want default", sOpts.Image)
	}
	if sOpts.Keep {
		t.Error("Keep should be false by default")
	}
	if len(passthrough) != 1 || passthrough[0] != "tests/hello-proof.md" {
		t.Errorf("passthrough = %v", passthrough)
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	content := "# Title\n\n```bash\ncurl http://localhost\njq .status\n```\n\nSome text\n\n```python\nprint('hi')\n```\n"
	blocks := extractCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1 (only bash)", len(blocks))
	}
	assertStringContains(t, blocks[0], "curl")
	assertStringContains(t, blocks[0], "jq")
}

func TestExtractCodeBlocksSh(t *testing.T) {
	content := "```sh\nwget http://example.com\n```\n"
	blocks := extractCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	assertStringContains(t, blocks[0], "wget")
}

func TestResolvePassthroughPaths(t *testing.T) {
	cwd := "/home/user/project"

	// Relative path stays as-is.
	args, err := resolvePassthroughPaths([]string{"tests/hello.md"}, cwd)
	if err != nil {
		t.Fatal(err)
	}
	if args[0] != "tests/hello.md" {
		t.Errorf("relative path changed: %q", args[0])
	}

	// Flags pass through unchanged.
	args, err = resolvePassthroughPaths([]string{"--report", "json"}, cwd)
	if err != nil {
		t.Fatal(err)
	}
	if args[0] != "--report" || args[1] != "json" {
		t.Errorf("flags changed: %v", args)
	}
}

func TestResolvePassthroughPathsOutsideCWD(t *testing.T) {
	// Create a temp file outside a fake CWD.
	tmp, err := os.CreateTemp("", "mdproof-test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	_, err = resolvePassthroughPaths([]string{tmp.Name()}, "/nonexistent/project")
	if err == nil {
		t.Error("expected error for path outside CWD")
	}
}

func TestAppleBuildArgs(t *testing.T) {
	a := &AppleRuntime{}
	opts := RunOpts{
		Image:      "debian:bookworm-slim",
		WorkDir:    "/Users/dev/project",
		MountRO:    false,
		Keep:       false,
		BinaryPath: "/tmp/mdproof-linux-arm64",
		Deps:       []string{"ca-certificates", "jq"},
		Args:       []string{"tests/"},
	}

	args := a.buildArgs(opts)

	assertContains(t, args, "run")
	assertContains(t, args, "--rm")
	assertContains(t, args, "debian:bookworm-slim")

	foundMount := false
	for _, arg := range args {
		if arg == "type=bind,src=/Users/dev/project,dst=/workspace" {
			foundMount = true
		}
	}
	if !foundMount {
		t.Errorf("expected --mount bind in args: %v", args)
	}
}

func TestAppleTargetPlatform(t *testing.T) {
	a := &AppleRuntime{}
	goos, goarch := a.TargetPlatform()
	if goos != "linux" || goarch != "arm64" {
		t.Errorf("got %s/%s, want linux/arm64", goos, goarch)
	}
}
