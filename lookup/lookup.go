package lookup

import (
	"github.com/sword-drill/parser"
)

// Verse represents a single verse returned from a lookup.
type Verse struct {
	Book    string
	Chapter int
	Number  int
	Text    string
}

// LookupResult holds the result of a scripture lookup.
type LookupResult struct {
	Reference   string
	Translation string
	Verses      []Verse
	SourceURL   *string // nil for local sources, set for remote APIs
}

// BibleLookup is the interface for retrieving scripture text.
// Implementations may use a remote API, local database, etc.
type BibleLookup interface {
	Lookup(ref parser.ScriptureRef, translation string) (LookupResult, error)
}
