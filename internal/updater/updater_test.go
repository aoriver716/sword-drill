package updater

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.5.0", "v0.4.1", true},
		{"v0.4.2", "v0.4.1", true},
		{"v1.0.0", "v0.9.9", true},
		{"v0.4.1", "v0.4.1", false},
		{"v0.4.0", "v0.4.1", false},
		{"v0.3.9", "v0.4.1", false},
		{"v2.0.0", "v1.9.9", true},
		{"v0.10.0", "v0.9.0", true},
		// RC scenarios
		{"v0.7.0", "v0.7.0-rc1", true},      // stable is newer than RC of same version
		{"v0.7.0", "v0.7.0-rc3", true},      // stable is newer than any RC
		{"v0.6.0", "v0.7.0-rc1", false},     // older stable is not newer than RC of higher version
		{"v0.7.0-rc1", "v0.7.0", false},     // RC is not newer than stable of same version
		{"v0.7.0-rc2", "v0.7.0-rc1", false}, // same base, both RC — not newer
		{"v0.8.0", "v0.7.0-rc1", true},      // higher stable is newer than lower RC
	}

	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestFindPlatformAsset(t *testing.T) {
	assets := []githubAsset{
		{Name: "sword-drill-windows-amd64.exe", BrowserDownloadURL: "https://example.com/windows.exe"},
		{Name: "sword-drill-macos-universal.app.zip", BrowserDownloadURL: "https://example.com/macos.zip"},
	}

	url := findPlatformAsset(assets)
	if url == "" {
		t.Error("expected a platform asset URL, got empty string")
	}
}

func TestExtractNightlySHA(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"nightly-20260601-abc1234", "abc1234"},
		{"nightly-20260601-deadbeef", "deadbeef"},
		{"nightly-20260601", ""},
		{"v0.6.0", ""},
		{"dev", ""},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := extractNightlySHA(tt.version)
			if got != tt.want {
				t.Errorf("extractNightlySHA(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
