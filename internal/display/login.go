// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package display

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func RunLoginInput(customEndpoint bool) map[string]string {
	var inputs []textinput.Model

	if customEndpoint {
		endpointInput := textinput.New()

		endpointInput.Placeholder = "https://eu.api.ovh.com/v1"
		endpointInput.Focus()
		endpointInput.Width = 60
		endpointInput.Prompt = ""

		inputs = append(inputs, endpointInput)
	}

	appKeyInput := textinput.New()
	appKeyInput.Placeholder = "aaaaaaaaaaaaaaaa"
	if !customEndpoint {
		appKeyInput.Focus()
	}
	appKeyInput.CharLimit = 16
	appKeyInput.Width = 32
	appKeyInput.Prompt = ""
	inputs = append(inputs, appKeyInput)

	appSecretInput := textinput.New()
	appSecretInput.Placeholder = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	appSecretInput.CharLimit = 32
	appSecretInput.Width = 32
	appSecretInput.Prompt = ""
	inputs = append(inputs, appSecretInput)

	consumerKeyInput := textinput.New()
	consumerKeyInput.Placeholder = "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"
	consumerKeyInput.CharLimit = 32
	consumerKeyInput.Width = 32
	consumerKeyInput.Prompt = ""
	inputs = append(inputs, consumerKeyInput)

	mod := inputModel{
		withCustomEndpoint: customEndpoint,
		inputs:             inputs,
		focused:            0,
		err:                nil,
	}

	p := tea.NewProgram(mod)

	if _, err := p.Run(); err != nil {
		exitError(err.Error())
	}

	if customEndpoint {
		return map[string]string{
			"endpoint":           mod.inputs[0].Value(),
			"application_key":    mod.inputs[1].Value(),
			"application_secret": mod.inputs[2].Value(),
			"consumer_key":       mod.inputs[3].Value(),
		}
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
	withCustomEndpoint bool
	inputs             []textinput.Model
	focused            int
	err                error
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
	if m.withCustomEndpoint {
		return fmt.Sprintf(
			`
 %s
 %s

 %s
 %s

 %s
 %s

 %s
 %s

 %s
`,
			inputStyle.Width(30).Render("API endpoint"),
			m.inputs[0].View(),
			inputStyle.Width(30).Render("Application key"),
			m.inputs[1].View(),
			inputStyle.Width(30).Render("Application secret"),
			m.inputs[2].View(),
			inputStyle.Width(30).Render("Consumer key"),
			m.inputs[3].View(),
			continueStyle.Render("Press enter to validate ->"),
		) + "\n"
	}

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
