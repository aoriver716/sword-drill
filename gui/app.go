package gui

import (
	"context"
	"fmt"
	"sync"

	"github.com/aoriver716/sword-drill/internal/formatter"
	"github.com/aoriver716/sword-drill/internal/lookup"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogEntry represents a scripture lookup result for the frontend.
type LogEntry struct {
	Reference string `json:"reference"`
	Text      string `json:"text"`
	IsError   bool   `json:"isError"`
}

// BrowserTab represents a chapter tab for the frontend.
type BrowserTab struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// App is the Wails application backend.
type App struct {
	ctx      context.Context
	fmtOpts  formatter.Options
	openTabs map[string]bool
	mu       sync.Mutex
	skipNext bool
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		openTabs: make(map[string]bool),
	}
}

// SetFormatOptions sets the formatter options.
func (a *App) SetFormatOptions(opts formatter.Options) {
	a.fmtOpts = opts
}

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Ctx returns the app context.
func (a *App) Ctx() context.Context {
	return a.ctx
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

// AppendLog sends a scripture result to the frontend log.
func (a *App) AppendLog(result lookup.LookupResult) {
	entry := LogEntry{
		Reference: result.Reference,
		Text:      formatter.Format(result, a.fmtOpts),
	}
	runtime.EventsEmit(a.ctx, "log:append", entry)
}

// AppendLogError sends an error entry to the frontend log.
func (a *App) AppendLogError(reference string, err error) {
	entry := LogEntry{
		Reference: reference,
		Text:      fmt.Sprintf("Error: %v", err),
		IsError:   true,
	}
	runtime.EventsEmit(a.ctx, "log:append", entry)
}

// OpenTab sends a chapter tab to the frontend.
func (a *App) OpenTab(name string, result lookup.LookupResult) {
	a.mu.Lock()
	a.openTabs[name] = true
	a.mu.Unlock()

	tab := BrowserTab{
		Name: name,
		Text: formatter.Format(result, a.fmtOpts),
	}
	runtime.EventsEmit(a.ctx, "browser:openTab", tab)
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
// It sets the skip flag before the clipboard write happens.
func (a *App) CopyText(text string) {
	a.SkipNext()
	runtime.ClipboardSetText(a.ctx, text)
}

// Quit closes the application.
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}
