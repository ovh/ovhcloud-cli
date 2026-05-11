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

// ─── Options ──────────────────────────────────────────────────────────────────

// lbL7RuleTypeOptions maps rule type values to labels and which comparison types are valid.
var lbL7RuleTypeOptions = []struct {
	label    string
	value    string
	desc     string
	needsKey bool // HEADER / COOKIE require a key
}{
	{"Cookie", "cookie", "Match on a cookie name/value", true},
	{"File Type", "file_type", "Match on the file extension", false},
	{"Header", "header", "Match on an HTTP header name/value", true},
	{"Host Name", "host_name", "Match on the request hostname", false},
	{"Path", "path", "Match on the request URI path", false},
	{"SSL Conn Has Cert", "ssl_conn_has_cert", "TLS connection has a client certificate", false},
	{"SSL DN Field", "ssl_dn_field", "Match on an SSL certificate DN field", true},
}

// compareTypeOptions lists all valid comparison types.
var lbL7RuleCompareOptions = []struct {
	label string
	value string
}{
	{"Contains", "contains"},
	{"Ends With", "ends_with"},
	{"Equal To", "equal_to"},
	{"Regex", "regex"},
	{"Starts With", "starts_with"},
}

// ─── Render ──────────────────────────────────────────────────────────────────

func (m Model) renderLBL7RuleWizardTypeStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7B68EE")).Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Select Rule Type") + "\n\n")
	b.WriteString(descStyle.Render("  Policy: "+m.wizard.l7RulePolicyName) + "\n\n")

	for i, opt := range lbL7RuleTypeOptions {
		line := fmt.Sprintf("  %-20s  %s", opt.label, opt.desc)
		if i == m.wizard.l7RuleTypeIdx {
			b.WriteString(selectedStyle.Render("▶ "+line) + "\n")
		} else {
			b.WriteString(unselectedStyle.Render("  "+line) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  ↑↓: Navigate • Enter: Select\n"))
	return b.String()
}

func (m Model) renderLBL7RuleWizardCompareStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7B68EE")).Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Select Comparison Type") + "\n\n")
	b.WriteString(descStyle.Render("  Rule Type: "+m.wizard.l7RuleType) + "\n\n")

	for i, opt := range lbL7RuleCompareOptions {
		if i == m.wizard.l7RuleCompareIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
		} else {
			b.WriteString(unselectedStyle.Render("  "+opt.label) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  ↑↓: Navigate • Enter: Select • ←: Back\n"))
	return b.String()
}

func (m Model) renderLBL7RuleWizardKeyStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Enter Key") + "\n\n")
	b.WriteString(descStyle.Render(fmt.Sprintf("  Rule Type: %s — the key identifies the header or cookie name", m.wizard.l7RuleType)) + "\n\n")
	b.WriteString("  Key: " + inputStyle.Render(m.wizard.l7RuleKeyInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back\n"))
	return b.String()
}

func (m Model) renderLBL7RuleWizardValueStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Enter Value") + "\n\n")
	b.WriteString(descStyle.Render(fmt.Sprintf("  Type: %s  Comparison: %s", m.wizard.l7RuleType, m.wizard.l7RuleCompare)) + "\n\n")
	if m.wizard.l7RuleKey != "" {
		b.WriteString(descStyle.Render("  Key: "+m.wizard.l7RuleKey) + "\n\n")
	}
	b.WriteString("  Value: " + inputStyle.Render(m.wizard.l7RuleValueInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back\n"))
	return b.String()
}

func (m Model) renderLBL7RuleWizardInvertStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	onStyle := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	offStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)

	b.WriteString(titleStyle.Render("  Invert Rule?") + "\n\n")
	b.WriteString(descStyle.Render("  When enabled, the rule matches if the condition is NOT met.") + "\n\n")

	if m.wizard.l7RuleInvert {
		b.WriteString("  " + onStyle.Render("ON") + "\n\n")
	} else {
		b.WriteString("  " + offStyle.Render("OFF") + "\n\n")
	}
	b.WriteString(descStyle.Render("  Space/Enter: Toggle • ←: Back\n"))
	return b.String()
}

func (m Model) renderLBL7RuleWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(titleStyle.Render("  Confirm L7 Rule Creation") + "\n\n")
	b.WriteString(labelStyle.Render("  Policy:") + valStyle.Render(m.wizard.l7RulePolicyName) + "\n")
	b.WriteString(labelStyle.Render("  Type:") + valStyle.Render(m.wizard.l7RuleType) + "\n")
	b.WriteString(labelStyle.Render("  Comparison:") + valStyle.Render(m.wizard.l7RuleCompare) + "\n")
	if m.wizard.l7RuleKey != "" {
		b.WriteString(labelStyle.Render("  Key:") + valStyle.Render(m.wizard.l7RuleKey) + "\n")
	}
	b.WriteString(labelStyle.Render("  Value:") + valStyle.Render(m.wizard.l7RuleValue) + "\n")
	invertStr := "No"
	if m.wizard.l7RuleInvert {
		invertStr = "Yes"
	}
	b.WriteString(labelStyle.Render("  Invert:") + valStyle.Render(invertStr) + "\n\n")

	confirmStyle := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	cancelStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 2)
	if m.wizard.l7RuleConfirmIdx == 1 {
		confirmStyle = lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 2)
		cancelStyle = lipgloss.NewStyle().Background(lipgloss.Color("#FF4444")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	}
	b.WriteString("  " + confirmStyle.Render("Confirm") + "  " + cancelStyle.Render("Cancel") + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  ←→: Select • Enter: Execute • ←: Back\n"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBL7RuleWizardTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.l7RuleTypeIdx > 0 {
			m.wizard.l7RuleTypeIdx--
		}
	case "down", "j":
		if m.wizard.l7RuleTypeIdx < len(lbL7RuleTypeOptions)-1 {
			m.wizard.l7RuleTypeIdx++
		}
	case "enter":
		opt := lbL7RuleTypeOptions[m.wizard.l7RuleTypeIdx]
		m.wizard.l7RuleType = opt.value
		m.wizard.l7RuleCompareIdx = 0
		m.wizard.step = LBL7RuleWizardStepCompare
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardCompareKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.l7RuleCompareIdx > 0 {
			m.wizard.l7RuleCompareIdx--
		}
	case "down", "j":
		if m.wizard.l7RuleCompareIdx < len(lbL7RuleCompareOptions)-1 {
			m.wizard.l7RuleCompareIdx++
		}
	case "enter":
		m.wizard.l7RuleCompare = lbL7RuleCompareOptions[m.wizard.l7RuleCompareIdx].value
		// Determine if this type needs a key field
		needsKey := false
		for _, t := range lbL7RuleTypeOptions {
			if t.value == m.wizard.l7RuleType {
				needsKey = t.needsKey
				break
			}
		}
		if needsKey {
			m.wizard.step = LBL7RuleWizardStepKey
		} else {
			m.wizard.l7RuleKey = ""
			m.wizard.step = LBL7RuleWizardStepValue
		}
	case "left", "h":
		m.wizard.step = LBL7RuleWizardStepType
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardKeyKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		m.wizard.l7RuleKey = strings.TrimSpace(m.wizard.l7RuleKeyInput)
		m.wizard.step = LBL7RuleWizardStepValue
	case "backspace":
		if len(m.wizard.l7RuleKeyInput) > 0 {
			m.wizard.l7RuleKeyInput = m.wizard.l7RuleKeyInput[:len(m.wizard.l7RuleKeyInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.l7RuleKeyInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardValueKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		val := strings.TrimSpace(m.wizard.l7RuleValueInput)
		if val == "" {
			return m, nil
		}
		m.wizard.l7RuleValue = val
		m.wizard.step = LBL7RuleWizardStepInvert
	case "backspace":
		if len(m.wizard.l7RuleValueInput) > 0 {
			m.wizard.l7RuleValueInput = m.wizard.l7RuleValueInput[:len(m.wizard.l7RuleValueInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.l7RuleValueInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardInvertKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case " ", "enter":
		if key == " " {
			m.wizard.l7RuleInvert = !m.wizard.l7RuleInvert
		} else {
			// enter confirms and advances
			m.wizard.l7RuleConfirmIdx = 0
			m.wizard.step = LBL7RuleWizardStepConfirm
		}
	case "left", "h":
		m.wizard.step = LBL7RuleWizardStepValue
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.l7RuleConfirmIdx = 0
	case "right", "l":
		m.wizard.l7RuleConfirmIdx = 1
	case "enter":
		if m.wizard.l7RuleConfirmIdx == 1 {
			// Cancel → back to invert step
			m.wizard.step = LBL7RuleWizardStepInvert
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating L7 Rule..."
		return m, m.createLBL7Rule()
	}
	return m, nil
}

// ─── API ──────────────────────────────────────────────────────────────────────

func (m Model) createLBL7Rule() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbL7RuleCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.l7RulePolicyId == "" || m.wizard.l7RuleLBRegion == "" {
			return lbL7RuleCreatedMsg{err: fmt.Errorf("policy ID or region missing")}
		}

		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule",
			m.cloudProject, url.PathEscape(m.wizard.l7RuleLBRegion), url.PathEscape(m.wizard.l7RulePolicyId))

		body := map[string]interface{}{
			"ruleType":    m.wizard.l7RuleType,
			"compareType": m.wizard.l7RuleCompare,
			"value":       m.wizard.l7RuleValue,
			"invert":      m.wizard.l7RuleInvert,
		}
		if m.wizard.l7RuleKey != "" {
			body["key"] = m.wizard.l7RuleKey
		}

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbL7RuleCreatedMsg{policyID: m.wizard.l7RulePolicyId, err: fmt.Errorf("creation failed: %w", err)}
		}
		return lbL7RuleCreatedMsg{policyID: m.wizard.l7RulePolicyId}
	}
}
