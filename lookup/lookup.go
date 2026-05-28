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

// BibleLookup is the interface for retrieving scripture text.
// Implementations may use a remote API, local database, etc.
type BibleLookup interface {
	Lookup(ref parser.ScriptureRef, translation string) (ScriptureText, error)
}
