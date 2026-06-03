package lookup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/aoriver716/sword-drill/internal/detector"
)

// esvAPIKey is the ESV API key, set at compile time via:
//
//	-ldflags "-X github.com/aoriver716/sword-drill/internal/lookup.esvAPIKey=YOUR_KEY"
var esvAPIKey string

// ESVKeyAvailable returns true if an ESV API key was compiled in.
func ESVKeyAvailable() bool {
	return esvAPIKey != ""
}

// ESVClient implements BibleLookup using the ESV API (api.esv.org).
type ESVClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewESVClient creates a client for the ESV API.
func NewESVClient() *ESVClient {
	return &ESVClient{
		BaseURL:    "https://api.esv.org/v3",
		HTTPClient: http.DefaultClient,
	}
}

type esvTextResponse struct {
	Query       string            `json:"query"`
	Canonical   string            `json:"canonical"`
	Parsed      [][]int           `json:"parsed"`
	PassageMeta []esvPassageMeta  `json:"passage_meta"`
	Passages    []string          `json:"passages"`
}

type esvPassageMeta struct {
	Canonical    string `json:"canonical"`
	PrevVerse    int    `json:"prev_verse"`
	NextVerse    int    `json:"next_verse"`
}

// verseNumRe matches verse number markers like "[16] " in ESV text output.
var verseNumRe = regexp.MustCompile(`\[(\d+)\]\s`)

// Lookup fetches scripture verses from the ESV API.
// The translation parameter is ignored since ESV only serves one translation.
func (c *ESVClient) Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error) {
	query := formatESVQuery(ref)
	reqURL := fmt.Sprintf("%s/passage/text/?q=%s&include-passage-references=false&include-footnotes=false&include-headings=false&include-short-copyright=true&include-verse-numbers=true",
		c.BaseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return LookupResult{}, fmt.Errorf("esv request build failed: %w", err)
	}
	req.Header.Set("Authorization", "Token "+esvAPIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return LookupResult{}, fmt.Errorf("esv request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return LookupResult{StatusCode: resp.StatusCode}, fmt.Errorf("esv returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp esvTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return LookupResult{}, fmt.Errorf("esv response decode failed: %w", err)
	}

	if len(apiResp.Passages) == 0 {
		return LookupResult{
			Reference:   apiResp.Canonical,
			Translation: "ESV",
			StatusCode:  resp.StatusCode,
		}, nil
	}

	verses := parseESVVerses(apiResp.Passages[0], ref)

	return LookupResult{
		Reference:   apiResp.Canonical,
		Translation: "ESV",
		Verses:      verses,
		SourceURL:   &reqURL,
		StatusCode:  resp.StatusCode,
	}, nil
}

// parseESVVerses splits the ESV text response into individual Verse structs
// by finding [N] markers that denote verse numbers.
func parseESVVerses(passage string, ref detector.ScriptureRef) []Verse {
	// Find all verse number positions
	matches := verseNumRe.FindAllStringSubmatchIndex(passage, -1)
	if len(matches) == 0 {
		// No verse markers found; return entire text as single verse
		text := strings.TrimSpace(passage)
		if text == "" {
			return nil
		}
		return []Verse{{
			Book:    ref.Book,
			Chapter: ref.StartChapter,
			Number:  ref.StartVerse,
			Text:    text,
		}}
	}

	var verses []Verse
	for i, match := range matches {
		verseNum, _ := strconv.Atoi(passage[match[2]:match[3]])

		// Text starts after the "[N] " marker
		textStart := match[1]
		// Text ends at the start of the next marker, or end of passage
		var textEnd int
		if i+1 < len(matches) {
			textEnd = matches[i+1][0]
		} else {
			textEnd = len(passage)
		}

		text := strings.TrimSpace(passage[textStart:textEnd])

		// Determine the chapter for this verse. For single-chapter refs it's
		// straightforward. For cross-chapter refs, detect when verse numbers
		// reset (decrease) to identify a chapter boundary.
		chapter := ref.StartChapter
		if i > 0 && verseNum < verses[len(verses)-1].Number {
			chapter = verses[len(verses)-1].Chapter + 1
		} else if i > 0 {
			chapter = verses[len(verses)-1].Chapter
		}

		verses = append(verses, Verse{
			Book:    ref.Book,
			Chapter: chapter,
			Number:  verseNum,
			Text:    text,
		})
	}

	return verses
}

// formatESVQuery converts a ScriptureRef to an ESV API query string.
// e.g., "John 3:16-18" or "Psalm 23"
func formatESVQuery(ref detector.ScriptureRef) string {
	var b strings.Builder
	b.WriteString(ref.Book)
	b.WriteString(fmt.Sprintf(" %d", ref.StartChapter))

	if ref.StartVerse > 0 {
		b.WriteString(fmt.Sprintf(":%d", ref.StartVerse))

		if ref.EndChapter != ref.StartChapter {
			b.WriteString(fmt.Sprintf("-%d:%d", ref.EndChapter, ref.EndVerse))
		} else if ref.EndVerse > 0 && ref.EndVerse != ref.StartVerse {
			b.WriteString(fmt.Sprintf("-%d", ref.EndVerse))
		}
	}
	return b.String()
}

// Translations returns the single ESV translation.
func (c *ESVClient) Translations() ([]Translation, error) {
	return []Translation{
		{Name: "English Standard Version", Key: "esv"},
	}, nil
}

// RefreshTranslations is a no-op for the ESV API (only one translation available).
func (c *ESVClient) RefreshTranslations() error {
	return nil
}
