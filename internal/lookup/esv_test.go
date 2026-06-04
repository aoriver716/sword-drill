package lookup

import (
	"strings"
	"testing"

	"github.com/aoriver716/sword-drill/internal/detector"
)

func TestESVClient(t *testing.T) {
	if !ESVKeyAvailable() {
		t.Skip("ESV API key not available; skipping integration test")
	}
	client := NewESVClient()

	t.Run("John 3:16 single verse", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "John",
			StartChapter: 3,
			StartVerse:   16,
			EndChapter:   3,
			EndVerse:     16,
		}
		result, err := client.Lookup(ref, "esv")
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
		if v.Text == "" {
			t.Error("Verse text is empty")
		}
		t.Logf("Text: %q", v.Text)
	})

	t.Run("Romans 8:28-30 multi-verse", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "Romans",
			StartChapter: 8,
			StartVerse:   28,
			EndChapter:   8,
			EndVerse:     30,
		}
		result, err := client.Lookup(ref, "esv")
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
	})

	t.Run("Psalm 23 whole chapter", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "Psalm",
			StartChapter: 23,
			EndChapter:   23,
		}
		result, err := client.Lookup(ref, "esv")
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}
		if len(result.Verses) < 6 {
			t.Fatalf("Expected at least 6 verses for Psalm 23, got %d", len(result.Verses))
		}
		t.Logf("Got %d verses", len(result.Verses))
	})

	t.Run("3 John 14 single-chapter book", func(t *testing.T) {
		ref := detector.ScriptureRef{
			Book:         "3 John",
			StartChapter: 1,
			StartVerse:   14,
			EndChapter:   1,
			EndVerse:     14,
		}
		result, err := client.Lookup(ref, "esv")
		if err != nil {
			t.Fatalf("Lookup failed: %v", err)
		}
		if len(result.Verses) == 0 {
			t.Fatal("Expected at least 1 verse, got 0")
		}
		v := result.Verses[0]
		if v.Chapter != 1 || v.Number != 14 {
			t.Errorf("Expected chapter 1 verse 14, got chapter %d verse %d", v.Chapter, v.Number)
		}
		if v.Text == "" {
			t.Error("Verse text is empty")
		}
		t.Logf("Text: %q", v.Text)
	})

	t.Run("Translations", func(t *testing.T) {
		translations, err := client.Translations()
		if err != nil {
			t.Fatalf("Translations failed: %v", err)
		}
		if len(translations) != 1 {
			t.Fatalf("Expected 1 translation, got %d", len(translations))
		}
		if translations[0].Key != "esv" {
			t.Errorf("Expected key 'esv', got %q", translations[0].Key)
		}
	})
}

func TestParseESVVerses(t *testing.T) {
	ref := detector.ScriptureRef{
		Book:         "Romans",
		StartChapter: 8,
		StartVerse:   28,
		EndChapter:   8,
		EndVerse:     30,
	}

	passage := `  [28] And we know that for those who love God all things work together for good, for those who are called according to his purpose. [29] For those whom he foreknew he also predestined to be conformed to the image of his Son, in order that he might be the firstborn among many brothers. [30] And those whom he predestined he also called, and those whom he called he also justified, and those whom he justified he also glorified. (ESV)`

	verses := parseESVVerses(passage, ref)

	if len(verses) != 3 {
		for i, v := range verses {
			t.Logf("  verse[%d]: chapter=%d number=%d text=%q", i, v.Chapter, v.Number, v.Text)
		}
		t.Fatalf("Expected 3 verses, got %d", len(verses))
	}

	if verses[0].Number != 28 || verses[1].Number != 29 || verses[2].Number != 30 {
		t.Errorf("Verse numbers: got %d, %d, %d", verses[0].Number, verses[1].Number, verses[2].Number)
	}

	if verses[0].Chapter != 8 {
		t.Errorf("Expected chapter 8, got %d", verses[0].Chapter)
	}

	// Last verse should include the (ESV) short copyright
	if verses[2].Text == "" {
		t.Error("Last verse text is empty")
	}

	// Verse 28 should contain expected text
	if !strings.Contains(verses[0].Text, "all things work together for good") {
		t.Errorf("Verse 28 doesn't contain expected text: %q", verses[0].Text)
	}
}

func TestParseESVVerses_CrossChapter(t *testing.T) {
	ref := detector.ScriptureRef{
		Book:         "Romans",
		StartChapter: 8,
		StartVerse:   38,
		EndChapter:   9,
		EndVerse:     1,
	}

	// Simulated text with a chapter boundary (verse numbers reset)
	passage := `  [38] For I am sure that neither death nor life. [39] nor anything else in all creation. [1] I am speaking the truth in Christ. (ESV)`

	verses := parseESVVerses(passage, ref)

	if len(verses) != 3 {
		t.Fatalf("Expected 3 verses, got %d", len(verses))
	}

	if verses[0].Chapter != 8 || verses[1].Chapter != 8 {
		t.Errorf("First two verses should be chapter 8, got %d and %d", verses[0].Chapter, verses[1].Chapter)
	}
	if verses[2].Chapter != 9 {
		t.Errorf("Third verse should be chapter 9, got %d", verses[2].Chapter)
	}
}

func TestFormatESVQuery(t *testing.T) {
	tests := []struct {
		name string
		ref  detector.ScriptureRef
		want string
	}{
		{
			name: "single verse",
			ref:  detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 16},
			want: "John 3:16",
		},
		{
			name: "verse range",
			ref:  detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 18},
			want: "John 3:16-18",
		},
		{
			name: "chapter only",
			ref:  detector.ScriptureRef{Book: "Psalm", StartChapter: 23, EndChapter: 23},
			want: "Psalm 23",
		},
		{
			name: "cross-chapter",
			ref:  detector.ScriptureRef{Book: "Romans", StartChapter: 8, EndChapter: 9, StartVerse: 38, EndVerse: 1},
			want: "Romans 8:38-9:1",
		},
		{
			name: "single-chapter book verse",
			ref:  detector.ScriptureRef{Book: "3 John", StartChapter: 1, EndChapter: 1, StartVerse: 14, EndVerse: 14},
			want: "3 John 14",
		},
		{
			name: "single-chapter book range",
			ref:  detector.ScriptureRef{Book: "Jude", StartChapter: 1, EndChapter: 1, StartVerse: 4, EndVerse: 6},
			want: "Jude 4-6",
		},
		{
			name: "single-chapter book whole chapter",
			ref:  detector.ScriptureRef{Book: "Philemon", StartChapter: 1, EndChapter: 1, StartVerse: 0, EndVerse: 0},
			want: "Philemon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatESVQuery(tt.ref)
			if got != tt.want {
				t.Errorf("formatESVQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}
