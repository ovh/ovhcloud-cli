// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// ─── Options ──────────────────────────────────────────────────────────────────

var lbListenerProtoOptions = []struct{ label, value, desc string }{
	{"HTTP", "http", "Protocole HTTP (couche 7)"},
	{"HTTPS", "https", "Protocole HTTPS (terminaison TLS côté LB)"},
	{"TCP", "tcp", "Protocole TCP (couche 4)"},
	{"UDP", "udp", "Protocole UDP (couche 4)"},
	{"Terminated HTTPS", "terminatedHTTPS", "TLS terminé par le LB, backend reçoit HTTP"},
	{"PROMETHEUS", "prometheus", "Métriques Prometheus"},
	{"SCTP", "sctp", "Protocole SCTP"},
}

// ─── Render ──────────────────────────────────────────────────────────────────

func (m Model) renderLBListenerWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(40)

	b.WriteString(titleStyle.Render("Nom du listener") + "\n\n")
	b.WriteString(descStyle.Render("LB : "+m.wizard.lbListenerLBName+" ("+m.wizard.lbListenerLBRegion+")") + "\n\n")
	b.WriteString("  Nom : " + inputStyle.Render(m.wizard.lbListenerNameInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Valider • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBListenerWizardProtoStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Protocole du listener") + "\n\n")

	for i, opt := range lbListenerProtoOptions {
		if i == m.wizard.lbListenerProtoIdx {
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

func (m Model) renderLBListenerWizardPortStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(12)

	b.WriteString(titleStyle.Render("Port d'écoute") + "\n\n")
	b.WriteString(descStyle.Render("Protocole : "+m.wizard.lbListenerProto) + "\n\n")
	b.WriteString("  Port (1-65535) : " + inputStyle.Render(m.wizard.lbListenerPortInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le port • Enter : Valider • ← : Retour • Esc : Annuler"))
	return b.String()
}

// lbCompatiblePoolProtos returns pool protocols compatible with a given listener protocol.
func lbCompatiblePoolProtos(listenerProto string) []string {
	switch listenerProto {
	case "http":
		return []string{"http"}
	case "https":
		return []string{"https"}
	case "terminatedHTTPS":
		return []string{"http"}
	case "tcp", "sctp", "prometheus":
		return []string{"tcp", "proxy", "proxyV2"}
	case "udp":
		return []string{"udp"}
	default:
		return nil // no filter
	}
}

func lbPoolCompatible(poolProto, listenerProto string) bool {
	compatible := lbCompatiblePoolProtos(listenerProto)
	if compatible == nil {
		return true
	}
	for _, p := range compatible {
		if p == poolProto {
			return true
		}
	}
	return false
}

func (m Model) renderLBListenerWizardPoolStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Pool par défaut (facultatif)") + "\n\n")

	compatProtos := lbCompatiblePoolProtos(m.wizard.lbListenerProto)
	compatNote := ""
	if len(compatProtos) > 0 {
		compatNote = "Protocoles compatibles avec " + m.wizard.lbListenerProto + " : " + strings.Join(compatProtos, ", ")
	}
	b.WriteString(descStyle.Render("Associez un pool existant à ce listener, ou ignorez cette étape.") + "\n")
	if compatNote != "" {
		b.WriteString(descStyle.Render(compatNote) + "\n")
	}
	b.WriteString("\n")

	allPools := m.lbPools[m.wizard.lbListenerLBId]
	// Filter to compatible pools only
	var pools []map[string]interface{}
	for _, p := range allPools {
		if lbPoolCompatible(getStringValue(p, "protocol", ""), m.wizard.lbListenerProto) {
			pools = append(pools, p)
		}
	}

	// Option 0: no pool
	if m.wizard.lbListenerPoolIdx == 0 {
		b.WriteString(selectedStyle.Render("▶ Aucun pool (ignorer)") + "\n\n")
	} else {
		b.WriteString(dimStyle.Render("  Aucun pool (ignorer)") + "\n\n")
	}

	for i, p := range pools {
		pName := getStringValue(p, "name", "?")
		pID := getStringValue(p, "id", "")
		label := fmt.Sprintf("%s (%s)", pName, truncate(pID, 8))
		idx := i + 1
		if idx == m.wizard.lbListenerPoolIdx {
			b.WriteString(selectedStyle.Render("▶ "+label) + "\n\n")
		} else {
			b.WriteString(dimStyle.Render("  "+label) + "\n\n")
		}
	}

	if len(pools) == 0 {
		msg := "  (aucun pool disponible)"
		if len(allPools) > 0 {
			msg = "  (aucun pool compatible — créez un pool avec un protocole compatible)"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render(msg) + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ : Naviguer • Enter : Sélectionner • ← : Retour • Esc : Annuler"))
	return b.String()
}

func (m Model) renderLBListenerWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	isEdit := m.wizard.lbListenerEditId != ""
	if isEdit {
		b.WriteString(titleStyle.Render("Confirmer la modification du Listener :") + "\n\n")
	} else {
		b.WriteString(titleStyle.Render("Confirmer la création du Listener :") + "\n\n")
	}
	b.WriteString(labelStyle.Render("  Load Balancer :") + valStyle.Render(m.wizard.lbListenerLBName) + "\n")
	b.WriteString(labelStyle.Render("  Région :") + valStyle.Render(m.wizard.lbListenerLBRegion) + "\n")
	b.WriteString(labelStyle.Render("  Nom :") + valStyle.Render(m.wizard.lbListenerName) + "\n")
	if !isEdit {
		b.WriteString(labelStyle.Render("  Protocole :") + valStyle.Render(m.wizard.lbListenerProto) + "\n")
		b.WriteString(labelStyle.Render("  Port :") + valStyle.Render(strconv.Itoa(m.wizard.lbListenerPort)) + "\n")
	}

	poolLabel := "Aucun"
	if m.wizard.lbListenerPoolId != "" {
		pools := m.lbPools[m.wizard.lbListenerLBId]
		for _, p := range pools {
			if getStringValue(p, "id", "") == m.wizard.lbListenerPoolId {
				poolLabel = getStringValue(p, "name", m.wizard.lbListenerPoolId)
				break
			}
		}
	}
	b.WriteString(labelStyle.Render("  Pool par défaut :") + valStyle.Render(poolLabel) + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ " + m.wizard.loadingMessage))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	actionLabel := " Créer "
	if isEdit {
		actionLabel = " Modifier "
	}
	btnAction := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2).Render(actionLabel)
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.lbListenerConfirmIdx == 1 {
		btnAction = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(actionLabel)
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	b.WriteString(btnAction + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ : Sélectionner • Enter : Confirmer • Esc : Annuler"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBListenerWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbListenerNameInput)
		if name == "" {
			m.wizard.errorMsg = "Le nom ne peut pas être vide"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.lbListenerName = name
		if m.wizard.lbListenerEditId != "" {
			// Edit mode: skip proto/port, go directly to pool
			m.wizard.lbListenerPoolIdx = 0
			m.wizard.step = LBListenerWizardStepPool
		} else {
			m.wizard.lbListenerProtoIdx = 0
			m.wizard.step = LBListenerWizardStepProto
		}
	case "backspace":
		if len(m.wizard.lbListenerNameInput) > 0 {
			m.wizard.lbListenerNameInput = m.wizard.lbListenerNameInput[:len(m.wizard.lbListenerNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbListenerNameInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBListenerWizardProtoKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbListenerProtoIdx > 0 {
			m.wizard.lbListenerProtoIdx--
		}
	case "down", "j":
		if m.wizard.lbListenerProtoIdx < len(lbListenerProtoOptions)-1 {
			m.wizard.lbListenerProtoIdx++
		}
	case "enter":
		m.wizard.lbListenerProto = lbListenerProtoOptions[m.wizard.lbListenerProtoIdx].value
		m.wizard.lbListenerPortInput = ""
		m.wizard.step = LBListenerWizardStepPort
	case "left":
		m.wizard.step = LBListenerWizardStepName
	}
	return m, nil
}

func (m Model) handleLBListenerWizardPortKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		portStr := strings.TrimSpace(m.wizard.lbListenerPortInput)
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			m.wizard.errorMsg = "Port invalide (1-65535)"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.lbListenerPort = port
		m.wizard.lbListenerPoolIdx = 0
		m.wizard.step = LBListenerWizardStepPool
	case "backspace":
		if len(m.wizard.lbListenerPortInput) > 0 {
			m.wizard.lbListenerPortInput = m.wizard.lbListenerPortInput[:len(m.wizard.lbListenerPortInput)-1]
		}
	case "left":
		m.wizard.step = LBListenerWizardStepProto
	default:
		// Only allow digits
		if len(key) == 1 && key >= "0" && key <= "9" {
			if len(m.wizard.lbListenerPortInput) < 5 {
				m.wizard.lbListenerPortInput += key
			}
		}
	}
	return m, nil
}

func (m Model) handleLBListenerWizardPoolKeys(key string) (tea.Model, tea.Cmd) {
	allPools := m.lbPools[m.wizard.lbListenerLBId]
	var pools []map[string]interface{}
	for _, p := range allPools {
		if lbPoolCompatible(getStringValue(p, "protocol", ""), m.wizard.lbListenerProto) {
			pools = append(pools, p)
		}
	}
	maxIdx := len(pools) // 0 = none, 1..n = pool index

	switch key {
	case "up", "k":
		if m.wizard.lbListenerPoolIdx > 0 {
			m.wizard.lbListenerPoolIdx--
		}
	case "down", "j":
		if m.wizard.lbListenerPoolIdx < maxIdx {
			m.wizard.lbListenerPoolIdx++
		}
	case "enter":
		if m.wizard.lbListenerPoolIdx == 0 {
			m.wizard.lbListenerPoolId = ""
		} else {
			m.wizard.lbListenerPoolId = getStringValue(pools[m.wizard.lbListenerPoolIdx-1], "id", "")
		}
		m.wizard.lbListenerConfirmIdx = 0
		m.wizard.step = LBListenerWizardStepConfirm
	case "left":
		if m.wizard.lbListenerEditId != "" {
			// Edit mode: back to name step
			m.wizard.step = LBListenerWizardStepName
		} else {
			m.wizard.step = LBListenerWizardStepPort
		}
	}
	return m, nil
}

func (m Model) handleLBListenerWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.lbListenerConfirmIdx = 0
	case "right", "l":
		m.wizard.lbListenerConfirmIdx = 1
	case "enter":
		if m.wizard.lbListenerConfirmIdx == 1 {
			// Cancel → go back to pool step
			m.wizard.step = LBListenerWizardStepPool
			return m, nil
		}
		m.wizard.isLoading = true
		if m.wizard.lbListenerEditId != "" {
			m.wizard.loadingMessage = "Mise à jour du listener..."
			return m, m.updateLBListener()
		}
		m.wizard.loadingMessage = "Création du listener..."
		return m, m.createLBListener()
	}
	return m, nil
}

// ─── API ──────────────────────────────────────────────────────────────────────

func (m Model) createLBListener() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbListenerCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.lbListenerLBId == "" || m.wizard.lbListenerLBRegion == "" {
			return lbListenerCreatedMsg{err: fmt.Errorf("LB ID ou région manquant")}
		}

		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/listener",
			m.cloudProject, url.PathEscape(m.wizard.lbListenerLBRegion))

		body := map[string]interface{}{
			"loadbalancerId": m.wizard.lbListenerLBId,
			"name":           m.wizard.lbListenerName,
			"protocol":       m.wizard.lbListenerProto,
			"port":           m.wizard.lbListenerPort,
		}
		if m.wizard.lbListenerPoolId != "" {
			body["defaultPoolId"] = m.wizard.lbListenerPoolId
		}

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbListenerCreatedMsg{listenerName: m.wizard.lbListenerName, err: fmt.Errorf("échec de la création: %w", err)}
		}
		return lbListenerCreatedMsg{listenerName: m.wizard.lbListenerName}
	}
}
