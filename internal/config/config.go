package config

// FormattingOptions controls how scripture text is displayed.
type FormattingOptions struct {
	VerseByVerse  bool `json:"verse_by_verse"`
	ShowVerseNums bool `json:"show_verse_nums"`
}

// Config holds the application configuration.
type Config struct {
	ConfigVersion       int               `json:"config_version"`
	DefaultTranslation  string            `json:"default_translation"`
	ParallelTranslation string            `json:"parallel_translation"`
	BibleTextAPI        string            `json:"bible_text_api"`
	FormattingOptions   FormattingOptions `json:"formatting_options"`
	TabOpenBehavior     string            `json:"tab_open_behavior"`
	CheckForUpdates     bool              `json:"check_for_updates"`
	UpdateChannel       string            `json:"update_channel"`
	CacheTTLDays        int               `json:"cache_ttl_days"`
	SkippedVersion      string            `json:"skipped_version,omitempty"`
}
