// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// lbHMTypeOptions lists the valid health monitor types.
var lbHMTypeOptions = []struct {
	label string
	value string
}{
	{"HTTP", "http"},
	{"HTTPS", "https"},
	{"TCP", "tcp"},
	{"Ping", "ping"},
	{"TLS Hello", "tls-hello"},
	{"UDP Connect", "udp-connect"},
	{"SCTP", "sctp"},
}

// lbHMTypeNeedsHttpConfig returns true for types that require an httpConfiguration.
func lbHMTypeNeedsHttpConfig(t string) bool {
	return t == "http" || t == "https"
}

// lbHMHttpMethodOptions lists the valid HTTP methods for health monitors.
var lbHMHttpMethodOptions = []string{
	"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT", "TRACE",
}

// hmNameValid returns true if the name only contains letters, digits, underscores, dashes or dots.
var hmNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

func hmNameValid(name string) bool {
	return hmNameRegex.MatchString(name)
}

// ─── renderLBHealthMonitorView ────────────────────────────────────────────────

func (m Model) renderLBHealthMonitorView(width int) string {
	var content strings.Builder

	if m.selectedLBPool == nil {
		return "No pool selected"
	}

	poolID := getStringValue(m.selectedLBPool, "id", "")
	poolName := getStringValue(m.selectedLBPool, "name", "N/A")
	hm := m.lbHealthMonitors[poolID]

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Actions bar
	var actionsContent string
	if hm != nil {
		editBtn := lipgloss.NewStyle().
			Background(lipgloss.Color("#7B68EE")).Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 1).Render("✏ Edit  [e]")
		var deleteBtnContent string
		if m.lbHMConfirm {
			deleteBtnContent = lipgloss.NewStyle().
				Background(lipgloss.Color("#FF4444")).Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render("⚠ Confirm delete? Press [d] again • Esc to cancel")
		} else {
			deleteBtnContent = lipgloss.NewStyle().
				Background(lipgloss.Color("#CC3333")).Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render("🗑 Delete  [d]")
		}
		actionsContent = lipgloss.JoinHorizontal(lipgloss.Top, editBtn, "  ", deleteBtnContent)
	} else {
		createBtn := lipgloss.NewStyle().
			Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 1).Render("+ Create  [Enter]")
		actionsContent = createBtn
	}
	actionsBox := renderBox("Actions", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	titleLine := fmt.Sprintf("Health Monitor — Pool: %s", poolName)

	if hm == nil {
		empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  No health monitor configured for this pool.")
		content.WriteString(renderBox(titleLine, empty, width-4))
		return content.String()
	}

	hmID := getStringValue(hm, "id", "N/A")
	hmName := getStringValue(hm, "name", "N/A")
	hmMonitorType := getStringValue(hm, "monitorType", "N/A")
	hmDelay := fmt.Sprintf("%v s", hm["delay"])
	hmTimeout := fmt.Sprintf("%v s", hm["timeout"])
	hmMaxRetries := fmt.Sprintf("%v", hm["maxRetries"])
	hmMaxRetriesDown := fmt.Sprintf("%v", hm["maxRetriesDown"])
	hmOpStatus := getStringValue(hm, "operatingStatus", "N/A")
	hmProvStatus := getStringValue(hm, "provisioningStatus", "N/A")
	// Extract HTTP-specific fields from httpConfiguration if present
	hmHttpMethod := ""
	hmUrlPath := ""
	hmExpectedCodes := ""
	if httpCfg, ok := hm["httpConfiguration"].(map[string]interface{}); ok {
		hmHttpMethod = getStringValue(httpCfg, "httpMethod", "")
		hmUrlPath = getStringValue(httpCfg, "urlPath", "")
		hmExpectedCodes = getStringValue(httpCfg, "expectedCodes", "")
	}

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(hmOpStatus) != "online" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(hmOpStatus) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}

	var detailContent strings.Builder
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(hmID, 36))))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Name"), valueSt.Render(hmName)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Type"), valueSt.Render(hmMonitorType)))
	if hmHttpMethod != "" {
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("HTTP Method"), valueSt.Render(hmHttpMethod)))
	}
	if hmUrlPath != "" {
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("URL Path"), valueSt.Render(hmUrlPath)))
	}
	if hmExpectedCodes != "" {
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Expected Codes"), valueSt.Render(hmExpectedCodes)))
	}
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Delay"), valueSt.Render(hmDelay)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Timeout"), valueSt.Render(hmTimeout)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Max Retries"), valueSt.Render(hmMaxRetries)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Max Retries Down"), valueSt.Render(hmMaxRetriesDown)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(hmOpStatus)))
	detailContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Prov. Status"), valueSt.Render(hmProvStatus)))

	content.WriteString(renderBox(titleLine, detailContent.String(), width-4))
	return content.String()
}

// ─── Wizard render steps ──────────────────────────────────────────────────────

func (m Model) renderLBHMWizardNameStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	action := "Create"
	if m.wizard.lbHMEditId != "" {
		action = "Edit"
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s Health Monitor — Name", action)) + "\n\n")
	b.WriteString(descStyle.Render("  Only letters, digits, underscores, dashes or dots allowed.") + "\n\n")
	b.WriteString("  Name: " + inputStyle.Render(m.wizard.lbHMNameInput+"█") + "\n")
	if len(m.wizard.lbHMNameInput) > 0 && !hmNameValid(m.wizard.lbHMNameInput) {
		b.WriteString("  " + warnStyle.Render("⚠ Invalid characters") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardTypeStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7B68EE")).Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — Type") + "\n\n")
	b.WriteString(descStyle.Render("  Protocol used to probe the backend members.") + "\n\n")
	for i, opt := range lbHMTypeOptions {
		if i == m.wizard.lbHMTypeIdx {
			b.WriteString(selectedStyle.Render("▶ "+opt.label) + "\n")
		} else {
			b.WriteString(unselectedStyle.Render("  "+opt.label) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  ↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardHttpMethodStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7B68EE")).Padding(0, 1)
	unselectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — HTTP Method") + "\n\n")
	b.WriteString(descStyle.Render("  HTTP method used to probe the backend members.") + "\n\n")
	for i, method := range lbHMHttpMethodOptions {
		if i == m.wizard.lbHMHttpMethodIdx {
			b.WriteString(selectedStyle.Render("▶ "+method) + "\n")
		} else {
			b.WriteString(unselectedStyle.Render("  "+method) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  ↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardUrlPathStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	b.WriteString(titleStyle.Render("  Health Monitor — URL Path") + "\n\n")
	b.WriteString(descStyle.Render("  HTTP path to probe on the backend members (must start with /).") + "\n\n")
	b.WriteString("  URL Path: " + inputStyle.Render(m.wizard.lbHMUrlPathInput+"█") + "\n")
	if v := strings.TrimSpace(m.wizard.lbHMUrlPathInput); len(v) > 0 && v[0] != '/' {
		b.WriteString("  " + warnStyle.Render("⚠ Must start with /") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardExpectedCodesStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — Expected HTTP Codes") + "\n\n")
	b.WriteString(descStyle.Render("  HTTP status code considered healthy (e.g. 200).") + "\n\n")
	b.WriteString("  Expected Codes: " + inputStyle.Render(m.wizard.lbHMExpectedCodesInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardDelayStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — Delay") + "\n\n")
	b.WriteString(descStyle.Render("  Time in seconds between two consecutive checks (≥ 1).") + "\n\n")
	b.WriteString("  Delay (s): " + inputStyle.Render(m.wizard.lbHMDelayInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardMaxRetriesStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — Max Retries") + "\n\n")
	b.WriteString(descStyle.Render("  Number of successful checks before marking the member as UP (1–10).") + "\n\n")
	b.WriteString("  Max Retries: " + inputStyle.Render(m.wizard.lbHMMaxRetriesInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardMaxRetriesDownStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)

	b.WriteString(titleStyle.Render("  Health Monitor — Max Retries Down") + "\n\n")
	b.WriteString(descStyle.Render("  Number of failed checks before marking the member as DOWN (1–10).") + "\n\n")
	b.WriteString("  Max Retries Down: " + inputStyle.Render(m.wizard.lbHMMaxRetriesDownInput+"█") + "\n\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardTimeoutStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#1A1A2E")).Padding(0, 1)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	b.WriteString(titleStyle.Render("  Health Monitor — Timeout") + "\n\n")
	maxTimeout := m.wizard.lbHMDelay
	if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) && maxTimeout > 4 {
		maxTimeout = 4
	}
	b.WriteString(descStyle.Render(fmt.Sprintf("  Maximum wait time in seconds per check. Must be ≥ 1 and ≤ %d s.", maxTimeout)) + "\n\n")
	b.WriteString("  Timeout (s): " + inputStyle.Render(m.wizard.lbHMTimeoutInput+"█") + "\n")
	if v, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMTimeoutInput)); err == nil {
		maxT := m.wizard.lbHMDelay
		if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) && maxT > 4 {
			maxT = 4
		}
		if v > maxT {
			b.WriteString("  " + warnStyle.Render(fmt.Sprintf("⚠ Timeout must be ≤ %d s", maxT)) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(descStyle.Render("  Enter: Confirm • Backspace: Delete • ←: Back • Esc: Cancel\n"))
	return b.String()
}

func (m Model) renderLBHMWizardConfirmStep(width int) string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	readOnlySt := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	action := "Create"
	if m.wizard.lbHMEditId != "" {
		action = "Update"
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("  Confirm %s Health Monitor", action)) + "\n\n")
	b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Name"), valueSt.Render(m.wizard.lbHMName)))
	if m.wizard.lbHMEditId != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Type"), readOnlySt.Render(m.wizard.lbHMType+" (read-only)")))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Type"), valueSt.Render(m.wizard.lbHMType)))
	}
	if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) {
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  HTTP Method"), valueSt.Render(m.wizard.lbHMHttpMethod)))
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  URL Path"), valueSt.Render(m.wizard.lbHMUrlPath)))
		b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Expected Codes"), valueSt.Render(m.wizard.lbHMExpectedCodes)))
	}
	b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Delay"), valueSt.Render(strconv.Itoa(m.wizard.lbHMDelay)+" s")))
	b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Max Retries"), valueSt.Render(strconv.Itoa(m.wizard.lbHMMaxRetries))))
	b.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("  Max Retries Down"), valueSt.Render(strconv.Itoa(m.wizard.lbHMMaxRetriesDown))))
	b.WriteString(fmt.Sprintf("%s %s\n\n", labelSt.Render("  Timeout"), valueSt.Render(strconv.Itoa(m.wizard.lbHMTimeout)+" s")))

	saveStyle := lipgloss.NewStyle().Background(lipgloss.Color("#00AA55")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	cancelStyle := lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	if m.wizard.lbHMConfirmIdx == 1 {
		saveStyle = lipgloss.NewStyle().Background(lipgloss.Color("#555555")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
		cancelStyle = lipgloss.NewStyle().Background(lipgloss.Color("#CC3333")).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 2)
	}
	b.WriteString("  " + saveStyle.Render(action) + "  " + cancelStyle.Render("Cancel") + "\n\n")
	b.WriteString(descStyle.Render("  ←→: Select • Enter: Confirm • ←: Back • Esc: Cancel\n"))
	return b.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBHMWizardNameKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbHMNameInput)
		if name == "" || !hmNameValid(name) {
			return m, nil
		}
		m.wizard.lbHMName = name
		if m.wizard.lbHMEditId != "" {
			// In edit mode, type is immutable
			if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) {
				m.wizard.step = LBHMWizardStepHttpMethod
			} else {
				m.wizard.step = LBHMWizardStepDelay
			}
		} else {
			m.wizard.step = LBHMWizardStepType
		}
	case "backspace":
		if len(m.wizard.lbHMNameInput) > 0 {
			m.wizard.lbHMNameInput = m.wizard.lbHMNameInput[:len(m.wizard.lbHMNameInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbHMNameInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBHMWizardTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbHMTypeIdx > 0 {
			m.wizard.lbHMTypeIdx--
		}
	case "down", "j":
		if m.wizard.lbHMTypeIdx < len(lbHMTypeOptions)-1 {
			m.wizard.lbHMTypeIdx++
		}
	case "enter":
		m.wizard.lbHMType = lbHMTypeOptions[m.wizard.lbHMTypeIdx].value
		if m.wizard.lbHMDelayInput == "" {
			m.wizard.lbHMDelayInput = "5"
		}
		if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) {
			m.wizard.step = LBHMWizardStepHttpMethod
		} else {
			m.wizard.step = LBHMWizardStepDelay
		}
	case "left", "h":
		m.wizard.step = LBHMWizardStepName
	}
	return m, nil
}

func (m Model) handleLBHMWizardHttpMethodKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbHMHttpMethodIdx > 0 {
			m.wizard.lbHMHttpMethodIdx--
		}
	case "down", "j":
		if m.wizard.lbHMHttpMethodIdx < len(lbHMHttpMethodOptions)-1 {
			m.wizard.lbHMHttpMethodIdx++
		}
	case "enter":
		m.wizard.lbHMHttpMethod = lbHMHttpMethodOptions[m.wizard.lbHMHttpMethodIdx]
		if m.wizard.lbHMUrlPathInput == "" {
			m.wizard.lbHMUrlPathInput = "/"
		}
		if m.wizard.lbHMExpectedCodesInput == "" {
			m.wizard.lbHMExpectedCodesInput = "200"
		}
		m.wizard.step = LBHMWizardStepUrlPath
	case "left", "h":
		if m.wizard.lbHMEditId != "" {
			m.wizard.step = LBHMWizardStepName
		} else {
			m.wizard.step = LBHMWizardStepType
		}
	}
	return m, nil
}

func (m Model) handleLBHMWizardUrlPathKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v := strings.TrimSpace(m.wizard.lbHMUrlPathInput)
		if v == "" || v[0] != '/' {
			return m, nil
		}
		m.wizard.lbHMUrlPath = v
		m.wizard.step = LBHMWizardStepExpectedCodes
	case "left", "h":
		m.wizard.step = LBHMWizardStepHttpMethod
	case "backspace":
		if len(m.wizard.lbHMUrlPathInput) > 0 {
			m.wizard.lbHMUrlPathInput = m.wizard.lbHMUrlPathInput[:len(m.wizard.lbHMUrlPathInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbHMUrlPathInput += key
		}
	}
	return m, nil
}

func (m Model) handleLBHMWizardExpectedCodesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v := strings.TrimSpace(m.wizard.lbHMExpectedCodesInput)
		if v == "" {
			return m, nil
		}
		m.wizard.lbHMExpectedCodes = v
		if m.wizard.lbHMDelayInput == "" {
			m.wizard.lbHMDelayInput = "5"
		}
		m.wizard.step = LBHMWizardStepDelay
	case "left", "h":
		m.wizard.step = LBHMWizardStepUrlPath
	case "backspace":
		if len(m.wizard.lbHMExpectedCodesInput) > 0 {
			m.wizard.lbHMExpectedCodesInput = m.wizard.lbHMExpectedCodesInput[:len(m.wizard.lbHMExpectedCodesInput)-1]
		}
	default:
		if len(key) == 1 {
			m.wizard.lbHMExpectedCodesInput += key
		}
	}
	return m, nil
}

func intInput(current string, key string) string {
	switch key {
	case "backspace":
		if len(current) > 0 {
			return current[:len(current)-1]
		}
	default:
		if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			return current + key
		}
	}
	return current
}

func (m Model) handleLBHMWizardDelayKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMDelayInput))
		if err != nil || v < 1 {
			return m, nil
		}
		m.wizard.lbHMDelay = v
		// Clamp timeout to min(delay, 4 for http/https)
		if tv, err2 := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMTimeoutInput)); err2 == nil {
			maxT := v
			if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) && maxT > 4 {
				maxT = 4
			}
			if tv > maxT {
				m.wizard.lbHMTimeoutInput = strconv.Itoa(maxT)
			}
		}
		if m.wizard.lbHMMaxRetriesInput == "" {
			m.wizard.lbHMMaxRetriesInput = "3"
		}
		m.wizard.step = LBHMWizardStepMaxRetries
	case "left", "h":
		if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) {
			m.wizard.step = LBHMWizardStepExpectedCodes
		} else if m.wizard.lbHMEditId != "" {
			m.wizard.step = LBHMWizardStepName
		} else {
			m.wizard.step = LBHMWizardStepType
		}
	default:
		m.wizard.lbHMDelayInput = intInput(m.wizard.lbHMDelayInput, key)
	}
	return m, nil
}

func (m Model) handleLBHMWizardMaxRetriesKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMMaxRetriesInput))
		if err != nil || v < 1 || v > 10 {
			return m, nil
		}
		m.wizard.lbHMMaxRetries = v
		if m.wizard.lbHMMaxRetriesDownInput == "" {
			m.wizard.lbHMMaxRetriesDownInput = "3"
		}
		m.wizard.step = LBHMWizardStepMaxRetriesDown
	case "left", "h":
		m.wizard.step = LBHMWizardStepDelay
	default:
		m.wizard.lbHMMaxRetriesInput = intInput(m.wizard.lbHMMaxRetriesInput, key)
	}
	return m, nil
}

func (m Model) handleLBHMWizardMaxRetriesDownKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMMaxRetriesDownInput))
		if err != nil || v < 1 || v > 10 {
			return m, nil
		}
		m.wizard.lbHMMaxRetriesDown = v
		// Default timeout = min(5, effective max)
		if m.wizard.lbHMTimeoutInput == "" {
			defaultTimeout := 5
			if m.wizard.lbHMDelay > 0 && defaultTimeout > m.wizard.lbHMDelay {
				defaultTimeout = m.wizard.lbHMDelay
			}
			if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) && defaultTimeout > 4 {
				defaultTimeout = 4
			}
			m.wizard.lbHMTimeoutInput = strconv.Itoa(defaultTimeout)
		}
		m.wizard.step = LBHMWizardStepTimeout
	case "left", "h":
		m.wizard.step = LBHMWizardStepMaxRetries
	default:
		m.wizard.lbHMMaxRetriesDownInput = intInput(m.wizard.lbHMMaxRetriesDownInput, key)
	}
	return m, nil
}

func (m Model) handleLBHMWizardTimeoutKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		v, err := strconv.Atoi(strings.TrimSpace(m.wizard.lbHMTimeoutInput))
		if err != nil || v < 1 {
			return m, nil
		}
		maxTimeout := m.wizard.lbHMDelay
		if lbHMTypeNeedsHttpConfig(m.wizard.lbHMType) && maxTimeout > 4 {
			maxTimeout = 4
		}
		// timeout must be <= min(delay, 4 for http/https)
		if m.wizard.lbHMDelay > 0 && v > maxTimeout {
			return m, nil
		}
		m.wizard.lbHMTimeout = v
		m.wizard.lbHMConfirmIdx = 0
		m.wizard.step = LBHMWizardStepConfirm
	case "left", "h":
		m.wizard.step = LBHMWizardStepMaxRetriesDown
	default:
		m.wizard.lbHMTimeoutInput = intInput(m.wizard.lbHMTimeoutInput, key)
	}
	return m, nil
}

func (m Model) handleLBHMWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.lbHMConfirmIdx = 0
	case "right", "l":
		m.wizard.lbHMConfirmIdx = 1
	case "enter":
		if m.wizard.lbHMConfirmIdx == 1 {
			m.wizard = WizardData{}
			m.mode = LBHealthMonitorView
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Saving health monitor..."
		return m, m.saveHealthMonitor()
	}
	return m, nil
}
