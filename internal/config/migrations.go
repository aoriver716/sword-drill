package config

import "strings"

// migrations is the ordered list of config migration functions.
// Each function transforms the config from version N to N+1.
// A config at version 0 (or missing the field entirely) has never been
// migrated; it will run all migrations in sequence.
//
// To add a new migration:
//  1. Append a new function to this slice.
//  2. The function receives a *Config and mutates it in place.
//  3. Do NOT change the order of existing entries.
var migrations = []func(*Config){
	migrateV0ToV1, // prefix translation keys with provider
}

// runMigrations applies any pending migrations to the config and returns
// true if any were applied.
func runMigrations(cfg *Config) bool {
	if cfg.ConfigVersion >= len(migrations) {
		return false
	}
	for i := cfg.ConfigVersion; i < len(migrations); i++ {
		migrations[i](cfg)
	}
	cfg.ConfigVersion = len(migrations)
	return true
}

// migrateV0ToV1 prefixes bare translation keys with the configured provider.
// Old configs stored keys like "de4e12af7f28f599-02"; the multi-provider
// system expects "api.bible/de4e12af7f28f599-02".
func migrateV0ToV1(cfg *Config) {
	provider := cfg.BibleTextAPI
	if provider == "" {
		provider = "api.bible"
	}

	if cfg.DefaultTranslation != "" && !strings.Contains(cfg.DefaultTranslation, "/") {
		cfg.DefaultTranslation = provider + "/" + cfg.DefaultTranslation
	}
	if cfg.ParallelTranslation != "" && !strings.Contains(cfg.ParallelTranslation, "/") {
		cfg.ParallelTranslation = provider + "/" + cfg.ParallelTranslation
	}
}
