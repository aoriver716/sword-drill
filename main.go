package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.design/x/clipboard"
)

func onClipboardChange(text string) {
	fmt.Println("──────────────────────────────────")
	fmt.Printf("[%s] Clipboard changed!\n", time.Now().Format("15:04:05"))
	fmt.Println(text)
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
