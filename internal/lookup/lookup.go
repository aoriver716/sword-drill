package lookup

import (
	"github.com/aoriver716/sword-drill/internal/detector"
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
	StatusCode  int     // HTTP status code (0 for local sources)
}

// Translation represents an available Bible translation.
type Translation struct {
	Name string // display name shown to the user
	Key  string // key used when querying the API
}

// BibleLookup is the interface for retrieving scripture text.
// Implementations may use a remote API, local database, etc.
type BibleLookup interface {
	Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error)
	Translations() ([]Translation, error)
	RefreshTranslations() error
}
