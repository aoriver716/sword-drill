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

	// ChannelStable selects the latest stable release.
	ChannelStable = "stable"
	// ChannelNightly selects the rolling nightly release.
	ChannelNightly = "nightly"
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
	// IsDowngrade is true when the offered release is on a different channel
	// that the user has switched to (e.g. on a nightly build but the selected
	// channel is stable). The UI should phrase the prompt as a downgrade.
	IsDowngrade bool   `json:"isDowngrade,omitempty"`
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

// CheckForUpdates queries the GitHub API for an available release on the
// requested channel and compares it against the current build. The channel
// argument selects which release stream to check (ChannelStable or
// ChannelNightly); when empty it defaults to stable. If the current build
// is on a different channel than the one selected, the returned UpdateInfo
// will indicate a cross-channel switch (with IsDowngrade set when moving
// from nightly back to stable).
func CheckForUpdates(channel string) UpdateInfo {
	info := UpdateInfo{Current: Version}

	if Version == "dev" || strings.HasPrefix(Version, "pr-") {
		info.Error = "Cannot check for updates in dev builds"
		return info
	}

	if channel == "" {
		channel = ChannelStable
	}

	currentIsNightly := strings.HasPrefix(Version, "nightly-")

	switch channel {
	case ChannelNightly:
		if currentIsNightly {
			return checkNightlyUpdate(info)
		}
		// Cross-channel: user on stable wants to switch to nightly.
		return checkNightlySwitch(info)
	case ChannelStable:
		if currentIsNightly {
			// Cross-channel: user on nightly wants to return to stable.
			return checkStableDowngrade(info)
		}
		return checkStableUpdate(info)
	default:
		info.Error = fmt.Sprintf("Unknown update channel: %q", channel)
		return info
	}
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

// checkStableDowngrade is called when the running build is a nightly but the
// user has selected the stable channel. Any existing stable release is
// reported as available with IsDowngrade=true so the UI can prompt the user
// to switch back to stable.
func checkStableDowngrade(info UpdateInfo) UpdateInfo {
	release, err := fetchLatestRelease(fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName))
	if err != nil {
		if err == errReleaseNotFound {
			return info
		}
		info.Error = err.Error()
		return info
	}

	info.Available = true
	info.IsDowngrade = true
	info.Latest = release.TagName
	info.ReleaseURL = release.HTMLURL
	info.DownloadURL = findPlatformAsset(release.Assets)
	return info
}

// checkNightlySwitch is called when the running build is stable but the
// user has selected the nightly channel. If a nightly release exists it is
// reported as available so the UI can prompt the user to switch.
func checkNightlySwitch(info UpdateInfo) UpdateInfo {
	release, err := fetchLatestRelease(fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/tags/nightly", repoOwner, repoName))
	if err != nil {
		if err == errReleaseNotFound {
			return info
		}
		info.Error = err.Error()
		return info
	}

	info.Available = true
	info.Latest = release.TagName
	info.ReleaseURL = release.HTMLURL
	info.DownloadURL = findPlatformAsset(release.Assets)
	return info
}

// errReleaseNotFound is returned by fetchLatestRelease when the GitHub API
// responds with 404 (no release for the requested tag).
var errReleaseNotFound = fmt.Errorf("release not found")

// fetchLatestRelease performs a GET against a GitHub releases endpoint and
// decodes the response. Returns errReleaseNotFound for 404 responses.
func fetchLatestRelease(url string) (githubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return githubRelease{}, fmt.Errorf("Failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return githubRelease{}, errReleaseNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("Failed to parse response: %v", err)
	}
	return release, nil
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
// Expects semver-like tags: v1.2.3 or v1.2.3-rc1
// An RC version is considered older than the same base version without RC.
func isNewer(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Split off pre-release suffix
	latestBase, latestPre := splitPrerelease(latest)
	currentBase, currentPre := splitPrerelease(current)

	latestParts := strings.Split(latestBase, ".")
	currentParts := strings.Split(currentBase, ".")

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

	if len(latestParts) != len(currentParts) {
		return len(latestParts) > len(currentParts)
	}

	// Same base version: stable is newer than RC
	if currentPre != "" && latestPre == "" {
		return true
	}

	return false
}

// splitPrerelease splits "1.2.3-rc1" into ("1.2.3", "rc1").
func splitPrerelease(version string) (string, string) {
	if idx := strings.Index(version, "-"); idx != -1 {
		return version[:idx], version[idx+1:]
	}
	return version, ""
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
