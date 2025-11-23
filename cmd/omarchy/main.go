package main

import (
	"fmt"
	"log"
	"os"

	"omarchy-tui/internal/config"
	"omarchy-tui/internal/tui"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Create and initialize TUI application
	app, err := tui.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize TUI: %v", err)
		os.Exit(1)
	}

	// Run the application
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
