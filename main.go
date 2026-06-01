package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

// initDataDir sets the working directory to the application's data directory.
// On macOS, apps launched from Finder have CWD set to "/" which breaks relative
// file paths (config.json, tabs.json, etc.). This ensures a consistent location.
func initDataDir() {
	// If wails.json exists in CWD, we're running from the project directory
	// (development or build step) — don't relocate.
	if _, err := os.Stat("wails.json"); err == nil {
		return
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		// Fall back to home directory
		dir, err = os.UserHomeDir()
		if err != nil {
			return
		}
	}

	dataDir := filepath.Join(dir, "sword-drill")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("WARNING: could not create data directory %s: %v", dataDir, err)
		return
	}

	if err := os.Chdir(dataDir); err != nil {
		log.Printf("WARNING: could not change to data directory %s: %v", dataDir, err)
	}
}

func buildAppMenu(app *gui.App) *menu.Menu {
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New Tab", keys.CmdOrCtrl("n"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.Ctx(), "menu:new-tab")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Preferences…", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.Ctx(), "menu:preferences")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.Ctx(), "menu:quit")
	})

	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("About Sword Drill", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.Ctx(), "menu:about")
	})

	return appMenu
}

func main() {
	initDataDir()
	initConfig()

	app := gui.NewApp(lookupChapter, registry)

	err := wails.Run(&options.App{
		Title:  "Sword Drill",
		Width:  800,
		Height: 600,
		Menu:   buildAppMenu(app),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.Startup(ctx)

			// Initialize clipboard after the application event loop is running.
			// On macOS, clipboard access requires the NSApplication loop which
			// Wails sets up before calling OnStartup.
			if err := clipboard.Init(); err != nil {
				log.Printf("WARNING: clipboard init failed: %v", err)
				return
			}
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
