package gui

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aoriver716/sword-drill/internal/config"
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
	Name      string         `json:"name"`
	Verses    []ChapterVerse `json:"verses"`
	Highlight [][2]int       `json:"highlight"`
}

// focusTab represents a focus+highlight event for an existing tab.
type focusTab struct {
	Name      string   `json:"name"`
	Highlight [][2]int `json:"highlight"`
}

// FormatOptions controls how verses are rendered in the browser.
type FormatOptions struct {
	VerseByVerse  bool `json:"verseByVerse"`
	ShowVerseNums bool `json:"showVerseNums"`
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

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Ctx returns the app context.
func (a *App) Ctx() context.Context {
	return a.ctx
}

// GetFormatOptions returns the current formatting options (called from frontend).
func (a *App) GetFormatOptions() FormatOptions {
	cfg := a.registry.Config()
	return FormatOptions{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
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
			info = &tabInfo{isNew: !a.HasTab(tabName)}
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
			verses, err := a.chapterLookup(r.Book, r.Chapter)
			if err != nil {
				log.Printf("ERROR chapter lookup %s: %v", tabName, err)
				delete(tabMap, tabName)
				continue
			}

			a.mu.Lock()
			a.openTabs[tabName] = true
			a.mu.Unlock()

			runtime.EventsEmit(a.ctx, "browser:openTab", browserTab{
				Name:      tabName,
				Verses:    verses,
				Highlight: info.highlights,
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

// LoadChapter is called from the frontend to load a book/chapter.
// Returns the structured verse data.
func (a *App) LoadChapter(book string, chapter int) ([]ChapterVerse, error) {
	return a.chapterLookup(book, chapter)
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
	runtime.Quit(a.ctx)
}
