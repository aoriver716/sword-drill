package lookup

import (
	"fmt"
	"strings"

	"github.com/aoriver716/sword-drill/internal/detector"
)

// MultiClient aggregates multiple BibleLookup providers and dispatches
// lookups to the correct provider based on a prefixed translation key.
// Translation keys are formatted as "providerKey/translationKey".
type MultiClient struct {
	providers map[string]BibleLookup
	order     []string // provider keys in registration order
	labels    map[string]string
	cached    []Translation // in-memory cache for the current session
}

// NewMultiClient creates a MultiClient from a set of named providers.
func NewMultiClient() *MultiClient {
	return &MultiClient{
		providers: make(map[string]BibleLookup),
		labels:    make(map[string]string),
	}
}

// AddProvider registers a BibleLookup provider under the given key.
func (m *MultiClient) AddProvider(key, label string, client BibleLookup) {
	m.providers[key] = client
	m.labels[key] = label
	m.order = append(m.order, key)
}

// Lookup dispatches to the correct provider based on the translation key prefix.
// The translation parameter should be in the form "providerKey/translationKey".
func (m *MultiClient) Lookup(ref detector.ScriptureRef, translation string) (LookupResult, error) {
	providerKey, innerKey := splitTranslationKey(translation)
	client, ok := m.providers[providerKey]
	if !ok {
		return LookupResult{}, fmt.Errorf("unknown provider %q in translation key %q", providerKey, translation)
	}
	return client.Lookup(ref, innerKey)
}

// Translations returns all translations from all providers, with each group
// prefixed by a header entry. Header entries have an empty Value and a bold
// label (prefixed with "—") to indicate they are non-selectable group headers.
// Results are cached in memory for the session lifetime.
func (m *MultiClient) Translations() ([]Translation, error) {
	if m.cached != nil {
		return m.cached, nil
	}
	var all []Translation
	for _, provKey := range m.order {
		client := m.providers[provKey]
		translations, err := client.Translations()
		if err != nil {
			continue
		}
		if len(translations) == 0 {
			continue
		}
		// Add group header
		all = append(all, Translation{
			Name:    m.labels[provKey],
			Key:     "", // empty key signals non-selectable header
			IsGroup: true,
		})
		// Add translations with prefixed keys
		for _, t := range translations {
			all = append(all, Translation{
				Name: t.Name,
				Key:  provKey + "/" + t.Key,
			})
		}
	}
	m.cached = all
	return all, nil
}

// RefreshTranslations clears the in-memory cache and refreshes all
// providers' translation lists.
func (m *MultiClient) RefreshTranslations() error {
	m.cached = nil
	for _, client := range m.providers {
		if r, ok := client.(interface{ RefreshTranslations() error }); ok {
			_ = r.RefreshTranslations()
		}
	}
	return nil
}

// splitTranslationKey splits a "provider/key" translation into its parts.
// If no "/" is found, the whole string is treated as the inner key with an
// empty provider (for backwards compatibility).
func splitTranslationKey(compositeKey string) (provider, key string) {
	if idx := strings.Index(compositeKey, "/"); idx >= 0 {
		return compositeKey[:idx], compositeKey[idx+1:]
	}
	return "", compositeKey
}

// ProviderForTranslation returns the provider key for a composite translation key.
func ProviderForTranslation(compositeKey string) string {
	provider, _ := splitTranslationKey(compositeKey)
	return provider
}
