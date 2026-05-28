# Sword Drill

Don't you hate it when you're reading a great article, and they have a scripture reference, but they don't quote the scripture? Even worse, they don't even link the reference to a popular platform where you can read the text. Well, gone are those times. The Sword Drill app seeks to solve this problem in a most elegant way. Simply highligh the scripture reference and press ctrl+c. Sword Drill will automatically find the text for you. Multiple scripture references are supported too. You can even copy the entire article to get every scripture reference at once.

This entire project was shamelessly vibe coded, and everything that follows is not my own words (my agent might even modify this introduction at some point as the project is updated).

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
- **Pause/Resume** — Toggle clipboard processing from the menu bar
- **Draggable Tabs** — Reorder browser tabs by drag and drop
- **Resizable Panels** — Adjust the split between browser and log
- **Configurable Formatting** — Verse-by-verse or paragraph mode, optional verse numbers

Close the window or use File → Quit (Ctrl+Q) to exit.

## Configuration

Settings are read from `config.json` in the working directory. If the file doesn't exist, defaults are used.

```json
{
  "default_translation": "kjv",
  "bible_text_api": "bible-api.com",
  "formatting_options": {
    "verse_by_verse": false,
    "show_verse_nums": false
  }
}
```

| Key | Description | Default |
|---|---|---|
| `default_translation` | Bible translation ID (e.g. `kjv`, `web`) | `kjv` |
| `bible_text_api` | API source for scripture text | `bible-api.com` |
| `formatting_options.verse_by_verse` | Display each verse on its own line | `false` |
| `formatting_options.show_verse_nums` | Prefix each verse with its number | `false` |

## Project Plan

See [doc/PLAN.md](doc/PLAN.md) for the full implementation checklist and roadmap.
