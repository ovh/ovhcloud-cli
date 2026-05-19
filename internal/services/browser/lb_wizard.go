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

// ─── API / fetch ──────────────────────────────────────────────────────────────

// fetchLBRegions loads regions that support the Octavia load balancer service.
func (m Model) fetchLBRegions() tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
		var allNames []string
		if err := httpLib.Client.Get(endpoint, &allNames); err != nil {
			return lbRegionsLoadedMsg{err: err}
		}
		ids := make([]any, len(allNames))
		for i, n := range allNames {
			ids[i] = n
		}
		details, _ := httpLib.FetchObjectsParallel[map[string]any](endpoint+"/%s", ids, true)
		var regions []string
		for i, r := range details {
			if r == nil {
				continue
			}
			if services, ok := r["services"].([]interface{}); ok {
				for _, svc := range services {
					if sm, ok := svc.(map[string]interface{}); ok {
						if sm["name"] == "octavialoadbalancer" && sm["status"] == "UP" {
							regions = append(regions, allNames[i])
							break
						}
					}
				}
			}
		}
		sort.Strings(regions)
		return lbRegionsLoadedMsg{regions: regions}
	}
}

// fetchLBFlavors loads available load balancer flavors (sizes) for the selected region.
func (m Model) fetchLBFlavors() tea.Cmd {
	region := m.wizard.lbRegion
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/flavor",
			m.cloudProject, url.PathEscape(region))
		var flavors []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &flavors); err != nil {
			return lbFlavorsLoadedMsg{err: err}
		}
		return lbFlavorsLoadedMsg{flavors: flavors}
	}
}

// fetchLBNetworks loads private networks available in the selected region.
func (m Model) fetchLBNetworks() tea.Cmd {
	region := m.wizard.lbRegion
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network",
			m.cloudProject, url.PathEscape(region))
		var nets []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &nets); err != nil {
			return lbNetworksLoadedMsg{err: err}
		}
		var filtered []map[string]interface{}
		for _, n := range nets {
			if getStringValue(n, "visibility", "") == "private" {
				filtered = append(filtered, n)
			}
		}
		return lbNetworksLoadedMsg{networks: filtered}
	}
}

// fetchLBSubnet fetches the first subnet of the chosen private network.
func (m Model) fetchLBSubnet(networkID string) tea.Cmd {
	region := m.wizard.lbRegion
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
			m.cloudProject, url.PathEscape(region), url.PathEscape(networkID))
		var subnets []map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &subnets); err != nil || len(subnets) == 0 {
			return lbSubnetLoadedMsg{subnetID: ""}
		}
		return lbSubnetLoadedMsg{subnetID: getStringValue(subnets[0], "id", "")}
	}
}

// createLBFromWizard calls the OVHcloud API to create the load balancer.
// The API requires the network field with nested private.network.{id, subnetId}.
func (m Model) createLBFromWizard() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return lbCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		if m.wizard.lbNetworkId == "" || m.wizard.lbSubnetId == "" {
			return lbCreatedMsg{err: fmt.Errorf("a private network with subnet is required to create a load balancer")}
		}
		body := map[string]interface{}{
			"name":     m.wizard.lbName,
			"flavorId": m.wizard.lbFlavorId,
			"network": map[string]interface{}{
				"private": map[string]interface{}{
					"network": map[string]interface{}{
						"id":       m.wizard.lbNetworkId,
						"subnetId": m.wizard.lbSubnetId,
					},
				},
			},
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/loadbalancing/loadbalancer",
			m.cloudProject, url.PathEscape(m.wizard.lbRegion))
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return lbCreatedMsg{err: fmt.Errorf("load balancer creation failed: %w", err)}
		}
		return lbCreatedMsg{lb: result}
	}
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderLBWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Nom du Load Balancer :") + "\n\n")
	content.WriteString(descStyle.Render("Choisissez un nom unique pour identifier votre load balancer.") + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(40)
	content.WriteString(inputStyle.Render(m.wizard.lbNameInput+"▌") + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Continuer • Esc : Annuler"))
	return content.String()
}

func (m Model) renderLBWizardRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Choose a region:") + "\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf("Nom : %s", m.wizard.lbName)) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading regions..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.lbAvailableRegions) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No region compatible with Load Balancer (Octavia) found.") + "\n")
	} else {
		maxVisible := 14
		startIdx := 0
		if m.wizard.lbRegionIdx >= maxVisible {
			startIdx = m.wizard.lbRegionIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.wizard.lbAvailableRegions) {
			endIdx = len(m.wizard.lbAvailableRegions)
		}
		for i := startIdx; i < endIdx; i++ {
			r := m.wizard.lbAvailableRegions[i]
			if i == m.wizard.lbRegionIdx {
				content.WriteString(selectedStyle.Render("▶ "+r) + "\n")
			} else {
				content.WriteString(dimStyle.Render("  "+r) + "\n")
			}
		}
		if len(m.wizard.lbAvailableRegions) > maxVisible {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
				Render(fmt.Sprintf("\n  %d / %d regions", m.wizard.lbRegionIdx+1, len(m.wizard.lbAvailableRegions))))
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderLBWizardFlavorStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Choisir la taille du Load Balancer :") + "\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf("Name: %s  •  Region: %s",
		m.wizard.lbName, m.wizard.lbRegion)) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Chargement des tailles disponibles..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.lbFlavors) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No size available in this region.") + "\n")
	} else {
		for i, f := range m.wizard.lbFlavors {
			name := getStringValue(f, "name", getStringValue(f, "id", "unknown"))
			if i == m.wizard.lbFlavorIdx {
				content.WriteString(selectedStyle.Render("▶ "+strings.ToUpper(name)) + "\n")
			} else {
				content.WriteString(dimStyle.Render("  "+strings.ToUpper(name)) + "\n")
			}
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderLBWizardNetworkStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

	flavorDisplay := m.wizard.lbFlavorName
	if flavorDisplay == "" {
		flavorDisplay = m.wizard.lbFlavorId
	}

	content.WriteString(titleStyle.Render("Select a private network:") + "\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf(
		"Name: %s  •  Region: %s  •  Size: %s",
		m.wizard.lbName, m.wizard.lbRegion, strings.ToUpper(flavorDisplay),
	)) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading networks..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.lbNetworks) == 0 {
		content.WriteString(warnStyle.Render("⚠️  No private network available in region "+m.wizard.lbRegion+".") + "\n")
		content.WriteString(descStyle.Render("Create a private network in this region first, then restart the wizard.") + "\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("← : Retour • Esc : Annuler"))
		return content.String()
	}

	for i, n := range m.wizard.lbNetworks {
		name := getStringValue(n, "name", getStringValue(n, "id", "unknown"))
		if i == m.wizard.lbNetworkIdx {
			content.WriteString(selectedStyle.Render("▶ "+name) + "\n")
		} else {
			content.WriteString(dimStyle.Render("  "+name) + "\n")
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderLBWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	lbLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("Confirm Load Balancer creation:") + "\n\n")
	content.WriteString(lbLabelStyle.Render("  Nom :") + valStyle.Render(m.wizard.lbName) + "\n")
	content.WriteString(lbLabelStyle.Render("  Region:") + valStyle.Render(m.wizard.lbRegion) + "\n")

	flavorDisplay := m.wizard.lbFlavorName
	if flavorDisplay == "" {
		flavorDisplay = m.wizard.lbFlavorId
	}
	content.WriteString(lbLabelStyle.Render("  Taille :") + valStyle.Render(strings.ToUpper(flavorDisplay)) + "\n")

	if m.wizard.lbNetworkName != "" {
		content.WriteString(lbLabelStyle.Render("  Private network:") + valStyle.Render(m.wizard.lbNetworkName) + "\n")
	}
	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Creating..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.lbConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleLBWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.lbNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
			return m, nil
		}
		m.wizard.lbName = name
		m.wizard.errorMsg = ""
		m.wizard.step = LBWizardStepRegion
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading regions..."
		return m, m.fetchLBRegions()
	case "backspace":
		if len(m.wizard.lbNameInput) > 0 {
			m.wizard.lbNameInput = m.wizard.lbNameInput[:len(m.wizard.lbNameInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.wizard.lbNameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleLBWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbRegionIdx > 0 {
			m.wizard.lbRegionIdx--
		}
	case "down", "j":
		if m.wizard.lbRegionIdx < len(m.wizard.lbAvailableRegions)-1 {
			m.wizard.lbRegionIdx++
		}
	case "enter":
		if len(m.wizard.lbAvailableRegions) > 0 {
			m.wizard.lbRegion = m.wizard.lbAvailableRegions[m.wizard.lbRegionIdx]
			m.wizard.lbFlavorIdx = 0
			m.wizard.step = LBWizardStepFlavor
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Chargement des tailles..."
			return m, m.fetchLBFlavors()
		}
	case "left":
		m.wizard.step = LBWizardStepName
	}
	return m, nil
}

func (m Model) handleLBWizardFlavorKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.lbFlavorIdx > 0 {
			m.wizard.lbFlavorIdx--
		}
	case "down", "j":
		if m.wizard.lbFlavorIdx < len(m.wizard.lbFlavors)-1 {
			m.wizard.lbFlavorIdx++
		}
	case "enter":
		if len(m.wizard.lbFlavors) > 0 {
			f := m.wizard.lbFlavors[m.wizard.lbFlavorIdx]
			m.wizard.lbFlavorId = getStringValue(f, "id", "")
			m.wizard.lbFlavorName = getStringValue(f, "name", "")
			m.wizard.lbNetworkIdx = 0
			m.wizard.step = LBWizardStepNetwork
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading networks..."
			return m, m.fetchLBNetworks()
		}
	case "left":
		m.wizard.step = LBWizardStepRegion
	}
	return m, nil
}

func (m Model) handleLBWizardNetworkKeys(key string) (tea.Model, tea.Cmd) {
	total := len(m.wizard.lbNetworks)
	switch key {
	case "up", "k":
		if m.wizard.lbNetworkIdx > 0 {
			m.wizard.lbNetworkIdx--
		}
	case "down", "j":
		if m.wizard.lbNetworkIdx < total-1 {
			m.wizard.lbNetworkIdx++
		}
	case "enter":
		if total == 0 {
			// No networks available — can't proceed
			return m, nil
		}
		net := m.wizard.lbNetworks[m.wizard.lbNetworkIdx]
		m.wizard.lbNetworkId = getStringValue(net, "id", "")
		m.wizard.lbNetworkName = getStringValue(net, "name", getStringValue(net, "id", "unknown"))
		// Try to find subnet ID embedded in network data
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
			m.wizard.lbSubnetId = subnetID
			m.wizard.step = LBWizardStepConfirm
		} else {
			// Need to fetch subnet separately
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Checking subnet..."
			return m, m.fetchLBSubnet(m.wizard.lbNetworkId)
		}
	case "left":
		m.wizard.step = LBWizardStepFlavor
	}
	return m, nil
}

func (m Model) handleLBWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.lbConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.lbConfirmBtnIdx = 1
	case "enter":
		if m.wizard.lbConfirmBtnIdx == 1 {
			// Cancel → back to network selection
			m.wizard.step = LBWizardStepNetwork
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating Load Balancer..."
		return m, m.createLBFromWizard()
	}
	return m, nil
}
