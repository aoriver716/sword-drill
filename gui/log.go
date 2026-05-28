package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ScriptureLog is a scrollable log of scripture lookup results.
type ScriptureLog struct {
	container *fyne.Container
	scroll    *container.Scroll
}

// NewScriptureLog creates a new empty scripture log.
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
	timestamp := time.Now().Format("15:04:05")
	header := widget.NewRichTextFromMarkdown(fmt.Sprintf("**[%s] %s**", timestamp, reference))
	body := widget.NewLabel(text)
	body.Wrapping = fyne.TextWrapWord
	sep := widget.NewSeparator()

	sl.container.Add(header)
	sl.container.Add(body)
	sl.container.Add(sep)

	// Scroll to bottom
	sl.scroll.ScrollToBottom()
}
