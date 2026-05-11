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

	b.WriteString(titleStyle.Render("Sélectionnez un Workflow") + "\n\n")
	b.WriteString(descStyle.Render("Un Workflow décrit une ou plusieurs actions.") + "\n\n")

	inner := selectedStyle.Render("▶ Sauvegarde automatisée des instances") + "\n\n" +
		descStyle.Render("Ce Workflow générera des sauvegardes d'Instance, les sauvegardes\npourront être utilisées pour démarrer de nouvelles Instances.")

	b.WriteString(boxStyle.Render(inner) + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Enter : Continuer • Esc : Annuler"))
	return b.String()
}

func (m Model) renderWorkflowWizardInstanceStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Sélectionnez une Instance à sauvegarder :") + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ Chargement des instances..."))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}
	if len(m.wizard.wfInstances) == 0 {
		b.WriteString(descStyle.Render("Aucune instance disponible.") + "\n")
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
		Render("↑↓ : Naviguer • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderWorkflowWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(40)

	b.WriteString(titleStyle.Render("Donnez un nom à ce Workflow :") + "\n\n")
	b.WriteString("  Nom : " + inputStyle.Render(m.wizard.wfNameInput+"█") + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Valider • ← : Retour • Esc : Annuler"))
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

	b.WriteString(titleStyle.Render("Définir l'ordonnancement") + "\n\n")

	options := []struct{ label, desc string }{
		{"Rotation 7", "Conserver les 7 dernières sauvegardes"},
		{"Rotation 14", "Conserver les 14 dernières sauvegardes"},
		{"Personnalisé", "Définir votre propre planification"},
	}

	for i, opt := range options {
		if i == m.wizard.wfScheduleIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n")
			if i == 2 {
				// Custom cron input
				b.WriteString("\n   Cron : " + inputStyle.Render(m.wizard.wfCronInput+"█") + "\n")
				b.WriteString("   " + descStyle.Render("ex: 0 0 * * * (tous les jours à minuit)") + "\n")
				b.WriteString("   Rotation : " + inputStyle.Render(fmt.Sprintf("%d", m.wizard.wfRotation)) + "\n")
			}
			b.WriteString("\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ : Choisir • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderWorkflowWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(titleStyle.Render("Confirmer la création du Workflow :") + "\n\n")
	b.WriteString(labelStyle.Render("  Workflow :") + valStyle.Render("Sauvegarde automatisée") + "\n")
	instName := m.wizard.wfInstanceName
	if instName == "" {
		instName = m.wizard.wfInstanceId
	}
	b.WriteString(labelStyle.Render("  Instance :") + valStyle.Render(instName) + "\n")
	b.WriteString(labelStyle.Render("  Région :") + valStyle.Render(m.wizard.wfRegion) + "\n")
	b.WriteString(labelStyle.Render("  Nom :") + valStyle.Render(m.wizard.wfName) + "\n")
	b.WriteString(labelStyle.Render("  Cron :") + valStyle.Render(m.wizard.wfCron) + "\n")
	b.WriteString(labelStyle.Render("  Rotation :") + valStyle.Render(fmt.Sprintf("%d sauvegardes", m.wizard.wfRotation)) + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ Création en cours..."))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Créer ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.wfConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Créer ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	b.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ : Sélectionner • Enter : Confirmer • Esc : Annuler"))
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
		m.wizard.loadingMessage = "Chargement des instances..."
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
		m.wizard.loadingMessage = "Création du Workflow..."
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
