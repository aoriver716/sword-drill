package gui

import (
	"context"
	"fmt"
	"log"
	"sync"

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

// browserTab represents a chapter tab for the frontend.
type browserTab struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// App is the Wails application backend. It implements ScriptureDisplay.
type App struct {
	ctx           context.Context
	chapterLookup ChapterLookupFunc
	openTabs      map[string]bool
	mu            sync.Mutex
	skipNext      bool
}

// NewApp creates a new App instance with the given chapter lookup callback.
func NewApp(chapterLookup ChapterLookupFunc) *App {
	return &App{
		chapterLookup: chapterLookup,
		openTabs:      make(map[string]bool),
	}
}

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Ctx returns the app context.
func (a *App) Ctx() context.Context {
	return a.ctx
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

	seen := make(map[string]bool)
	for _, r := range results {
		if r.IsError {
			continue
		}
		tabName := fmt.Sprintf("%s %d", r.Book, r.Chapter)
		if seen[tabName] || a.HasTab(tabName) {
			continue
		}
		seen[tabName] = true

		text, err := a.chapterLookup(r.Book, r.Chapter)
		if err != nil {
			log.Printf("ERROR chapter lookup %s: %v", tabName, err)
			continue
		}

		a.mu.Lock()
		a.openTabs[tabName] = true
		a.mu.Unlock()

		runtime.EventsEmit(a.ctx, "browser:openTab", browserTab{
			Name: tabName,
			Text: text,
		})
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

// CloseTab is called from the frontend when a tab is closed.
func (a *App) CloseTab(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.openTabs, name)
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
