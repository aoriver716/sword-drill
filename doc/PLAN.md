# Sword Drill — Scripture Reference Detector

## Architecture Overview

A clipboard-monitoring app that detects scripture references in copied text and displays the corresponding biblical text.

## Implementation Checklist

### 1. Project Setup
- [x] Choose language/framework and initialize project structure — **Go** (`golang.design/x/clipboard`)
- [x] Set up dependencies and build configuration

### 2. Clipboard Monitor Service
- [x] Implement a service that polls/watches the system clipboard for changes
- [x] Trigger a callback whenever new text is copied to the clipboard

### 3. Scripture Reference Detection
- [x] Build a robust regex/parser that detects scripture references in arbitrary text (`internal/detector`)
- [x] Handle varied reference formats:
  - Full book names (`Genesis 1:1`, `1 Chronicles 15:10-13`)
  - Standard abbreviations (`Gen. 1:1`, `1 Chron. 15:10-13`)
  - Informal abbreviations (`Gen 1:1`, `1 Chr 15:10`)
  - Verse ranges (`John 3:16-18`)
  - Chapter-only references (`Psalm 23`)
  - Multi-chapter ranges (`Isaiah 52:13-53:12`)
  - Multiple references in one string (`Rom. 8:28; John 3:16`)
- [x] Emit detected references as events (book, chapter, start verse, end verse)

### 4. Bible Text API Integration
- [x] Define `BibleLookup` interface with `Lookup(ref, translation)` method
- [x] Implement `BibleAPIClient` (bible-api.com) — default online provider
- [x] Wire lookup into clipboard callback to display scripture text
- [x] Support a translation parameter (default: KJV)
- [x] Add shared test suite for `BibleLookup` implementations (John 3:16 + Romans 8:28-30 multi-verse)
- [x] Implement `APIBibleClient` (API.Bible) — alternate provider with 2,500+ translations
- [x] API key injection via compile-time `-ldflags`, env var, or config file
- [x] `Translations()` method on `BibleLookup` interface (static for bible-api.com, cached for API.Bible)
- [x] `RefreshTranslations()` method for fetching/caching available translations from API.Bible
- [ ] Implement local SQLite DB lookup (`BibleLookup` implementation)

### 5. GUI — Scripture Display
- [x] Create a GUI window that listens for scripture detection events — **Wails v2** (Go + HTML/CSS/JS)
- [x] Query the API for each detected reference
- [x] Display the returned scripture text in a readable format
- [x] Handle multiple references in a single clipboard event
- [x] Scripture Browser with tabbed chapter view and per-tab navigation (book dropdown, chapter arrows)
- [x] Scripture Log with copy-all and clear toolbar
- [x] Verse highlighting for referenced verses (soft blue)
- [x] Focus existing tabs and re-highlight on duplicate references
- [x] Draggable tab reordering
- [x] Resizable split between browser and log panels
- [x] Pause/resume clipboard processing toggle

### 6. Polish & Integration
- [x] Wire all components together end-to-end
- [x] Error handling and edge cases (invalid references, API failures, empty clipboard)
- [x] JSON config file with defaults fallback
- [x] ScriptureDisplay interface decoupling GUI from core logic
- [ ] System tray / background operation support

### 7. Future Enhancements
- [ ] Preferences menu (formatting options, translation, theme)
- [ ] Session persistence (save/restore open tabs and panel layout on restart)
- [ ] Reopen closed tabs with Ctrl+Shift+T
- [ ] User highlighting in Scripture Browser (select text to highlight with custom colors)
- [ ] Annotations in Scripture Browser (attach notes to verses)
- [ ] Cross-reference and footnote support
- [ ] Dark mode option
