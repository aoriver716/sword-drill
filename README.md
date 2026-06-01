# Sword Drill

__Disclaimer__: This entire project was shamelessly vibe coded.

Don't you hate it when you're reading a great article and the author drops a scripture reference without quoting the text — or worse, doesn't even link it? Gone are those days. Sword Drill watches your clipboard for scripture references and instantly pulls up the full text. Just highlight a reference and press Ctrl+C. It handles multiple references at once, so you can even copy an entire article and get every scripture in one shot.

## Download

| Channel | Link | Description |
|---------|------|-------------|
| **Stable** | [Latest Release](https://github.com/aoriver716/sword-drill/releases/latest) | Recommended for most users |
| **Nightly** | [Nightly Build](https://github.com/aoriver716/sword-drill/releases/tag/nightly) | Latest from `main` — may be unstable |

Available for **Windows** and **macOS**.

Windows releases include both a portable executable (`sword-drill-windows-amd64.exe`) and an NSIS installer (`sword-drill-windows-amd64-setup.exe`).

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
CGO_ENABLED=1 go build -tags "desktop,production" -ldflags "-X github.com/aoriver716/sword-drill/internal/lookup.apiKey=YOUR_KEY" -o sword-drill.exe .
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

- **Scripture Browser** — Navigate and read scripture in context.
- **Scripture Log** — Scrollable log of all detected references.
- **Multiple Bible APIs** — Supports [bible-api.com](https://bible-api.com) and [API.Bible](https://scripture.api.bible)

Close the window or use File → Quit (Ctrl+Q) to exit.

### API.Bible Setup

Sword Drill uses API.Bible among its bible API's, which requires an API key. Official releases come pre-compiled with an API.Bible key — no setup needed.

However, if you're building from source and want to use [API.Bible](https://scripture.api.bible) (2,500+ translations), sign up at [api.bible/sign-up](https://api.bible/sign-up) for a free API key, then build with:

```sh
CGO_ENABLED=1 go build -tags "desktop,production" -ldflags "-X github.com/aoriver716/sword-drill/internal/lookup.apiKey=your-key" -o sword-drill.exe .
```
