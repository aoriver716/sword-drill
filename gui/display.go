package gui

// ChapterVerse is a single verse in a chapter.
type ChapterVerse struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

// ScriptureResult represents a formatted scripture for display.
type ScriptureResult struct {
	Reference  string
	Text       string
	Book       string
	Chapter    int
	StartVerse int
	EndVerse   int
	IsError    bool
}

// ChapterLookupFunc is called by the display to request chapter verse data
// for a given book, chapter, and translation.
type ChapterLookupFunc func(book string, chapter int, translation string) ([]ChapterVerse, error)

// ScriptureDisplay is the interface main.go uses to push scripture to the GUI.
type ScriptureDisplay interface {
	// ShowResults displays scripture lookup results in the log panel.
	// For each unique book/chapter, the display requests the full chapter
	// via ChapterLookupFunc and opens a browser tab.
	// If a tab already exists, it is focused and the relevant verses highlighted.
	ShowResults(results []ScriptureResult)

	// ShouldSkip checks and resets the skip flag, used to ignore
	// clipboard events triggered by in-app copy operations.
	ShouldSkip() bool

	// IsPaused returns true if clipboard processing is paused.
	IsPaused() bool
}
