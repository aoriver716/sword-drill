package gui

import (
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// App holds the GUI application components.
type App struct {
	fyneApp  fyne.App
	window   fyne.Window
	Log      *ScriptureLog
	skipNext atomic.Bool
}

// SkipNext sets a one-shot flag to ignore the next clipboard event.
func (a *App) SkipNext() {
	a.skipNext.Store(true)
}

// ShouldSkip checks and resets the skip flag. Returns true if the next event should be ignored.
func (a *App) ShouldSkip() bool {
	return a.skipNext.CompareAndSwap(true, false)
}

// New creates and configures the GUI application.
func New() *App {
	a := app.New()
	w := a.NewWindow("Sword Drill")
	w.Resize(fyne.NewSize(600, 500))

	scriptureLog := NewScriptureLog()

	guiApp := &App{
		fyneApp: a,
		window:  w,
		Log:     scriptureLog,
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			text := scriptureLog.PlainText()
			if text != "" {
				guiApp.SkipNext()
				w.Clipboard().SetContent(text)
			}
		}),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			scriptureLog.Clear()
		}),
	)

	content := container.NewBorder(toolbar, nil, nil, nil, scriptureLog.Widget())

	w.SetMainMenu(NewMenuBar(w))
	w.SetContent(content)

	return guiApp
}

// Run starts the GUI event loop. This blocks until the window is closed.
func (a *App) Run() {
	a.window.ShowAndRun()
}
