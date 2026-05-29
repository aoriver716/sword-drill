package main

import (
	"context"
	"embed"
	"fmt"
	"log"

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
	registry *config.Registry
	bible    lookup.BibleLookup
)

func initConfig() {
	registry = config.NewRegistry("config.json")
	config.RegisterFields(registry)

	if err := registry.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bible = registry.BibleLookup()
}

func lookupChapter(book string, chapter int, translation string) ([]gui.ChapterVerse, error) {
	ref := detector.ScriptureRef{
		Book:         book,
		StartChapter: chapter,
		EndChapter:   chapter,
	}
	result, err := bible.Lookup(ref, translation)
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
			lr, err := bible.Lookup(ref, registry.Config().DefaultTranslation)
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
					Text:       formatter.Format(lr, registry.Config()),
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

	app := gui.NewApp(lookupChapter, registry)

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
		OnBeforeClose: app.BeforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Wails error: %v", err)
	}
}
