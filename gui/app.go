package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/aoriver716/sword-drill/internal/cache"
	"github.com/aoriver716/sword-drill/internal/config"
	"github.com/aoriver716/sword-drill/internal/updater"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Compile-time check that App implements ScriptureDisplay.
var _ ScriptureDisplay = (*App)(nil)

// logEntry represents a scripture lookup result for the frontend.
type logEntry struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
	IsError   bool   `json:"isError"`
}

// browserTab represents a chapter tab event for the frontend.
type browserTab struct {
	Name        string         `json:"name"`
	Verses      []ChapterVerse `json:"verses"`
	Highlight   [][2]int       `json:"highlight"`
	Translation string         `json:"translation"`
}

// focusTab represents a focus+highlight event for an existing tab.
type focusTab struct {
	Name      string   `json:"name"`
	Highlight [][2]int `json:"highlight"`
}

// TabStateEntry describes a single tab for persistence.
// Array order encodes position; closed tabs include a Position hint for reopening.
type TabStateEntry struct {
	Book                string `json:"book"`
	Chapter             int    `json:"chapter"`
	Translation         string `json:"translation"`
	ParallelMode        bool   `json:"parallelMode,omitempty"`
	ParallelTranslation string `json:"parallelTranslation,omitempty"`
	Position            int    `json:"position,omitempty"`
}

// TabsFile is the on-disk format for tab persistence.
type TabsFile struct {
	Open      []TabStateEntry `json:"open"`
	ActiveIdx int             `json:"activeIdx"`
	Closed    []TabStateEntry `json:"closed"`
}

// App is the Wails application backend. It implements ScriptureDisplay.
type App struct {
	ctx           context.Context
	chapterLookup ChapterLookupFunc
	openTabs      map[string]bool
	mu            sync.Mutex
	skipNext      bool
	paused        bool
	registry      *config.Registry
	closing       atomic.Bool
}

// NewApp creates a new App instance with the given chapter lookup callback and config registry.
func NewApp(chapterLookup ChapterLookupFunc, registry *config.Registry) *App {
	return &App{
		chapterLookup: chapterLookup,
		openTabs:      make(map[string]bool),
		registry:      registry,
	}
}

// GetConfigSchema returns all config field schemas for the preferences UI.
func (a *App) GetConfigSchema() []config.FieldSchema {
	if a.registry == nil {
		return nil
	}
	return a.registry.Schema()
}

// UpdateConfigField updates a config field by key and saves to disk.
func (a *App) UpdateConfigField(key string, value any) error {
	if a.registry == nil {
		return nil
	}
	return a.registry.Update(key, value)
}

// ResetConfigToDefaults resets all config fields to defaults and saves.
func (a *App) ResetConfigToDefaults() error {
	if a.registry == nil {
		return nil
	}
	return a.registry.ResetToDefaults()
}

// RefreshTranslations invalidates any cached translation lists so the next
// call refetches from all providers.
func (a *App) RefreshTranslations() error {
	if a.registry == nil {
		return nil
	}
	multi := a.registry.MultiLookup()
	return multi.RefreshTranslations()
}

// InvokeFieldAction triggers the Action callback associated with a config
// field (used by button-type widgets such as "Clear Scripture Cache").
// Returns a status message on success.
func (a *App) InvokeFieldAction(key string) (string, error) {
	if a.registry == nil {
		return "", nil
	}
	return a.registry.InvokeAction(key)
}

// CacheStats reports the current state of the scripture cache. Returns a
// zero-value Stats when the cache is disabled.
func (a *App) CacheStats() cache.Stats {
	if a.registry == nil {
		return cache.Stats{}
	}
	c := a.registry.Cache()
	if c == nil {
		return cache.Stats{}
	}
	s, _ := c.Stats()
	return s
}

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if a.registry != nil {
		a.registry.OnChange(func(cfg *config.Config) {
			a.emitFormatOptions(cfg)
		})
	}
}

// emitFormatOptions pushes the current formatting options to the frontend.
func (a *App) emitFormatOptions(cfg *config.Config) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "config:formatChanged", map[string]bool{
		"verseByVerse":  cfg.FormattingOptions.VerseByVerse,
		"showVerseNums": cfg.FormattingOptions.ShowVerseNums,
	})
}

// Ctx returns the app context.
func (a *App) Ctx() context.Context {
	return a.ctx
}

// GetFormatOptions returns the current formatting options (called from frontend).
func (a *App) GetFormatOptions() map[string]bool {
	cfg := a.registry.Config()
	return map[string]bool{
		"verseByVerse":  cfg.FormattingOptions.VerseByVerse,
		"showVerseNums": cfg.FormattingOptions.ShowVerseNums,
	}
}

// ShowResults displays scripture lookup results and opens chapter tabs.
func (a *App) ShowResults(results []ScriptureResult) {
	for _, r := range results {
		entry := logEntry{
			Reference: r.Reference,
			Text:      r.Text,
			IsError:   r.IsError,
		}
		runtime.EventsEmit(a.ctx, "log:append", entry)
	}

	tabBehavior := a.registry.Config().TabOpenBehavior
	if tabBehavior == "never" {
		return
	}

	// Collect highlight ranges per tab
	type tabInfo struct {
		highlights [][2]int
		isNew      bool
	}
	tabMap := make(map[string]*tabInfo)

	for _, r := range results {
		if r.IsError {
			continue
		}
		tabName := fmt.Sprintf("%s %d", r.Book, r.Chapter)
		info, exists := tabMap[tabName]
		if !exists {
			forceNew := tabBehavior == "always_new"
			info = &tabInfo{isNew: forceNew || !a.HasTab(tabName)}
			tabMap[tabName] = info
		}
		if r.StartVerse > 0 {
			info.highlights = append(info.highlights, [2]int{r.StartVerse, r.EndVerse})
		}
	}

	for _, r := range results {
		if r.IsError {
			continue
		}
		tabName := fmt.Sprintf("%s %d", r.Book, r.Chapter)
		info, ok := tabMap[tabName]
		if !ok {
			continue
		}

		if info.isNew {
			translation := a.registry.Config().DefaultTranslation
			verses, err := a.chapterLookup(r.Book, r.Chapter, translation)
			if err != nil {
				log.Printf("ERROR chapter lookup %s: %v", tabName, err)
				delete(tabMap, tabName)
				continue
			}

			a.mu.Lock()
			a.openTabs[tabName] = true
			a.mu.Unlock()

			runtime.EventsEmit(a.ctx, "browser:openTab", browserTab{
				Name:        tabName,
				Verses:      verses,
				Highlight:   info.highlights,
				Translation: translation,
			})
		} else {
			runtime.EventsEmit(a.ctx, "browser:focusTab", focusTab{
				Name:      tabName,
				Highlight: info.highlights,
			})
		}

		// Only process each tab once
		delete(tabMap, tabName)
	}
}

// SkipNext sets a one-shot flag to ignore the next clipboard event.
func (a *App) SkipNext() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.skipNext = true
}

// ShouldSkip checks and resets the skip flag.
func (a *App) ShouldSkip() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.skipNext {
		a.skipNext = false
		return true
	}
	return false
}

// HasTab returns true if a tab with the given name is open.
func (a *App) HasTab(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.openTabs[name]
}

// LoadChapter is called from the frontend to load a book/chapter/translation.
// Returns the structured verse data.
func (a *App) LoadChapter(book string, chapter int, translation string) ([]ChapterVerse, error) {
	return a.chapterLookup(book, chapter, translation)
}

// GetTranslations returns the available translations from all Bible APIs.
func (a *App) GetTranslations() []config.Option {
	if a.registry == nil {
		return nil
	}
	multi := a.registry.MultiLookup()
	translations, err := multi.Translations()
	if err != nil {
		return nil
	}
	opts := make([]config.Option, 0, len(translations))
	for _, t := range translations {
		opts = append(opts, config.Option{
			Label:   t.Name,
			Value:   t.Key,
			IsGroup: t.IsGroup,
		})
	}
	return opts
}

// GetDefaultTranslation returns the currently configured default translation key.
func (a *App) GetDefaultTranslation() string {
	if a.registry == nil {
		return ""
	}
	return a.registry.Config().DefaultTranslation
}

// GetParallelTranslation returns the currently configured parallel translation key.
func (a *App) GetParallelTranslation() string {
	if a.registry == nil {
		return ""
	}
	return a.registry.Config().ParallelTranslation
}

// RenameTab updates the backend tab tracking when a tab changes book/chapter.
func (a *App) RenameTab(oldName string, newName string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.openTabs, oldName)
	a.openTabs[newName] = true
}

// CloseTab is called from the frontend when a tab is closed.
func (a *App) CloseTab(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.openTabs, name)
}

// OpenTab registers a tab as open in the backend tracking.
func (a *App) OpenTab(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.openTabs[name] = true
}

// SetPaused is called from the frontend to toggle clipboard processing.
func (a *App) SetPaused(paused bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.paused = paused
}

// IsPaused returns true if clipboard processing is paused.
func (a *App) IsPaused() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.paused
}

// CopyText is called from the frontend to copy text to clipboard.
func (a *App) CopyText(text string) {
	a.SkipNext()
	runtime.ClipboardSetText(a.ctx, text)
}

// Quit closes the application.
func (a *App) Quit() {
	a.closing.Store(true)
	runtime.Quit(a.ctx)
}

// BeforeClose is called by Wails when the window is about to close.
// It returns true to prevent close (so the frontend can save state first),
// unless Quit() was called explicitly (closing flag set).
func (a *App) BeforeClose(ctx context.Context) bool {
	if a.closing.Load() {
		if a.registry != nil {
			if c := a.registry.Cache(); c != nil {
				_ = c.Close()
			}
		}
		return false
	}
	runtime.EventsEmit(ctx, "app:beforeClose")
	return true
}

// SaveTabState writes the tab state to tabs.json.
func (a *App) SaveTabState(state TabsFile) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("tabs.json", data, 0644)
}

// LoadTabState reads the tab state from tabs.json.
func (a *App) LoadTabState() *TabsFile {
	data, err := os.ReadFile("tabs.json")
	if err != nil {
		return nil
	}
	var state TabsFile
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("Failed to parse tabs.json: %v", err)
		return nil
	}
	return &state
}

// CheckForUpdates queries GitHub for the latest release and returns update info.
func (a *App) CheckForUpdates() updater.UpdateInfo {
	channel := updater.ChannelStable
	if a.registry != nil {
		if c := a.registry.Config().UpdateChannel; c != "" {
			channel = c
		}
	}
	return updater.CheckForUpdates(channel)
}

// GetVersion returns the current application version.
func (a *App) GetVersion() string {
	return updater.Version
}

// ShouldCheckForUpdates returns whether auto-update check is enabled.
func (a *App) ShouldCheckForUpdates() bool {
	if a.registry == nil {
		return false
	}
	return a.registry.Config().CheckForUpdates
}
