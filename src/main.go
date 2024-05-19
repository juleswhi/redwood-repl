package main

import (
	"log"
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
	Title       lipgloss.Style
	Output      lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Output = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Title = lipgloss.NewStyle().Bold(true)
	return s
}

type model struct {
	prompt      Prompt
	width       int
	height      int
	answerField textinput.Model
	styles      *Styles
	output      string
}

type Prompt struct {
	prompt string
	answer string
}

func NewPrompt(prompt string) Prompt {
	return Prompt{prompt: prompt}
}

func New(prompt Prompt) *model {
	styles := DefaultStyles()
	answerField := textinput.New()
	answerField.Placeholder = "Your Answer Here"
	answerField.Focus()
	return &model{
		prompt:      prompt,
		answerField: answerField,
		styles:      styles,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	current := &m.prompt
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			current.answer = m.answerField.Value()
			m.answerField.SetValue("")
			m.output = "just repled"
			return m, nil
		}
	}

	m.answerField, cmd = m.answerField.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if len(m.output) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.Title.Render(m.prompt.prompt),
			m.styles.InputField.Render(m.answerField.View()),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.styles.Title.Render(m.prompt.prompt),
		m.styles.InputField.Render(m.answerField.View()),
		m.styles.Output.Render(m.output),
	)
}

func main() {

	redwood_name := "redwood"

	path, err := exec.LookPath(redwood_name)
	if err != nil {
		// log.Printf("%s is not installed on path\n", redwood_name)
	} else {
        log.Printf("%s is installed at %s\n", redwood_name, path)
    }

	m := New(NewPrompt("Redwood Interactive REPL"))
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
