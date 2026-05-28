package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		log.Fatalf("Failed to load config: %v", err)
	}

	bible = lookup.NewBibleAPIClient()
	fmtOpts = formatter.Options{
		VerseByVerse:  cfg.FormattingOptions.VerseByVerse,
		ShowVerseNums: cfg.FormattingOptions.ShowVerseNums,
	}
}

func onClipboardChange(text string) {
	fmt.Println("──────────────────────────────────")
	fmt.Printf("[%s] Clipboard changed!\n", time.Now().Format("15:04:05"))
	fmt.Println(text)

	refs := parser.ParseReferences(text)
	if len(refs) == 0 {
		fmt.Println("  (no scripture references detected)")
	} else {
		for _, ref := range refs {
			fmt.Printf("  → %s\n", ref)
			result, err := bible.Lookup(ref, cfg.DefaultTranslation)
			if err != nil {
				fmt.Printf("    ✗ %v\n", err)
			} else {
				fmt.Printf("    %s\n", formatter.Format(result, fmtOpts))
			}
		}
	}
	fmt.Println("──────────────────────────────────")
}

func main() {
	initConfig()

	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize clipboard: %v", err)
	}

	fmt.Println("Sword Drill — Clipboard Monitor")
	fmt.Println("Watching clipboard for changes... (Ctrl+C to quit)")

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
			fmt.Println("\nShutting down.")
			return
		}
	}
}
