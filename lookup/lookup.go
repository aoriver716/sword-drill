package lookup

import (
	"fmt"

	"github.com/sword-drill/parser"
)

// ScriptureText holds the result of a scripture lookup.
type ScriptureText struct {
	Reference   string
	Text        string
	Translation string
}

func (s ScriptureText) String() string {
	return fmt.Sprintf("[%s] %s\n%s", s.Translation, s.Reference, s.Text)
}

// LookupOptions controls how scripture text is returned.
type LookupOptions struct {
	VerseByVerse bool // If true, each verse on its own line; if false, paragraph style
	ShowVerseNums bool // If true, prefix each verse with its number
}

// DefaultOptions returns sensible defaults (paragraph mode, no verse numbers).
func DefaultOptions() LookupOptions {
	return LookupOptions{
		VerseByVerse: false,
		ShowVerseNums: false,
	}
}

// BibleLookup is the interface for retrieving scripture text.
// Implementations may use a remote API, local database, etc.
type BibleLookup interface {
	Lookup(ref parser.ScriptureRef, translation string, opts LookupOptions) (ScriptureText, error)
}
