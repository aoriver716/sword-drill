package lookup

import (
	"fmt"
	"log"
	"strings"

	"github.com/aoriver716/sword-drill/internal/cache"
	"github.com/aoriver716/sword-drill/internal/detector"
)

// CachedLookup is a decorator over BibleLookup that memoizes per-chapter
// scripture results and translation lists in an on-disk cache.
//
// Cache key format is per-chapter so verse-range and chapter-only requests
// share a single entry per chapter. A verse-range miss "upgrades" to a
// full-chapter fetch before caching, ensuring partial chapters are never
// stored. Multi-chapter ranges (e.g. Rom 8:28-9:5) are split into one
// sub-request per chapter and each chapter is cached independently.
type CachedLookup struct {
	inner  BibleLookup
	cache  *cache.Cache
	source string
}

// chapterPayload is the on-disk representation of a single cached chapter.
type chapterPayload struct {
	Reference   string  `json:"reference"`
	Translation string  `json:"translation"`
	SourceURL   *string `json:"source_url,omitempty"`
	StatusCode  int     `json:"status_code"`
	Verses      []Verse `json:"verses"`
}

// WithCache wraps inner with a cache decorator. If cache is nil, inner is
// returned unchanged.
func WithCache(inner BibleLookup, c *cache.Cache, source string) BibleLookup {
	if c == nil {
		return inner
	}
	return &CachedLookup{inner: inner, cache: c, source: source}
}

// Lookup serves a scripture request from the cache where possible,
// upgrading verse-range misses to full-chapter fetches so future requests
// for overlapping ranges are free.
func (c *CachedLookup) Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error) {
	startCh := ref.StartChapter
	endCh := ref.EndChapter
	if endCh == 0 {
		endCh = startCh
	}

	allVerses := make([]Verse, 0)
	var lastSourceURL *string
	var lastStatus int

	for ch := startCh; ch <= endCh; ch++ {
		payload, err := c.fetchChapter(ref.Book, ch, translation)
		if err != nil {
			return LookupResult{}, err
		}
		if payload.SourceURL != nil {
			lastSourceURL = payload.SourceURL
		}
		if payload.StatusCode != 0 {
			lastStatus = payload.StatusCode
		}
		allVerses = append(allVerses, filterVerses(payload.Verses, ref, ch, startCh, endCh)...)
	}

	return LookupResult{
		Reference:   ref.String(),
		Translation: translation,
		Verses:      allVerses,
		SourceURL:   lastSourceURL,
		StatusCode:  lastStatus,
	}, nil
}

// fetchChapter returns the cached chapter payload, fetching and caching it
// on miss. The fetched payload always represents the whole chapter,
// regardless of the original request's verse range.
func (c *CachedLookup) fetchChapter(book string, chapter int, translation string) (chapterPayload, error) {
	key := chapterKey(c.source, translation, book, chapter)
	var payload chapterPayload
	hit, _ := c.cache.Get(key, &payload)
	if hit {
		log.Printf("cache HIT  %s", key)
		// A cache hit yields a chapter with no API origin to report.
		payload.SourceURL = nil
		payload.StatusCode = 0
		return payload, nil
	}
	log.Printf("cache MISS %s", key)

	chapterRef := detector.ScriptureRef{
		Book:         book,
		StartChapter: chapter,
		EndChapter:   chapter,
	}
	result, err := c.inner.Lookup(chapterRef, translation)
	if err != nil {
		return chapterPayload{}, err
	}
	payload = chapterPayload{
		Reference:   result.Reference,
		Translation: result.Translation,
		SourceURL:   result.SourceURL,
		StatusCode:  result.StatusCode,
		Verses:      result.Verses,
	}
	_ = c.cache.Put(key, payload)
	return payload, nil
}

// filterVerses keeps only the verses from a chapter payload that fall
// inside the requested range.
func filterVerses(verses []Verse, ref detector.ScriptureRef, currentCh, startCh, endCh int) []Verse {
	// Chapter-only request: no filtering.
	if ref.StartVerse == 0 {
		return verses
	}
	out := make([]Verse, 0, len(verses))
	for _, v := range verses {
		switch {
		case currentCh == startCh && currentCh == endCh:
			if v.Number < ref.StartVerse {
				continue
			}
			if ref.EndVerse != 0 && v.Number > ref.EndVerse {
				continue
			}
		case currentCh == startCh:
			if v.Number < ref.StartVerse {
				continue
			}
		case currentCh == endCh:
			if ref.EndVerse != 0 && v.Number > ref.EndVerse {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

// Translations returns the cached translation list, falling through to the
// inner provider on miss and caching the result on success.
func (c *CachedLookup) Translations() ([]Translation, error) {
	key := translationsKey(c.source)
	var cached []Translation
	if hit, _ := c.cache.Get(key, &cached); hit {
		log.Printf("cache HIT  %s", key)
		return cached, nil
	}
	log.Printf("cache MISS %s", key)
	translations, err := c.inner.Translations()
	if err != nil {
		return nil, err
	}
	_ = c.cache.Put(key, translations)
	return translations, nil
}

// RefreshTranslations invalidates the cached translation list so the next
// call to Translations refetches from the inner provider. This is not part
// of the BibleLookup interface; call it via a type assertion.
func (c *CachedLookup) RefreshTranslations() error {
	key := translationsKey(c.source)
	log.Printf("cache INVALIDATE %s", key)
	return c.cache.Delete(key)
}

func chapterKey(source, translation, book string, chapter int) string {
	return fmt.Sprintf("v1|%s|%s|%s|%d", source, translation, strings.ToLower(book), chapter)
}

func translationsKey(source string) string {
	return fmt.Sprintf("v1|translations|%s", source)
}
