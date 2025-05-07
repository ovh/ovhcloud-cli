package display

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func RunLoginInput() map[string]string {
	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "aaaaaaaaaaaaaaaa"
	inputs[0].Focus()
	inputs[0].CharLimit = 16
	inputs[0].Width = 30
	inputs[0].Prompt = ""

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	inputs[1].CharLimit = 32
	inputs[1].Width = 30
	inputs[1].Prompt = ""

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"
	inputs[2].CharLimit = 32
	inputs[2].Width = 30
	inputs[2].Prompt = ""

	mod := inputModel{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}

	p := tea.NewProgram(mod)

	if _, err := p.Run(); err != nil {
		ExitError(err.Error())
	}

	return map[string]string{
		"application_key":    mod.inputs[0].Value(),
		"application_secret": mod.inputs[1].Value(),
		"consumer_key":       mod.inputs[2].Value(),
	}
}

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type inputModel struct {
	inputs  []textinput.Model
	focused int
	err     error
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, tea.Quit
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		`
 %s
 %s

 %s
 %s

 %s
 %s

 %s
`,
		inputStyle.Width(30).Render("Application key"),
		m.inputs[0].View(),
		inputStyle.Width(30).Render("Application secret"),
		m.inputs[1].View(),
		inputStyle.Width(30).Render("Consumer key"),
		m.inputs[2].View(),
		continueStyle.Render("Press enter to validate ->"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *inputModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *inputModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
