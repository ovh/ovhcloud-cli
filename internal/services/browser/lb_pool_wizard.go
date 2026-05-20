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

	b.WriteString(titleStyle.Render("Pool name") + "\n\n")
	b.WriteString(descStyle.Render("LB: "+m.wizard.lbPoolLBName+" ("+m.wizard.lbPoolLBRegion+")") + "\n\n")
	b.WriteString("  Name: " + inputStyle.Render(m.wizard.lbPoolNameInput+"█") + "\n\n")
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type name • Enter: Confirm • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBPoolWizardAlgoStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Load Balancer algorithm") + "\n\n")

	options := []struct{ label, value, desc string }{
		{"Round Robin", "roundRobin", "Distributes requests evenly in rotation"},
		{"Least Connections", "leastConnections", "Routes to the member with the fewest active connections"},
		{"Source IP", "sourceIP", "Same source IP always routed to the same member"},
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
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBPoolWizardProtoStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Pool protocol") + "\n\n")

	options := []struct{ label, value, desc string }{
		{"HTTP", "http", "HTTP protocol (layer 7)"},
		{"HTTPS", "https", "HTTPS protocol (layer 7, TLS terminated at backend)"},
		{"TCP", "tcp", "TCP protocol (layer 4)"},
		{"UDP", "udp", "UDP protocol (layer 4)"},
		{"PROXY", "proxy", "PROXY protocol (passes client information)"},
		{"PROXY v2", "proxyV2", "PROXY protocol v2 (binary version)"},
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
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBPoolWizardSessionStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	b.WriteString(titleStyle.Render("Session persistence") + "\n\n")

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
		Render("↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return b.String()
}

func (m Model) renderLBPoolWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	confirmTitle := "Confirm pool creation:"
	if m.wizard.lbPoolEditPoolId != "" {
		confirmTitle = "Confirm pool update:"
	}
	b.WriteString(titleStyle.Render(confirmTitle) + "\n\n")
	b.WriteString(labelStyle.Render("  Load Balancer:") + valStyle.Render(m.wizard.lbPoolLBName) + "\n")
	b.WriteString(labelStyle.Render("  Region:") + valStyle.Render(m.wizard.lbPoolLBRegion) + "\n")
	b.WriteString(labelStyle.Render("  Pool name:") + valStyle.Render(m.wizard.lbPoolName) + "\n")
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
	b.WriteString(labelStyle.Render("  Algorithm:") + valStyle.Render(algoLabel) + "\n")
	b.WriteString(labelStyle.Render("  Protocol:") + valStyle.Render(protoLabel) + "\n")
	sessionLabel := "None (disabled)"
	for _, o := range lbPoolSessionOptions {
		if o.value == m.wizard.lbPoolSession {
			sessionLabel = o.label
			break
		}
	}
	b.WriteString(labelStyle.Render("  Session:") + valStyle.Render(sessionLabel) + "\n\n")

	if m.wizard.isLoading {
		b.WriteString(loadingStyle.Render("⏳ "+m.wizard.loadingMessage))
		return b.String()
	}
	if m.wizard.errorMsg != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	btnLabel := " Create "
	if m.wizard.lbPoolEditPoolId != "" {
		btnLabel = " Update "
	}
	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2).Render(btnLabel)
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.lbPoolConfirmIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(btnLabel)
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	b.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBPoolWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbPoolNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
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
	{"None (disabled)", "disabled", "No session persistence"},
	{"Source IP", "sourceIP", "Persistence based on client source IP address"},
	{"HTTP Cookie", "httpCookie", "Persistence via HTTP cookie inserted by the load balancer"},
	{"App Cookie", "appCookie", "Persistence via existing cookie in the application response"},
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
			m.wizard.loadingMessage = "Updating pool..."
			return m, m.updateLBPool()
		}
		m.wizard.loadingMessage = "Creating pool..."
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
			return lbPoolCreatedMsg{err: fmt.Errorf("LB ID or region missing")}
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
			return lbPoolCreatedMsg{poolName: m.wizard.lbPoolName, err: fmt.Errorf("creation failed: %w", err)}
		}
		return lbPoolCreatedMsg{poolName: m.wizard.lbPoolName}
	}
}
