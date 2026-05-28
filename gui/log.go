package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ScriptureLog is a scrollable log of scripture lookup results.
type ScriptureLog struct {
	container *fyne.Container
	scroll    *container.Scroll
}

// NewScriptureLog creates a new scripture log.
func NewScriptureLog() *ScriptureLog {
	c := container.NewVBox()
	scroll := container.NewVScroll(c)
	return &ScriptureLog{container: c, scroll: scroll}
}

// Widget returns the scrollable widget to embed in a layout.
func (sl *ScriptureLog) Widget() fyne.CanvasObject {
	return sl.scroll
}

// Append adds a scripture lookup result to the log.
func (sl *ScriptureLog) Append(reference string, text string) {
	header := widget.NewRichTextFromMarkdown("**" + reference + "**")
	body := widget.NewLabel(text)
	body.Wrapping = fyne.TextWrapWord
	sep := widget.NewSeparator()

	fyne.Do(func() {
		sl.container.Add(header)
		sl.container.Add(body)
		sl.container.Add(sep)
		sl.scroll.ScrollToBottom()
	})
}

// PlainText returns all log entries as plain text.
func (sl *ScriptureLog) PlainText() string {
	var result string
	// Iterate over widgets: pattern is header, body, separator repeating
	objects := sl.container.Objects
	for i := 0; i+2 < len(objects); i += 3 {
		if header, ok := objects[i].(*widget.RichText); ok {
			result += header.String() + "\n"
		}
		if body, ok := objects[i+1].(*widget.Label); ok {
			result += body.Text + "\n\n"
		}
	}
	return result
}

// Clear removes all entries from the log.
func (sl *ScriptureLog) Clear() {
	sl.container.RemoveAll()
}
