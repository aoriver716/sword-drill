package config

import (
	"fmt"
	"strings"

	"github.com/aoriver716/sword-drill/internal/lookup"
	"github.com/aoriver716/sword-drill/internal/updater"
)

// defaultUpdateChannel returns the update channel a fresh install should
// default to. Nightly builds default to the nightly channel so they don't
// immediately prompt the user to downgrade to stable; everything else
// defaults to stable.
func defaultUpdateChannel() string {
	if strings.HasPrefix(updater.Version, "nightly-") {
		return updater.ChannelNightly
	}
	return updater.ChannelStable
}

// RegisterFields registers all lookup providers and config fields on the given registry.
func RegisterFields(r *Registry) {
	// -- Lookup providers --
	r.RegisterProvider(LookupProvider{
		Key:                "api.bible",
		Label:              "API.Bible",
		DefaultTranslation: "de4e12af7f28f599-02",
		Factory: func(cfg *Config) lookup.BibleLookup {
			return lookup.NewAPIBibleClient()
		},
		Available: lookup.APIKeyAvailable,
	})

	r.RegisterProvider(LookupProvider{
		Key:                "bible-api.com",
		Label:              "bible-api.com",
		DefaultTranslation: "kjv",
		Factory: func(cfg *Config) lookup.BibleLookup {
			return lookup.NewBibleAPIClient()
		},
		Available: func() bool { return false }, // disabled; kept for future use
	})

	r.RegisterProvider(LookupProvider{
		Key:                "esv",
		Label:              "ESV (api.esv.org)",
		DefaultTranslation: "esv",
		Factory: func(cfg *Config) lookup.BibleLookup {
			return lookup.NewESVClient()
		},
		Available: lookup.ESVKeyAvailable,
	})

	// -- Config fields --
	r.Register(FieldDef{
		Key: "bible_text_api", Label: "Bible API", Group: "API",
		Widget: WidgetSelect, Default: "api.bible",
		Hidden: true, // hidden from UI; translations now span all providers
		RequiresRestart: func(*Registry) bool { return true },
		Getter:          func(c *Config) any { return c.BibleTextAPI },
		Setter:          func(c *Config, v any) { c.BibleTextAPI, _ = v.(string) },
	})

	r.Register(FieldDef{
		Key: "default_translation", Label: "Default Translation", Group: "API",
		Widget: WidgetSelect, Default: "api.bible/de4e12af7f28f599-02",
		OptionsFunc: func() []Option {
			multi := r.MultiLookup()
			translations, err := multi.Translations()
			if err != nil {
				return nil
			}
			opts := make([]Option, 0, len(translations))
			for _, t := range translations {
				opts = append(opts, Option{
					Label:   t.Name,
					Value:   t.Key,
					IsGroup: t.IsGroup,
				})
			}
			return opts
		},
		Getter: func(c *Config) any { return c.DefaultTranslation },
		Setter: func(c *Config, v any) { c.DefaultTranslation, _ = v.(string) },
	})

	r.Register(FieldDef{
		Key: "parallel_translation", Label: "Preferred Parallel Translation", Group: "API",
		Widget: WidgetSelect, Default: "esv/esv",
		OptionsFunc: func() []Option {
			multi := r.MultiLookup()
			translations, err := multi.Translations()
			if err != nil {
				return nil
			}
			opts := make([]Option, 0, len(translations))
			for _, t := range translations {
				opts = append(opts, Option{
					Label:   t.Name,
					Value:   t.Key,
					IsGroup: t.IsGroup,
				})
			}
			return opts
		},
		Getter: func(c *Config) any { return c.ParallelTranslation },
		Setter: func(c *Config, v any) { c.ParallelTranslation, _ = v.(string) },
	})

	r.Register(FieldDef{
		Key: "formatting_options.verse_by_verse", Label: "Verse-by-Verse", Group: "Formatting",
		Description: "Display each verse on its own line",
		Widget:      WidgetToggle, Default: true,
		Getter: func(c *Config) any { return c.FormattingOptions.VerseByVerse },
		Setter: func(c *Config, v any) { c.FormattingOptions.VerseByVerse, _ = v.(bool) },
	})

	r.Register(FieldDef{
		Key: "formatting_options.show_verse_nums", Label: "Show Verse Numbers", Group: "Formatting",
		Description: "Prefix each verse with its number",
		Widget:      WidgetToggle, Default: true,
		Getter: func(c *Config) any { return c.FormattingOptions.ShowVerseNums },
		Setter: func(c *Config, v any) { c.FormattingOptions.ShowVerseNums, _ = v.(bool) },
	})

	r.Register(FieldDef{
		Key: "tab_open_behavior", Label: "Scripture Tab Behavior", Group: "Browser",
		Description: "How tabs open when scripture is detected from the clipboard",
		Widget:      WidgetSelect, Default: "focus",
		Options: []Option{
			{Label: "Focus existing tab if possible", Value: "focus"},
			{Label: "Always open new tab", Value: "always_new"},
			{Label: "Never open a tab", Value: "never"},
		},
		Getter: func(c *Config) any { return c.TabOpenBehavior },
		Setter: func(c *Config, v any) { c.TabOpenBehavior, _ = v.(string) },
	})

	r.Register(FieldDef{
		Key: "check_for_updates", Label: "Check for Updates on Startup", Group: "General",
		Description: "Automatically check for new versions when the app starts",
		Widget:      WidgetToggle, Default: true,
		Getter: func(c *Config) any { return c.CheckForUpdates },
		Setter: func(c *Config, v any) { c.CheckForUpdates, _ = v.(bool) },
	})

	r.Register(FieldDef{
		Key: "update_channel", Label: "Update Channel", Group: "General",
		Description: "Which release channel to check for updates (Stable or Nightly)",
		Widget:      WidgetSelect, Default: defaultUpdateChannel(),
		Options: []Option{
			{Label: "Stable", Value: "stable"},
			{Label: "Nightly", Value: "nightly"},
		},
		Getter: func(c *Config) any { return c.UpdateChannel },
		Setter: func(c *Config, v any) { c.UpdateChannel, _ = v.(string) },
	})

	r.Register(FieldDef{
		Key: "clear_scripture_cache", Label: "Clear Scripture Cache", Group: "General",
		Description: "Delete all cached scripture lookups. Next requests will hit the API.",
		Widget:      WidgetButton,
		Action: func() (string, error) {
			c := r.Cache()
			if c == nil {
				return "No cache configured", nil
			}
			stats, _ := c.Stats()
			if err := c.Clear(); err != nil {
				return "", err
			}
			if stats.Entries == 0 {
				return "Cache already empty", nil
			}
			return fmt.Sprintf("Cleared %d entries", stats.Entries), nil
		},
	})

	r.Register(FieldDef{
		Key:    "skipped_version",
		Hidden: true,
		Getter: func(c *Config) any { return c.SkippedVersion },
		Setter: func(c *Config, v any) { c.SkippedVersion, _ = v.(string) },
	})
}
