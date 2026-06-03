package config

import "testing"

func TestMigrateV0ToV1_PrefixesTranslationKeys(t *testing.T) {
	cfg := &Config{
		ConfigVersion:       0,
		BibleTextAPI:        "api.bible",
		DefaultTranslation:  "de4e12af7f28f599-02",
		ParallelTranslation: "de4e12af7f28f599-02",
	}

	migrateV0ToV1(cfg)

	if cfg.DefaultTranslation != "api.bible/de4e12af7f28f599-02" {
		t.Errorf("DefaultTranslation = %q, want %q", cfg.DefaultTranslation, "api.bible/de4e12af7f28f599-02")
	}
	if cfg.ParallelTranslation != "api.bible/de4e12af7f28f599-02" {
		t.Errorf("ParallelTranslation = %q, want %q", cfg.ParallelTranslation, "api.bible/de4e12af7f28f599-02")
	}
}

func TestMigrateV0ToV1_SkipsAlreadyPrefixed(t *testing.T) {
	cfg := &Config{
		ConfigVersion:       0,
		BibleTextAPI:        "api.bible",
		DefaultTranslation:  "esv/esv",
		ParallelTranslation: "api.bible/de4e12af7f28f599-02",
	}

	migrateV0ToV1(cfg)

	if cfg.DefaultTranslation != "esv/esv" {
		t.Errorf("DefaultTranslation = %q, want %q", cfg.DefaultTranslation, "esv/esv")
	}
	if cfg.ParallelTranslation != "api.bible/de4e12af7f28f599-02" {
		t.Errorf("ParallelTranslation = %q, want %q", cfg.ParallelTranslation, "api.bible/de4e12af7f28f599-02")
	}
}

func TestMigrateV0ToV1_DefaultsToAPIBible(t *testing.T) {
	cfg := &Config{
		ConfigVersion:      0,
		BibleTextAPI:       "", // missing
		DefaultTranslation: "de4e12af7f28f599-02",
	}

	migrateV0ToV1(cfg)

	if cfg.DefaultTranslation != "api.bible/de4e12af7f28f599-02" {
		t.Errorf("DefaultTranslation = %q, want %q", cfg.DefaultTranslation, "api.bible/de4e12af7f28f599-02")
	}
}

func TestRunMigrations_AppliesAll(t *testing.T) {
	cfg := &Config{
		ConfigVersion:       0,
		BibleTextAPI:        "api.bible",
		DefaultTranslation:  "de4e12af7f28f599-02",
		ParallelTranslation: "de4e12af7f28f599-02",
	}

	migrated := runMigrations(cfg)

	if !migrated {
		t.Error("Expected migrations to run")
	}
	if cfg.ConfigVersion != len(migrations) {
		t.Errorf("ConfigVersion = %d, want %d", cfg.ConfigVersion, len(migrations))
	}
	if cfg.DefaultTranslation != "api.bible/de4e12af7f28f599-02" {
		t.Errorf("DefaultTranslation not migrated: %q", cfg.DefaultTranslation)
	}
}

func TestRunMigrations_SkipsIfCurrent(t *testing.T) {
	cfg := &Config{
		ConfigVersion:      len(migrations),
		DefaultTranslation: "esv/esv",
	}

	migrated := runMigrations(cfg)

	if migrated {
		t.Error("Expected no migrations to run")
	}
	if cfg.DefaultTranslation != "esv/esv" {
		t.Errorf("DefaultTranslation changed unexpectedly: %q", cfg.DefaultTranslation)
	}
}
