package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const Release = 1

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "sqview rel %d - SQLite table viewer\n", Release)
		fmt.Fprintf(os.Stderr, "Usage: %s <database.db>\n", os.Args[0])
		os.Exit(1)
	}

	dbPath := os.Args[1]

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", dbPath)
		os.Exit(1)
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	model := NewModel(db)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
