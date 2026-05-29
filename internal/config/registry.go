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
	Key             string     `json:"key"`
	Label           string     `json:"label"`
	Description     string     `json:"description,omitempty"`
	Group           string     `json:"group"`
	Widget          WidgetType `json:"widget"`
	Value           any        `json:"value"`
	Default         any        `json:"default"`
	Options         []Option   `json:"options,omitempty"`
	RequiresRestart bool       `json:"requiresRestart,omitempty"`
}

// OptionsFunc returns dynamic options for a select/radio field.
type OptionsFunc func() []Option

// FieldDef is the definition of a config field.
type FieldDef struct {
	Key             string
	Label           string
	Description     string
	Group           string
	Widget          WidgetType
	Default         any
	Options         []Option    // static options (nil if dynamic)
	OptionsFunc     OptionsFunc // called at schema time if non-nil
	RequiresRestart func(*Registry) bool // evaluated at runtime; nil means false
	Getter          func(*Config) any
	Setter          func(*Config, any)
}

// requiresRestart returns whether this field currently requires a restart.
func (f *FieldDef) requiresRestart(r *Registry) bool {
	if f.RequiresRestart != nil {
		return f.RequiresRestart(r)
	}
	return false
}

// Registry holds all config field definitions and the current config.
type Registry struct {
	fields    []FieldDef
	providers []LookupProvider
	cfg       Config
	pending   map[string]any // buffered values for RequiresRestart fields
	path      string
	onChange  func(*Config)
}

// NewRegistry creates a registry that loads/saves from the given path.
func NewRegistry(path string) *Registry {
	return &Registry{
		path:    path,
		pending: make(map[string]any),
	}
}

// Pending returns the pending value for a field key, or nil if none.
func (r *Registry) Pending(key string) (any, bool) {
	v, ok := r.pending[key]
	return v, ok
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

// PendingBibleLookup returns the BibleLookup client for the pending API
// provider if one has been selected but not yet applied, otherwise falls
// back to the live config. Used by the UI to populate translation options.
func (r *Registry) PendingBibleLookup() lookup.BibleLookup {
	apiKey := r.cfg.BibleTextAPI
	if v, ok := r.pending["bible_text_api"]; ok {
		if s, ok := v.(string); ok {
			apiKey = s
		}
	}
	for _, p := range r.providers {
		if p.Key == apiKey {
			return p.Factory(&r.cfg)
		}
	}
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
// Clears any pending (buffered) values since a load/restart makes them live.
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

	// Clear pending — restart applies buffered values.
	r.pending = make(map[string]any)
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

// Save writes the current config to disk, merging any pending values so
// the on-disk file always reflects the user's intended state.
func (r *Registry) Save() error {
	// Build a copy of cfg with pending values applied for serialization.
	tmp := r.cfg
	for _, f := range r.fields {
		if v, ok := r.pending[f.Key]; ok {
			f.Setter(&tmp, v)
		}
	}
	data, err := json.MarshalIndent(tmp, "", "  ")
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
// For RequiresRestart fields, the value shown is the pending value if one exists.
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
		// Use pending value for the UI if one is buffered.
		value := f.Getter(&r.cfg)
		if v, ok := r.pending[f.Key]; ok {
			value = v
		}
		schemas[i] = FieldSchema{
			Key:             f.Key,
			Label:           f.Label,
			Description:     f.Description,
			Group:           f.Group,
			Widget:          f.Widget,
			Value:           value,
			Default:         f.Default,
			Options:         opts,
			RequiresRestart: f.requiresRestart(r),
		}
	}
	return schemas
}

// OnChange registers a callback that fires whenever the config is updated.
func (r *Registry) OnChange(fn func(*Config)) {
	r.onChange = fn
}

// notifyChange calls the onChange callback if one is registered.
func (r *Registry) notifyChange() {
	if r.onChange != nil {
		r.onChange(&r.cfg)
	}
}

// Update sets a config field by key and saves to disk.
// For RequiresRestart fields the value is buffered in pending rather than
// applied to the live config. Non-restart fields update the live config
// immediately.
func (r *Registry) Update(key string, value any) error {
	for _, f := range r.fields {
		if f.Key != key {
			continue
		}

		if f.requiresRestart(r) {
			r.pending[key] = value
		} else {
			f.Setter(&r.cfg, value)
		}

		// When the API provider changes, auto-reset the translation
		// to the new provider's default (also buffered in pending).
		if key == "bible_text_api" {
			newAPI, _ := value.(string)
			for _, p := range r.providers {
				if p.Key == newAPI {
					r.pending["default_translation"] = p.DefaultTranslation
					r.pending["parallel_translation"] = p.DefaultTranslation
					break
				}
			}
		}

		if err := r.Save(); err != nil {
			return err
		}
		r.notifyChange()
		return nil
	}
	return nil
}

// ResetToDefaults resets all fields to their default values and saves.
func (r *Registry) ResetToDefaults() error {
	r.cfg = Config{}
	r.pending = make(map[string]any)
	r.applyDefaults()
	r.ensureAvailableProvider()
	if err := r.Save(); err != nil {
		return err
	}
	r.notifyChange()
	return nil
}

// applyDefaults sets all config fields to their registered default values.
func (r *Registry) applyDefaults() {
	for _, f := range r.fields {
		f.Setter(&r.cfg, f.Default)
	}
}
