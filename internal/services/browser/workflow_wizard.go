// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// ─── Render ──────────────────────────────────────────────────────────────────

func (m Model) renderWorkflowWizardTypeStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(1, 2).
		Width(width - 8)

	b.WriteString(titleStyle.Render("Select a Workflow") + "\n\n")
	b.WriteString(descStyle.Render("A Workflow describes one or more actions.") + "\n\n")

	inner := selectedStyle.Render("▶ Automated instance backup") + "\n\n" +
		descStyle.Render("This Workflow will generate Instance backups, which\ncan be used to start new Instances.")

	b.WriteString(boxStyle.Render(inner) + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Enter: Continue • Esc: Cancel"))
	return b.String()
}

func (m Model) renderWorkflowWizardInstanceStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Select an Instance to back up:") + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ Loading instances..."))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}
	if len(m.wizard.wfInstances) == 0 {
		b.WriteString(descStyle.Render("No instances available.") + "\n")
	} else {
		maxVisible := 14
		start := 0
		if m.wizard.wfInstanceIdx >= maxVisible {
			start = m.wizard.wfInstanceIdx - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(m.wizard.wfInstances) {
			end = len(m.wizard.wfInstances)
		}
		for i := start; i < end; i++ {
			inst := m.wizard.wfInstances[i]
			label := getStringValue(inst, "name", getStringValue(inst, "id", "unknown"))
			region := getStringValue(inst, "region", "")
			if region != "" {
				label += "  " + descStyle.Render("("+region+")")
			}
			if i == m.wizard.wfInstanceIdx {
				b.WriteString(selectedStyle.Render("▶ "+label) + "\n")
			} else {
				b.WriteString(dimStyle.Render("  "+label) + "\n")
			}
		}
	}
	b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderWorkflowWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(40)

	b.WriteString(titleStyle.Render("Give this Workflow a name:") + "\n\n")
	b.WriteString("  Nom : " + inputStyle.Render(m.wizard.wfNameInput+"█") + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type a name • Enter: Confirm • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderWorkflowWizardScheduleStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(30)

	b.WriteString(titleStyle.Render("Define the schedule") + "\n\n")

	options := []struct{ label, desc string }{
		{"Rotation 7", "Keep the 7 most recent backups"},
		{"Rotation 14", "Keep the 14 most recent backups"},
		{"Custom", "Define your own schedule"},
	}

	for i, opt := range options {
		if i == m.wizard.wfScheduleIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n")
			if i == 2 {
				// Custom cron input
				b.WriteString("\n   Cron : " + inputStyle.Render(m.wizard.wfCronInput+"█") + "\n")
				b.WriteString("   " + descStyle.Render("e.g. 0 0 * * * (every day at midnight)") + "\n")
				b.WriteString("   Rotation : " + inputStyle.Render(fmt.Sprintf("%d", m.wizard.wfRotation)) + "\n")
			}
			b.WriteString("\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓: Choose • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderWorkflowWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(titleStyle.Render("Confirm workflow creation:") + "\n\n")
	b.WriteString(labelStyle.Render("  Workflow:") + valStyle.Render("Automated backup") + "\n")
	instName := m.wizard.wfInstanceName
	if instName == "" {
		instName = m.wizard.wfInstanceId
	}
	b.WriteString(labelStyle.Render("  Instance :") + valStyle.Render(instName) + "\n")
	b.WriteString(labelStyle.Render("  Region:") + valStyle.Render(m.wizard.wfRegion) + "\n")
	b.WriteString(labelStyle.Render("  Nom :") + valStyle.Render(m.wizard.wfName) + "\n")
	b.WriteString(labelStyle.Render("  Cron :") + valStyle.Render(m.wizard.wfCron) + "\n")
	b.WriteString(labelStyle.Render("  Rotation :") + valStyle.Render(fmt.Sprintf("%d backups", m.wizard.wfRotation)) + "\n\n")

	if m.wizard.isLoading {
			b.WriteString(loadingStyle.Render("⏳ Creating..."))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.wfConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	b.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleWorkflowWizardTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		// Only one type available — load instances and go to next step
		m.wizard.wfInstanceIdx = 0
		m.wizard.step = WorkflowWizardStepInstance
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading instances..."
		return m, m.fetchWorkflowInstances()
	}
	return m, nil
}

func (m Model) handleWorkflowWizardInstanceKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.wfInstanceIdx > 0 {
			m.wizard.wfInstanceIdx--
		}
	case "down", "j":
		if m.wizard.wfInstanceIdx < len(m.wizard.wfInstances)-1 {
			m.wizard.wfInstanceIdx++
		}
	case "enter":
		if len(m.wizard.wfInstances) > 0 {
			inst := m.wizard.wfInstances[m.wizard.wfInstanceIdx]
			m.wizard.wfInstanceId = getStringValue(inst, "id", "")
			m.wizard.wfInstanceName = getStringValue(inst, "name", m.wizard.wfInstanceId)
			m.wizard.wfRegion = getStringValue(inst, "region", "")
			m.wizard.wfNameInput = ""
			m.wizard.step = WorkflowWizardStepName
		}
	case "left":
		m.wizard.step = WorkflowWizardStepType
	}
	return m, nil
}

func (m Model) handleWorkflowWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.wfNameInput)
		if name == "" {
			return m, nil
		}
		m.wizard.wfName = name
		m.wizard.wfScheduleIdx = 0
		m.wizard.wfCronInput = "0 0 * * *"
		m.wizard.wfRotation = 7
		m.wizard.step = WorkflowWizardStepSchedule
	case "left":
		m.wizard.step = WorkflowWizardStepInstance
	case "backspace":
		if len(m.wizard.wfNameInput) > 0 {
			m.wizard.wfNameInput = m.wizard.wfNameInput[:len(m.wizard.wfNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.wfNameInput += key
		}
	}
	return m, nil
}

func (m Model) handleWorkflowWizardScheduleKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.wfScheduleIdx > 0 {
			m.wizard.wfScheduleIdx--
		}
	case "down", "j":
		if m.wizard.wfScheduleIdx < 2 {
			m.wizard.wfScheduleIdx++
		}
	case "enter":
		switch m.wizard.wfScheduleIdx {
		case 0:
			m.wizard.wfCron = "0 0 * * *"
			m.wizard.wfRotation = 7
		case 1:
			m.wizard.wfCron = "0 0 * * *"
			m.wizard.wfRotation = 14
		case 2:
			cron := strings.TrimSpace(m.wizard.wfCronInput)
			if cron == "" {
				return m, nil
			}
			m.wizard.wfCron = cron
		}
		m.wizard.wfConfirmBtnIdx = 0
		m.wizard.step = WorkflowWizardStepConfirm
	case "left":
		if m.wizard.wfScheduleIdx != 2 {
			m.wizard.step = WorkflowWizardStepName
		}
	case "backspace":
		if m.wizard.wfScheduleIdx == 2 && len(m.wizard.wfCronInput) > 0 {
			m.wizard.wfCronInput = m.wizard.wfCronInput[:len(m.wizard.wfCronInput)-1]
		}
	default:
		if m.wizard.wfScheduleIdx == 2 && len(key) == 1 {
			m.wizard.wfCronInput += key
		}
	}
	return m, nil
}

func (m Model) handleWorkflowWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.wfConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.wfConfirmBtnIdx = 1
	case "enter":
		if m.wizard.wfConfirmBtnIdx == 1 {
			m.wizard.step = WorkflowWizardStepSchedule
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating workflow..."
		return m, m.createWorkflow()
	}
	return m, nil
}

// ─── API ──────────────────────────────────────────────────────────────────────

type workflowInstancesLoadedMsg struct {
	instances []map[string]interface{}
	err       error
}

type workflowCreatedMsg struct {
	name string
	err  error
}

func (m Model) fetchWorkflowInstances() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return workflowInstancesLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		var instances []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance", m.cloudProject)
		if err := httpLib.Client.Get(endpoint, &instances); err != nil {
			return workflowInstancesLoadedMsg{err: err}
		}
		return workflowInstancesLoadedMsg{instances: instances}
	}
}

func (m Model) createWorkflow() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return workflowCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/workflow/backup",
			m.cloudProject, url.PathEscape(m.wizard.wfRegion))
		body := map[string]interface{}{
			"name":       m.wizard.wfName,
			"instanceId": m.wizard.wfInstanceId,
			"cron":       m.wizard.wfCron,
			"rotation":   m.wizard.wfRotation,
		}
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return workflowCreatedMsg{err: err}
		}
		return workflowCreatedMsg{name: m.wizard.wfName}
	}
}
