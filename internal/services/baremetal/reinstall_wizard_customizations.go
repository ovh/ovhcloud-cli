// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package baremetal

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const customSSHKeyChoice = "Enter a custom key"

func (m *reinstallWizardModel) fetchOsDetailsCmd(osName string) tea.Cmd {
	return func() tea.Msg {
		details, err := fetchOsDetails(osName)
		if err != nil {
			return osDetailsLoadedMsg{err: err}
		}

		var sshKeys []string
		if slices.ContainsFunc(details.Inputs, func(input osInput) bool {
			return input.Type == "sshPubKey"
		}) {
			// A missing SSH key list is not worth failing the whole wizard for:
			// the user can always type a key by hand.
			sshKeys, _ = fetchSSHKeyNames()
		}

		return osDetailsLoadedMsg{details: details, sshKeys: sshKeys}
	}
}

func (m *reinstallWizardModel) fetchSSHKeyContentCmd(keyName string) tea.Cmd {
	return func() tea.Msg {
		content, err := fetchSSHKeyContent(keyName)
		return sshKeyLoadedMsg{content: content, err: err}
	}
}

// inputTypeFor returns the API-declared type of a customization, by name, so
// the confirm screen can format (or mask) its answer accordingly.
func (m *reinstallWizardModel) inputTypeFor(name string) string {
	if m.details == nil {
		return ""
	}

	for _, input := range m.details.Inputs {
		if input.Name == name {
			return input.Type
		}
	}

	return ""
}

// focusedInput is the question the form cursor currently sits on.
func (m *reinstallWizardModel) focusedInput() *osInput {
	if m.details == nil || m.formIdx < 0 || m.formIdx >= len(m.details.Inputs) {
		return nil
	}

	return &m.details.Inputs[m.formIdx]
}

// answerString gives the live value of an answer, as text to display or edit.
func (m *reinstallWizardModel) answerString(name string) string {
	answer, ok := m.answers[name]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%v", answer)
}

func (m *reinstallWizardModel) setFocusedAnswer(value string) {
	if input := m.focusedInput(); input != nil {
		m.answers[input.Name] = value
	}
}

// toggleFocusedBoolean flips a yes/no answer in place, without leaving the form
// row: only two values, so there is nothing to pick from a list.
func (m *reinstallWizardModel) toggleFocusedBoolean() {
	input := m.focusedInput()
	if input == nil {
		return
	}

	m.errorMsg = ""
	m.answers[input.Name] = strconv.FormatBool(!isTrueValue(m.answerString(input.Name)))
}

// openInputsForm shows the customization form, or skips straight to the storage
// step when the OS asks nothing.
func (m *reinstallWizardModel) openInputsForm() (tea.Model, tea.Cmd) {
	if m.details == nil || len(m.details.Inputs) == 0 {
		return m.finishInputsForm()
	}

	m.errorMsg = ""
	m.formIdx = 0
	m.step = reinstallStepInputs

	return m, nil
}

// focusNextInput moves the form focus one row down, or leaves step 2 when the
// last question was just answered.
func (m *reinstallWizardModel) focusNextInput() (tea.Model, tea.Cmd) {
	if m.details != nil && m.formIdx < len(m.details.Inputs)-1 {
		m.errorMsg = ""
		m.formIdx++
		m.step = reinstallStepInputs
		return m, nil
	}

	return m.finishInputsForm()
}

// finishInputsForm validates every answer, not only the focused one: the user
// may have walked past a question without ever validating it. The answers left
// at their default value are then dropped, so the request stays minimal.
func (m *reinstallWizardModel) finishInputsForm() (tea.Model, tea.Cmd) {
	var inputs []osInput
	if m.details != nil {
		inputs = m.details.Inputs
	}

	if index, err := firstInvalidAnswer(inputs, m.answers); err != nil {
		m.errorMsg = err.Error()
		m.formIdx = index
		m.step = reinstallStepInputs
		return m, nil
	}

	m.errorMsg = ""
	m.customizations = finalizeCustomizations(inputs, m.answers)

	return m, m.loadStorage()
}

func (m *reinstallWizardModel) handleKeyValueKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.kvIdx > 0 {
			m.kvIdx--
		}
	case "down", "j":
		if m.kvIdx < len(m.kvPairs)-1 {
			m.kvIdx++
		}
	case "a":
		m.errorMsg = ""
		if len(m.kvPairs) >= maxKeyValuePairs {
			m.errorMsg = fmt.Sprintf("a maximum of %d pairs is allowed", maxKeyValuePairs)
			return m, nil
		}
		m.startKeyValueEdition(len(m.kvPairs))
	case "e":
		if len(m.kvPairs) > 0 {
			m.startKeyValueEdition(m.kvIdx)
		}
	case "d":
		if len(m.kvPairs) > 0 {
			m.kvPairs = slices.Delete(m.kvPairs, m.kvIdx, m.kvIdx+1)
			m.kvIdx = min(m.kvIdx, max(len(m.kvPairs)-1, 0))
			m.errorMsg = ""
		}
	case "left":
		// Back to the form, on the same question, leaving it untouched.
		m.errorMsg = ""
		m.step = reinstallStepInputs
	case "enter":
		input := m.focusedInput()
		if input == nil {
			return m, nil
		}
		values := kvMapFromPairs(m.kvPairs)
		if values == nil {
			values = map[string]string{}
		}
		m.answers[input.Name] = values
		return m.focusNextInput()
	}

	return m, nil
}

func (m *reinstallWizardModel) startKeyValueEdition(index int) {
	m.errorMsg = ""
	m.kvEditIdx = index
	m.kvEditFieldIdx = 0

	if index < len(m.kvPairs) {
		m.kvEditKey = m.kvPairs[index].Key
		m.kvEditValue = m.kvPairs[index].Value
	} else {
		m.kvEditKey = ""
		m.kvEditValue = ""
	}

	m.step = reinstallStepKeyValueEdit
}

func (m *reinstallWizardModel) handleKeyValueEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.errorMsg = ""
		m.step = reinstallStepKeyValue
		return m, nil
	case "tab", "down", "right", "shift+tab", "up", "left":
		// Only two fields: either direction just toggles between them.
		m.kvEditFieldIdx = (m.kvEditFieldIdx + 1) % 2
		return m, nil
	case "enter":
		trimmedKey := strings.TrimSpace(m.kvEditKey)
		if trimmedKey == "" {
			m.errorMsg = "the key cannot be empty"
			return m, nil
		}
		for i, pair := range m.kvPairs {
			if i != m.kvEditIdx && pair.Key == trimmedKey {
				m.errorMsg = fmt.Sprintf("%q is already used", trimmedKey)
				return m, nil
			}
		}

		pair := keyValuePair{Key: trimmedKey, Value: m.kvEditValue}
		if m.kvEditIdx < len(m.kvPairs) {
			m.kvPairs[m.kvEditIdx] = pair
		} else {
			m.kvPairs = append(m.kvPairs, pair)
		}

		m.errorMsg = ""
		m.kvIdx = m.kvEditIdx
		m.step = reinstallStepKeyValue
		return m, nil
	}

	switch key {
	case "backspace":
		if m.kvEditFieldIdx == 0 {
			m.kvEditKey = trimLastRune(m.kvEditKey)
		} else {
			m.kvEditValue = trimLastRune(m.kvEditValue)
		}
	default:
		if len(msg.Runes) == 0 {
			return m, nil
		}
		if m.kvEditFieldIdx == 0 {
			m.kvEditKey += string(msg.Runes)
		} else {
			m.kvEditValue += string(msg.Runes)
		}
	}

	return m, nil
}

// submitFocusedInput is what Enter does on a form row: open the editor a
// question needs one for, otherwise validate the answer and move on.
func (m *reinstallWizardModel) submitFocusedInput() (tea.Model, tea.Cmd) {
	input := m.focusedInput()
	if input == nil {
		return m, nil
	}

	switch input.Type {
	case "enum":
		if len(input.Enum) > 0 {
			m.errorMsg = ""
			m.enumList.Title = inputTitle(*input)
			fillPicker(&m.enumList, labelPickerItems(input.Enum),
				max(slices.Index(input.Enum, m.answerString(input.Name)), 0))
			m.step = reinstallStepEnum
			return m, nil
		}
	case "sshPubKey":
		// Only offer the account keys while the question is unanswered: once a
		// key is there, Enter validates it like any other typed value.
		if m.answerString(input.Name) == "" && len(m.sshKeyNames) > 0 {
			m.errorMsg = ""
			m.sshKeyList.Title = inputTitle(*input)
			fillPicker(&m.sshKeyList,
				labelPickerItems(append(slices.Clone(m.sshKeyNames), customSSHKeyChoice)), 0)
			m.step = reinstallStepSSHKey
			return m, nil
		}
	case "keyValue":
		m.errorMsg = ""
		m.kvPairs = kvPairsFromAnswer(m.answers[input.Name])
		m.kvIdx = 0
		m.step = reinstallStepKeyValue
		return m, nil
	}

	if err := validateAnswer(*input, m.answers[input.Name]); err != nil {
		m.errorMsg = err.Error()
		return m, nil
	}

	return m.focusNextInput()
}

func (m *reinstallWizardModel) handleInputsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	input := m.focusedInput()
	if input == nil {
		return m, nil
	}

	key := msg.String()

	switch key {
	case "down", "tab":
		// Moving around the form never validates anything.
		m.errorMsg = ""
		m.formIdx = (m.formIdx + 1) % len(m.details.Inputs)
		return m, nil
	case "up", "shift+tab":
		m.errorMsg = ""
		m.formIdx = (m.formIdx + len(m.details.Inputs) - 1) % len(m.details.Inputs)
		return m, nil
	case "enter":
		return m.submitFocusedInput()
	}

	// Left only changes a value on a yes/no question, otherwise it leaves the
	// form for the OS list.
	if input.Type == "boolean" {
		switch key {
		case "left", "right":
			m.toggleFocusedBoolean()
		}
		return m, nil
	}

	if key == "left" {
		m.errorMsg = ""
		m.step = reinstallStepOs
		return m, nil
	}

	// An enum value and a key/value list are only picked from their own screen.
	if input.Type == "enum" || input.Type == "keyValue" {
		return m, nil
	}

	switch key {
	case "backspace":
		m.setFocusedAnswer(trimLastRune(m.answerString(input.Name)))
	case "ctrl+u":
		m.setFocusedAnswer("")
	case "alt+enter":
		if input.Type == "text" {
			m.setFocusedAnswer(m.answerString(input.Name) + "\n")
		}
	default:
		if len(msg.Runes) > 0 {
			m.setFocusedAnswer(m.answerString(input.Name) + string(msg.Runes))
		}
	}

	return m, nil
}

func (m *reinstallWizardModel) handleEnumKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.filteringPicker() {
		switch msg.String() {
		case "left":
			// Back to the form, on the same question, leaving it untouched.
			m.errorMsg = ""
			m.step = reinstallStepInputs
			return m, nil
		case "enter":
			selected, ok := selectedPickerItem(&m.enumList)
			if !ok {
				return m, nil
			}
			m.errorMsg = ""
			m.setFocusedAnswer(selected.label)
			return m.focusNextInput()
		}
	}

	return m.updateActivePicker(msg)
}

func (m *reinstallWizardModel) handleSSHKeyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.filteringPicker() {
		switch msg.String() {
		case "left":
			m.errorMsg = ""
			m.step = reinstallStepInputs
			return m, nil
		case "enter":
			selected, ok := selectedPickerItem(&m.sshKeyList)
			if !ok {
				return m, nil
			}
			m.errorMsg = ""
			if selected.index >= len(m.sshKeyNames) {
				// The key is typed by hand: back to its row, focus unchanged.
				m.setFocusedAnswer("")
				m.step = reinstallStepInputs
				return m, nil
			}
			m.isLoading = true
			m.loadingMessage = "Loading SSH key…"
			return m, m.fetchSSHKeyContentCmd(m.sshKeyNames[selected.index])
		}
	}

	return m.updateActivePicker(msg)
}

func (m *reinstallWizardModel) renderInputsStep() string {
	focused := m.focusedInput()
	if focused == nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render("Installation parameters of "+m.osName+":") + "\n")
	content.WriteString(wizardDescStyle.Render(
		"Fill in every parameter, then press Enter on the last one to continue.") + "\n\n")

	for i, input := range m.details.Inputs {
		if i == m.formIdx {
			content.WriteString(m.renderFocusedInput(input))
			continue
		}

		value := compactSummary(input.Type, m.answers[input.Name])
		if value == "" {
			value = "(not set)"
		}
		content.WriteString(wizardLabelStyle.Render("  "+inputRowLabel(input)) +
			wizardDimStyle.Render("  "+value) + "\n")
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(m.inputsHint(*focused)))

	return content.String()
}

// renderFocusedInput draws the row being answered: the value is editable in
// place, so the whole form stays on a single screen.
func (m *reinstallWizardModel) renderFocusedInput(input osInput) string {
	label := wizardSelectedStyle.Render("▶ "+inputTitle(input)) + "  "
	value := m.answerString(input.Name)

	var row string
	switch input.Type {
	case "enum":
		choice := value
		if choice == "" {
			choice = "(not set)"
		}
		row = label + wizardSelectedStyle.Render(choice+" (Enter to choose)")
	case "boolean":
		choice := "No"
		if isTrueValue(value) {
			choice = "Yes"
		}
		row = label + wizardSelectedStyle.Render("◀ "+choice+" ▶")
	case "keyValue":
		summary := compactSummary(input.Type, m.answers[input.Name])
		if summary == "" {
			summary = "no pair defined"
		}
		row = label + wizardSelectedStyle.Render(summary+" (Enter to edit)")
	case "text", "sshPubKey":
		row = label + "\n" + wizardTextAreaStyle.Render(value+"▌")
	case "number":
		row = lipgloss.JoinHorizontal(lipgloss.Center, label,
			wizardNumberInputStyle.Render(visibleTail(value, 8)+"▌"))
	default:
		row = lipgloss.JoinHorizontal(lipgloss.Center, label,
			wizardInputStyle.Render(visibleTail(value, 56)+"▌"))
	}

	return row + "\n"
}

// inputRowLabel names a question on its form row, marking the ones the API
// requires an answer for.
func inputRowLabel(input osInput) string {
	if input.Mandatory {
		return input.Name + " *"
	}

	return input.Name
}

func (m *reinstallWizardModel) inputsHint(input osInput) string {
	switch input.Type {
	case "boolean":
		return "↑↓/Tab: Field • ←→: Change value • Enter: Continue • Esc: Cancel"
	case "enum":
		return "↑↓/Tab: Field • Enter: Choose a value • ←: Back to the OS • Esc: Cancel"
	case "keyValue":
		return "↑↓/Tab: Field • Enter: Edit the pairs • ←: Back to the OS • Esc: Cancel"
	case "text":
		return "↑↓/Tab: Field • Type: Edit • Alt+Enter: New line • Enter: Continue • ←: Back to the OS • Esc: Cancel"
	case "sshPubKey":
		if m.answerString(input.Name) == "" && len(m.sshKeyNames) > 0 {
			return "Enter: Pick one of your saved SSH keys, or type one • ←: Back to the OS • Esc: Cancel"
		}
	}

	return "↑↓/Tab: Field • Type: Edit • Enter: Continue • ←: Back to the OS • Esc: Cancel"
}

func (m *reinstallWizardModel) renderEnumStep() string {
	var content strings.Builder
	content.WriteString(wizardDescStyle.Render("Select a value, typing to filter the list:") + "\n\n")
	content.WriteString(m.enumList.View() + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("Enter: Select • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderSSHKeyStep() string {
	var content strings.Builder
	content.WriteString(wizardDescStyle.Render("Select one of your SSH keys, or enter a key manually:") + "\n\n")
	content.WriteString(m.sshKeyList.View() + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("Enter: Select • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderKeyValueStep() string {
	input := m.focusedInput()
	if input == nil {
		return ""
	}

	var content strings.Builder
	content.WriteString(wizardTitleStyle.Render(inputTitle(*input)) + "\n")
	content.WriteString(wizardSubtleStyle.Render(fmt.Sprintf("%d/%d pairs", len(m.kvPairs), maxKeyValuePairs)) + "\n\n")

	if len(m.kvPairs) == 0 {
		content.WriteString(wizardDimStyle.Render("  No pair defined yet.") + "\n")
	}
	for i, pair := range m.kvPairs {
		label := fmt.Sprintf("%s = %s", pair.Key, pair.Value)
		if i == m.kvIdx {
			content.WriteString(wizardSelectedStyle.Render("▶ "+label) + "\n")
		} else {
			content.WriteString(wizardDimStyle.Render("  "+label) + "\n")
		}
	}
	content.WriteString("\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render(
		"↑↓ Navigate • a: Add • e: Edit • d: Delete • Enter: Continue • ←: Back • Esc: Cancel"))

	return content.String()
}

func (m *reinstallWizardModel) renderKeyValueEditStep() string {
	input := m.focusedInput()
	if input == nil {
		return ""
	}

	var content strings.Builder
	title := "Edit pair"
	if m.kvEditIdx >= len(m.kvPairs) {
		title = "New pair"
	}
	content.WriteString(wizardTitleStyle.Render(inputTitle(*input)+" — "+title+":") + "\n\n")

	keyBox, valueBox := wizardKVBoxStyle, wizardKVBoxStyle
	keyText, valueText := m.kvEditKey, m.kvEditValue
	if m.kvEditFieldIdx == 0 {
		keyBox = wizardKVBoxFocusedStyle
		keyText += "▌"
	} else {
		valueBox = wizardKVBoxFocusedStyle
		valueText += "▌"
	}

	// Headers must line up with the boxes below them: measure the boxes'
	// actual rendered width instead of assuming one, so the two rows stay in
	// sync even if the box style changes later.
	boxWidth := lipgloss.Width(wizardKVBoxStyle.Render(""))
	gap := strings.Repeat(" ", lipgloss.Width(" = "))
	headerStyle := lipgloss.NewStyle().Width(boxWidth)
	// The left indent is joined as its own block, not string-concatenated:
	// a plain "  "+multilineString prefix only lands on the block's first
	// line, leaving the rest of the box shifted 2 columns left of the header.
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		"  ", headerStyle.Render("Key"), gap, headerStyle.Render("Value")) + "\n")
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		"  ", keyBox.Render(keyText), " = ", valueBox.Render(valueText)) + "\n\n")

	content.WriteString(m.renderError())
	content.WriteString(wizardHintStyle.Render("Type a value • Tab/←→: Switch field • Enter: Save • Esc: Cancel"))

	return content.String()
}

func inputTitle(input osInput) string {
	title := input.Name
	if input.Description != "" {
		title = input.Description + " (" + input.Name + ")"
	}

	if input.Mandatory {
		title += " (mandatory)"
	}

	return title
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	return string(runes[:len(runes)-1])
}

// normalizeDigits strips leading zeros from a digit string being typed, so
// e.g. "0" then "2" then "0" then "4" then "8" then "0" reads as "20480"
// instead of "020480". A lone "0" is preserved as-is.
func normalizeDigits(value string) string {
	trimmed := strings.TrimLeft(value, "0")
	if trimmed == "" {
		return "0"
	}

	return trimmed
}
