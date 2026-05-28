package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aoriver716/sword-drill/gui"
	"github.com/aoriver716/sword-drill/internal/config"
	"github.com/aoriver716/sword-drill/internal/detector"
	"github.com/aoriver716/sword-drill/internal/formatter"
	"github.com/aoriver716/sword-drill/internal/lookup"
	"golang.design/x/clipboard"
)

var (
	cfg     config.Config
	bible   lookup.BibleLookup
	fmtOpts formatter.Options
)

func initConfig() {
	var err error
	cfg, err = config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bible = lookup.NewBibleAPIClient()
	fmtOpts = formatter.Options{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	}
}

func watchClipboard(ctx context.Context, app *gui.App) {
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	for data := range ch {
		if len(data) == 0 {
			continue
		}
		if app.ShouldSkip() {
			log.Println("Skipping self-triggered clipboard event")
			continue
		}
		text := string(data)
		refs := detector.ParseReferences(text)
		for _, ref := range refs {
			result, err := bible.Lookup(ref, cfg.DefaultTranslation)
			if err != nil {
				log.Printf("ERROR %s: %v", ref, err)
				app.Log.Append(fmt.Sprintf("%s", ref), fmt.Sprintf("Error: %v", err))
			} else {
				if result.SourceURL != nil {
					log.Printf("OK [%d] %s → %s", result.StatusCode, ref, *result.SourceURL)
				} else {
					log.Printf("OK %s (%d verses)", ref, len(result.Verses))
				}
				app.Log.Append(result.Reference, formatter.Format(result, fmtOpts))
			}
		}
	}
}

func main() {
	initConfig()

	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	app := gui.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchClipboard(ctx, app)

	app.Run()
}
