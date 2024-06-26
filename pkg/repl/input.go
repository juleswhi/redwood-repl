package repl

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor lipgloss.Color
	InputField  lipgloss.Style
	Output      lipgloss.Style
}

func SuccessStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("#3C3C3C")
	s.InputField = lipgloss.NewStyle()
	s.Output = lipgloss.NewStyle()
	return s
}

func FailureStyles() *Styles {
	s := new(Styles)
	s.BorderColor = lipgloss.Color("9")
	s.InputField = lipgloss.NewStyle()
	s.Output = lipgloss.NewStyle()
	return s
}

type ReplModel struct {
	ConfigFile     string
	Redwood        Redwood
	successCommand bool
	width          int
	height         int
	answerField    textinput.Model
	successStyles  *Styles
	failureStyles  *Styles
	cmds           []*RedwoodCmd
}

type RedwoodCmd struct {
	input  string
	output string
}

func NewReplModel(redwood *Redwood) *ReplModel {
	if redwood == nil {
		return nil
	}
	success := SuccessStyles()
	failure := FailureStyles()
	answerField := textinput.New()
	answerField.Placeholder = "Statement..."
	answerField.Focus()
	return &ReplModel{
		answerField:    answerField,
		successStyles:  success,
		failureStyles:  failure,
		Redwood:        *redwood,
		successCommand: true,
	}
}

func (m ReplModel) Init() tea.Cmd {
    return nil
}

func (m ReplModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = (msg.Width / 2) - 2
		m.height = msg.Height - 2
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			m.Redwood.Clear(m.Redwood.Buf1)
			m.cmds = nil
			m.successCommand = true
			return m, nil
		case "ctrl+r":
			return m, nil
		case "enter":
			m.Redwood.Add(m.answerField.Value(), m.Redwood.Buf1)
			m.Redwood.Add("\n", m.Redwood.Buf1)
			out, success := m.Redwood.Run()
			c := RedwoodCmd{input: m.answerField.Value(), output: out}
			m.cmds = append(m.cmds, &c)
			if success {
				m.Redwood.Clear(m.Redwood.Buf2)
				contents := m.Redwood.Read(m.Redwood.Buf1)
				m.Redwood.Add(contents, m.Redwood.Buf2)
			} else {
				m.Redwood.Clear(m.Redwood.Buf1)
				contents := m.Redwood.Read(m.Redwood.Buf2)
				m.Redwood.Add(contents, m.Redwood.Buf1)
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
	var styles Styles

	if m.successCommand {
		styles = *m.successStyles
	} else {
		styles = *m.failureStyles
	}

	outputs := make([]string, len(m.cmds))

	for i, cmd := range m.cmds {
		combined := "> " + cmd.input + "\n" + cmd.output
		outputs[i] = combined
	}

	out := strings.Join(outputs, "")

	if len(outputs) == 0 {
		out = "Redwood Compiler"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Output.Render(out),
		styles.InputField.Width(m.width).Render(m.answerField.View()),
	)
}

func Init() *Redwood {
	dir1 := "/tmp/e_buf_one.rw"
	dir2 := "/tmp/e_buf_two.rw"

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

		rw := Redwood{Bin: bin_path, Buf1: dir1, Buf2: dir2}
		return &rw
	} else {
		path := "/home/juleswhite/projects/redwood/zig-out/bin/redwood"

		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Fatal("Could not find binary.")
			return nil
		}

		rw := Redwood{Bin: path, Buf1: dir1, Buf2: dir2}
		return &rw
	}
}

type Redwood struct {
	Bin        string
	Buf1       string
	Buf2       string
	LedConfig string
}

func (r *Redwood) Run() (string, bool) {
	cmd := exec.Command(r.Bin, r.Buf1, r.LedConfig)

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
