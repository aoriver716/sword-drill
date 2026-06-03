package lookup

import (
	"testing"

	"github.com/aoriver716/sword-drill/internal/detector"
)

// mockClient is a minimal BibleLookup for testing MultiClient dispatch.
type mockClient struct {
	translationKey string
	called         bool
}

func (m *mockClient) Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error) {
	m.called = true
	m.translationKey = translation
	return LookupResult{
		Reference:   "Mock 1:1",
		Translation: translation,
		Verses:      []Verse{{Book: ref.Book, Chapter: 1, Number: 1, Text: "mock text"}},
	}, nil
}

func (m *mockClient) Translations() ([]Translation, error) {
	return []Translation{
		{Name: "Mock Translation", Key: "mock-t1"},
		{Name: "Mock Translation 2", Key: "mock-t2"},
	}, nil
}

func (m *mockClient) RefreshTranslations() error { return nil }

func TestMultiClient_Translations(t *testing.T) {
	mc := NewMultiClient()
	mc.AddProvider("provider-a", "Provider A", &mockClient{})
	mc.AddProvider("provider-b", "Provider B", &mockClient{})

	translations, err := mc.Translations()
	if err != nil {
		t.Fatalf("Translations() error: %v", err)
	}

	// Expect: header-a, t1, t2, header-b, t1, t2
	if len(translations) != 6 {
		for i, tr := range translations {
			t.Logf("  [%d] Name=%q Key=%q IsGroup=%v", i, tr.Name, tr.Key, tr.IsGroup)
		}
		t.Fatalf("Expected 6 entries, got %d", len(translations))
	}

	// First entry is group header
	if !translations[0].IsGroup || translations[0].Name != "Provider A" {
		t.Errorf("Expected group header 'Provider A', got %+v", translations[0])
	}
	if translations[0].Key != "" {
		t.Errorf("Group header should have empty key, got %q", translations[0].Key)
	}

	// Second entry is a prefixed translation
	if translations[1].Key != "provider-a/mock-t1" {
		t.Errorf("Expected prefixed key 'provider-a/mock-t1', got %q", translations[1].Key)
	}
}

func TestMultiClient_Lookup_Dispatch(t *testing.T) {
	clientA := &mockClient{}
	clientB := &mockClient{}

	mc := NewMultiClient()
	mc.AddProvider("alpha", "Alpha", clientA)
	mc.AddProvider("beta", "Beta", clientB)

	ref := detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 16}

	// Lookup with provider-b prefix
	_, err := mc.Lookup(ref, "beta/some-translation")
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}

	if clientA.called {
		t.Error("Provider A should not have been called")
	}
	if !clientB.called {
		t.Error("Provider B should have been called")
	}
	if clientB.translationKey != "some-translation" {
		t.Errorf("Expected inner key 'some-translation', got %q", clientB.translationKey)
	}
}

func TestMultiClient_Lookup_UnknownProvider(t *testing.T) {
	mc := NewMultiClient()
	mc.AddProvider("alpha", "Alpha", &mockClient{})

	ref := detector.ScriptureRef{Book: "John", StartChapter: 3, EndChapter: 3, StartVerse: 16, EndVerse: 16}

	_, err := mc.Lookup(ref, "unknown/key")
	if err == nil {
		t.Fatal("Expected error for unknown provider")
	}
}

func TestSplitTranslationKey(t *testing.T) {
	tests := []struct {
		input      string
		wantProv   string
		wantKey    string
	}{
		{"api.bible/de4e12af7f28f599-02", "api.bible", "de4e12af7f28f599-02"},
		{"esv/esv", "esv", "esv"},
		{"bible-api.com/kjv", "bible-api.com", "kjv"},
		{"legacy-key-no-slash", "", "legacy-key-no-slash"},
	}

	for _, tt := range tests {
		prov, key := splitTranslationKey(tt.input)
		if prov != tt.wantProv || key != tt.wantKey {
			t.Errorf("splitTranslationKey(%q) = (%q, %q), want (%q, %q)",
				tt.input, prov, key, tt.wantProv, tt.wantKey)
		}
	}
}
