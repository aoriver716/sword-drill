package formatter

import (
	"testing"

	"github.com/aoriver716/sword-drill/internal/config"
	"github.com/aoriver716/sword-drill/internal/lookup"
)

func TestFormat_BasicText(t *testing.T) {
	cfg := &config.Config{}
	result := lookup.LookupResult{
		Translation: "KJV",
		Verses: []lookup.Verse{
			{Number: 16, Text: "For God so loved the world"},
			{Number: 17, Text: "For God sent not his Son"},
		},
	}

	got := Format(result, cfg)
	want := "For God so loved the worldFor God sent not his Son"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_VerseByVerse(t *testing.T) {
	cfg := &config.Config{}
	cfg.FormattingOptions.VerseByVerse = true
	result := lookup.LookupResult{
		Translation: "KJV",
		Verses: []lookup.Verse{
			{Number: 1, Text: "First verse"},
			{Number: 2, Text: "Second verse"},
		},
	}

	got := Format(result, cfg)
	want := "First verse\nSecond verse"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_ShowVerseNums(t *testing.T) {
	cfg := &config.Config{}
	cfg.FormattingOptions.ShowVerseNums = true
	result := lookup.LookupResult{
		Translation: "KJV",
		Verses: []lookup.Verse{
			{Number: 3, Text: "Some text"},
		},
	}

	got := Format(result, cfg)
	want := "3 Some text"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_EmptyVerses_ReturnsOmittedMessage(t *testing.T) {
	cfg := &config.Config{}
	result := lookup.LookupResult{
		Translation: "ESV",
		Verses:      nil,
	}

	got := Format(result, cfg)
	want := "[Verse not included in ESV]"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_EmptyVerses_IncludesTranslationName(t *testing.T) {
	cfg := &config.Config{}
	result := lookup.LookupResult{
		Translation: "NIV",
		Verses:      []lookup.Verse{},
	}

	got := Format(result, cfg)
	want := "[Verse not included in NIV]"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}
