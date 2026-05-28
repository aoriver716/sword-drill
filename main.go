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
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/clipboard"
)

//go:embed gui/frontend/*
var assets embed.FS

var (
	cfg   config.Config
	bible lookup.BibleLookup
)

func initConfig() {
	var err error
	cfg, err = config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	bible = lookup.NewBibleAPIClient()
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

		type result struct {
			ref    detector.ScriptureRef
			lookup lookup.LookupResult
			err    error
		}
		var results []result

		for _, ref := range refs {
			lr, err := bible.Lookup(ref, cfg.DefaultTranslation)
			if err != nil {
				log.Printf("ERROR %s: %v", ref, err)
			} else {
				if lr.SourceURL != nil {
					log.Printf("OK [%d] %s → %s", lr.StatusCode, ref, *lr.SourceURL)
				} else {
					log.Printf("OK %s (%d verses)", ref, len(lr.Verses))
				}
			}
			results = append(results, result{ref: ref, lookup: lr, err: err})
		}

		for _, r := range results {
			if r.err != nil {
				app.AppendLogError(fmt.Sprintf("%s", r.ref), r.err)
			} else {
				app.AppendLog(r.lookup)
			}
		}

		seen := make(map[string]bool)
		for _, r := range results {
			if r.err != nil {
				continue
			}
			tabName := fmt.Sprintf("%s %d", r.ref.Book, r.ref.StartChapter)
			if seen[tabName] || app.HasTab(tabName) {
				continue
			}
			seen[tabName] = true

			chapterRef := detector.ScriptureRef{
				Book:         r.ref.Book,
				StartChapter: r.ref.StartChapter,
				EndChapter:   r.ref.StartChapter,
			}
			chResult, err := bible.Lookup(chapterRef, cfg.DefaultTranslation)
			if err != nil {
				log.Printf("ERROR chapter lookup %s: %v", tabName, err)
				continue
			}
			log.Printf("OK [%d] chapter %s → %s", chResult.StatusCode, tabName, *chResult.SourceURL)
			app.OpenTab(tabName, chResult)
		}
	}
}

func main() {
	initConfig()

	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	app := gui.NewApp()
	app.SetFormatOptions(formatter.Options{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	})

	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.Quit(app.Ctx())
	})

	err = wails.Run(&options.App{
		Title:  "Sword Drill",
		Width:  800,
		Height: 600,
		Menu:   appMenu,
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
