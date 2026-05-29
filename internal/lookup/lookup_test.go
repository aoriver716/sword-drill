package lookup

import (
	"regexp"
	"testing"

	"github.com/aoriver716/sword-drill/internal/detector"
)

// RunLookupTests runs the shared test suite against any BibleLookup implementation.
func RunLookupTests(t *testing.T, client BibleLookup) {
	t.Run("John 3:16 KJV", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "John",
			StartChapter: 3,
			StartVerse:   16,
			EndChapter:   3,
			EndVerse:     16,
		}

		result, err := client.Lookup(ref, "kjv")
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}

		if len(result.Verses) != 1 {
			t.Fatalf("Expected 1 verse, got %d", len(result.Verses))
		}

		v := result.Verses[0]
		if v.Chapter != 3 || v.Number != 16 {
			t.Errorf("Expected chapter 3 verse 16, got chapter %d verse %d", v.Chapter, v.Number)
		}

		expected := regexp.MustCompile(`(?s)^\s*\W*For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life\.\s*$`)
		if !expected.MatchString(v.Text) {
			t.Errorf("Unexpected text:\n  got:  %q\n  want: match for %s", v.Text, expected.String())
		}
	})

	t.Run("Romans 8:28-30 KJV multi-verse", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "Romans",
			StartChapter: 8,
			StartVerse:   28,
			EndChapter:   8,
			EndVerse:     30,
		}

		result, err := client.Lookup(ref, "kjv")
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}

		if len(result.Verses) != 3 {
			for i, v := range result.Verses {
				t.Logf("  verse[%d]: chapter=%d number=%d text=%q", i, v.Chapter, v.Number, v.Text)
			}
			t.Fatalf("Expected 3 verses, got %d", len(result.Verses))
		}

		expectedNums := []int{28, 29, 30}
		for i, wantNum := range expectedNums {
			v := result.Verses[i]
			if v.Number != wantNum {
				t.Errorf("Verse %d: expected number %d, got %d", i, wantNum, v.Number)
			}
			if v.Chapter != 8 {
				t.Errorf("Verse %d: expected chapter 8, got %d", i, v.Chapter)
			}
			if v.Text == "" {
				t.Errorf("Verse %d: text is empty", i)
			}
		}

		// Verify verse 28 contains expected text
		if !regexp.MustCompile(`all things work together for good`).MatchString(result.Verses[0].Text) {
			t.Errorf("Verse 28 text doesn't match expected:\n  got: %q", result.Verses[0].Text)
		}
	})
}
