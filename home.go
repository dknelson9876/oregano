package main

import tea "github.com/charmbracelet/bubbletea"

type homeModel struct {
	content string
	ready   bool
}

func NewHomeModel() homeModel {
	return homeModel{"hello world", false}
}

func (m homeModel) Init() tea.Cmd {
	return nil
}

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd

	// switch msg := msg.(type) {
	// case tea.KeyMsg:
	// 	// if msg.String() == "esc" {
	// 	// 	return m, tea.Quit
	// 	// }
	// }

	return m, tea.Batch(cmds...)
}

func (m homeModel) View() string {
	return "//"+m.content
}