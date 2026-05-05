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

// gatewayModels lists the available OVHcloud gateway sizes.
var gatewayModels = []string{"s", "m", "l", "xl", "2xl", "3xl"}

// ─── API call ─────────────────────────────────────────────────────────────────

// createGatewayFromWizard sends the POST request to create a gateway linked to the
// private network/subnet stored in the wizard state.
func (m Model) createGatewayFromWizard() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return gwCreatedMsg{err: fmt.Errorf("aucun projet cloud sélectionné")}
		}

		model := gatewayModels[m.wizard.gwModelIdx]
		body := map[string]interface{}{
			"model": model,
			"name":  m.wizard.gwName,
		}

		endpoint := fmt.Sprintf(
			"/v1/cloud/project/%s/region/%s/network/%s/subnet/%s/gateway",
			m.cloudProject,
			url.PathEscape(m.wizard.gwRegion),
			url.PathEscape(m.wizard.gwNetworkID),
			url.PathEscape(m.wizard.gwSubnetID),
		)

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return gwCreatedMsg{err: fmt.Errorf("failed to create gateway: %w", err)}
		}
		return gwCreatedMsg{gateway: result}
	}
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderGwWizardModelStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Choisir le modèle de la Gateway :") + "\n\n")
	content.WriteString(descStyle.Render(
		fmt.Sprintf("Réseau : %s  •  Région : %s", m.wizard.gwNetworkName, m.wizard.gwRegion),
	) + "\n\n")

	for i, model := range gatewayModels {
		if i == m.wizard.gwModelIdx {
			content.WriteString(selectedStyle.Render("▶ " + strings.ToUpper(model)) + "\n")
		} else {
			content.WriteString(dimStyle.Render("  "+strings.ToUpper(model)) + "\n")
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Naviguer • Enter : Sélectionner • Esc : Annuler"))
	return content.String()
}

func (m Model) renderGwWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Nom de la Gateway :") + "\n\n")
	content.WriteString(descStyle.Render(
		fmt.Sprintf("Modèle : %s  •  Réseau : %s", strings.ToUpper(gatewayModels[m.wizard.gwModelIdx]), m.wizard.gwNetworkName),
	) + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(40)
	content.WriteString(inputStyle.Render(m.wizard.gwNameInput+"▌") + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderGwWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("Confirmer la création de la Gateway :") + "\n\n")
	content.WriteString(labelStyle.Render("  Réseau :") + valueStyle.Render(m.wizard.gwNetworkName) + "\n")
	content.WriteString(labelStyle.Render("  Région :") + valueStyle.Render(m.wizard.gwRegion) + "\n")
	content.WriteString(labelStyle.Render("  Modèle :") + valueStyle.Render(strings.ToUpper(gatewayModels[m.wizard.gwModelIdx])) + "\n")
	content.WriteString(labelStyle.Render("  Nom :") + valueStyle.Render(m.wizard.gwName) + "\n")
	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Création en cours..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Créer ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.gwConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Créer ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ : Sélectionner • Enter : Confirmer • Esc : Annuler"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleGwWizardModelKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.gwModelIdx > 0 {
			m.wizard.gwModelIdx--
		}
	case "down", "j":
		if m.wizard.gwModelIdx < len(gatewayModels)-1 {
			m.wizard.gwModelIdx++
		}
	case "enter":
		m.wizard.step = GwWizardStepName
	}
	return m, nil
}

func (m Model) handleGwWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.gwNameInput)
		if name == "" {
			m.wizard.errorMsg = "Le nom ne peut pas être vide"
			return m, nil
		}
		m.wizard.gwName = name
		m.wizard.errorMsg = ""
		m.wizard.step = GwWizardStepConfirm
	case "left":
		m.wizard.step = GwWizardStepModel
	case "backspace":
		if len(m.wizard.gwNameInput) > 0 {
			m.wizard.gwNameInput = m.wizard.gwNameInput[:len(m.wizard.gwNameInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.wizard.gwNameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleGwWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.gwConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.gwConfirmBtnIdx = 1
	case "enter":
		if m.wizard.gwConfirmBtnIdx == 1 {
			// Cancel → go back to name step
			m.wizard.step = GwWizardStepName
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Création de la Gateway..."
		return m, m.createGatewayFromWizard()
	}
	return m, nil
}
