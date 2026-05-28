package gui

// ScriptureResult represents a formatted scripture for display.
type ScriptureResult struct {
	Reference string
	Text      string
	Book      string
	Chapter   int
	IsError   bool
}

// ChapterLookupFunc is called by the display to request formatted chapter text
// for a given book and chapter.
type ChapterLookupFunc func(book string, chapter int) (text string, err error)

// ScriptureDisplay is the interface main.go uses to push scripture to the GUI.
type ScriptureDisplay interface {
	// ShowResults displays scripture lookup results in the log panel.
	// For each unique book/chapter, the display requests the full chapter
	// via ChapterLookupFunc and opens a browser tab.
	ShowResults(results []ScriptureResult)

	// ShouldSkip checks and resets the skip flag, used to ignore
	// clipboard events triggered by in-app copy operations.
	ShouldSkip() bool
}
