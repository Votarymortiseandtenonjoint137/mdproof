package sandbox

import (
	"testing"
)

func TestDockerBuildArgs(t *testing.T) {
	d := &DockerRuntime{}
	opts := RunOpts{
		Image:      "debian:bookworm-slim",
		WorkDir:    "/home/user/project",
		MountRO:    false,
		Keep:       false,
		BinaryPath: "/tmp/mdproof-linux-amd64",
		Deps:       []string{"ca-certificates", "curl", "jq"},
		Args:       []string{"--report", "json", "tests/"},
		Env:        map[string]string{"FOO": "bar"},
	}

	args := d.buildArgs(opts)

	assertContains(t, args, "--rm")
	assertContains(t, args, "-w")
	assertContains(t, args, "/workspace")
	assertContains(t, args, "-e")
	assertContains(t, args, "MDPROOF_ALLOW_EXECUTE=1")
	assertContains(t, args, "debian:bookworm-slim")

	// Check mount is NOT read-only.
	for _, a := range args {
		if a == "/home/user/project:/workspace:ro" {
			t.Error("expected rw mount, got ro")
		}
	}
}

func TestDockerBuildArgsReadOnly(t *testing.T) {
	d := &DockerRuntime{}
	opts := RunOpts{
		Image:      "debian:bookworm-slim",
		WorkDir:    "/home/user/project",
		MountRO:    true,
		BinaryPath: "/tmp/mdproof-linux-amd64",
		Deps:       []string{"ca-certificates"},
		Args:       []string{"tests/"},
	}

	args := d.buildArgs(opts)

	found := false
	for _, a := range args {
		if a == "/home/user/project:/workspace:ro" {
			found = true
		}
	}
	if !found {
		t.Error("expected ro mount, not found")
	}
}

func TestDockerBuildArgsKeep(t *testing.T) {
	d := &DockerRuntime{}
	opts := RunOpts{
		Image:      "debian:bookworm-slim",
		WorkDir:    "/home/user/project",
		Keep:       true,
		BinaryPath: "/tmp/mdproof-linux-amd64",
		Deps:       []string{"ca-certificates"},
		Args:       []string{"tests/"},
	}

	args := d.buildArgs(opts)

	for _, a := range args {
		if a == "--rm" {
			t.Error("--rm should not be present when Keep=true")
		}
	}
}

func TestDockerTargetPlatform(t *testing.T) {
	d := &DockerRuntime{}
	goos, goarch := d.TargetPlatform()
	if goos != "linux" || goarch != "amd64" {
		t.Errorf("got %s/%s, want linux/amd64", goos, goarch)
	}
}

func TestBuildContainerCommand(t *testing.T) {
	cmd := buildContainerCommand([]string{"jq", "curl"}, []string{"--report", "json", "tests/"})
	if cmd == "" {
		t.Fatal("empty command")
	}
	// Should contain apt-get and mdproof.
	assertStringContains(t, cmd, "apt-get")
	assertStringContains(t, cmd, "mdproof --report json tests/")
}

func TestBuildContainerCommandNoDeps(t *testing.T) {
	cmd := buildContainerCommand(nil, []string{"tests/"})
	// Should NOT contain apt-get.
	if containsString(cmd, "apt-get") {
		t.Errorf("should not contain apt-get when no deps: %q", cmd)
	}
	assertStringContains(t, cmd, "mdproof tests/")
}

func TestDetectRuntime(t *testing.T) {
	rt, err := DetectRuntime("")
	if err != nil {
		t.Skipf("no runtime available: %v", err)
	}
	if rt == nil {
		t.Fatal("expected non-nil runtime")
	}
}

func TestDetectRuntimeInvalidOverride(t *testing.T) {
	_, err := DetectRuntime("podman")
	if err == nil {
		t.Error("expected error for unknown runtime")
	}
}

func assertContains(t *testing.T, slice []string, want string) {
	t.Helper()
	for _, s := range slice {
		if s == want {
			return
		}
	}
	t.Errorf("slice %v does not contain %q", slice, want)
}

func assertStringContains(t *testing.T, s, sub string) {
	t.Helper()
	if !containsString(s, sub) {
		t.Errorf("%q does not contain %q", s, sub)
	}
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
