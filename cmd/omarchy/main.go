package main

import (
	"fmt"
	"log"
	"os"

	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"
	"omarchy-tui/internal/tui"
)

func main() {
	// Initialize logger
	if err := logger.Init("./app.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Log("Application starting")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log("Failed to load configuration: %v", err)
		log.Fatalf("Failed to load configuration: %v", err)
		os.Exit(1)
	}
	logger.Log("Configuration loaded successfully")

	// Create and initialize TUI application
	app, err := tui.NewApp(cfg)
	if err != nil {
		logger.Log("Failed to initialize TUI: %v", err)
		log.Fatalf("Failed to initialize TUI: %v", err)
		os.Exit(1)
	}
	logger.Log("TUI application initialized")

	// Run the application
	logger.Log("Starting application event loop")
	if err := app.Run(); err != nil {
		logger.Log("Application error: %v", err)
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}

	logger.Log("Application exited normally")
}
