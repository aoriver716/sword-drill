package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/aoriver716/sword-drill/gui"
	"github.com/aoriver716/sword-drill/internal/config"
	"github.com/aoriver716/sword-drill/internal/detector"
	"github.com/aoriver716/sword-drill/internal/formatter"
	"github.com/aoriver716/sword-drill/internal/lookup"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"golang.design/x/clipboard"
)

//go:embed gui/frontend/*
var assets embed.FS

var (
	cfg     config.Config
	bible   lookup.BibleLookup
	fmtOpts formatter.Options

	// Set at compile time via: -ldflags "-X main.apiBibleKey=YOUR_KEY"
	apiBibleKey string
)

func initConfig() {
	var err error
	cfg, err = config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	switch cfg.BibleTextAPI {
	case "api.bible":
		apiKey := apiBibleKey // compile-time default
		if cfg.APIBibleKey != "" {
			apiKey = cfg.APIBibleKey // config file overrides
		}
		if envKey := os.Getenv("API_BIBLE_KEY"); envKey != "" {
			apiKey = envKey // env var wins
		}
		if apiKey == "" {
			log.Fatal("API key required: set API_BIBLE_KEY env var, api_bible_key in config.json, or compile with -ldflags \"-X main.apiBibleKey=KEY\"")
		}
		bibleID := cfg.APIBibleID
		if bibleID == "" {
			bibleID = "de4e12af7f28f599-02" // default KJV
		}
		bible = lookup.NewAPIBibleClient(apiKey, bibleID)
		log.Printf("Using API.Bible (bibleId=%s)", bibleID)
	default:
		bible = lookup.NewBibleAPIClient()
		log.Println("Using bible-api.com")
	}

	fmtOpts = formatter.Options{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	}
}

func lookupChapter(book string, chapter int) ([]gui.ChapterVerse, error) {
	ref := detector.ScriptureRef{
		Book:         book,
		StartChapter: chapter,
		EndChapter:   chapter,
	}
	result, err := bible.Lookup(ref, cfg.DefaultTranslation)
	if err != nil {
		return nil, err
	}
	if result.SourceURL != nil {
		log.Printf("OK [%d] chapter %s %d → %s", result.StatusCode, book, chapter, *result.SourceURL)
	}
	verses := make([]gui.ChapterVerse, len(result.Verses))
	for i, v := range result.Verses {
		verses[i] = gui.ChapterVerse{Number: v.Number, Text: v.Text}
	}
	return verses, nil
}

func watchClipboard(ctx context.Context, display gui.ScriptureDisplay) {
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		if len(data) == 0 {
			continue
		}
		if display.ShouldSkip() {
			log.Println("Skipping self-triggered clipboard event")
			continue
		}
		if display.IsPaused() {
			continue
		}
		text := string(data)
		refs := detector.ParseReferences(text)

		var results []gui.ScriptureResult
		for _, ref := range refs {
			lr, err := bible.Lookup(ref, cfg.DefaultTranslation)
			if err != nil {
				log.Printf("ERROR %s: %v", ref, err)
				results = append(results, gui.ScriptureResult{
					Reference: fmt.Sprintf("%s", ref),
					Text:      fmt.Sprintf("Error: %v", err),
					Book:      ref.Book,
					Chapter:   ref.StartChapter,
					IsError:   true,
				})
			} else {
				if lr.SourceURL != nil {
					log.Printf("OK [%d] %s → %s", lr.StatusCode, ref, *lr.SourceURL)
				} else {
					log.Printf("OK %s (%d verses)", ref, len(lr.Verses))
				}
				results = append(results, gui.ScriptureResult{
					Reference:  lr.Reference,
					Text:       formatter.Format(lr, fmtOpts),
					Book:       ref.Book,
					Chapter:    ref.StartChapter,
					StartVerse: ref.StartVerse,
					EndVerse:   ref.EndVerse,
				})
			}
		}

		if len(results) > 0 {
			display.ShowResults(results)
		}
	}
}

func main() {
	initConfig()

	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	app := gui.NewApp(lookupChapter)
	app.SetFormatOptions(gui.FormatOptions{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	})

	err = wails.Run(&options.App{
		Title:  "Sword Drill",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)
			go watchClipboard(ctx, app)
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Wails error: %v", err)
	}
}
