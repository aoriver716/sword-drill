package detector

import (
	"testing"
)

func TestParseReferences(t *testing.T) {
	tests := []struct {
		input string
		want  []ScriptureRef
	}{
		{"Genesis 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"1 Chronicles 15:10-13", []ScriptureRef{
			{Book: "1 Chronicles", StartChapter: 15, StartVerse: 10, EndChapter: 15, EndVerse: 13},
		}},
		{"Gen. 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"Gen 1:1", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"John 3:16-18", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		{"Psalm 23", []ScriptureRef{
			{Book: "Psalms", StartChapter: 23, StartVerse: 0, EndChapter: 23, EndVerse: 0},
		}},
		{"Isaiah 52:13-53:12", []ScriptureRef{
			{Book: "Isaiah", StartChapter: 52, StartVerse: 13, EndChapter: 53, EndVerse: 12},
		}},
		{"Rom. 8:28; John 3:16", []ScriptureRef{
			{Book: "Romans", StartChapter: 8, StartVerse: 28, EndChapter: 8, EndVerse: 28},
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 16},
		}},
		{"Hello world", nil},
		{"As we see in Rev. 21:4, there will be no more tears.", []ScriptureRef{
			{Book: "Revelation", StartChapter: 21, StartVerse: 4, EndChapter: 21, EndVerse: 4},
		}},
		{"John 3:16-18\n", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		{"John 3:16-18\r\n", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		{"  Genesis 1:1  ", []ScriptureRef{
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"John 3:16\u201318", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		{"John 3:16\u201418", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 18},
		}},
		{"John 3:16\nGenesis 1:1", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 16},
			{Book: "Genesis", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"1Cor 13:4-7", []ScriptureRef{
			{Book: "1 Corinthians", StartChapter: 13, StartVerse: 4, EndChapter: 13, EndVerse: 7},
		}},
		{"JOHN 3:16", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 16},
		}},
		{"jOhN 3:16", []ScriptureRef{
			{Book: "John", StartChapter: 3, StartVerse: 16, EndChapter: 3, EndVerse: 16},
		}},
		// Single-chapter books: bare number is a verse, not a chapter
		{"3 John 14", []ScriptureRef{
			{Book: "3 John", StartChapter: 1, StartVerse: 14, EndChapter: 1, EndVerse: 14},
		}},
		{"Jude 4-6", []ScriptureRef{
			{Book: "Jude", StartChapter: 1, StartVerse: 4, EndChapter: 1, EndVerse: 6},
		}},
		{"Obadiah 1", []ScriptureRef{
			{Book: "Obadiah", StartChapter: 1, StartVerse: 1, EndChapter: 1, EndVerse: 1},
		}},
		{"Philemon 9", []ScriptureRef{
			{Book: "Philemon", StartChapter: 1, StartVerse: 9, EndChapter: 1, EndVerse: 9},
		}},
		{"2 John 7-11", []ScriptureRef{
			{Book: "2 John", StartChapter: 1, StartVerse: 7, EndChapter: 1, EndVerse: 11},
		}},
		// Single-chapter book with explicit chapter:verse still works
		{"3 John 1:14", []ScriptureRef{
			{Book: "3 John", StartChapter: 1, StartVerse: 14, EndChapter: 1, EndVerse: 14},
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
