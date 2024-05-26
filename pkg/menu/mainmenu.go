package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuModel struct {
	width  int
	height int
	help   HelpModel
}

func CreateMenu() MenuModel {
	return MenuModel{
		width:  0,
		height: 0,
		help:   NewHelp(),
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

    mhelp, cmd := m.help.Update(msg)

    m.help = mhelp.(HelpModel)

    if cmd != nil {
        return m, cmd
    }


	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	return lipgloss.PlaceVertical(
		m.height,
		lipgloss.Bottom,
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.help.View(),
		),
	)
}
