package main

import (
	"testing"
)

func TestParseReferences(t *testing.T) {
	tests := []struct {
		input string
		want  []ScriptureRef
	}{
		// Full book name with chapter:verse
		{"Genesis 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		// Numbered book with verse range
		{"1 Chronicles 15:10-13", []ScriptureRef{
			{Book: "1 Chronicles", StartChapter: 15, StartVerse: 10, EndChapter: 15, EndVerse: 13},
		}},
		// Standard abbreviation with dot
		{"Gen. 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		// Informal abbreviation
		{"Gen 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		// Verse range
		{"John 3:16-18", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		// Chapter-only
		{"Psalm 23", []ScriptureRef{
			{Book: "Psalms", StartChapter: 23, StartVerse: 0, EndChapter: 23, EndVerse: 0},
		}},
		// Multi-chapter range
		{"Isaiah 52:13-53:12", []ScriptureRef{
			{Book: "Isaiah", StartChapter: 52, StartVerse: 13, EndChapter: 53, EndVerse: 12},
		}},
		// Multiple references separated by semicolon
		{"Rom. 8:28; John 3:16", []ScriptureRef{
			{Book: "Romans", StartChapter: 8, StartVerse: 28, EndChapter: 8, EndVerse: 28},
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 16},
		}},
		// No reference
		{"Hello world", nil},
		// Reference embedded in prose
		{"As we see in Rev. 21:4, there will be no more tears.", []ScriptureRef{
			{Book: "Revelation", StartChapter: 21, StartVerse: 4, EndChapter: 21, EndVerse: 4},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseReferences(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseReferences(%q) returned %d refs, want %d\n  got: %+v", tt.input, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ref[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
