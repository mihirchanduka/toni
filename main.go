package main

import (
	"fmt"
	"os"

	"toni/cmd"
	"toni/internal/db"
	"toni/internal/search"
	"toni/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Parse CLI flags
	config, err := cmd.ParseFlags(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize Yelp client
	var yelpClient *search.YelpClient
	if config.YelpAPIKey != "" {
		yelpClient = search.NewYelpClient(config.YelpAPIKey)
	} else if !config.YelpEnabled {
		fmt.Fprintln(os.Stderr, "ℹ  Yelp autocomplete disabled in onboarding settings")
	} else {
		fmt.Fprintln(os.Stderr, "ℹ  No YELP_API_KEY set — restaurant autocomplete disabled")
	}

	// Detect terminal capabilities
	termCaps := ui.DetectTerminalCapabilities()

	// Open database
	database, err := db.Open(config.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create and run Bubble Tea app
	p := tea.NewProgram(ui.New(database, yelpClient, termCaps), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
