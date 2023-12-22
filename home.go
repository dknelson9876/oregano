package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	acctStyle   = lipgloss.NewStyle()
	budgetStyle = lipgloss.NewStyle()
	transStyle  = lipgloss.NewStyle()
)

type homeModel struct {
	content     string
	ready       bool
	width       int
	height      int
	sizeCounter int
}

func NewHomeModel() homeModel {
	return homeModel{
		content: "hello world",
		ready:   false,
	}
}

func (m homeModel) Init() tea.Cmd {
	return nil
}

func (m homeModel) Update(msg tea.Msg) (homeModel, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "right" {
			m.width++
		} else if msg.String() == "left" {
			m.width--
		} else if msg.String() == "up" {
			m.height--
		} else if msg.String() == "down" {
			m.height++
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *homeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.sizeCounter++
}

func verticalLine(count int) string {
	return strings.TrimSpace(strings.Repeat("|\n", count))
}

func horizontalLine(count int) string {
	return strings.Repeat("-", count)
}

func (m homeModel) View() string {
	acctWidth := m.width / 3 - 1 // -1 to accoutn for border line
	budgetWidth := m.width - acctWidth
	transWidth := budgetWidth

	acctHeight := m.height - 3
	budgetHeight := m.height / 2 - 3
	transHeight := m.height - budgetHeight - 3 // -1 to account for border line

	acctView := acctStyle.Width(acctWidth).Height(acctHeight).Render(fmt.Sprintf("//TODO accounts... sizeCount %d", m.sizeCounter))
	budgetView := budgetStyle.Width(budgetWidth).Height(budgetHeight).Render(fmt.Sprintf("//TODO bugets... local width %d", m.width))
	transView := transStyle.Width(transWidth).Height(transHeight).Render(fmt.Sprintf("//TODO transactions... local height %d", m.height))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		acctView,
		verticalLine(acctHeight),
		lipgloss.JoinVertical(lipgloss.Left,
			budgetView,
			horizontalLine(budgetWidth),
			transView),
	)

	// return "/// TODOO......"
}
