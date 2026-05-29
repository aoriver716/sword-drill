# Sword Drill

__Disclaimer__: This entire project was shamelessly vibe coded.

Don't you hate it when you're reading a great article and the author drops a scripture reference without quoting the text — or worse, doesn't even link it? Gone are those days. Sword Drill watches your clipboard for scripture references and instantly pulls up the full text. Just highlight a reference and press Ctrl+C. It handles multiple references at once, so you can even copy an entire article and get every scripture in one shot.

## Build

Requires [Go](https://go.dev/dl/) 1.21+ and a C compiler (CGo is needed for the clipboard package).

**Windows** — install [MSYS2](https://www.msys2.org/) and MinGW-w64 GCC:
```sh
pacman -S mingw-w64-x86_64-gcc
```
Then add `C:\msys64\mingw64\bin` to your PATH and set `CGO_ENABLED=1`.

**Linux** — install GCC and required development headers:
```sh
sudo apt install gcc libgtk-3-dev libwebkit2gtk-4.0-dev
```

**macOS** — install Xcode command line tools:
```sh
xcode-select --install
```

Build:
```sh
CGO_ENABLED=1 go build -tags "desktop,production" -o sword-drill.exe .
```

To embed an [API.Bible](https://scripture.api.bible) key at compile time:
```sh
CGO_ENABLED=1 go build -tags "desktop,production" -ldflags "-X main.apiBibleKey=YOUR_KEY" -o sword-drill.exe .
```

For development builds (with console output visible):
```sh
CGO_ENABLED=1 go build -tags desktop -o sword-drill.exe .
```

## Run

```sh
./sword-drill.exe
```

Copy any text containing a scripture reference (e.g. `John 3:16`, `Gen. 1:1`, `Rom 8:28-30`) and the app will detect it, look up the text, and display it in the GUI window.

### Features

- **Scripture Browser** — Tabbed chapter viewer with full chapter text, verse-level highlighting of referenced passages, and book/chapter navigation toolbar
- **Scripture Log** — Scrollable log of all detected references with copy-all and clear buttons
- **New Tab** — Open Genesis 1 with File → New Tab or Ctrl+N
- **Pause/Resume** — Toggle clipboard processing from the menu bar
- **Draggable Tabs** — Reorder browser tabs by drag and drop
- **Resizable Panels** — Adjust the split between browser and log
- **Configurable Formatting** — Verse-by-verse or paragraph mode, optional verse numbers
- **Multiple Bible APIs** — Supports [bible-api.com](https://bible-api.com) and [API.Bible](https://scripture.api.bible) with pluggable `BibleLookup` interface

Close the window or use File → Quit (Ctrl+Q) to exit.

## Configuration

Settings are read from `config.json` in the working directory. If the file doesn't exist, defaults are used.

```json
{
  "default_translation": "kjv",
  "bible_text_api": "api.bible",
  "formatting_options": {
    "verse_by_verse": false,
    "show_verse_nums": false
  }
}
```

| Key | Description | Default |
|---|---|---|
| `default_translation` | Bible translation ID (e.g. `kjv`, `web`) | `kjv` |
| `bible_text_api` | API source: `bible-api.com` or `api.bible` | `bible-api.com` |
| `formatting_options.verse_by_verse` | Display each verse on its own line | `false` |
| `formatting_options.show_verse_nums` | Prefix each verse with its number | `false` |

### API.Bible Setup

To use [API.Bible](https://scripture.api.bible) (2,500+ translations), sign up at [api.bible/sign-up](https://api.bible/sign-up) for a free API key. Provide the key via one of these methods (highest priority first):

1. **Environment variable**: `API_BIBLE_KEY=your-key ./sword-drill.exe`
2. **Config file**: add `"api_bible_key": "your-key"` to `config.json`
3. **Compile-time flag**: build with `-ldflags "-X main.apiBibleKey=your-key"`

## Project Plan

See [doc/PLAN.md](doc/PLAN.md) for the full implementation checklist and roadmap.
