# Sword Drill

__Disclaimer__: This entire project was shamelessly vibe coded.

Don't you hate it when you're reading a great article and the author drops a scripture reference without quoting the text — or worse, doesn't even link it? Gone are those days. Sword Drill watches your clipboard for scripture references and instantly pulls up the full text. Just highlight a reference and press Ctrl+C. It handles multiple references at once, so you can even copy an entire article and get every scripture in one shot.

## Features

- **Clipboard Detection** — Automatically detects scripture references (e.g. `John 3:16`, `Gen. 1:1`, `Rom 8:28-30`) when you copy text.
- **Scripture Browser** — Navigate and read scripture in context with tabbed browsing.
- **Scripture Log** — Scrollable log of all detected references.
- **Multiple Bible APIs** — Supports [bible-api.com](https://bible-api.com) and [API.Bible](https://scripture.api.bible) (2,500+ translations).
- **Cross-platform** — Fully supported on Windows and macOS.

## Download

[![Latest Release](https://img.shields.io/github/v/release/aoriver716/sword-drill)](https://github.com/aoriver716/sword-drill/releases/latest)

### Windows

📦 [Latest Stable](https://github.com/aoriver716/sword-drill/releases/latest/download/sword-drill-windows-amd64-setup.exe) · 🌙 [Nightly](https://github.com/aoriver716/sword-drill/releases/download/nightly/sword-drill-windows-amd64-setup.exe)

### macOS

📦 [Latest Stable](https://github.com/aoriver716/sword-drill/releases/latest/download/sword-drill-macos-universal.app.zip) · 🌙 [Nightly](https://github.com/aoriver716/sword-drill/releases/download/nightly/sword-drill-macos-universal.app.zip)

> 💡 **Stable** is recommended for most users. **Nightly** builds include the latest changes from `main` but may be unstable.

## Building from Source

Requires [Go](https://go.dev/dl/) 1.21+ and the [Wails CLI](https://wails.io/docs/gettingstarted/installation).

### Prerequisites

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

### Build

```sh
wails build
```

To embed an [API.Bible](https://scripture.api.bible) key at compile time:
```sh
wails build -ldflags "-X github.com/aoriver716/sword-drill/internal/lookup.apiKey=YOUR_KEY"
```

### API.Bible Setup

Official releases come pre-compiled with an API.Bible key — no setup needed.

If you're building from source and want to use API.Bible, sign up at [api.bible/sign-up](https://api.bible/sign-up) for a free API key, then build with the `-ldflags` flag shown above.

## License

[MIT](LICENSE)
