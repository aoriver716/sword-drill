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

// Lookup fetches scripture verses from bible-api.com.
func (c *BibleAPIClient) Lookup(ref parser.ScriptureRef, translation string) (LookupResult, error) {
	query := formatRefForAPI(ref)
	reqURL := fmt.Sprintf("%s/%s?translation=%s", c.BaseURL, url.PathEscape(query), url.QueryEscape(translation))

	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return LookupResult{}, fmt.Errorf("bible-api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return LookupResult{}, fmt.Errorf("bible-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp bibleAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return LookupResult{}, fmt.Errorf("bible-api response decode failed: %w", err)
	}

	verses := make([]Verse, len(apiResp.Verses))
	for i, v := range apiResp.Verses {
		verses[i] = Verse{
			Book:    v.BookName,
			Chapter: v.Chapter,
			Number:  v.Verse,
			Text:    v.Text,
		}
	}

	return LookupResult{
		Reference:   apiResp.Reference,
		Translation: apiResp.TranslationName,
		Verses:      verses,
		SourceURL:   &reqURL,
	}, nil
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
