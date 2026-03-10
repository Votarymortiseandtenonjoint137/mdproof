// Package upgrade implements self-update by downloading the latest GitHub release.
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	releaseURL = "https://api.github.com/repos/runkids/mdproof/releases/latest"
	userAgent  = "mdproof-upgrade"
)

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Run checks for the latest release and replaces the current binary if a newer version exists.
func Run(currentVersion string) error {
	fmt.Println("Checking for updates...")

	rel, err := fetchLatest()
	if err != nil {
		return fmt.Errorf("check latest version: %w", err)
	}

	if rel.TagName == currentVersion {
		fmt.Printf("Already up to date (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("Updating %s → %s\n", currentVersion, rel.TagName)

	a, err := findAsset(rel)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", a.Name)

	archivePath, err := downloadToTemp(a.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer os.Remove(archivePath)

	binPath, err := extractBinary(archivePath, a.Name)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	defer os.Remove(binPath)

	if err := replaceSelf(binPath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	fmt.Printf("Upgraded to %s\n", rel.TagName)
	return nil
}

func fetchLatest() (*release, error) {
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func findAsset(rel *release) (*asset, error) {
	name := expectedAssetName(rel.TagName)
	for i := range rel.Assets {
		if rel.Assets[i].Name == name {
			return &rel.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no release for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, name)
}

func expectedAssetName(tag string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("mdproof-%s-%s-%s.zip", tag, runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("mdproof-%s-%s-%s.tar.gz", tag, runtime.GOOS, runtime.GOARCH)
}

func downloadToTemp(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "mdproof-download-*")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}

func extractBinary(archivePath, assetName string) (string, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(archivePath)
	}
	return extractTarGz(archivePath)
}

func extractTarGz(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	hdr, err := tr.Next()
	if err != nil {
		return "", fmt.Errorf("read tar entry: %w", err)
	}

	tmp, err := os.CreateTemp("", "mdproof-bin-*")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmp, tr); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), hdr.FileInfo().Mode()|0o755); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

func extractZip(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	if len(r.File) == 0 {
		return "", fmt.Errorf("zip archive is empty")
	}

	rc, err := r.File[0].Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	tmp, err := os.CreateTemp("", "mdproof-bin-*.exe")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmp, rc); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil
}

func replaceSelf(newBinary string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find current executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	// Create temp file in the same directory for atomic rename.
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".mdproof-upgrade-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	src, err := os.Open(newBinary)
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}

	if _, err := io.Copy(tmp, src); err != nil {
		src.Close()
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	src.Close()
	tmp.Close()

	// Preserve original permissions, ensure executable.
	info, err := os.Stat(execPath)
	if err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, info.Mode()|0o755); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic replace.
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
