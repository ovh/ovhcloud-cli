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

var lbL7PolicyActionOptions = []struct{ label, value, desc string }{
	{"Redirect to Pool", "redirectToPool", "Redirects requests to an existing pool"},
	{"Redirect to URL", "redirectToURL", "Redirects requests to an absolute URL"},
	{"Redirect Prefix", "redirectPrefix", "Redirects requests by prefixing the URL"},
	{"Reject", "reject", "Rejects requests (returns 403)"},
}

// ─── Render ──────────────────────────────────────────────────────────────────

func (m Model) renderLBL7PolicyWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(40)

	b.WriteString(titleStyle.Render("L7 Policy name") + "\n\n")
	b.WriteString(descStyle.Render("Listener: "+m.wizard.l7PolicyListenerName) + "\n\n")
	b.WriteString("  Name: " + inputStyle.Render(m.wizard.l7PolicyNameInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type name • Enter: Confirm • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBL7PolicyWizardPositionStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(10)

	b.WriteString(titleStyle.Render("L7 Policy position") + "\n\n")
	b.WriteString(descStyle.Render("Positions start at 1. Policies are evaluated in ascending order.") + "\n\n")
	b.WriteString("  Position (>= 1): " + inputStyle.Render(m.wizard.l7PolicyPositionInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type position • Enter: Confirm • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBL7PolicyWizardActionStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("L7 Policy action") + "\n\n")
	for i, opt := range lbL7PolicyActionOptions {
		if i == m.wizard.l7PolicyActionIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		} else {
			b.WriteString(dimStyle.Render("  "+opt.label) + "\n")
			b.WriteString("   " + descStyle.Render(opt.desc) + "\n\n")
		}
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBL7PolicyWizardRedirectPoolStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	b.WriteString(titleStyle.Render("Target pool (Redirect to Pool)") + "\n\n")

	pools := m.lbPools[m.wizard.l7PolicyLBId]
	if len(pools) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("  (no pool available — create a pool first)") + "\n\n")
	} else {
		for i, p := range pools {
			pName := getStringValue(p, "name", "?")
			pID := getStringValue(p, "id", "")
			label := fmt.Sprintf("%s (%s)", pName, truncate(pID, 8))
			if i == m.wizard.l7PolicyRedirectPoolIdx {
				b.WriteString(selectedStyle.Render("▶ "+label) + "\n\n")
			} else {
				b.WriteString(dimStyle.Render("  "+label) + "\n\n")
			}
		}
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBL7PolicyWizardRedirectUrlStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E")).
		Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Width(50)

	label := "Redirect URL"
	hint := "e.g. https://example.com/new-path"
	if m.wizard.l7PolicyAction == "redirectPrefix" {
		label = "Redirect prefix"
		hint = "e.g. https://new.example.com"
	}
	b.WriteString(titleStyle.Render(label) + "\n\n")
	b.WriteString(descStyle.Render(hint) + "\n\n")
	b.WriteString("  URL: " + inputStyle.Render(m.wizard.l7PolicyRedirectUrlInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type URL • Enter: Confirm • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBL7PolicyWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(24)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	b.WriteString(titleStyle.Render("Confirm L7 Policy creation:") + "\n\n")
	b.WriteString(labelStyle.Render("  Listener:") + valStyle.Render(m.wizard.l7PolicyListenerName) + "\n")
	b.WriteString(labelStyle.Render("  Name:") + valStyle.Render(m.wizard.l7PolicyName) + "\n")
	b.WriteString(labelStyle.Render("  Position:") + valStyle.Render(strconv.Itoa(m.wizard.l7PolicyPosition)) + "\n")
	b.WriteString(labelStyle.Render("  Action:") + valStyle.Render(m.wizard.l7PolicyAction) + "\n")

	if m.wizard.l7PolicyAction == "redirectToPool" && m.wizard.l7PolicyRedirectPoolId != "" {
		poolLabel := m.wizard.l7PolicyRedirectPoolId
		pools := m.lbPools[m.wizard.l7PolicyLBId]
		for _, p := range pools {
			if getStringValue(p, "id", "") == m.wizard.l7PolicyRedirectPoolId {
				poolLabel = getStringValue(p, "name", poolLabel)
				break
			}
		}
		b.WriteString(labelStyle.Render("  Target pool:") + valStyle.Render(poolLabel) + "\n")
	} else if m.wizard.l7PolicyAction == "redirectToURL" || m.wizard.l7PolicyAction == "redirectPrefix" {
		b.WriteString(labelStyle.Render("  URL:") + valStyle.Render(m.wizard.l7PolicyRedirectUrl) + "\n")
	}
	b.WriteString("\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ " + m.wizard.loadingMessage))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.l7PolicyConfirmIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	b.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBL7PolicyWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.l7PolicyNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.l7PolicyName = name
		m.wizard.step = LBL7PolicyWizardStepPosition
	case "backspace":
		if len(m.wizard.l7PolicyNameInput) > 0 {
			m.wizard.l7PolicyNameInput = m.wizard.l7PolicyNameInput[:len(m.wizard.l7PolicyNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.l7PolicyNameInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBL7PolicyWizardPositionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		posStr := strings.TrimSpace(m.wizard.l7PolicyPositionInput)
		pos, err := strconv.Atoi(posStr)
		if err != nil || pos < 1 {
			m.wizard.errorMsg = "Invalid position (integer >= 1)"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.l7PolicyPosition = pos
		m.wizard.l7PolicyActionIdx = 0
		m.wizard.step = LBL7PolicyWizardStepAction
	case "backspace":
		if len(m.wizard.l7PolicyPositionInput) > 0 {
			m.wizard.l7PolicyPositionInput = m.wizard.l7PolicyPositionInput[:len(m.wizard.l7PolicyPositionInput)-1]
		}
	case "left":
		m.wizard.step = LBL7PolicyWizardStepName
	default:
		if len(key) == 1 && key >= "0" && key <= "9" {
			if len(m.wizard.l7PolicyPositionInput) < 5 {
				m.wizard.l7PolicyPositionInput += key
			}
		}
	}
	return m, nil
}

func (m Model) handleLBL7PolicyWizardActionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.l7PolicyActionIdx > 0 {
			m.wizard.l7PolicyActionIdx--
		}
	case "down", "j":
		if m.wizard.l7PolicyActionIdx < len(lbL7PolicyActionOptions)-1 {
			m.wizard.l7PolicyActionIdx++
		}
	case "enter":
		m.wizard.l7PolicyAction = lbL7PolicyActionOptions[m.wizard.l7PolicyActionIdx].value
		switch m.wizard.l7PolicyAction {
		case "redirectToPool":
			m.wizard.l7PolicyRedirectPoolIdx = 0
			m.wizard.step = LBL7PolicyWizardStepRedirectPool
		case "redirectToURL", "redirectPrefix":
			m.wizard.l7PolicyRedirectUrlInput = ""
			m.wizard.step = LBL7PolicyWizardStepRedirectUrl
		default: // reject
			m.wizard.l7PolicyConfirmIdx = 0
			m.wizard.step = LBL7PolicyWizardStepConfirm
		}
	case "left":
		m.wizard.step = LBL7PolicyWizardStepPosition
	}
	return m, nil
}

func (m Model) handleLBL7PolicyWizardRedirectPoolKeys(key string) (tea.Model, tea.Cmd) {
	pools := m.lbPools[m.wizard.l7PolicyLBId]
	maxIdx := len(pools) - 1
	if maxIdx < 0 {
		maxIdx = 0
	}
	switch key {
	case "up", "k":
		if m.wizard.l7PolicyRedirectPoolIdx > 0 {
			m.wizard.l7PolicyRedirectPoolIdx--
		}
	case "down", "j":
		if m.wizard.l7PolicyRedirectPoolIdx < maxIdx {
			m.wizard.l7PolicyRedirectPoolIdx++
		}
	case "enter":
		if len(pools) == 0 {
			return m, nil
		}
		m.wizard.l7PolicyRedirectPoolId = getStringValue(pools[m.wizard.l7PolicyRedirectPoolIdx], "id", "")
		m.wizard.l7PolicyConfirmIdx = 0
		m.wizard.step = LBL7PolicyWizardStepConfirm
	case "left":
		m.wizard.step = LBL7PolicyWizardStepAction
	}
	return m, nil
}

func (m Model) handleLBL7PolicyWizardRedirectUrlKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		u := strings.TrimSpace(m.wizard.l7PolicyRedirectUrlInput)
		if u == "" {
			m.wizard.errorMsg = "URL cannot be empty"
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.l7PolicyRedirectUrl = u
		m.wizard.l7PolicyConfirmIdx = 0
		m.wizard.step = LBL7PolicyWizardStepConfirm
	case "backspace":
		if len(m.wizard.l7PolicyRedirectUrlInput) > 0 {
			m.wizard.l7PolicyRedirectUrlInput = m.wizard.l7PolicyRedirectUrlInput[:len(m.wizard.l7PolicyRedirectUrlInput)-1]
		}
	case "left":
		m.wizard.step = LBL7PolicyWizardStepAction
	default:
		if len(key) == 1 {
			m.wizard.l7PolicyRedirectUrlInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBL7PolicyWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.l7PolicyConfirmIdx = 0
	case "right", "l":
		m.wizard.l7PolicyConfirmIdx = 1
	case "enter":
		if m.wizard.l7PolicyConfirmIdx == 1 {
			// Cancel → back to action step
			m.wizard.step = LBL7PolicyWizardStepAction
			return m, nil
		}
		m.wizard.isLoading = true
		if m.wizard.l7PolicyEditId != "" {
			m.wizard.loadingMessage = "Updating L7 Policy..."
			return m, m.updateLBL7Policy()
		}
		m.wizard.loadingMessage = "Creating L7 Policy..."
		return m, m.createLBL7Policy()
	}
	return m, nil
}

// ─── API ──────────────────────────────────────────────────────────────────────

func (m Model) createLBL7Policy() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbL7PolicyCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.l7PolicyListenerId == "" || m.wizard.l7PolicyLBRegion == "" {
			return lbL7PolicyCreatedMsg{err: fmt.Errorf("listener ID or region missing")}
		}

		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy",
			m.cloudProject, url.PathEscape(m.wizard.l7PolicyLBRegion))

		body := map[string]interface{}{
			"listenerId": m.wizard.l7PolicyListenerId,
			"name":       m.wizard.l7PolicyName,
			"position":   m.wizard.l7PolicyPosition,
			"action":     m.wizard.l7PolicyAction,
		}
		if m.wizard.l7PolicyAction == "redirectToPool" && m.wizard.l7PolicyRedirectPoolId != "" {
			body["redirectPoolId"] = m.wizard.l7PolicyRedirectPoolId
		}
		if m.wizard.l7PolicyAction == "redirectToURL" {
			body["redirectUrl"] = m.wizard.l7PolicyRedirectUrl
		}
		if m.wizard.l7PolicyAction == "redirectPrefix" {
			body["redirectPrefix"] = m.wizard.l7PolicyRedirectUrl
		}

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbL7PolicyCreatedMsg{policyName: m.wizard.l7PolicyName, err: fmt.Errorf("creation failed: %w", err)}
		}
		return lbL7PolicyCreatedMsg{policyName: m.wizard.l7PolicyName}
	}
}

func (m Model) fetchLBL7Policies(listenerID, region string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbL7PoliciesLoadedMsg{listenerID: listenerID, err: fmt.Errorf("no cloud project selected")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy",
			m.cloudProject, url.PathEscape(region))
		var all []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &all); err != nil {
			return lbL7PoliciesLoadedMsg{listenerID: listenerID, err: err}
		}
		// Filter by listenerId
		var policies []map[string]interface{}
		for _, p := range all {
			if getStringValue(p, "listenerId", "") == listenerID {
				policies = append(policies, p)
			}
		}
		return lbL7PoliciesLoadedMsg{listenerID: listenerID, policies: policies}
	}
}

func (m Model) executeDeleteLBL7Policy() tea.Cmd {
	return func() tea.Msg {
		if m.selectedLBL7Policy == nil {
			return lbL7PolicyDeletedMsg{err: fmt.Errorf("no L7 policy selected")}
		}
		if m.cloudProject == "" {
			return lbL7PolicyDeletedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		policyID := getStringValue(m.selectedLBL7Policy, "id", "")
		policyName := getStringValue(m.selectedLBL7Policy, "name", "")
		region := getStringValue(m.detailData, "region", "")
		if policyID == "" || region == "" {
			return lbL7PolicyDeletedMsg{err: fmt.Errorf("L7 policy ID or region not found")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s",
			m.cloudProject, url.PathEscape(region), url.PathEscape(policyID))
		if err := httpLib.Client.Delete(endpoint, nil); err != nil {
			return lbL7PolicyDeletedMsg{policyName: policyName, err: fmt.Errorf("deletion failed: %w", err)}
		}
		return lbL7PolicyDeletedMsg{policyName: policyName}
	}
}

func (m Model) updateLBL7Policy() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbL7PolicyUpdatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.l7PolicyEditId == "" || m.wizard.l7PolicyLBRegion == "" {
			return lbL7PolicyUpdatedMsg{err: fmt.Errorf("L7 policy ID or region missing")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/l7Policy/%s",
			m.cloudProject, url.PathEscape(m.wizard.l7PolicyLBRegion), url.PathEscape(m.wizard.l7PolicyEditId))
		body := map[string]interface{}{
			"name":     m.wizard.l7PolicyName,
			"position": m.wizard.l7PolicyPosition,
			"action":   m.wizard.l7PolicyAction,
		}
		if m.wizard.l7PolicyAction == "redirectToPool" && m.wizard.l7PolicyRedirectPoolId != "" {
			body["redirectPoolId"] = m.wizard.l7PolicyRedirectPoolId
		}
		if m.wizard.l7PolicyAction == "redirectToURL" {
			body["redirectUrl"] = m.wizard.l7PolicyRedirectUrl
		}
		if m.wizard.l7PolicyAction == "redirectPrefix" {
			body["redirectPrefix"] = m.wizard.l7PolicyRedirectUrl
		}
		var result map[string]interface{}
		if err := httpLib.Client.Put(endpoint, body, &result); err != nil {
			return lbL7PolicyUpdatedMsg{policyName: m.wizard.l7PolicyName, err: fmt.Errorf("update failed: %w", err)}
		}
		return lbL7PolicyUpdatedMsg{policyName: m.wizard.l7PolicyName}
	}
}
