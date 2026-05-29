package lookup

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aoriver716/sword-drill/internal/detector"
)

func TestFormatPassageID(t *testing.T) {
	tests := []struct {
		name string
		ref  detector.ScriptureRef
		want string
	}{
		{
			name: "single verse",
			ref:  detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 16},
			want: "JHN.3.16",
		},
		{
			name: "verse range",
			ref:  detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 18},
			want: "JHN.3.16-JHN.3.18",
		},
		{
			name: "chapter only",
			ref:  detector.ScriptureRef{Book: "Psalm", StartChapter: 23, EndChapter: 23},
			want: "PSA.23",
		},
		{
			name: "cross-chapter range",
			ref:  detector.ScriptureRef{Book: "Isaiah", StartChapter: 52, EndChapter: 53, StartVerse: 13, EndVerse: 12},
			want: "ISA.52.13-ISA.53.12",
		},
		{
			name: "numbered book",
			ref:  detector.ScriptureRef{Book: "1 Corinthians", StartChapter: 13, EndChapter: 13, StartVerse: 4, EndVerse: 7},
			want: "1CO.13.4-1CO.13.7",
		},
		{
			name: "single verse same start/end",
			ref:  detector.ScriptureRef{Book: "Genesis", StartChapter: 1, EndChapter: 1, StartVerse: 1, EndVerse: 1},
			want: "GEN.1.1",
		},
		{
			name: "single verse no end",
			ref:  detector.ScriptureRef{Book: "Revelation", StartChapter: 21, EndChapter: 21, StartVerse: 4, EndVerse: 0},
			want: "REV.21.4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatPassageID(tc.ref)
			if got != tc.want {
				t.Errorf("formatPassageID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseVersesFromJSON(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		book      string
		wantCount int
		wantFirst int
		wantLast  int
		wantChapter int
	}{
		{
			name: "multi-verse passage",
			content: `{
				"type": "tag", "name": "para",
				"attrs": {"style": "p"},
				"items": [
					{"type": "text", "text": "And we know that all things work together for good.", "attrs": {"verseId": "ROM.8.28"}},
					{"type": "text", "text": "For whom he did foreknow, he also did predestinate.", "attrs": {"verseId": "ROM.8.29"}},
					{"type": "text", "text": "Moreover whom he did predestinate, them he also called.", "attrs": {"verseId": "ROM.8.30"}}
				]
			}`,
			book:        "Romans",
			wantCount:   3,
			wantFirst:   28,
			wantLast:    30,
			wantChapter: 8,
		},
		{
			name: "single verse",
			content: `{
				"type": "tag", "name": "para",
				"attrs": {"style": "p"},
				"items": [
					{"type": "text", "text": "For God so loved the world.", "attrs": {"verseId": "JHN.3.16"}}
				]
			}`,
			book:        "John",
			wantCount:   1,
			wantFirst:   16,
			wantLast:    16,
			wantChapter: 3,
		},
		{
			name: "nested tags concatenate text",
			content: `{
				"type": "tag", "name": "para",
				"attrs": {"style": "p"},
				"items": [
					{"type": "text", "text": "And we know that ", "attrs": {"verseId": "ROM.8.28"}},
					{"type": "tag", "name": "char", "attrs": {"style": "add"}, "items": [
						{"type": "text", "text": "all things", "attrs": {"verseId": "ROM.8.28"}}
					]},
					{"type": "text", "text": " work together.", "attrs": {"verseId": "ROM.8.28"}}
				]
			}`,
			book:        "Romans",
			wantCount:   1,
			wantFirst:   28,
			wantLast:    28,
			wantChapter: 8,
		},
		{
			name:      "empty content",
			content:   `{}`,
			book:      "Genesis",
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verses := parseVersesFromJSON(json.RawMessage(tc.content), tc.book)
			if len(verses) != tc.wantCount {
				for i, v := range verses {
					t.Logf("  verse[%d]: ch=%d num=%d text=%q", i, v.Chapter, v.Number, v.Text)
				}
				t.Fatalf("got %d verses, want %d", len(verses), tc.wantCount)
			}
			if tc.wantCount > 0 {
				if verses[0].Number != tc.wantFirst {
					t.Errorf("first verse number = %d, want %d", verses[0].Number, tc.wantFirst)
				}
				if verses[len(verses)-1].Number != tc.wantLast {
					t.Errorf("last verse number = %d, want %d", verses[len(verses)-1].Number, tc.wantLast)
				}
				if verses[0].Chapter != tc.wantChapter {
					t.Errorf("first verse chapter = %d, want %d", verses[0].Chapter, tc.wantChapter)
				}
			}
		})
	}
}

func TestBookNameToID(t *testing.T) {
	tests := map[string]string{
		"Genesis":         "GEN",
		"Exodus":          "EXO",
		"Psalm":           "PSA",
		"Psalms":          "PSA",
		"Song of Solomon": "SNG",
		"1 Corinthians":   "1CO",
		"Revelation":      "REV",
		"John":            "JHN",
		"Jude":            "JUD",
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			got := bookNameToID(name)
			if got != want {
				t.Errorf("bookNameToID(%q) = %q, want %q", name, got, want)
			}
		})
	}
}

func TestAPIBibleClient(t *testing.T) {
	key := os.Getenv("API_BIBLE_KEY")
	if key == "" {
		t.Skip("API_BIBLE_KEY not set, skipping integration test")
	}
	client := NewAPIBibleClient(key, "de4e12af7f28f599-02")
	RunLookupTests(t, client)
}
