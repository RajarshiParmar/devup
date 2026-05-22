// Package updater implements self-update functionality using GitHub Releases.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	// githubAPI is the base URL for the GitHub Releases API.
	githubAPI = "https://api.github.com/repos/%s/%s/releases/latest"
)

// Release represents a GitHub release response (subset of fields).
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file in a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater handles checking for and applying updates from GitHub.
type Updater struct {
	Owner   string
	Repo    string
	Current string
	Client  *http.Client
}

// New creates an Updater for the given GitHub owner/repo with the current version.
func New(owner, repo, currentVersion string) *Updater {
	return &Updater{
		Owner:   owner,
		Repo:    repo,
		Current: currentVersion,
		Client:  http.DefaultClient,
	}
}

// CheckLatest queries the GitHub Releases API for the latest release.
func (u *Updater) CheckLatest(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf(githubAPI, u.Owner, u.Repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "devup-updater")

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s/%s", u.Owner, u.Repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release response: %w", err)
	}

	return &release, nil
}

// IsNewer returns true if the release tag represents a newer version than current.
func (u *Updater) IsNewer(release *Release) bool {
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(u.Current, "v")
	return latest != current && u.Current != "dev"
}

// AssetName returns the expected asset filename for the current platform.
func AssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("devup_%s_%s.tar.gz", os, arch)
}

// FindAsset locates the appropriate download asset for the current platform.
func FindAsset(release *Release) (*Asset, error) {
	target := AssetName()
	for i := range release.Assets {
		if release.Assets[i].Name == target {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no asset found for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, target)
}

// Download fetches an asset from its URL and returns the response body.
// The caller is responsible for closing the returned ReadCloser.
func (u *Updater) Download(ctx context.Context, asset *Asset) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}
	req.Header.Set("User-Agent", "devup-updater")

	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading asset: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// ReplaceBinary atomically replaces the currently running binary with new content.
func ReplaceBinary(newBinary io.Reader) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("determining executable path: %w", err)
	}

	// Write to a temporary file next to the current binary.
	tmpPath := execPath + ".new"
	tmp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := io.Copy(tmp, newBinary); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing new binary: %w", err)
	}
	tmp.Close()

	// Atomic rename.
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}
