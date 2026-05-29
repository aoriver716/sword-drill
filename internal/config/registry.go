package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aoriver716/sword-drill/internal/lookup"
)

// LookupProvider describes a Bible lookup API that can be registered with the config system.
type LookupProvider struct {
	Key                string                               // config value (e.g. "api.bible")
	Label              string                               // display name for UI
	DefaultTranslation string                               // default translation key for this provider
	Factory            func(cfg *Config) lookup.BibleLookup // creates a client for this provider
	Available          func() bool                          // returns true if this provider can be used
}

// WidgetType defines the GUI control used for a config field.
type WidgetType string

const (
	WidgetToggle WidgetType = "toggle"
	WidgetSelect WidgetType = "select"
	WidgetText   WidgetType = "text"
	WidgetNumber WidgetType = "number"
)

// Option represents a selectable value for select/radio widgets.
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// FieldSchema describes a single config field for the preferences UI.
type FieldSchema struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Description string     `json:"description,omitempty"`
	Group       string     `json:"group"`
	Widget      WidgetType `json:"widget"`
	Value       any        `json:"value"`
	Default     any        `json:"default"`
	Options     []Option   `json:"options,omitempty"`
}

// OptionsFunc returns dynamic options for a select/radio field.
type OptionsFunc func() []Option

// FieldDef is the definition of a config field.
type FieldDef struct {
	Key         string
	Label       string
	Description string
	Group       string
	Widget      WidgetType
	Default     any
	Options     []Option    // static options (nil if dynamic)
	OptionsFunc OptionsFunc // called at schema time if non-nil
	Getter      func(*Config) any
	Setter      func(*Config, any)
}

// Registry holds all config field definitions and the current config.
type Registry struct {
	fields    []FieldDef
	providers []LookupProvider
	cfg       Config
	path      string
}

// NewRegistry creates a registry that loads/saves from the given path.
func NewRegistry(path string) *Registry {
	return &Registry{
		path: path,
	}
}

// RegisterProvider adds a lookup provider to the registry.
func (r *Registry) RegisterProvider(p LookupProvider) {
	r.providers = append(r.providers, p)
}

// BibleLookup returns the BibleLookup client for the currently configured API.
func (r *Registry) BibleLookup() lookup.BibleLookup {
	for _, p := range r.providers {
		if p.Key == r.cfg.BibleTextAPI {
			return p.Factory(&r.cfg)
		}
	}
	// Fallback to first available provider
	for _, p := range r.providers {
		if p.Available() {
			return p.Factory(&r.cfg)
		}
	}
	return nil
}

// providerOptions returns Option entries for all available providers.
func (r *Registry) providerOptions() []Option {
	var opts []Option
	for _, p := range r.providers {
		if p.Available() {
			opts = append(opts, Option{Label: p.Label, Value: p.Key})
		}
	}
	return opts
}

// firstAvailableProvider returns the key of the first available provider, or "".
func (r *Registry) firstAvailableProvider() string {
	for _, p := range r.providers {
		if p.Available() {
			return p.Key
		}
	}
	return ""
}

// Register adds a field definition to the registry.
func (r *Registry) Register(f FieldDef) {
	r.fields = append(r.fields, f)
}

// Load reads the config from disk. If the file doesn't exist, generates it from defaults.
func (r *Registry) Load() error {
	// Start from field defaults
	r.applyDefaults()

	// If configured provider is unavailable, fall back
	r.ensureAvailableProvider()

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return r.Save()
		}
		return err
	}

	if err := json.Unmarshal(data, &r.cfg); err != nil {
		return err
	}

	// Re-check after loading from disk
	if r.ensureAvailableProvider() {
		_ = r.Save()
	}
	return nil
}

// ensureAvailableProvider checks if the configured provider is available.
// If not, falls back to the first available one. Returns true if a fallback occurred.
func (r *Registry) ensureAvailableProvider() bool {
	for _, p := range r.providers {
		if p.Key == r.cfg.BibleTextAPI && p.Available() {
			// Provider is available; ensure translation is compatible
			if r.cfg.DefaultTranslation == "" {
				r.cfg.DefaultTranslation = p.DefaultTranslation
				return true
			}
			return false
		}
	}
	for _, p := range r.providers {
		if p.Available() {
			log.Printf("%s not available; falling back to %s", r.cfg.BibleTextAPI, p.Key)
			r.cfg.BibleTextAPI = p.Key
			r.cfg.DefaultTranslation = p.DefaultTranslation
			return true
		}
	}
	return false
}

// Save writes the current config to disk.
func (r *Registry) Save() error {
	data, err := json.MarshalIndent(r.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0644)
}

// Config returns a pointer to the current config.
func (r *Registry) Config() *Config {
	return &r.cfg
}

// Schema returns all field schemas with current values and resolved options.
func (r *Registry) Schema() []FieldSchema {
	schemas := make([]FieldSchema, len(r.fields))
	for i, f := range r.fields {
		opts := f.Options
		if f.OptionsFunc != nil {
			opts = f.OptionsFunc()
		}
		// Filter options to only available providers for bible_text_api
		if f.Key == "bible_text_api" {
			opts = r.providerOptions()
		}
		schemas[i] = FieldSchema{
			Key:         f.Key,
			Label:       f.Label,
			Description: f.Description,
			Group:       f.Group,
			Widget:      f.Widget,
			Value:       f.Getter(&r.cfg),
			Default:     f.Default,
			Options:     opts,
		}
	}
	return schemas
}

// Update sets a config field by key and saves to disk.
func (r *Registry) Update(key string, value any) error {
	for _, f := range r.fields {
		if f.Key == key {
			f.Setter(&r.cfg, value)
			return r.Save()
		}
	}
	return nil
}

// ResetToDefaults resets all fields to their default values and saves.
func (r *Registry) ResetToDefaults() error {
	r.cfg = Config{}
	r.applyDefaults()
	r.ensureAvailableProvider()
	return r.Save()
}

// applyDefaults sets all config fields to their registered default values.
func (r *Registry) applyDefaults() {
	for _, f := range r.fields {
		f.Setter(&r.cfg, f.Default)
	}
}
