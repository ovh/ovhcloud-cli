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
	label         string
	value         string
	desc          string
	needsKey      bool     // HEADER / COOKIE / SSL_DN_FIELD require a key
	validCompares []string // nil = all compareTypes valid
	boolValue     bool     // value must be "true" or "false"
}{
	{"Cookie", "cookie", "Match on a cookie name/value", true, nil, false},
	{"File Type", "fileType", "Match on the file extension", false, []string{"equalTo", "regex"}, false},
	{"Header", "header", "Match on an HTTP header name/value", true, nil, false},
	{"Host Name", "hostName", "Match on the request hostname", false, nil, false},
	{"Path", "path", "Match on the request URI path", false, nil, false},
	{"SSL Conn Has Cert", "sslConnHasCert", "TLS connection has a client certificate", false, []string{"equalTo"}, true},
	{"SSL DN Field", "sslDNField", "Match on an SSL certificate DN field", true, nil, false},
	{"SSL Verify Result", "sslVerifyResult", "Match on the SSL verification result", false, []string{"equalTo"}, true},
}

// compareTypeOptions lists all valid comparison types.
var lbL7RuleCompareOptions = []struct {
	label string
	value string
}{
	{"Contains", "contains"},
	{"Ends With", "endsWith"},
	{"Equal To", "equalTo"},
	{"Regex", "regex"},
	{"Starts With", "startsWith"},
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

// validCompareOptionsForType returns the subset of lbL7RuleCompareOptions allowed for the current rule type.
func validCompareOptionsForType(ruleTypeValue string) []struct{ label, value string } {
	var validValues []string
	for _, t := range lbL7RuleTypeOptions {
		if t.value == ruleTypeValue && t.validCompares != nil {
			validValues = t.validCompares
			break
		}
	}
	if validValues == nil {
		return lbL7RuleCompareOptions
	}
	valid := make(map[string]bool, len(validValues))
	for _, v := range validValues {
		valid[v] = true
	}
	var result []struct{ label, value string }
	for _, opt := range lbL7RuleCompareOptions {
		if valid[opt.value] {
			result = append(result, opt)
		}
	}
	return result
}

func (m Model) renderLBL7RuleWizardCompareStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7B68EE")).Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	opts := validCompareOptionsForType(m.wizard.l7RuleType)

	b.WriteString(titleStyle.Render("  Select Comparison Type") + "\n\n")
	b.WriteString(descStyle.Render("  Rule Type: "+m.wizard.l7RuleType) + "\n\n")

	for i, opt := range opts {
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

func (m Model) isBoolValueType() bool {
	for _, t := range lbL7RuleTypeOptions {
		if t.value == m.wizard.l7RuleType {
			return t.boolValue
		}
	}
	return false
}

func (m Model) renderLBL7RuleWizardValueStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("  Enter Value") + "\n\n")
	b.WriteString(descStyle.Render(fmt.Sprintf("  Type: %s  Comparison: %s", m.wizard.l7RuleType, m.wizard.l7RuleCompare)) + "\n\n")
	if m.wizard.l7RuleKey != "" {
		b.WriteString(descStyle.Render("  Key: "+m.wizard.l7RuleKey) + "\n\n")
	}

	if m.isBoolValueType() {
		// Boolean toggle: true / false
		trueStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		falseStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		if m.wizard.l7RuleValueInput == "true" {
			trueStyle = lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		} else {
			falseStyle = lipgloss.NewStyle().Background(lipgloss.Color("#CC3333")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		}
		b.WriteString("  " + trueStyle.Render("true") + "   " + falseStyle.Render("false") + "\n\n")
		b.WriteString(descStyle.Render("  ←/→: Toggle • Enter: Confirm • ←: Back\n"))
	} else {
		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)
		b.WriteString("  Value: " + inputStyle.Render(m.wizard.l7RuleValueInput+"█") + "\n\n")
		b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back\n"))
	}
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
		// Clamp compareIdx to the valid options for this type
		validOpts := validCompareOptionsForType(opt.value)
		if m.wizard.l7RuleCompareIdx >= len(validOpts) {
			m.wizard.l7RuleCompareIdx = 0
		}
		// If only one compare option, auto-select it and skip the compare step
		if len(validOpts) == 1 {
			m.wizard.l7RuleCompare = validOpts[0].value
			m.wizard.l7RuleCompareIdx = 0
			if opt.needsKey {
				m.wizard.step = LBL7RuleWizardStepKey
			} else {
				m.wizard.l7RuleKey = ""
				m.wizard.step = LBL7RuleWizardStepValue
			}
		} else {
			m.wizard.step = LBL7RuleWizardStepCompare
		}
	}
	return m, nil
}

func (m Model) handleLBL7RuleWizardCompareKeys(key string) (tea.Model, tea.Cmd) {
	opts := validCompareOptionsForType(m.wizard.l7RuleType)
	switch key {
	case "up", "k":
		if m.wizard.l7RuleCompareIdx > 0 {
			m.wizard.l7RuleCompareIdx--
		}
	case "down", "j":
		if m.wizard.l7RuleCompareIdx < len(opts)-1 {
			m.wizard.l7RuleCompareIdx++
		}
	case "enter":
		if m.wizard.l7RuleCompareIdx < len(opts) {
			m.wizard.l7RuleCompare = opts[m.wizard.l7RuleCompareIdx].value
		} else {
			m.wizard.l7RuleCompare = opts[0].value
		}
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
	if m.isBoolValueType() {
		// Ensure a default is set
		if m.wizard.l7RuleValueInput != "true" && m.wizard.l7RuleValueInput != "false" {
			m.wizard.l7RuleValueInput = "true"
		}
		switch key {
		case "left", "h", "right", "l", "tab":
			if m.wizard.l7RuleValueInput == "true" {
				m.wizard.l7RuleValueInput = "false"
			} else {
				m.wizard.l7RuleValueInput = "true"
			}
		case "enter":
			m.wizard.l7RuleValue = m.wizard.l7RuleValueInput
			m.wizard.step = LBL7RuleWizardStepInvert
		}
		return m, nil
	}
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

		// Only include key for types that require it
		needsKey := false
		for _, t := range lbL7RuleTypeOptions {
			if t.value == m.wizard.l7RuleType {
				needsKey = t.needsKey
				break
			}
		}
		body := map[string]interface{}{
			"ruleType":    m.wizard.l7RuleType,
			"compareType": m.wizard.l7RuleCompare,
			"value":       m.wizard.l7RuleValue,
			"invert":      m.wizard.l7RuleInvert,
		}
		if needsKey && m.wizard.l7RuleKey != "" {
			body["key"] = m.wizard.l7RuleKey
		}

		policyID := m.wizard.l7RulePolicyId
		region := m.wizard.l7RuleLBRegion

		if m.wizard.l7RuleEditId != "" {
			// Edit mode: PUT
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule/%s",
				m.cloudProject, url.PathEscape(region), url.PathEscape(policyID), url.PathEscape(m.wizard.l7RuleEditId))
			var result map[string]interface{}
			if err := httpLib.Client.Put(endpoint, body, &result); err != nil {
				return lbL7RuleCreatedMsg{policyID: policyID, err: fmt.Errorf("update failed: %w", err)}
			}
			return lbL7RuleCreatedMsg{policyID: policyID}
		}

		// Create mode: POST
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule",
			m.cloudProject, url.PathEscape(region), url.PathEscape(policyID))
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbL7RuleCreatedMsg{policyID: policyID, err: fmt.Errorf("creation failed: %w", err)}
		}
		return lbL7RuleCreatedMsg{policyID: policyID}
	}
}

func (m Model) executeDeleteLBL7Rule(policyID, ruleID, region string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbL7RuleDeletedMsg{policyID: policyID, err: fmt.Errorf("no cloud project selected")}
		}
		if policyID == "" || ruleID == "" || region == "" {
			return lbL7RuleDeletedMsg{policyID: policyID, err: fmt.Errorf("policy ID, rule ID or region missing")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s/l7Rule/%s",
			m.cloudProject, url.PathEscape(region), url.PathEscape(policyID), url.PathEscape(ruleID))
		if err := httpLib.Client.Delete(endpoint, nil); err != nil {
			return lbL7RuleDeletedMsg{policyID: policyID, err: fmt.Errorf("deletion failed: %w", err)}
		}
		return lbL7RuleDeletedMsg{policyID: policyID}
	}
}
