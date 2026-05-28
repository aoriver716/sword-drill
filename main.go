package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sword-drill/lookup"
	"github.com/sword-drill/parser"
	"golang.design/x/clipboard"
)

var bible lookup.BibleLookup = lookup.NewBibleAPIClient()

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
			result, err := bible.Lookup(ref, "kjv")
			if err != nil {
				fmt.Printf("    ✗ %v\n", err)
			} else {
				fmt.Printf("    %s\n", result.Text)
			}
		}
	}
	fmt.Println("──────────────────────────────────")
}

func main() {
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
