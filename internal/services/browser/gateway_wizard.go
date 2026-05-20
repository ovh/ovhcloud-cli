// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// gatewayModels lists the gateway sizes available across all regions.
var gatewayModels = []string{"s", "m", "l", "xl", "2xl"}

// gwActiveModels returns the model list (always the static one; no per-region API exists).
func (m Model) gwActiveModels() []string {
	return gatewayModels
}

// ─── API / fetch ──────────────────────────────────────────────────────────────

// fetchGwRegions loads only the regions that have the "network" service UP.
func (m Model) fetchGwRegions() tea.Cmd {
	return func() tea.Msg {
		regions, err := m.fetchNetworkRegions()
		if err != nil {
			return gwRegionsLoadedMsg{err: err}
		}
		sort.Strings(regions)
		return gwRegionsLoadedMsg{regions: regions}
	}
}

// fetchGwSubnet fetches the first subnet of a regional network and returns its ID.
func (m Model) fetchGwSubnet(networkID string) tea.Cmd {
	region := m.wizard.gwRegion
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
			m.cloudProject, url.PathEscape(region), url.PathEscape(networkID))
		var subnets []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &subnets); err != nil || len(subnets) == 0 {
			// No subnet exists yet — network selected but not ready for gateway
			return gwSubnetLoadedMsg{subnetID: "", err: fmt.Errorf("this network has no compatible subnet (noGateway=true). First create a subnet using the 'OVH Gateway' option.")}
		}
		return gwSubnetLoadedMsg{subnetID: getStringValue(subnets[0], "id", "")}
	}
}

func (m Model) fetchGwNetworks() tea.Cmd {
	region := m.wizard.gwRegion
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network",
			m.cloudProject, url.PathEscape(region))
		var nets []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &nets); err != nil {
			return gwNetworksLoadedMsg{err: err}
		}
		// Exclude the external/public network
		var filtered []map[string]interface{}
		for _, n := range nets {
			name := getStringValue(n, "name", "")
			if name != "" && name != "Ext-Net" {
				filtered = append(filtered, n)
			}
		}
		return gwNetworksLoadedMsg{networks: filtered}
	}
}

// createGatewayFromWizard creates the gateway.
// When a network+subnet are selected, uses the subnet-specific endpoint.
// Otherwise creates a standalone gateway.
func (m Model) createGatewayFromWizard() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return gwCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		model := func() string {
			models := m.gwActiveModels()
			if m.wizard.gwModelIdx < len(models) {
				return models[m.wizard.gwModelIdx]
			}
			return "s"
		}()
		body := map[string]interface{}{
			"model": model,
			"name":  m.wizard.gwName,
		}

		var endpoint string
		if m.wizard.gwNetworkID != "" && m.wizard.gwSubnetID != "" {
			// Attach to a specific subnet
			endpoint = fmt.Sprintf(
				"/v1/cloud/project/%s/region/%s/network/%s/subnet/%s/gateway",
				m.cloudProject,
				url.PathEscape(m.wizard.gwRegion),
				url.PathEscape(m.wizard.gwNetworkID),
				url.PathEscape(m.wizard.gwSubnetID),
			)
		} else {
			// Standalone gateway (no network attachment)
			endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway",
				m.cloudProject, url.PathEscape(m.wizard.gwRegion))
		}

		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "gateway IP must not be used by a port") ||
				strings.Contains(errMsg, "gateway IP") && strings.Contains(errMsg, "port") {
				return gwCreatedMsg{err: fmt.Errorf(
					"the subnet already has a gateway IP in use. " +
						"The OVH Gateway can only be created on a subnet configured without a static gateway (mode 'OVH Gateway'). " +
						"Recreate the private network with this option")}
			}
			return gwCreatedMsg{err: fmt.Errorf("failed to create gateway: %w", err)}
		}
		return gwCreatedMsg{gateway: result}
	}
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderGwWizardRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	content.WriteString(titleStyle.Render("Choose region:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading regions..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.gwAvailableRegions) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No regions available.") + "\n")
	} else {
		for i, r := range m.wizard.gwAvailableRegions {
			if i == m.wizard.gwRegionIdx {
				content.WriteString(selectedStyle.Render("▶ "+r) + "\n")
			} else {
				content.WriteString(dimStyle.Render("  "+r) + "\n")
			}
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • Esc: Cancel"))
	return content.String()
}

func (m Model) renderGwWizardModelStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Choose gateway size:") + "\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf("Region: %s", m.wizard.gwRegion)) + "\n\n")

	models := m.gwActiveModels()
	for i, model := range models {
		if i == m.wizard.gwModelIdx {
			content.WriteString(selectedStyle.Render("▶ "+strings.ToUpper(model)) + "\n")
		} else {
			content.WriteString(dimStyle.Render("  "+strings.ToUpper(model)) + "\n")
		}
	}

	backHint := "←: Back "
	if len(m.wizard.gwAvailableRegions) == 0 {
		backHint = ""
	}
	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • "+backHint+"• Esc: Cancel"))
	return content.String()
}

func (m Model) renderGwWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	models := m.gwActiveModels()
	modelName := ""
	if m.wizard.gwModelIdx < len(models) {
		modelName = strings.ToUpper(models[m.wizard.gwModelIdx])
	}

	content.WriteString(titleStyle.Render("Gateway name:") + "\n\n")
	content.WriteString(descStyle.Render(
		fmt.Sprintf("Region: %s  •  Size: %s",
			m.wizard.gwRegion, modelName),
	) + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(40)
	content.WriteString(inputStyle.Render(m.wizard.gwNameInput+"▌") + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Type a name • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderGwWizardNetworkStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Attach to a private network:") + "\n\n")
	content.WriteString(descStyle.Render(
		fmt.Sprintf("Region: %s  •  Size: %s  •  Name: %s",
			m.wizard.gwRegion, func() string {
				models := m.gwActiveModels()
				if m.wizard.gwModelIdx < len(models) {
					return strings.ToUpper(models[m.wizard.gwModelIdx])
				}
				return ""
			}(), m.wizard.gwName),
	) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading networks..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.gwAvailableNetworks) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No private network available in this region.") + "\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("←: Back • Esc: Cancel"))
	} else {
		for i, net := range m.wizard.gwAvailableNetworks {
			name := getStringValue(net, "name", getStringValue(net, "id", "unknown"))
			if i == m.wizard.gwNetworkIdx {
				content.WriteString(selectedStyle.Render("▶ "+name) + "\n")
			} else {
				content.WriteString(dimStyle.Render("  "+name) + "\n")
			}
		}
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("↑↓ Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	}
	return content.String()
}

func (m Model) renderGwWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("Confirm gateway creation:") + "\n\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.gwRegion) + "\n")
	content.WriteString(labelStyle.Render("  Size:") + valueStyle.Render(func() string {
		models := m.gwActiveModels()
		if m.wizard.gwModelIdx < len(models) {
			return strings.ToUpper(models[m.wizard.gwModelIdx])
		}
		return ""
	}()) + "\n")
	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.gwName) + "\n")
	if m.wizard.gwNetworkName != "" {
		content.WriteString(labelStyle.Render("  Network:") + valueStyle.Render(m.wizard.gwNetworkName) + "\n")
	}
	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Creating..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.gwConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleGwWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.gwRegionIdx > 0 {
			m.wizard.gwRegionIdx--
		}
	case "down", "j":
		if m.wizard.gwRegionIdx < len(m.wizard.gwAvailableRegions)-1 {
			m.wizard.gwRegionIdx++
		}
	case "enter":
		if len(m.wizard.gwAvailableRegions) > 0 {
			m.wizard.gwRegion = m.wizard.gwAvailableRegions[m.wizard.gwRegionIdx]
			// Attach mode: look up openstackId + subnetId for the selected region
			if m.wizard.gwNetworkRegionMap != nil {
				if regionData, ok := m.wizard.gwNetworkRegionMap[m.wizard.gwRegion]; ok {
					m.wizard.gwNetworkID = regionData["openstackId"]
					m.wizard.gwSubnetID = regionData["subnetId"]
				}
			}
			m.wizard.gwModelIdx = 0
			m.wizard.step = GwWizardStepModel
		}
	}
	return m, nil
}

func (m Model) handleGwWizardModelKeys(key string) (tea.Model, tea.Cmd) {
	models := m.gwActiveModels()
	switch key {
	case "up", "k":
		if m.wizard.gwModelIdx > 0 {
			m.wizard.gwModelIdx--
		}
	case "down", "j":
		if m.wizard.gwModelIdx < len(models)-1 {
			m.wizard.gwModelIdx++
		}
	case "enter":
		m.wizard.step = GwWizardStepName
	case "left":
		// Only go back to region step if we came from there (full wizard)
		if len(m.wizard.gwAvailableRegions) > 0 {
			m.wizard.step = GwWizardStepRegion
		}
	}
	return m, nil
}

func (m Model) handleGwWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.gwNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
			return m, nil
		}
		m.wizard.gwName = name
		m.wizard.errorMsg = ""
		// Attach mode (region map populated) or standalone with network pre-set
		// → skip network-selection step, go straight to confirm
		if m.wizard.gwNetworkRegionMap != nil || (m.wizard.gwNetworkID != "" && len(m.wizard.gwAvailableNetworks) == 0) {
			m.wizard.step = GwWizardStepConfirm
			return m, nil
		}
		// Full wizard: load networks for selected region
		m.wizard.step = GwWizardStepNetwork
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading networks..."
		return m, m.fetchGwNetworks()
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

func (m Model) handleGwWizardNetworkKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.gwNetworkIdx > 0 {
			m.wizard.gwNetworkIdx--
		}
	case "down", "j":
		if m.wizard.gwNetworkIdx < len(m.wizard.gwAvailableNetworks)-1 {
			m.wizard.gwNetworkIdx++
		}
	case "enter":
		if len(m.wizard.gwAvailableNetworks) > 0 {
			net := m.wizard.gwAvailableNetworks[m.wizard.gwNetworkIdx]
			m.wizard.gwNetworkID = getStringValue(net, "id", "")
			m.wizard.gwNetworkName = getStringValue(net, "name", getStringValue(net, "id", "unknown"))
			// Regional network API returns subnets as a list of ID strings — extract the first one
			subnetID := ""
			if subnets, ok := net["subnets"].([]interface{}); ok && len(subnets) > 0 {
				switch v := subnets[0].(type) {
				case string:
					subnetID = v
				case map[string]interface{}:
					subnetID = getStringValue(v, "id", "")
				}
			}
			if subnetID != "" {
				m.wizard.gwSubnetID = subnetID
				m.wizard.step = GwWizardStepConfirm
				return m, nil
			}
			// Subnet IDs not embedded — fetch them
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Verifying subnet..."
			return m, m.fetchGwSubnet(m.wizard.gwNetworkID)
		}
	case "left":
		m.wizard.step = GwWizardStepName
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
			// Cancel → back to previous step
			if m.wizard.gwNetworkRegionMap != nil || (m.wizard.gwNetworkID != "" && len(m.wizard.gwAvailableNetworks) == 0) {
				m.wizard.step = GwWizardStepName
			} else {
				m.wizard.step = GwWizardStepNetwork
			}
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating gateway..."
		return m, m.createGatewayFromWizard()
	}
	return m, nil
}

