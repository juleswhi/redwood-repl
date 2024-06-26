package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	// "github.com/jam-computing/oak/pkg/components"
	"github.com/jam-computing/oak/pkg/menu"
)

func main() {
	m := menu.CreateMenu()

	// m := components.NewPicker()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
