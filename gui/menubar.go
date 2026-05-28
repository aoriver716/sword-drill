package gui

import "fyne.io/fyne/v2"

// NewMenuBar creates the application menu bar.
func NewMenuBar(window fyne.Window) *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Quit", func() {
			window.Close()
		}),
	)
	return fyne.NewMainMenu(fileMenu)
}
