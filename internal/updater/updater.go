package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner = "aoriver716"
	repoName  = "sword-drill"
)

// Version is set at build time via -ldflags:
//
//	-ldflags "-X github.com/aoriver716/sword-drill/internal/updater.Version=v0.4.1"
var Version = "dev"

// UpdateInfo contains information about an available update.
type UpdateInfo struct {
	Available   bool   `json:"available"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	DownloadURL string `json:"downloadURL"`
	ReleaseURL  string `json:"releaseURL"`
	Error       string `json:"error,omitempty"`
}

// githubRelease is a subset of the GitHub API response for a release.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	HTMLURL string        `json:"html_url"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset is a subset of the GitHub API response for a release asset.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdates queries the GitHub API for the latest release and compares
// it against the current version. Returns platform-specific download info.
// Nightly builds check against the nightly release (by commit SHA).
// Stable builds check against the latest stable release (by semver).
func CheckForUpdates() UpdateInfo {
	info := UpdateInfo{Current: Version}

	if Version == "dev" || strings.HasPrefix(Version, "pr-") {
		info.Error = "Cannot check for updates in dev builds"
		return info
	}

	if strings.HasPrefix(Version, "nightly-") {
		return checkNightlyUpdate(info)
	}

	return checkStableUpdate(info)
}

// checkStableUpdate checks for a newer stable release via /releases/latest.
func checkStableUpdate(info UpdateInfo) UpdateInfo {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := client.Get(url)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to check for updates: %v", err)
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("GitHub API returned status %d", resp.StatusCode)
		return info
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		info.Error = fmt.Sprintf("Failed to parse response: %v", err)
		return info
	}

	info.Latest = release.TagName
	info.ReleaseURL = release.HTMLURL
	info.Available = isNewer(release.TagName, Version)

	if info.Available {
		info.DownloadURL = findPlatformAsset(release.Assets)
	}

	return info
}

// checkNightlyUpdate checks if the nightly release has a newer commit than
// the currently running nightly build. Compares the short SHA embedded in
// the version string (nightly-YYYYMMDD-<sha>) against the nightly release's
// target commit.
func checkNightlyUpdate(info UpdateInfo) UpdateInfo {
	currentSHA := extractNightlySHA(Version)
	if currentSHA == "" {
		info.Error = "Cannot parse commit SHA from nightly version"
		return info
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/nightly", repoOwner, repoName)

	resp, err := client.Get(url)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to check for updates: %v", err)
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No nightly release exists yet
		return info
	}
	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("GitHub API returned status %d", resp.StatusCode)
		return info
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		info.Error = fmt.Sprintf("Failed to parse response: %v", err)
		return info
	}

	// The nightly release tag_name is "nightly" but the body or name may
	// not contain the SHA. Use the target_commitish from the release which
	// is the full SHA the tag points to. We fetch it via the git tag API.
	latestSHA, err := getNightlyTagSHA(client)
	if err != nil {
		info.Error = fmt.Sprintf("Failed to resolve nightly tag: %v", err)
		return info
	}

	// Compare short SHAs (the current version has a 7-char short SHA)
	latestShort := latestSHA
	if len(latestShort) > len(currentSHA) {
		latestShort = latestShort[:len(currentSHA)]
	}

	if latestShort != currentSHA {
		info.Available = true
		info.Latest = release.TagName
		info.ReleaseURL = release.HTMLURL
		info.DownloadURL = findPlatformAsset(release.Assets)
	}

	return info
}

// getNightlyTagSHA resolves the nightly tag to its commit SHA.
func getNightlyTagSHA(client *http.Client) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/tags/nightly", repoOwner, repoName)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var ref struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return "", err
	}
	return ref.Object.SHA, nil
}

// extractNightlySHA extracts the commit SHA from a nightly version string.
// Format: nightly-YYYYMMDD-<sha>
func extractNightlySHA(version string) string {
	parts := strings.Split(version, "-")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// findPlatformAsset returns the download URL for the current platform's binary.
func findPlatformAsset(assets []githubAsset) string {
	var pattern string
	switch runtime.GOOS {
	case "windows":
		pattern = "windows"
	case "darwin":
		pattern = "macos"
	case "linux":
		pattern = "linux"
	default:
		return ""
	}

	for _, a := range assets {
		if strings.Contains(strings.ToLower(a.Name), pattern) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

// isNewer returns true if latest is a newer version than current.
// Expects semver-like tags: v1.2.3
func isNewer(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		l := toInt(latestParts[i])
		c := toInt(currentParts[i])
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}
	return len(latestParts) > len(currentParts)
}

// toInt converts a string to an integer, returning 0 on failure.
func toInt(s string) int {
	n := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			break
		}
	}
	return n
}
