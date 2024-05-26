package repl

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
	s.BorderColor = lipgloss.Color("#3C3C3C")
    s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.DoubleBorder()).Width(40)
	s.Output = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#3C3C3C")).BorderStyle(lipgloss.DoubleBorder()).Width(60)
	s.Title = lipgloss.NewStyle().Bold(true)
	return s
}

func FailureStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("9")
	s.InputField = lipgloss.NewStyle().BorderForeground(s.BorderColor).BorderStyle(lipgloss.RoundedBorder()).Width(60)
	s.Output = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#3C3C3C")).BorderStyle(lipgloss.DoubleBorder()).Width(60)
	s.Title = lipgloss.NewStyle().Bold(true)
	return s
}

type ReplModel struct {
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

func NewReplModel(redwood *Redwood) *ReplModel {
    if redwood == nil {
        return nil
    }
	success := SuccessStyles()
	failure := FailureStyles()
	answerField := textinput.New()
	answerField.Placeholder = "Redwood Statement"
	answerField.Focus()
	return &ReplModel{
		prompt:         NewPrompt(""),
		answerField:    answerField,
		successStyles:  success,
		failureStyles:  failure,
		redwood:        *redwood,
		successCommand: true,
	}
}

func (m ReplModel) Init() tea.Cmd {
	return nil
}

func (m ReplModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "ctrl+r":
			return m, nil
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

func (m ReplModel) View() string {
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
			lipgloss.Left,
			styles.InputField.Render(m.answerField.View()),
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputField.Render(m.answerField.View()),
		styles.Output.Render(m.output),
	)
}

func Init() *Redwood {
	dir1 := "/tmp/e_buf_one.rw"
	dir2 := "/tmp/e_buf_two.rw"

	config := "led.json"

	err := os.WriteFile(dir1, []byte{}, 0644)

	if err != nil {
		log.Fatal("Could not create file, e_buf_one.rw")
	}

	err = os.WriteFile(dir2, []byte{}, 0644)

	if err != nil {
		log.Fatal("Could not create file, e_buf_two.rw")
	}

	if len(os.Args) > 1 {
		bin_path := os.Args[1]

		if _, err := os.Stat(bin_path); errors.Is(err, os.ErrNotExist) {
			log.Fatal("Could not find binary.")
			return nil
		}

		rw := Redwood{bin: bin_path, buf1: dir1, buf2: dir2, led_config: config}
        return &rw
	} else {
		path := "/home/juleswhite/projects/redwood/zig-out/bin/redwood"

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Fatal("Could not find binary.")
			return nil
		}

		rw := Redwood{bin: path, buf1: dir1, buf2: dir2, led_config: config}
        return &rw
	}
}


type Redwood struct {
	bin        string
	buf1       string
	buf2       string
	led_config string
}

func (r *Redwood) Run() (string, bool) {
	cmd := exec.Command(r.bin, r.buf1, r.led_config)

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

