package lookup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sword-drill/parser"
)

// BibleAPIClient implements BibleLookup using bible-api.com.
type BibleAPIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewBibleAPIClient creates a client with default settings.
func NewBibleAPIClient() *BibleAPIClient {
	return &BibleAPIClient{
		BaseURL:    "https://bible-api.com",
		HTTPClient: http.DefaultClient,
	}
}

type bibleAPIResponse struct {
	Reference       string          `json:"reference"`
	Verses          []bibleAPIVerse `json:"verses"`
	Text            string          `json:"text"`
	TranslationID   string          `json:"translation_id"`
	TranslationName string          `json:"translation_name"`
}

type bibleAPIVerse struct {
	BookID   string `json:"book_id"`
	BookName string `json:"book_name"`
	Chapter  int    `json:"chapter"`
	Verse    int    `json:"verse"`
	Text     string `json:"text"`
}

// Lookup fetches scripture text from bible-api.com.
func (c *BibleAPIClient) Lookup(ref parser.ScriptureRef, translation string, opts LookupOptions) (ScriptureText, error) {
	query := formatRefForAPI(ref)
	reqURL := fmt.Sprintf("%s/%s?translation=%s", c.BaseURL, url.PathEscape(query), url.QueryEscape(translation))

	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return ScriptureText{}, fmt.Errorf("bible-api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ScriptureText{}, fmt.Errorf("bible-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp bibleAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return ScriptureText{}, fmt.Errorf("bible-api response decode failed: %w", err)
	}

	text := formatVerses(apiResp.Verses, opts)

	return ScriptureText{
		Reference:   apiResp.Reference,
		Text:        text,
		Translation: apiResp.TranslationName,
	}, nil
}

func formatVerses(verses []bibleAPIVerse, opts LookupOptions) string {
	var b strings.Builder
	for i, v := range verses {
		if opts.VerseByVerse && i > 0 {
			b.WriteString("\n")
		}
		if opts.ShowVerseNums {
			b.WriteString(fmt.Sprintf("%d ", v.Verse))
		}
		b.WriteString(v.Text)
	}
	return b.String()
}

func formatRefForAPI(ref parser.ScriptureRef) string {
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
