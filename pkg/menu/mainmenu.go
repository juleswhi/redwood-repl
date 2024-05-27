package menu

import (
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jam-computing/oak/pkg/components"
	"github.com/jam-computing/oak/pkg/repl"
)

type MenuModel struct {
	width  int
	height int

	list   components.ListModel
	repl   repl.ReplModel
	picker filepicker.Model
	help   HelpModel

	focus         bool
	loaded        bool
	pickingConfig bool

	focusStyle   lipgloss.Style
	unfocusStyle lipgloss.Style
}

func CreateMenu() MenuModel {
	focus, unfocus := CreateStyles()
	r := repl.NewReplModel(repl.Init())

    fp := filepicker.New()
    fp.ShowHidden = true
    fp.CurrentDirectory, _ = os.Getwd()

	if r == nil {
		panic("Totes could not create menu")
	}

	return MenuModel{
		width:         0,
		height:        0,
		list:          *components.NewListModel(),
		repl:          *r,
		picker:        fp,
		help:          NewHelp(),
		focusStyle:    focus,
		unfocusStyle:  unfocus,
		focus:         true,
		loaded:        false,
		pickingConfig: true,
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
	return m.picker.Init()
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

	if !m.focus || !m.loaded {
		mlist, c := m.list.Update(msg)

		m.list = mlist.(components.ListModel)
        cmd = c
	}

	if (m.focus && !m.pickingConfig) || !m.loaded {
		mrepl, c := m.repl.Update(msg)

		m.repl = mrepl.(repl.ReplModel)

        cmd = c
	}

	if (m.focus && m.pickingConfig) {
		mpicker, c := m.picker.Update(msg)
		m.picker = mpicker

        cmd = c
	}

	m.loaded = true

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.Update(msg)
		m.repl.Update(msg)
		m.picker.Update(msg)
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
	return m, cmd
}

func (m MenuModel) View() string {
	var list string
	var repl string

	if m.focus {
		list = m.unfocusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.list.View())
		if m.pickingConfig {
			repl = m.focusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.picker.View())
		} else {
			repl = m.focusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.repl.View())
		}
	} else {
		list = m.focusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.list.View())
		if m.pickingConfig {
			repl = m.focusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.picker.View())
		} else {
			repl = m.focusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.repl.View())
		}
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		repl,
		list,
	)
}
