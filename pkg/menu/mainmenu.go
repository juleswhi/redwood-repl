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

	list   *components.ListModel
	repl   *repl.ReplModel
	picker *filepicker.Model

	focus         bool
	loaded        bool
	pickingConfig bool

	popup         string
	popupselected bool
	popupyes      lipgloss.Style
	popupno       lipgloss.Style
	popping       bool
	popupStyle    lipgloss.Style

	focusStyle   lipgloss.Style
	unfocusStyle lipgloss.Style
}

func GetFilesFromServer() {

}

func CreateMenu() MenuModel {
	focus, unfocus := CreateStyles()
	r := repl.NewReplModel(repl.Init())

	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.CurrentDirectory, _ = os.Getwd()

	pstyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Bold(true)
	focusColor := lipgloss.Color("36")
	unfocusColor := lipgloss.Color("9")
	pyes := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Bold(true).Width(10).BorderForeground(focusColor).Align(lipgloss.Center)
	pno := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Bold(true).Width(10).BorderForeground(unfocusColor).Align(lipgloss.Center)

	if r == nil {
		panic("Totes could not create menu")
	}

	return MenuModel{
		width:         0,
		height:        0,
		list:          components.NewListModel(),
		repl:          r,
		picker:        &fp,
		focusStyle:    focus,
		unfocusStyle:  unfocus,
		focus:         true,
		loaded:        false,
		pickingConfig: true,
		popping:       true,
		popup:         "Would you like to contact the server?",
		popupStyle:    pstyle,
		popupyes:      pyes,
		popupno:       pno,
        popupselected: true,
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

	if m.popping {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y":
                m.popupselected = true
                m.popping = false
                GetFilesFromServer()
                return m, nil
            case "n":
                m.popupselected = false
                m.popping = false
                GetFilesFromServer()
                return m, nil
            case "h":
                m.popupselected = !m.popupselected
            case "l":
                m.popupselected = !m.popupselected
            case "enter":
                m.popping = false
                GetFilesFromServer()
                return m, nil
			}
		}
	}

	if (!m.focus || !m.loaded) && !m.popping {
		mlist, c := m.list.Update(msg)

		l := mlist.(components.ListModel)
		m.list = &l
		cmd = c
	}

	if ((m.focus && !m.pickingConfig) || !m.loaded) && !m.popping {
		mrepl, c := m.repl.Update(msg)

		r := mrepl.(repl.ReplModel)
		m.repl = &r
		cmd = c
	}

	if (m.focus && m.pickingConfig) || !m.loaded {
		mpicker, c := m.picker.Update(msg)
		m.picker = &mpicker

		cmd = c

		selected, st := m.picker.DidSelectFile(msg)

		if selected {
			m.pickingConfig = false
			m.repl.Redwood.LedConfig = st
		}
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
			repl = m.unfocusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.picker.View())
		} else {
			repl = m.unfocusStyle.Height(m.height - 2).Width((m.width / 2) - 2).Render(m.repl.View())
		}
	}

	panes := lipgloss.JoinHorizontal(
		lipgloss.Left,
		repl,
		list,
	)

    var yes lipgloss.Style
    var no lipgloss.Style

    if m.popupselected {
        yes = m.popupyes
        no = m.popupno
    } else {
        yes = m.popupno
        no = m.popupyes
    }

	if m.popping {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(
				lipgloss.Center,
				m.popupStyle.Render(m.popup),
				lipgloss.JoinHorizontal(
					lipgloss.Center,
					yes.Render("Yes"),
					no.Render("No"),
				),
			),
		)
	}
	return lipgloss.JoinVertical(
		lipgloss.Center,
		panes,
	)
}
