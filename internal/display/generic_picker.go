package display

import (
	"fmt"
	"maps"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type genericModel struct {
	list     list.Model
	choice   string
	quitting bool
	choices  map[string]string
}

func (m genericModel) Init() tea.Cmd {
	return nil
}

func (m *genericModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m genericModel) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s selected.", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Nothing selected.")
	}

	listAndValue := "\n" + m.list.View()

	var currentValue string
	i, ok := m.list.SelectedItem().(item)
	if ok {
		currentValue = helpStyle.Render("\n" + m.choices[string(i)])
	}

	return listAndValue + currentValue
}

func RunGenericChoicePicker(question string, choices map[string]string) (string, string, error) {
	var items []list.Item
	for _, key := range slices.Sorted(maps.Keys(choices)) {
		items = append(items, item(key))
	}

	const defaultWidth = 20
	const listHeight = 10

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = question
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := &genericModel{
		list:    l,
		choices: choices,
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return "", "", fmt.Errorf("failed to run picker: %w", err)
	}

	var chosenValue string
	if m.choice != "" {
		chosenValue = choices[m.choice]
	}

	return m.choice, chosenValue, nil
}
