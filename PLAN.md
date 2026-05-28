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
- [x] Build a robust regex/parser that detects scripture references in arbitrary text
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
- [ ] Select and integrate a Bible text API (or local data source)
- [ ] Implement a lookup function: given a reference, return the scripture text
- [ ] Support a translation parameter (default: KJV)

### 5. GUI — Scripture Display
- [ ] Create a GUI window/overlay that listens for scripture detection events
- [ ] Query the API for each detected reference
- [ ] Display the returned scripture text in a readable format
- [ ] Handle multiple references in a single clipboard event

### 6. Polish & Integration
- [ ] Wire all components together end-to-end
- [ ] Error handling and edge cases (invalid references, API failures, empty clipboard)
- [ ] System tray / background operation support
