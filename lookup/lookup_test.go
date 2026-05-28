package lookup

import (
	"regexp"
	"testing"

	"github.com/aoriver716/sword-drill/parser"
)

// RunLookupTests runs the shared test suite against any BibleLookup implementation.
func RunLookupTests(t *testing.T, client BibleLookup) {
	t.Run("John 3:16 KJV", func(t *testing.T) {
		ref := parser.ScriptureRef{
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

		expected := regexp.MustCompile(`(?s)^\s*For God so loved the world, that he gave his only begotten Son, that whosoever believeth in him should not perish, but have everlasting life\.\s*$`)
		if !expected.MatchString(v.Text) {
			t.Errorf("Unexpected text:\n  got:  %q\n  want: match for %s", v.Text, expected.String())
		}
	})
}
