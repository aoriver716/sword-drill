package formatter

import (
	"fmt"
	"strings"

	"github.com/aoriver716/sword-drill/internal/config"
	"github.com/aoriver716/sword-drill/internal/lookup"
)

// Format renders a LookupResult as a string according to the given config.
func Format(result lookup.LookupResult, cfg *config.Config) string {
	var b strings.Builder
	for i, v := range result.Verses {
		if cfg.FormattingOptions.VerseByVerse && i > 0 {
			b.WriteString("\n")
		}
		if cfg.FormattingOptions.ShowVerseNums {
			b.WriteString(fmt.Sprintf("%d ", v.Number))
		}
		b.WriteString(v.Text)
	}
	return b.String()
}
