package components

import (
	"encoding/json"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jam-computing/oak/pkg/tcp"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func GetItems() []list.Item {
	packet := tcp.NewFullPacket(tcp.NewMetaPacket(), nil, nil)
	packet.Meta.Status = 200
	packet.Meta.Command = 4
	recv := packet.SendRecv()

	jsonData := recv.Data.Data[:recv.Meta.Len]
	var animations []tcp.Animation
	err := json.Unmarshal([]byte(jsonData), &animations)

	if err != nil {
		os.Exit(1)
	}

	numItems := len(animations)

	items := make([]list.Item, numItems)

	for i := 0; i < numItems; i++ {
		items[i] = item{
			title: animations[i].Title,
			desc:  animations[i].Artist,
		}
	}

	return items
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type ListModel struct {
	List list.Model
}

func NewListModel() *ListModel {
	items := GetItems()

	m := ListModel{List: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.List.Title = "Animations"

	return &m
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width, msg.Height-(v*2))
        return m, nil
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return docStyle.Render(m.List.View())
}
