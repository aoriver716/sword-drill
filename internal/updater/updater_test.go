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
