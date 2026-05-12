// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Render ──────────────────────────────────────────────────────────────────

func (m Model) renderLBMemberWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	action := "Create"
	if m.wizard.lbMemberEditId != "" {
		action = "Edit"
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s Member — Name", action)) + "\n\n")
	b.WriteString(descStyle.Render("  A descriptive name for this member.") + "\n\n")
	b.WriteString("  Name: " + inputStyle.Render(m.wizard.lbMemberNameInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBMemberWizardIPStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Member — IP Address") + "\n\n")
	b.WriteString(descStyle.Render("  The IP address of the backend server (e.g. 10.0.0.5).") + "\n\n")
	b.WriteString("  IP: " + inputStyle.Render(m.wizard.lbMemberIPInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBMemberWizardPortStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Member — Protocol Port") + "\n\n")
	b.WriteString(descStyle.Render("  The TCP/UDP port the backend server listens on (1–65535).") + "\n\n")
	b.WriteString("  Port: " + inputStyle.Render(m.wizard.lbMemberPortInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBMemberWizardWeightStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Member — Weight") + "\n\n")
	b.WriteString(descStyle.Render("  Relative weight for load distribution (1–256, default 1). 0 = disabled.") + "\n\n")
	b.WriteString("  Weight: " + inputStyle.Render(m.wizard.lbMemberWeightInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBMemberWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(12)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	readOnlySt := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	action := "Create"
	if m.wizard.lbMemberEditId != "" {
		action = "Update"
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("  Confirm %s Member", action)) + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Name"), valueSt.Render(m.wizard.lbMemberName)))
	if m.wizard.lbMemberEditId != "" {
		// Show IP and port as read-only (immutable)
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  IP"), readOnlySt.Render(m.wizard.lbMemberIP+" (read-only)")))
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Port"), readOnlySt.Render(strconv.Itoa(m.wizard.lbMemberPort)+" (read-only)")))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  IP"), valueSt.Render(m.wizard.lbMemberIP)))
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Port"), valueSt.Render(strconv.Itoa(m.wizard.lbMemberPort))))
	}
	b.WriteString(fmt.Sprintf("%s %s\n\n", labelSt.Render("  Weight"), valueSt.Render(strconv.Itoa(m.wizard.lbMemberWeight))))

	saveStyle := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	cancelStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	if m.wizard.lbMemberConfirmIdx == 1 {
		saveStyle = lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		cancelStyle = lipgloss.NewStyle().Background(lipgloss.Color("#CC3333")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	}

	b.WriteString("  " + saveStyle.Render(action) + "  " + cancelStyle.Render("Cancel") + "\n\n")
	b.WriteString(descStyle.Render("  ←→: Select • Enter: Confirm • ←: Back • Esc: Cancel\n"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBMemberWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbMemberNameInput)
		if name == "" {
			return m, nil
		}
		m.wizard.lbMemberName = name
		if m.wizard.lbMemberEditId != "" {
			// In edit mode, IP and port are immutable — skip to weight
			m.wizard.step = LBMemberWizardStepWeight
		} else {
			m.wizard.step = LBMemberWizardStepIP
		}
	case "backspace":
		if len(m.wizard.lbMemberNameInput) > 0 {
			m.wizard.lbMemberNameInput = m.wizard.lbMemberNameInput[:len(m.wizard.lbMemberNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbMemberNameInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBMemberWizardIPKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		ip := strings.TrimSpace(m.wizard.lbMemberIPInput)
		if ip == "" {
			return m, nil
		}
		m.wizard.lbMemberIP = ip
		m.wizard.step = LBMemberWizardStepPort
	case "backspace":
		if len(m.wizard.lbMemberIPInput) > 0 {
			m.wizard.lbMemberIPInput = m.wizard.lbMemberIPInput[:len(m.wizard.lbMemberIPInput)-1]
		}
	case "left", "h":
		m.wizard.step = LBMemberWizardStepName
	default:
		if len(key) == 1 {
			m.wizard.lbMemberIPInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBMemberWizardPortKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		port, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbMemberPortInput))
		if err != nil || port < 1 || port > 65535 {
			return m, nil
		}
		m.wizard.lbMemberPort = port
		if m.wizard.lbMemberWeightInput == "" {
			m.wizard.lbMemberWeightInput = "1"
		}
		m.wizard.step = LBMemberWizardStepWeight
	case "backspace":
		if len(m.wizard.lbMemberPortInput) > 0 {
			m.wizard.lbMemberPortInput = m.wizard.lbMemberPortInput[:len(m.wizard.lbMemberPortInput)-1]
		}
	case "left", "h":
		m.wizard.step = LBMemberWizardStepIP
	default:
		if len(key) == 1 && (key[0] >= '0' && key[0] <= '9') {
			m.wizard.lbMemberPortInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBMemberWizardWeightKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		w, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbMemberWeightInput))
		if err != nil || w < 0 || w > 256 {
			return m, nil
		}
		m.wizard.lbMemberWeight = w
		m.wizard.lbMemberConfirmIdx = 0
		m.wizard.step = LBMemberWizardStepConfirm
	case "backspace":
		if len(m.wizard.lbMemberWeightInput) > 0 {
			m.wizard.lbMemberWeightInput = m.wizard.lbMemberWeightInput[:len(m.wizard.lbMemberWeightInput)-1]
		}
	case "left", "h":
		if m.wizard.lbMemberEditId != "" {
			// In edit mode, IP and port are immutable — go back to name
			m.wizard.step = LBMemberWizardStepName
		} else {
			m.wizard.step = LBMemberWizardStepPort
		}
	default:
		if len(key) == 1 && (key[0] >= '0' && key[0] <= '9') {
			m.wizard.lbMemberWeightInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBMemberWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.lbMemberConfirmIdx = 0
	case "right", "l":
		m.wizard.lbMemberConfirmIdx = 1
	case "enter":
		if m.wizard.lbMemberConfirmIdx == 1 {
			// Cancel → back to LBPoolMembersView
			m.wizard = WizardData{}
			m.mode = LBPoolMembersView
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Saving member..."
		return m, m.saveLBMember()
	}
	return m, nil
}
