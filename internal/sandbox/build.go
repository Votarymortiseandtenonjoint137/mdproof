package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BuildBinary cross-compiles mdproof for the target OS/arch.
// Returns the path to the temporary binary.
func BuildBinary(targetOS, targetArch string) (string, error) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("go not found in PATH: %w", err)
	}

	projectRoot, err := findProjectRoot(".")
	if err != nil {
		return "", fmt.Errorf("find project root: %w", err)
	}

	outFile, err := os.CreateTemp("", fmt.Sprintf("mdproof-%s-%s-*", targetOS, targetArch))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	outPath := outFile.Name()
	outFile.Close()

	cmd := exec.Command(goPath, "build", "-o", outPath, "./cmd/mdproof")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"GOOS="+targetOS,
		"GOARCH="+targetArch,
		"CGO_ENABLED=0",
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("go build: %w", err)
	}

	if err := os.Chmod(outPath, 0755); err != nil {
		os.Remove(outPath)
		return "", err
	}

	return outPath, nil
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
