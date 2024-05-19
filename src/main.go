package main

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"

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

func SuccessStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("36")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Output = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Title = lipgloss.NewStyle().Bold(true)
	return s
}

func FailureStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("9")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Output = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Title = lipgloss.NewStyle().Bold(true)
	return s
}

type model struct {
	redwood        Redwood
	successCommand bool
	prompt         Prompt
	width          int
	height         int
	answerField    textinput.Model
	successStyles  *Styles
	failureStyles  *Styles
	output         string
}

type Prompt struct {
	prompt string
	answer string
}

func NewPrompt(prompt string) Prompt {
	return Prompt{prompt: prompt}
}

func New(prompt Prompt, redwood Redwood) *model {
	success := SuccessStyles()
	failure := FailureStyles()
	answerField := textinput.New()
	answerField.Placeholder = "Function / Statement"
	answerField.Focus()
	return &model{
		prompt:         prompt,
		answerField:    answerField,
		successStyles:  success,
		failureStyles:  failure,
		redwood:        redwood,
		successCommand: true,
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
		case "ctrl+l":
			m.redwood.Clear(m.redwood.buf1)
			m.output = "Cleared Buffer"
			m.successCommand = true
			return m, nil
		case "ctrl+n":
			// Next Line
		case "ctrl+p":
			// Previous Line
		case "enter":
			current.answer = m.answerField.Value()
			m.redwood.Add(m.answerField.Value(), m.redwood.buf1)
			m.redwood.Add("\n", m.redwood.buf1)
			out, success := m.redwood.Run()
			m.output = out
			if success {
				m.redwood.Clear(m.redwood.buf2)
				contents := m.redwood.Read(m.redwood.buf1)
				m.redwood.Add(contents, m.redwood.buf2)
			} else {
				m.redwood.Clear(m.redwood.buf1)
				contents := m.redwood.Read(m.redwood.buf2)
				m.redwood.Add(contents, m.redwood.buf1)
			}
			m.successCommand = success
			m.answerField.SetValue("")
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

	var styles Styles

	if m.successCommand {
		styles = *m.successStyles
	} else {
		styles = *m.failureStyles
	}

	if len(m.output) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Title.Render(m.prompt.prompt),
			styles.InputField.Render(m.answerField.View()),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Title.Render(m.prompt.prompt),
		styles.InputField.Render(m.answerField.View()),
		styles.Output.Render(m.output),
	)
}

func main() {
	dir1 := "/tmp/e_buf_one.rw"
	dir2 := "/tmp/e_buf_two.rw"

	err := os.WriteFile(dir1, []byte{}, 0644)

	if err != nil {
		log.Fatal("Could not create file, e_buf_one.rw")
	} else {
		log.Info("Created file")
	}

	err = os.WriteFile(dir2, []byte{}, 0644)

	if err != nil {
		log.Fatal("Could not create file, e_buf_two.rw")
	}

	if len(os.Args) > 1 {
		bin_path := os.Args[1]

		if _, err := os.Stat(bin_path); errors.Is(err, os.ErrNotExist) {
			log.Fatal("Could not find binary.")
			return
		}

		rw := Redwood{bin: bin_path, buf1: dir1, buf2: dir2}
		rw.Clear(rw.buf1)
		rw.Clear(rw.buf2)
		StartRepl(rw)
	} else {
		log.Warn("No binary provided. Looking on path.")
	}
}

type Redwood struct {
	bin  string
	buf1 string
	buf2 string
}

func (r *Redwood) Run() (string, bool) {
	cmd := exec.Command(r.bin, r.buf1)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Fatal("Could not run program")
	}

	if len(out) > 0 {
		if out[0] == '|' {
			return string(out), false
		}
	}

	return string(out), true
}

func (r *Redwood) Add(data, buf string) {
	file, err := os.OpenFile(buf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()

	if err != nil {
		log.Fatal("Could not open repl buffer. May be a permissions issue?")
	}

	_, err = io.WriteString(file, data)

	if err != nil {
		log.Fatal("Could not write to repl buffer. May be a permissions issue?")
	}
}

func (r *Redwood) Read(b string) string {
	file, err := os.Open(b)
	defer file.Close()

	if err != nil {
		log.Fatal("Could not open repl buffer. May be a permissions issue?")
	}

	buffer, err := io.ReadAll(file)

	if err != nil {
		log.Fatal("Could not read from repl buffer. Maybe a permissions issue?")
	}

	return string(buffer)
}

func (r *Redwood) Clear(buf string) {
	if _, err := os.Stat(buf); !errors.Is(err, os.ErrNotExist) {
		err := os.Remove(buf)
		if err != nil {
			log.Fatal("Could not access temp dir for repl buffer.")
			return
		}
	}
	err := os.WriteFile(buf, []byte{}, 0644)
	if err != nil {
		log.Fatal("Could not access temp dir for repl buffer.")
	}
}

func StartRepl(rw Redwood) {
	m := New(NewPrompt("Redwood Interactive REPL"), rw)
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
