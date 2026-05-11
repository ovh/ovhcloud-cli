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

func (m Model) renderLBPoolWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(40)

	b.WriteString(titleStyle.Render("Nom du pool") + "\n\n")
	b.WriteString(descStyle.Render("LB : "+m.wizard.lbPoolLBName+" ("+m.wizard.lbPoolLBRegion+")") + "\n\n")
	b.WriteString("  Nom : " + inputStyle.Render(m.wizard.lbPoolNameInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Valider • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBPoolWizardAlgoStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Algorithme du Load Balancer") + "\n\n")

	options := []struct{ label, value, desc string }{
		{"Round Robin", "roundRobin", "Distribue les requêtes de façon équitable en rotation"},
		{"Least Connections", "leastConnections", "Envoie vers le membre avec le moins de connexions actives"},
		{"Source IP", "sourceIP", "Même IP source toujours dirigée vers le même membre"},
	}

	for i, opt := range options {
		if i == m.wizard.lbPoolAlgoIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ : Naviguer • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBPoolWizardProtoStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Protocole du pool") + "\n\n")

	options := []struct{ label, value, desc string }{
		{"HTTP", "http", "Protocole HTTP (couche 7)"},
		{"HTTPS", "https", "Protocole HTTPS (couche 7, TLS terminé côté backend)"},
		{"TCP", "tcp", "Protocole TCP (couche 4)"},
		{"UDP", "udp", "Protocole UDP (couche 4)"},
		{"PROXY", "proxy", "Protocole PROXY (passe les informations client)"},
		{"PROXY v2", "proxyV2", "Protocole PROXY v2 (version binaire)"},
	}

	for i, opt := range options {
		if i == m.wizard.lbPoolProtoIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ : Naviguer • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBPoolWizardSessionStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Session permanente (Persistance de session)") + "\n\n")

	options := lbPoolSessionOptions

	for i, opt := range options {
		if i == m.wizard.lbPoolSessionIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ : Naviguer • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBPoolWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	confirmTitle := "Confirmer la création du Pool :"
	if m.wizard.lbPoolEditPoolId != "" {
		confirmTitle = "Confirmer la modification du Pool :"
	}
	b.WriteString(titleStyle.Render(confirmTitle) + "\n\n")
	b.WriteString(labelStyle.Render("  Load Balancer :") + valStyle.Render(m.wizard.lbPoolLBName) + "\n")
	b.WriteString(labelStyle.Render("  Région :") + valStyle.Render(m.wizard.lbPoolLBRegion) + "\n")
	b.WriteString(labelStyle.Render("  Nom du pool :") + valStyle.Render(m.wizard.lbPoolName) + "\n")
	// Display friendly label
	algoLabel := m.wizard.lbPoolAlgo
	for _, o := range lbPoolAlgoOptions {
		if o.value == m.wizard.lbPoolAlgo {
			algoLabel = o.label
			break
		}
	}
	protoLabel := m.wizard.lbPoolProto
	for _, o := range lbPoolProtoOptions {
		if o.value == m.wizard.lbPoolProto {
			protoLabel = o.label
			break
		}
	}
	b.WriteString(labelStyle.Render("  Algorithme :") + valStyle.Render(algoLabel) + "\n")
	b.WriteString(labelStyle.Render("  Protocole :") + valStyle.Render(protoLabel) + "\n")
	sessionLabel := "Aucune (disabled)"
	for _, o := range lbPoolSessionOptions {
		if o.value == m.wizard.lbPoolSession {
			sessionLabel = o.label
			break
		}
	}
	b.WriteString(labelStyle.Render("  Session :") + valStyle.Render(sessionLabel) + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ "+m.wizard.loadingMessage))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	btnLabel := " Créer "
	if m.wizard.lbPoolEditPoolId != "" {
		btnLabel = " Modifier "
	}
	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2).Render(btnLabel)
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.lbPoolConfirmIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(btnLabel)
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	b.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ : Sélectionner • Enter : Confirmer • Esc : Annuler"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBPoolWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbPoolNameInput)
		if name == "" {
			m.wizard.errorMsg = "Le nom ne peut pas être vide"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.lbPoolName = name
		m.wizard.lbPoolAlgoIdx = 0
		m.wizard.step = LBPoolWizardStepAlgo
	case "backspace":
		if len(m.wizard.lbPoolNameInput) > 0 {
			m.wizard.lbPoolNameInput = m.wizard.lbPoolNameInput[:len(m.wizard.lbPoolNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbPoolNameInput += key
		}
	}
	return m, nil
}

var lbPoolAlgoOptions = []struct{ label, value string }{
	{"Round Robin", "roundRobin"},
	{"Least Connections", "leastConnections"},
	{"Source IP", "sourceIP"},
}

var lbPoolSessionOptions = []struct{ label, value, desc string }{
	{"Aucune (disabled)", "disabled", "Pas de persistance de session"},
	{"Source IP", "sourceIP", "Persistance basée sur l'adresse IP source"},
	{"Cookie HTTP", "httpCookie", "Persistance via cookie HTTP inséré par le load balancer"},
	{"Cookie applicatif", "appCookie", "Persistance via cookie existant dans la réponse applicative"},
}

var lbPoolProtoOptions = []struct{ label, value string }{
	{"HTTP", "http"},
	{"HTTPS", "https"},
	{"TCP", "tcp"},
	{"UDP", "udp"},
	{"PROXY", "proxy"},
	{"PROXY v2", "proxyV2"},
}

func (m Model) handleLBPoolWizardAlgoKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbPoolAlgoIdx > 0 {
			m.wizard.lbPoolAlgoIdx--
		}
	case "down", "j":
		if m.wizard.lbPoolAlgoIdx < len(lbPoolAlgoOptions)-1 {
			m.wizard.lbPoolAlgoIdx++
		}
	case "enter":
		m.wizard.lbPoolAlgo = lbPoolAlgoOptions[m.wizard.lbPoolAlgoIdx].value
		m.wizard.lbPoolProtoIdx = 0
		if m.wizard.lbPoolEditPoolId != "" {
			// Edit mode: skip protocol step (can't change after creation)
			m.wizard.lbPoolSessionIdx = 0
			m.wizard.step = LBPoolWizardStepSession
		} else {
			m.wizard.step = LBPoolWizardStepProto
		}
	case "left":
		m.wizard.step = LBPoolWizardStepName
	}
	return m, nil
}

func (m Model) handleLBPoolWizardProtoKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbPoolProtoIdx > 0 {
			m.wizard.lbPoolProtoIdx--
		}
	case "down", "j":
		if m.wizard.lbPoolProtoIdx < len(lbPoolProtoOptions)-1 {
			m.wizard.lbPoolProtoIdx++
		}
	case "enter":
		m.wizard.lbPoolProto = lbPoolProtoOptions[m.wizard.lbPoolProtoIdx].value
		m.wizard.lbPoolSessionIdx = 0
		m.wizard.step = LBPoolWizardStepSession
	case "left":
		if m.wizard.lbPoolEditPoolId != "" {
			// In edit mode, going back from proto step goes to algo
			m.wizard.step = LBPoolWizardStepAlgo
		} else {
			m.wizard.step = LBPoolWizardStepAlgo
		}
	}
	return m, nil
}

func (m Model) handleLBPoolWizardSessionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbPoolSessionIdx > 0 {
			m.wizard.lbPoolSessionIdx--
		}
	case "down", "j":
		if m.wizard.lbPoolSessionIdx < len(lbPoolSessionOptions)-1 {
			m.wizard.lbPoolSessionIdx++
		}
	case "enter":
		m.wizard.lbPoolSession = lbPoolSessionOptions[m.wizard.lbPoolSessionIdx].value
		m.wizard.lbPoolConfirmIdx = 0
		m.wizard.step = LBPoolWizardStepConfirm
	case "left":
		m.wizard.step = LBPoolWizardStepProto
	}
	return m, nil
}

func (m Model) handleLBPoolWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.lbPoolConfirmIdx = 0
	case "right", "l":
		m.wizard.lbPoolConfirmIdx = 1
	case "enter":
		if m.wizard.lbPoolConfirmIdx == 1 {
			// Cancel → go back to session step
			m.wizard.step = LBPoolWizardStepSession
			return m, nil
		}
		m.wizard.isLoading = true
		if m.wizard.lbPoolEditPoolId != "" {
			m.wizard.loadingMessage = "Mise à jour du pool..."
			return m, m.updateLBPool()
		}
		m.wizard.loadingMessage = "Création du pool..."
		return m, m.createLBPool()
	}
	return m, nil
}

// ─── API ──────────────────────────────────────────────────────────────────────

func (m Model) createLBPool() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbPoolCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.lbPoolLBId == "" || m.wizard.lbPoolLBRegion == "" {
			return lbPoolCreatedMsg{err: fmt.Errorf("LB ID ou région manquant")}
		}

		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/pool",
			m.cloudProject, url.PathEscape(m.wizard.lbPoolLBRegion))

		body := map[string]interface{}{
			"loadbalancerId": m.wizard.lbPoolLBId,
			"name":           m.wizard.lbPoolName,
			"algorithm":      m.wizard.lbPoolAlgo,
			"protocol":       m.wizard.lbPoolProto,
		}

		if m.wizard.lbPoolSession != "" && m.wizard.lbPoolSession != "disabled" {
			body["sessionPersistence"] = map[string]interface{}{
				"type": m.wizard.lbPoolSession,
			}
		}

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbPoolCreatedMsg{poolName: m.wizard.lbPoolName, err: fmt.Errorf("échec de la création: %w", err)}
		}
		return lbPoolCreatedMsg{poolName: m.wizard.lbPoolName}
	}
}
