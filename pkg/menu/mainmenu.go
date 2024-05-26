package menu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jam-computing/oak/pkg/components"
	"github.com/jam-computing/oak/pkg/repl"
)

type MenuModel struct {
	width  int
	height int

	list components.ListModel
	repl repl.ReplModel
	help HelpModel

	focus  bool
	loaded bool

	focusStyle   lipgloss.Style
	unfocusStyle lipgloss.Style
}

func CreateMenu() MenuModel {
	focus, unfocus := CreateStyles()
	r := repl.NewReplModel(repl.Init())

	if r == nil {
		panic("Totes could not create menu")
	}

	return MenuModel{
		width:        0,
		height:       0,
		list:         *components.NewListModel(),
		repl:         *r,
		help:         NewHelp(),
		focusStyle:   focus,
		unfocusStyle: unfocus,
		focus:        true,
		loaded:       false,
	}
}

func CreateStyles() (lipgloss.Style, lipgloss.Style) {
	focusColor := lipgloss.Color("36")
	unfocusColor := lipgloss.Color("9")
	focus := lipgloss.NewStyle().BorderForeground(focusColor).BorderStyle(lipgloss.RoundedBorder())
	unfocus := lipgloss.NewStyle().BorderForeground(unfocusColor).BorderStyle(lipgloss.RoundedBorder())

	return focus, unfocus
}

func (m MenuModel) Init() tea.Cmd {
	m.list.Init()
	m.repl.Init()
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if m.focus || !m.loaded {
		mlist, cmd := m.list.Update(msg)

		m.list = mlist.(components.ListModel)

		if cmd != nil {
			return m, cmd
		}
	}

	if !m.focus || !m.loaded {
		mrepl, cmd := m.repl.Update(msg)

		m.repl = mrepl.(repl.ReplModel)

		if cmd != nil {
			return m, cmd
		}
	}

    m.loaded = true

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
        case "ctrl+n":
            m.focus = !m.focus
            return m, nil
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	var list string
	var repl string

	if m.focus {
		list = m.focusStyle.Height(m.height - 2).Render(m.list.View())
		repl = m.unfocusStyle.Height(m.height - 2).Render(m.repl.View())
	} else {
		list = m.unfocusStyle.Height(m.height - 2).Render(m.list.View())
		repl = m.focusStyle.Height(m.height - 2).Render(m.repl.View())
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			list,
			repl,
		),
	)
}
