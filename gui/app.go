package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App holds the GUI application components.
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	Log     *ScriptureLog
}

// New creates and configures the GUI application.
func New() *App {
	a := app.New()
	w := a.NewWindow("Sword Drill")
	w.Resize(fyne.NewSize(600, 500))

	scriptureLog := NewScriptureLog()

	w.SetMainMenu(NewMenuBar(w))
	w.SetContent(scriptureLog.Widget())

	return &App{
		fyneApp: a,
		window:  w,
		Log:     scriptureLog,
	}
}

// Run starts the GUI event loop. This blocks until the window is closed.
func (a *App) Run() {
	a.window.ShowAndRun()
}
