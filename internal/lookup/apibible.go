package lookup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aoriver716/sword-drill/internal/detector"
)

// apiKey is the API.Bible key, set at compile time via:
//
//	-ldflags "-X github.com/aoriver716/sword-drill/internal/lookup.apiKey=YOUR_KEY"
var apiKey string

// APIKeyAvailable returns true if an API.Bible key was compiled in.
func APIKeyAvailable() bool {
	return apiKey != ""
}

// APIBibleClient implements BibleLookup using API.Bible (rest.api.bible).
type APIBibleClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAPIBibleClient creates a client for API.Bible.
func NewAPIBibleClient() *APIBibleClient {
	return &APIBibleClient{
		BaseURL:    "https://rest.api.bible/v1",
		HTTPClient: http.DefaultClient,
	}
}

type apiBibleResponse struct {
	Data apiBibleData `json:"data"`
}

type apiBibleData struct {
	ID         string          `json:"id"`
	Reference  string          `json:"reference"`
	Content    json.RawMessage `json:"content"`
	VerseCount int             `json:"verseCount"`
	Copyright  string          `json:"copyright"`
}

// Lookup fetches scripture verses from API.Bible.
// The translation parameter is the API.Bible bible ID (e.g. "de4e12af7f28f599-02" for KJV).
func (c *APIBibleClient) Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error) {
	passageID := formatPassageID(ref)
	reqURL := fmt.Sprintf("%s/bibles/%s/passages/%s?content-type=json&include-titles=false&include-verse-numbers=false&include-verse-spans=false",
		c.BaseURL, translation, passageID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return LookupResult{}, fmt.Errorf("api.bible request build failed: %w", err)
	}
	req.Header.Set("api-key", apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return LookupResult{}, fmt.Errorf("api.bible request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return LookupResult{StatusCode: resp.StatusCode}, fmt.Errorf("api.bible returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiBibleResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return LookupResult{}, fmt.Errorf("api.bible response decode failed: %w", err)
	}

	verses := parseVersesFromJSON(apiResp.Data.Content, ref.Book)

	return LookupResult{
		Reference:   apiResp.Data.Reference,
		Translation: translation,
		Verses:      verses,
		SourceURL:   &reqURL,
		StatusCode:  resp.StatusCode,
	}, nil
}

// jsonContentNode represents a node in the API.Bible JSON content tree.
type jsonContentNode struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text"`
	Attrs map[string]interface{} `json:"attrs"`
	Items []jsonContentNode      `json:"items"`
}

// parseVersesFromJSON extracts verses from API.Bible JSON content.
// The content is a tree of nodes; text nodes have a verseId attr like "ROM.8.28".
func parseVersesFromJSON(content json.RawMessage, book string) []Verse {
	var root jsonContentNode
	if err := json.Unmarshal(content, &root); err != nil {
		// Try as array of nodes
		var roots []jsonContentNode
		if err2 := json.Unmarshal(content, &roots); err2 != nil {
			return nil
		}
		// Collect from all root nodes
		textByVerse := map[string]*strings.Builder{}
		var order []string
		for _, node := range roots {
			collectVerseText(&node, textByVerse, &order)
		}
		return buildVersesFromMap(textByVerse, order, book)
	}

	textByVerse := map[string]*strings.Builder{}
	var order []string
	collectVerseText(&root, textByVerse, &order)
	return buildVersesFromMap(textByVerse, order, book)
}

// collectVerseText recursively walks the JSON content tree and groups text by verseId.
func collectVerseText(node *jsonContentNode, textByVerse map[string]*strings.Builder, order *[]string) {
	if node.Type == "text" && node.Text != "" {
		verseID, _ := node.Attrs["verseId"].(string)
		if verseID == "" {
			return
		}
		if _, exists := textByVerse[verseID]; !exists {
			textByVerse[verseID] = &strings.Builder{}
			*order = append(*order, verseID)
		}
		textByVerse[verseID].WriteString(node.Text)
	}
	for i := range node.Items {
		collectVerseText(&node.Items[i], textByVerse, order)
	}
}

// buildVersesFromMap converts the collected verse text map into a sorted Verse slice.
func buildVersesFromMap(textByVerse map[string]*strings.Builder, order []string, book string) []Verse {
	var verses []Verse
	for _, verseID := range order {
		sb := textByVerse[verseID]
		// Parse verseId like "ROM.8.28" → chapter=8, verse=28
		parts := strings.Split(verseID, ".")
		chapter := 0
		num := 0
		if len(parts) >= 2 {
			chapter, _ = strconv.Atoi(parts[1])
		}
		if len(parts) >= 3 {
			num, _ = strconv.Atoi(parts[2])
		}
		verses = append(verses, Verse{
			Book:    book,
			Chapter: chapter,
			Number:  num,
			Text:    strings.TrimSpace(sb.String()),
		})
	}
	return verses
}

// formatPassageID converts a ScriptureRef to an API.Bible passage ID.
// e.g., "John 3:16-18" → "JHN.3.16-JHN.3.18"
func formatPassageID(ref detector.ScriptureRef) string {
	bookID := bookNameToID(ref.Book)

	if ref.StartVerse == 0 {
		// Chapter-only reference (e.g. "Psalm 23" → "PSA.23")
		return fmt.Sprintf("%s.%d", bookID, ref.StartChapter)
	}

	start := fmt.Sprintf("%s.%d.%d", bookID, ref.StartChapter, ref.StartVerse)

	if ref.EndVerse == 0 || (ref.EndChapter == ref.StartChapter && ref.EndVerse == ref.StartVerse) {
		// Single verse
		return start
	}

	end := fmt.Sprintf("%s.%d.%d", bookID, ref.EndChapter, ref.EndVerse)
	return start + "-" + end
}

// bookIDs maps canonical book names (lowercase) to API.Bible 3-letter book IDs.
var bookIDs = map[string]string{
	"genesis":         "GEN",
	"exodus":          "EXO",
	"leviticus":       "LEV",
	"numbers":         "NUM",
	"deuteronomy":     "DEU",
	"joshua":          "JOS",
	"judges":          "JDG",
	"ruth":            "RUT",
	"1 samuel":        "1SA",
	"2 samuel":        "2SA",
	"1 kings":         "1KI",
	"2 kings":         "2KI",
	"1 chronicles":    "1CH",
	"2 chronicles":    "2CH",
	"ezra":            "EZR",
	"nehemiah":        "NEH",
	"esther":          "EST",
	"job":             "JOB",
	"psalms":          "PSA",
	"psalm":           "PSA",
	"proverbs":        "PRO",
	"ecclesiastes":    "ECC",
	"song of solomon": "SNG",
	"isaiah":          "ISA",
	"jeremiah":        "JER",
	"lamentations":    "LAM",
	"ezekiel":         "EZK",
	"daniel":          "DAN",
	"hosea":           "HOS",
	"joel":            "JOL",
	"amos":            "AMO",
	"obadiah":         "OBA",
	"jonah":           "JON",
	"micah":           "MIC",
	"nahum":           "NAM",
	"habakkuk":        "HAB",
	"zephaniah":       "ZEP",
	"haggai":          "HAG",
	"zechariah":       "ZEC",
	"malachi":         "MAL",
	"matthew":         "MAT",
	"mark":            "MRK",
	"luke":            "LUK",
	"john":            "JHN",
	"acts":            "ACT",
	"romans":          "ROM",
	"1 corinthians":   "1CO",
	"2 corinthians":   "2CO",
	"galatians":       "GAL",
	"ephesians":       "EPH",
	"philippians":     "PHP",
	"colossians":      "COL",
	"1 thessalonians": "1TH",
	"2 thessalonians": "2TH",
	"1 timothy":       "1TI",
	"2 timothy":       "2TI",
	"titus":           "TIT",
	"philemon":        "PHM",
	"hebrews":         "HEB",
	"james":           "JAS",
	"1 peter":         "1PE",
	"2 peter":         "2PE",
	"1 john":          "1JN",
	"2 john":          "2JN",
	"3 john":          "3JN",
	"jude":            "JUD",
	"revelation":      "REV",
}

func bookNameToID(name string) string {
	if id, ok := bookIDs[strings.ToLower(name)]; ok {
		return id
	}
	// Fallback: use first 3 chars uppercased
	if len(name) >= 3 {
		return strings.ToUpper(name[:3])
	}
	return strings.ToUpper(name)
}

const translationsCacheFile = ".apibible_translations.json"

// Translations returns the available Bible translations from the local cache.
// Returns an error if the cache file does not exist — call RefreshTranslations first.
func (c *APIBibleClient) Translations() ([]Translation, error) {
	data, err := os.ReadFile(translationsCacheFile)
	if err != nil {
		return nil, fmt.Errorf("translations cache not found; call RefreshTranslations to populate it: %w", err)
	}
	var translations []Translation
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse translations cache: %w", err)
	}
	return translations, nil
}

// RefreshTranslations fetches available Bibles from API.Bible and caches them locally.
func (c *APIBibleClient) RefreshTranslations() error {
	reqURL := fmt.Sprintf("%s/bibles?language=eng", c.BaseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("api.bible translations request build failed: %w", err)
	}
	req.Header.Set("api-key", apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("api.bible translations request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api.bible translations returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Data []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation"`
			Description  string `json:"description"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("api.bible translations response decode failed: %w", err)
	}

	// De-duplicate translations that share the same base DBL ID (the part
	// before the hyphen-suffix).  These are canon variants (Protestant,
	// Catholic, Orthodox, Ecumenical) of the same translation text.
	// Keep "Protestant" when available; otherwise keep whichever entry
	// comes first so every translation is represented at least once.
	type candidate struct {
		Name        string
		Key         string
		Description string
	}
	best := make(map[string]candidate) // keyed by dblId (base ID)
	order := []string{}                // preserve insertion order
	for _, b := range apiResp.Data {
		base := b.ID
		if idx := strings.LastIndex(b.ID, "-"); idx != -1 {
			base = b.ID[:idx]
		}
		prev, exists := best[base]
		if !exists {
			order = append(order, base)
			best[base] = candidate{Name: b.Name, Key: b.ID, Description: b.Description}
		} else if prev.Description != "Protestant" && b.Description == "Protestant" {
			best[base] = candidate{Name: b.Name, Key: b.ID, Description: b.Description}
		}
	}

	translations := make([]Translation, 0, len(best))
	for _, base := range order {
		cand := best[base]
		// Only include translations that have all 66 Protestant-canon books.
		if !c.hasFullCanon(cand.Key) {
			continue
		}
		translations = append(translations, Translation{Name: cand.Name, Key: cand.Key})
	}

	data, err := json.MarshalIndent(translations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal translations: %w", err)
	}
	if err := os.WriteFile(translationsCacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write translations cache: %w", err)
	}

	return nil
}

// hasFullCanon returns true if the Bible has at least 66 books (full Protestant canon).
func (c *APIBibleClient) hasFullCanon(bibleID string) bool {
	reqURL := fmt.Sprintf("%s/bibles/%s/books", c.BaseURL, bibleID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("api-key", apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var booksResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&booksResp); err != nil {
		return false
	}
	return len(booksResp.Data) >= 66
}
