package sandbox

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCacheDir(t *testing.T) {
	dir, err := cacheDir()
	if err != nil {
		t.Fatalf("cacheDir() error: %v", err)
	}
	if !strings.Contains(dir, "mdproof") {
		t.Errorf("cache dir should contain 'mdproof', got: %s", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("cache dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("cache dir should be a directory")
	}
}

func TestCachedBuild_DevSkipsCache(t *testing.T) {
	// version="dev" should return cached=false.
	// We can't easily test the full build without a Go project,
	// so we just verify the function signature and dev-skip logic
	// by checking that it attempts a build (which will fail outside
	// the project root, confirming it didn't use cache).
	_, cached, err := CachedBuild("dev", runtime.GOOS, runtime.GOARCH)
	if err == nil && cached {
		t.Errorf("dev version should not be cached")
	}
	// Error is expected if not run from project root — that's fine.
}

func TestCachedBuild_CacheHit(t *testing.T) {
	dir, err := cacheDir()
	if err != nil {
		t.Fatalf("cacheDir() error: %v", err)
	}

	// Place a fake binary in the cache.
	version := "test-cache-hit-v999"
	name := "mdproof-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH
	fakePath := filepath.Join(dir, name)

	if err := os.WriteFile(fakePath, []byte("fake"), 0755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	defer os.Remove(fakePath)

	path, cached, err := CachedBuild(version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("CachedBuild() error: %v", err)
	}
	if !cached {
		t.Errorf("expected cached=true for existing binary")
	}
	if path != fakePath {
		t.Errorf("expected path=%s, got=%s", fakePath, path)
	}
}
