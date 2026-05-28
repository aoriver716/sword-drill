package config

import (
	"encoding/json"
	"os"
)

// FormattingOptions controls how scripture text is displayed.
type FormattingOptions struct {
	VerseByVerse  bool `json:"verse_by_verse"`
	ShowVerseNums bool `json:"show_verse_nums"`
}

// Config holds the application configuration.
type Config struct {
	DefaultTranslation string            `json:"default_translation"`
	BibleTextAPI       string            `json:"bible_text_api"`
	FormattingOptions  FormattingOptions `json:"formatting_options"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		DefaultTranslation: "kjv",
		BibleTextAPI:       "bible-api.com",
		FormattingOptions: FormattingOptions{
			VerseByVerse:  false,
			ShowVerseNums: false,
		},
	}
}

// Load reads configuration from the given JSON file path.
// If the file does not exist, it returns the default config.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes the configuration to the given JSON file path.
func Save(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
