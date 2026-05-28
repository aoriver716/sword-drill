package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sword-drill/config"
	"github.com/sword-drill/formatter"
	"github.com/sword-drill/lookup"
	"github.com/sword-drill/parser"
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
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	bible = lookup.NewBibleAPIClient()
	fmtOpts = formatter.Options{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	}
}

func onClipboardChange(text string) {
	slog.Info("Clipboard changed", "text", text)

	refs := parser.ParseReferences(text)
	if len(refs) == 0 {
		slog.Debug("No scripture references detected")
		return
	}

	for _, ref := range refs {
		slog.Info("Scripture reference detected", "reference", ref.String())

		result, err := bible.Lookup(ref, cfg.DefaultTranslation)
		if err != nil {
			slog.Error("Lookup failed", "reference", ref.String(), "error", err)
			continue
		}

		attrs := []any{
			"reference", result.Reference,
			"translation", result.Translation,
			"text", formatter.Format(result, fmtOpts),
		}
		if result.SourceURL != nil {
			attrs = append(attrs, "source_url", *result.SourceURL)
		}
		slog.Info("Scripture text retrieved", attrs...)
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	initConfig()

	err := clipboard.Init()
	if err != nil {
		slog.Error("Failed to initialize clipboard", "error", err)
		os.Exit(1)
	}

	slog.Info("Sword Drill started", "translation", cfg.DefaultTranslation, "api", cfg.BibleTextAPI)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := clipboard.Watch(ctx, clipboard.FmtText)

	// Graceful shutdown on Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case data := <-ch:
			if len(data) > 0 {
				onClipboardChange(string(data))
			}
		case <-sig:
			slog.Info("Shutting down")
			return
		}
	}
}
