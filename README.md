# Sword Drill

Don't you hate it when you're reading a great article, and they have a scripture reference, but they don't quote the scripture? Even worse, they don't even link the reference to a popular platform where you can read the text. Well, gone are those times. The Sword Drill app seeks to solve this problem in a most elegant way. Simply highligh the scripture reference and press ctrl+c. Sword Drill will automatically find the text for you. Multiple scripture references are supported too. You can even copy the entire article to get every scripture reference at once.

This entire project was shamelessly vibe coded, and everything that follows is not my own words (my agent might even modify this introduction at some point as the project is updated).

## Build

Requires [Go](https://go.dev/dl/) 1.21+.

```sh
go build -o sword-drill.exe .
```

## Run

```sh
./sword-drill.exe
```

Copy any text containing a scripture reference (e.g. `John 3:16`, `Gen. 1:1`, `Rom 8:28-30`) and the app will detect it, look up the text, and display it in the console. Press `Ctrl+C` to quit.

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
