package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// buildBinaryTo cross-compiles mdproof to outPath for the given OS/arch.
func buildBinaryTo(outPath, targetOS, targetArch string) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}

	projectRoot, err := findProjectRoot(".")
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	cmd := exec.Command(goPath, "build", "-o", outPath, "./cmd/mdproof")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"GOOS="+targetOS,
		"GOARCH="+targetArch,
		"CGO_ENABLED=0",
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	return os.Chmod(outPath, 0755)
}

// BuildBinary cross-compiles mdproof for the target OS/arch.
// Returns the path to a temporary binary (caller must clean up).
func BuildBinary(targetOS, targetArch string) (string, error) {
	outFile, err := os.CreateTemp("", fmt.Sprintf("mdproof-%s-%s-*", targetOS, targetArch))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()

	if err := buildBinaryTo(outPath, targetOS, targetArch); err != nil {
		os.Remove(outPath)
		return "", err
	}

	return outPath, nil
}

// cacheDir returns the mdproof binary cache directory, creating it if needed.
func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "mdproof")
	return dir, os.MkdirAll(dir, 0755)
}

// CachedBuild returns a cached binary for the given version/os/arch, building
// it on cache miss. Dev builds bypass the cache entirely.
// Returns (path, cached, error).
func CachedBuild(version, targetOS, targetArch string) (string, bool, error) {
	// Dev builds: always build fresh (no cache).
	if version == "dev" || version == "" {
		p, err := BuildBinary(targetOS, targetArch)
		return p, false, err
	}

	dir, err := cacheDir()
	if err != nil {
		// Cache dir unavailable — fall back to temp build.
		p, err := BuildBinary(targetOS, targetArch)
		return p, false, err
	}

	name := fmt.Sprintf("mdproof-%s-%s-%s", version, targetOS, targetArch)
	cachedPath := filepath.Join(dir, name)

	// Cache hit.
	if info, statErr := os.Stat(cachedPath); statErr == nil && info.Mode().IsRegular() {
		return cachedPath, true, nil
	}

	// Cache miss — build directly into cache dir.
	if err := buildBinaryTo(cachedPath, targetOS, targetArch); err != nil {
		return "", false, err
	}

	return cachedPath, true, nil
}

// findProjectRoot walks up from startDir looking for go.mod.
func findProjectRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found (searched from %s)", startDir)
		}
		dir = parent
	}
}

// DownloadBinary downloads a pre-built mdproof binary from GitHub releases.
// Used as fallback when go toolchain is not available.
func DownloadBinary(version, targetOS, targetArch string) (string, error) {
	return "", fmt.Errorf(
		"go toolchain not found and release download not yet implemented\n" +
			"  Install Go: https://go.dev/dl/\n" +
			"  Or specify a custom image with mdproof pre-installed: --image <image>",
	)
}
