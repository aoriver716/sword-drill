package formatter

import (
	"fmt"
	"strings"

	"github.com/aoriver716/sword-drill/internal/lookup"
)

// Options controls how scripture text is formatted.
type Options struct {
	VerseByVerse  bool // If true, each verse on its own line; if false, paragraph style
	ShowVerseNums bool // If true, prefix each verse with its number
}

// DefaultOptions returns sensible defaults (paragraph mode, no verse numbers).
func DefaultOptions() Options {
	return Options{
		VerseByVerse:  false,
		ShowVerseNums: false,
	}
}

// Format renders a LookupResult as a string according to the given options.
func Format(result lookup.LookupResult, opts Options) string {
	var b strings.Builder
	for i, v := range result.Verses {
		if opts.VerseByVerse && i > 0 {
			b.WriteString("\n")
		}
		if opts.ShowVerseNums {
			b.WriteString(fmt.Sprintf("%d ", v.Number))
		}
		b.WriteString(v.Text)
	}
	return b.String()
}
