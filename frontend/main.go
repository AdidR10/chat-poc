package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	appBus = NewBus()

	// Initialize the TUI program with the initial model
	// WithAltScreen() creates a full-screen terminal application
	p := tea.NewProgram(initialModel(appBus), tea.WithAltScreen())
	
	// Set global program reference for streaming
	globalProgram = p
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("❌ Error starting chat application: %v\n", err)
		os.Exit(1)
	}
}