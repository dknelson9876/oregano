package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles
var (
	borderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))
	focusedBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("87"))

	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// Define list items
type item string

// TODO find the purpose of FilterValue and itemDelegate for lists
func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Begin mainModel stuff
type mainModel struct {
	ready bool
	// content  string
	// viewport viewport.Model
	navList   list.Model
	content   tea.Model
	focusList bool
}

func newModel() mainModel {
	return mainModel{
		ready:     false,
		focusList: true,
	}
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		} else if msg.String() == " " {
			m.focusList = !m.focusList
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			items := []list.Item{
				item("Home"),
				item("Transactions"),
			}
			m.navList = list.New(items, itemDelegate{}, msg.Width/2, msg.Height-4)
			m.navList.SetShowTitle(false)
			m.navList.SetShowHelp(false)
			m.navList.SetShowFilter(false)
			m.navList.SetFilteringEnabled(false)
			m.navList.SetShowStatusBar(false)
			m.content = NewHomeModel()
			m.ready = true
		} else {
			m.navList.SetSize(msg.Width/4, msg.Height-4)
		}
	}

	// m.viewport, cmd = m.viewport.Update(msg)
	if m.focusList {
		m.navList, cmd = m.navList.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.content, cmd = m.content.Update(msg)
		cmds = append(cmds, cmd)
	}

	//TODO else send update to content

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	if !m.ready {
		return "\n Initializing...."
	}
	// return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	var navListStr string
	var contentStr string
	if m.focusList {
		navListStr = focusedBorderStyle.Render(m.navList.View())
		contentStr = borderStyle.Render(m.content.View())
	} else {
		navListStr = borderStyle.Render(m.navList.View())
		contentStr = focusedBorderStyle.Render(m.content.View())
	}

	// navListStr = m.navList.View()
	return lipgloss.JoinHorizontal(lipgloss.Top,
		navListStr,
		contentStr)
	// return navListStr

	//┌─┬─┬───┐
	//│ │ │ C │
	//│A│B├───┤
	//│ │ │ D │
	//└─┴─┴───┘
	// a.height = height
	// b.height = height
	// c.height = height/2
	// d.height = height/2
	// a.width = width/4
	// b.width = width/4
	// c.widht = width/2
	// d.width = width/2

}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
