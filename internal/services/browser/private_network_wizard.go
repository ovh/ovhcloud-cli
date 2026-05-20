// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// ─── Private Network wizard ───────────────────────────────────────────────────

// fetchPrivateNetRegionsCmd returns a tea.Cmd that loads network-capable regions.
func (m Model) fetchPrivateNetRegionsCmd() tea.Cmd {
	return func() tea.Msg {
		regions, err := m.fetchPrivateNetRegions()
		return privNetRegionsLoadedMsg{regions: regions, err: err}
	}
}

// createPrivateNetworkFromWizard sends the POST request to create the network (+optional subnet).
func (m Model) createPrivateNetworkFromWizard() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return privNetCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Build region list from selected region
		region := m.wizard.selectedRegion

		var network map[string]interface{}

		if m.wizard.privNetIsLocalZone {
			// Local zones use the regional network API (isolated, not vRack-based)
			body := map[string]interface{}{
				"name": m.wizard.privNetName,
			}
			if m.wizard.privNetVlanID > 0 {
				body["vlanId"] = m.wizard.privNetVlanID
			}
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network",
				m.cloudProject, url.PathEscape(region))
			var op map[string]interface{}
			if err := httpLib.Client.Post(endpoint, body, &op); err != nil {
				return privNetCreatedMsg{err: fmt.Errorf("failed to create network: %w", err)}
			}
			// Regional API may return an operation object with resourceId or a network with id
			netID := getString(op, "resourceId")
			if netID == "" {
				netID = getString(op, "id")
			}
			if netID == "" {
				return privNetCreatedMsg{err: fmt.Errorf("network created but ID missing in response")}
			}
			network = map[string]interface{}{"id": netID, "name": m.wizard.privNetName}

			// Optionally create a subnet using the regional subnet API
			if m.wizard.privNetEnableSubnet && m.wizard.privNetCIDR != "" {
				parts := strings.Split(m.wizard.privNetCIDR, "/")
				ipParts := strings.Split(parts[0], ".")
				var gatewayIP string
				if len(ipParts) == 4 {
					gatewayIP = ipParts[0] + "." + ipParts[1] + "." + ipParts[2] + ".1"
				}
				subnetBody := map[string]interface{}{
					"name":            m.wizard.privNetName + "-subnet",
					"cidr":            m.wizard.privNetCIDR,
					"ipVersion":       4,
					"enableDhcp":      m.wizard.privNetEnableDHCP,
					"enableGatewayIp": gatewayIP != "",
				}
				if gatewayIP != "" {
					subnetBody["gatewayIp"] = gatewayIP
				}
				subnetEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
					m.cloudProject, url.PathEscape(region), url.PathEscape(netID))
				var subnet map[string]interface{}
				// Retry briefly to let the network activate
				var subnetErr error
				for i := 0; i < 10; i++ {
					subnetErr = httpLib.Client.Post(subnetEndpoint, subnetBody, &subnet)
					if subnetErr == nil {
						break
					}
					time.Sleep(2 * time.Second)
				}
				if subnetErr != nil {
					return privNetCreatedMsg{
						network: network,
						err:     fmt.Errorf("network created but subnet failed (CIDR: %s): %w", m.wizard.privNetCIDR, subnetErr),
					}
				}
			}
		} else {
			// Standard vRack regions use the legacy global private network API
			body := map[string]interface{}{
				"name":    m.wizard.privNetName,
				"regions": []string{region},
			}
			if m.wizard.privNetVlanID > 0 {
				body["vlanId"] = m.wizard.privNetVlanID
			}
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private", m.cloudProject)
			if err := httpLib.Client.Post(endpoint, body, &network); err != nil {
				return privNetCreatedMsg{err: fmt.Errorf("failed to create network: %w", err)}
			}

			// Optionally create a subnet
			if m.wizard.privNetEnableSubnet && m.wizard.privNetCIDR != "" {
				netID, _ := network["id"].(string)
				if netID == "" {
					return privNetCreatedMsg{err: fmt.Errorf("network created but ID missing, cannot create subnet")}
				}

				// Poll until the network region becomes ACTIVE (OVH creates async)
				networkEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s",
					m.cloudProject, url.PathEscape(netID))
				const maxAttempts = 15
				regionActive := false
				for i := 0; i < maxAttempts; i++ {
					var netData map[string]interface{}
					if err := httpLib.Client.Get(networkEndpoint, &netData); err == nil {
						if regions, ok := netData["regions"].([]interface{}); ok {
							for _, r := range regions {
								if rMap, ok := r.(map[string]interface{}); ok {
									if rMap["region"] == region {
										if rMap["status"] == "ACTIVE" {
											regionActive = true
										}
									}
								}
							}
						}
					}
					if regionActive {
						break
					}
					time.Sleep(3 * time.Second)
				}
				if !regionActive {
					return privNetCreatedMsg{
						network: network,
						err:     fmt.Errorf("network created but region '%s' did not become active in time — subnet not created. Retry from the OVH interface.", region),
					}
				}

				noGateway := m.wizard.privNetGatewayMode == 1 // mode 1 = will attach OVH Gateway service

				subnetBody := map[string]interface{}{
					"dhcp":      m.wizard.privNetEnableDHCP,
					"network":   m.wizard.privNetCIDR,
					"noGateway": noGateway,
					"region":    region,
					"start":     m.wizard.privNetAllocStart,
					"end":       m.wizard.privNetAllocEnd,
				}
				var subnet map[string]interface{}
				subnetEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet",
					m.cloudProject, url.PathEscape(netID))
				if err := httpLib.Client.Post(subnetEndpoint, subnetBody, &subnet); err != nil {
					return privNetCreatedMsg{
						network: network,
						err:     fmt.Errorf("network created but subnet failed (%s, CIDR: %s): %w", netID, m.wizard.privNetCIDR, err),
					}
				}
			}
		}

		return privNetCreatedMsg{network: network}
	}
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderPrivNetWizardRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Choose private network location:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading regions..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	// Group by type
	type entry struct {
		name  string
		rtype string
	}
	var vrack, localz []entry
	for _, r := range m.wizard.privNetRegions {
		n, _ := r["name"].(string)
		t, _ := r["type"].(string)
		e := entry{name: n, rtype: t}
		if t == "localzone" {
			localz = append(localz, e)
		} else {
			vrack = append(vrack, e)
		}
	}

	allEntries := append(vrack, localz...)
	sectionStart := len(vrack)

	for i, e := range allEntries {
		if i == 0 {
			content.WriteString(typeStyle.Render("  ── Regions (vRack) ──") + "\n")
		} else if i == sectionStart {
			content.WriteString(typeStyle.Render("  ── Local Zones ──") + "\n")
		}
		label := e.name
		if i == m.wizard.privNetRegionIdx {
			content.WriteString(selectedStyle.Render("▶ " + label) + "\n")
		} else {
			content.WriteString(dimStyle.Render("  " + label) + "\n")
		}
	}

	if len(allEntries) == 0 {
		content.WriteString(dimStyle.Render("  No region available.") + "\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0).
		Render("↑↓ Navigate • Enter: Select • Esc: Cancel"))
	return content.String()
}

func (m Model) renderPrivNetWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Private network name:") + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(40)
	content.WriteString(inputStyle.Render(m.wizard.privNetNameInput+"▌") + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tapez le nom • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardVlanStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	content.WriteString(titleStyle.Render("Layer 2 network option – VLAN ID:") + "\n\n")

	// Show already-used VLAN IDs as a hint
	if len(m.wizard.privNetUsedVlanIDs) > 0 {
		var used []string
		for id := range m.wizard.privNetUsedVlanIDs {
			used = append(used, fmt.Sprintf("%d", id))
		}
		sort.Strings(used)
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		content.WriteString(warnStyle.Render("  VLANs already in use: "+strings.Join(used, ", ")) + "\n\n")
	}

	// Option 0 : auto
	if !m.wizard.privNetDefineVlan {
		content.WriteString(selectedStyle.Render("▶ Pas de VLAN (attribution automatique)") + "\n")
	} else {
		content.WriteString(dimStyle.Render("  Pas de VLAN (attribution automatique)") + "\n")
	}

	// Option 1 : define VLAN
	if m.wizard.privNetDefineVlan {
		content.WriteString(selectedStyle.Render("▶ Set a VLAN ID") + "\n\n")
		content.WriteString(descStyle.Render("  VLAN ID (plage valide : 1 – 4094) :") + "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).Width(20)
		val := m.wizard.privNetVlanInput
		if val == "" {
			val = "(vide)"
		}
		content.WriteString(inputStyle.Render(val+"▌") + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  Set a VLAN ID") + "\n\n")
	}

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Naviguer • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardSubnetStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))

	if m.wizard.privNetAddSubnetMode {
		content.WriteString(titleStyle.Render("Subnet CIDR for new region:") + "\n\n")

		// Show existing subnets so user knows what CIDRs are already taken
		if subnets, ok := m.detailData["_subnets"].([]map[string]any); ok && len(subnets) > 0 {
			content.WriteString(warnStyle.Render("⚠  Existing subnets (use a different CIDR):") + "\n")
			for _, sub := range subnets {
				cidr := getStringValue(sub, "cidr", "")
				region := getStringValue(sub, "region", "")
				if cidr != "" {
					content.WriteString(dimStyle.Render(fmt.Sprintf("  • %s  (%s)", cidr, region)) + "\n")
				}
			}
			content.WriteString("\n")
		}
	} else {
		content.WriteString(titleStyle.Render("Configure subnet:") + "\n\n")

		// Toggle: enable/disable subnet
		selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F"))
		enableLabel := "○ Create a subnet"
		if m.wizard.privNetEnableSubnet {
			enableLabel = "● Create a subnet  ✓"
		}
		content.WriteString(selectedStyle.Render(enableLabel) + "\n\n")
		if !m.wizard.privNetEnableSubnet {
			content.WriteString(dimStyle.Render("  No subnet will be created.") + "\n\n")
			if m.wizard.errorMsg != "" {
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
			}
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
				Render("Space: Enable/Disable • Enter: Continue • ←: Back • Esc: Cancel"))
			return content.String()
		}
	}

	// CIDR input
	cidrExample := "10.0.0.0/16"
	if m.wizard.privNetVlanID > 0 {
		cidrExample = fmt.Sprintf("10.%d.0.0/16", m.wizard.privNetVlanID)
	}
	content.WriteString(descStyle.Render("Subnet CIDR (e.g. "+cidrExample+"):") + "\n")
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).Width(30)
	cidr := m.wizard.privNetCIDRInput
	if cidr == "" {
		cidr = "(empty)"
	}
	content.WriteString(inputStyle.Render(cidr+"▌") + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	hint := "Space: Enable/Disable • Enter: Continue • ←: Back • Esc: Cancel"
	if m.wizard.privNetAddSubnetMode {
		hint = "Enter: Continue • ←: Back • Esc: Cancel"
	}
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(hint))
	return content.String()
}

func (m Model) renderPrivNetWizardDHCPStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("DHCP configuration:") + "\n\n")

	enabled := m.wizard.privNetEnableDHCP

	enabledStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 1)

	if enabled {
		content.WriteString(enabledStyle.Render("▶ Enabled  ✓") + "\n")
		content.WriteString(dimStyle.Render("  Disabled") + "\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
			Render("IP addresses will be assigned automatically to instances.") + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  Enabled") + "\n")
		content.WriteString(enabledStyle.Render("▶ Disabled") + "\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).
			Render("IP addresses must be configured manually.") + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓/Space: Toggle • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderPrivNetWizardGatewayStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	content.WriteString(titleStyle.Render("Network gateway options:") + "\n\n")

	// Option 0
	label0 := "Announce the first address of a given CIDR as the default gateway"
	// Option 1
	label1 := "Assign a Gateway and connect to the private network"

	if m.wizard.privNetGatewayMode == 0 {
		content.WriteString(selectedStyle.Render("▶ "+label0) + "\n")
		content.WriteString(dimStyle.Render("  "+label1) + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  "+label0) + "\n")
		content.WriteString(selectedStyle.Render("▶ "+label1) + "\n\n")
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		content.WriteString(descStyle.Render("  Gateway IP address (empty = first IP of CIDR):") + "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).Width(24)
		gw := m.wizard.privNetGatewayInput
		if gw == "" {
			gw = "(auto)"
		}
		content.WriteString(inputStyle.Render(gw+"▌") + "\n\n")
	}

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Naviguer • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(26)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	if m.wizard.privNetAddSubnetMode {
		content.WriteString(titleStyle.Render("Confirm adding subnet to: "+m.wizard.privNetName) + "\n\n")
		content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
	} else {
		content.WriteString(titleStyle.Render("Confirm private network creation:") + "\n\n")
		content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
		content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.privNetName) + "\n")
		vlanStr := "automatic"
		if m.wizard.privNetVlanID > 0 {
			vlanStr = fmt.Sprintf("%d", m.wizard.privNetVlanID)
		}
		content.WriteString(labelStyle.Render("  VLAN ID:") + valueStyle.Render(vlanStr) + "\n")
	}

	if m.wizard.privNetEnableSubnet || m.wizard.privNetAddSubnetMode {
		content.WriteString(labelStyle.Render("  Subnet (CIDR):") + valueStyle.Render(m.wizard.privNetCIDR) + "\n")
		dhcpStr := "disabled"
		if m.wizard.privNetEnableDHCP {
			dhcpStr = "enabled"
		}
		content.WriteString(labelStyle.Render("  DHCP:") + valueStyle.Render(dhcpStr) + "\n")
		if m.wizard.privNetAllocStart != "" || m.wizard.privNetAllocEnd != "" {
			allocStr := m.wizard.privNetAllocStart + " – " + m.wizard.privNetAllocEnd
			content.WriteString(labelStyle.Render("  IP address allocated:") + valueStyle.Render(allocStr) + "\n")
		}
		var gwStr string
		if m.wizard.privNetGatewayMode == 0 {
			gwStr = "First IP of CIDR (auto)"
		} else {
			gwStr = "Assigned IP"
			if m.wizard.privNetGateway != "" {
				gwStr = m.wizard.privNetGateway
			}
		}
		content.WriteString(labelStyle.Render("  Gateway:") + valueStyle.Render(gwStr) + "\n")
	} else {
		content.WriteString(labelStyle.Render("  Subnet:") + valueStyle.Render("none") + "\n")
	}

	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Creating..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	confirmLabel := " Create "
	if m.wizard.privNetAddSubnetMode {
		confirmLabel = " Add Subnet "
	}
	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(confirmLabel)
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.privNetConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(confirmLabel)
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • ←: Back • Esc: Cancel"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handlePrivNetWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	total := len(m.wizard.privNetRegions)
	switch key {
	case "up", "k":
		if m.wizard.privNetRegionIdx > 0 {
			m.wizard.privNetRegionIdx--
		}
	case "down", "j":
		if m.wizard.privNetRegionIdx < total-1 {
			m.wizard.privNetRegionIdx++
		}
	case "enter":
		if total > 0 {
			r := m.wizard.privNetRegions[m.wizard.privNetRegionIdx]
			m.wizard.selectedRegion, _ = r["name"].(string)
			rtype, _ := r["type"].(string)
			m.wizard.privNetIsLocalZone = (rtype == "localzone")
			if m.wizard.privNetAddSubnetMode {
				m.wizard.step = PrivNetWizardStepSubnet
			} else {
				m.wizard.step = PrivNetWizardStepName
			}
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		name := strings.TrimSpace(m.wizard.privNetNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty"
			return m, nil
		}
		m.wizard.privNetName = name
		m.wizard.errorMsg = ""
		m.wizard.step = PrivNetWizardStepVlanID
	case "left":
		m.wizard.step = PrivNetWizardStepRegion
	case "backspace":
		if len(m.wizard.privNetNameInput) > 0 {
			m.wizard.privNetNameInput = m.wizard.privNetNameInput[:len(m.wizard.privNetNameInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.wizard.privNetNameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardVlanKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.wizard.privNetDefineVlan {
			m.wizard.privNetDefineVlan = false
			m.wizard.privNetVlanInput = ""
			m.wizard.privNetVlanID = 0
			m.wizard.errorMsg = ""
		}
	case "down", "j":
		if !m.wizard.privNetDefineVlan {
			m.wizard.privNetDefineVlan = true
			m.wizard.errorMsg = ""
		}
	case "enter":
		if m.wizard.privNetDefineVlan {
			input := strings.TrimSpace(m.wizard.privNetVlanInput)
			if input == "" {
				m.wizard.errorMsg = "Entrez un VLAN ID (1–4094)"
				return m, nil
			}
			var v int
			if _, err := fmt.Sscanf(input, "%d", &v); err != nil || v < 1 || v > 4094 {
				m.wizard.errorMsg = "VLAN ID invalide (1–4094)"
				return m, nil
			}
			if m.wizard.privNetUsedVlanIDs[v] {
				m.wizard.errorMsg = fmt.Sprintf("VLAN ID %d is already used by an existing network", v)
				return m, nil
			}
			m.wizard.privNetVlanID = v
		} else {
			m.wizard.privNetVlanID = 0 // auto
		}
		m.wizard.errorMsg = ""
		// Pre-fill CIDR input with a dynamic example based on VLAN ID
		if m.wizard.privNetVlanID > 0 {
			m.wizard.privNetCIDRInput = fmt.Sprintf("10.%d.0.0/16", m.wizard.privNetVlanID)
		} else {
			m.wizard.privNetCIDRInput = "10.0.0.0/16"
		}
		m.wizard.step = PrivNetWizardStepSubnet
	case "left":
		m.wizard.step = PrivNetWizardStepName
	case "backspace":
		if m.wizard.privNetDefineVlan && len(m.wizard.privNetVlanInput) > 0 {
			m.wizard.privNetVlanInput = m.wizard.privNetVlanInput[:len(m.wizard.privNetVlanInput)-1]
		}
	default:
		if m.wizard.privNetDefineVlan && len(msg.Runes) > 0 {
			r := msg.Runes[0]
			if r >= '0' && r <= '9' {
				m.wizard.privNetVlanInput += string(r)
			}
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardSubnetKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case " ":
		m.wizard.privNetEnableSubnet = !m.wizard.privNetEnableSubnet
	case "enter":
		if m.wizard.privNetEnableSubnet {
			cidr := strings.TrimSpace(m.wizard.privNetCIDRInput)
			if cidr == "" {
				m.wizard.errorMsg = "CIDR cannot be empty"
				return m, nil
			}
			m.wizard.privNetCIDR = cidr
		}
		m.wizard.errorMsg = ""
		m.wizard.step = PrivNetWizardStepDHCP
	case "left":
		if m.wizard.privNetAddSubnetMode {
			m.wizard.step = PrivNetWizardStepRegion
		} else {
			m.wizard.step = PrivNetWizardStepVlanID
		}
	case "backspace":
		if m.wizard.privNetEnableSubnet && len(m.wizard.privNetCIDRInput) > 0 {
			m.wizard.privNetCIDRInput = m.wizard.privNetCIDRInput[:len(m.wizard.privNetCIDRInput)-1]
		}
	default:
		if m.wizard.privNetEnableSubnet && len(msg.Runes) > 0 {
			m.wizard.privNetCIDRInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardDHCPKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case " ", "up", "down", "k", "j":
		if m.wizard.privNetEnableSubnet {
			m.wizard.privNetEnableDHCP = !m.wizard.privNetEnableDHCP
		}
	case "enter":
		// Pre-fill allocation pool from CIDR if not yet set
		if m.wizard.privNetEnableSubnet && m.wizard.privNetCIDR != "" && m.wizard.privNetAllocStart == "" {
			start, end, err := cidrToFirstLast(m.wizard.privNetCIDR, true)
			if err == nil {
				m.wizard.privNetAllocStart = start
				m.wizard.privNetAllocEnd = end
			}
		}
		m.wizard.privNetAllocField = 0
		m.wizard.step = PrivNetWizardStepAllocPool
	case "left":
		m.wizard.step = PrivNetWizardStepSubnet
	}
	return m, nil
}

func (m Model) renderPrivNetWizardAllocPoolStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16)
	activeStyle := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Padding(0, 1)

	content.WriteString(titleStyle.Render("IP address allocation pool:") + "\n\n")
	if m.wizard.privNetCIDR != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("CIDR: "+m.wizard.privNetCIDR) + "\n\n")
	}

	startStr := m.wizard.privNetAllocStart
	endStr := m.wizard.privNetAllocEnd
	if m.wizard.privNetAllocField == 0 {
		content.WriteString(labelStyle.Render("Start IP:") + activeStyle.Render(startStr+"▌") + "\n")
		content.WriteString(labelStyle.Render("End IP:") + dimStyle.Render(endStr) + "\n")
	} else {
		content.WriteString(labelStyle.Render("Start IP:") + dimStyle.Render(startStr) + "\n")
		content.WriteString(labelStyle.Render("End IP:") + activeStyle.Render(endStr+"▌") + "\n")
	}

	if m.wizard.errorMsg != "" {
		content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n")
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Tab/↑↓: Switch field • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) handlePrivNetWizardAllocPoolKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "tab", "down", "j":
		m.wizard.privNetAllocField = 1 - m.wizard.privNetAllocField
	case "up", "k":
		m.wizard.privNetAllocField = 1 - m.wizard.privNetAllocField
	case "enter":
		// Basic validation
		if net.ParseIP(m.wizard.privNetAllocStart) == nil {
			m.wizard.errorMsg = "Invalid start IP: " + m.wizard.privNetAllocStart
			return m, nil
		}
		if net.ParseIP(m.wizard.privNetAllocEnd) == nil {
			m.wizard.errorMsg = "Invalid end IP: " + m.wizard.privNetAllocEnd
			return m, nil
		}
		m.wizard.errorMsg = ""
		if m.wizard.privNetIsLocalZone {
			m.wizard.step = PrivNetWizardStepConfirm
		} else {
			m.wizard.step = PrivNetWizardStepGateway
		}
	case "left":
		m.wizard.errorMsg = ""
		m.wizard.step = PrivNetWizardStepDHCP
	case "backspace":
		if m.wizard.privNetAllocField == 0 && len(m.wizard.privNetAllocStart) > 0 {
			m.wizard.privNetAllocStart = m.wizard.privNetAllocStart[:len(m.wizard.privNetAllocStart)-1]
		} else if m.wizard.privNetAllocField == 1 && len(m.wizard.privNetAllocEnd) > 0 {
			m.wizard.privNetAllocEnd = m.wizard.privNetAllocEnd[:len(m.wizard.privNetAllocEnd)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			ch := string(msg.Runes)
			if m.wizard.privNetAllocField == 0 {
				m.wizard.privNetAllocStart += ch
			} else {
				m.wizard.privNetAllocEnd += ch
			}
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardGatewayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "up", "k":
		if m.wizard.privNetGatewayMode > 0 {
			m.wizard.privNetGatewayMode--
			m.wizard.privNetGatewayInput = ""
			m.wizard.privNetGateway = ""
			m.wizard.errorMsg = ""
		}
	case "down", "j":
		if m.wizard.privNetGatewayMode < 1 {
			m.wizard.privNetGatewayMode++
			m.wizard.errorMsg = ""
		}
	case "enter":
		if m.wizard.privNetGatewayMode == 1 {
			m.wizard.privNetGateway = strings.TrimSpace(m.wizard.privNetGatewayInput)
		}
		m.wizard.errorMsg = ""
		m.wizard.step = PrivNetWizardStepConfirm
	case "left":
		m.wizard.step = PrivNetWizardStepAllocPool
		if m.wizard.privNetGatewayMode == 1 && len(m.wizard.privNetGatewayInput) > 0 {
			m.wizard.privNetGatewayInput = m.wizard.privNetGatewayInput[:len(m.wizard.privNetGatewayInput)-1]
		}
	default:
		if m.wizard.privNetGatewayMode == 1 && len(msg.Runes) > 0 {
			m.wizard.privNetGatewayInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handlePrivNetWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.privNetConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.privNetConfirmBtnIdx = 1
	case "enter":
		if m.wizard.privNetConfirmBtnIdx == 1 {
			// Cancel button → go back to previous step
			if m.wizard.privNetIsLocalZone {
				m.wizard.step = PrivNetWizardStepAllocPool
			} else {
				m.wizard.step = PrivNetWizardStepGateway
			}
			return m, nil
		}
		m.wizard.isLoading = true
		if m.wizard.privNetAddSubnetMode {
			m.wizard.loadingMessage = "Adding subnet..."
			return m, m.createSubnetForNetwork()
		}
		m.wizard.loadingMessage = "Creating private network..."
		return m, m.createPrivateNetworkFromWizard()
	}
	return m, nil
}

// cidrToFirstLast derives the first usable IP and last usable IP (broadcast-1)
// from an IPv4 CIDR block. If reserveGateway is true, the first IP (network+1)
// is reserved for the gateway and the pool starts at network+2.
func cidrToFirstLast(cidr string, reserveGateway bool) (first, last string, err error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	ip := ipNet.IP.To4()
	if ip == nil {
		return "", "", fmt.Errorf("only IPv4 CIDR is supported")
	}
	mask := []byte(ipNet.Mask)
	if len(mask) == 16 {
		mask = mask[12:]
	}
	broadcast := net.IP{ip[0] | ^mask[0], ip[1] | ^mask[1], ip[2] | ^mask[2], ip[3] | ^mask[3]}
	offset := byte(1)
	if reserveGateway {
		offset = 2 // skip network+1 which is reserved for the gateway
	}
	firstIP := net.IP{ip[0], ip[1], ip[2], ip[3] + offset}
	lastIP := net.IP{broadcast[0], broadcast[1], broadcast[2], broadcast[3] - 1}
	return firstIP.String(), lastIP.String(), nil
}

func (m Model) createSubnetForNetwork() tea.Cmd {
	return func() tea.Msg {
		netID := m.wizard.privNetTargetNetworkID
		region := m.wizard.selectedRegion

		// Check if region is already activated on the network; if not, activate it first
		networkEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s", m.cloudProject, url.PathEscape(netID))
		var netData map[string]interface{}
		regionActive := false
		openstackID := ""
		if err := httpLib.Client.Get(networkEndpoint, &netData); err == nil {
			if regions, ok := netData["regions"].([]interface{}); ok {
				for _, rv := range regions {
					if rm, ok := rv.(map[string]interface{}); ok {
						if rm["region"] == region {
							regionActive = true
							openstackID, _ = rm["openstackId"].(string)
						}
					}
				}
			}
		}
		if !regionActive {
			// Activate the region on the network first
			activateEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/region", m.cloudProject, url.PathEscape(netID))
			var op map[string]interface{}
			if err := httpLib.Client.Post(activateEndpoint, map[string]interface{}{"region": region}, &op); err != nil {
				return subnetAddedMsg{networkID: netID, err: fmt.Errorf("failed to activate region %s on network: %w", region, err)}
			}
			// Poll until ACTIVE and capture the openstackId
			for i := 0; i < 20; i++ {
				time.Sleep(3 * time.Second)
				var nd map[string]interface{}
				if err := httpLib.Client.Get(networkEndpoint, &nd); err == nil {
					if regs, ok := nd["regions"].([]interface{}); ok {
						for _, rv := range regs {
							if rm, ok := rv.(map[string]interface{}); ok {
								if rm["region"] == region && rm["status"] == "ACTIVE" {
									regionActive = true
									openstackID, _ = rm["openstackId"].(string)
								}
							}
						}
					}
				}
				if regionActive {
					break
				}
			}
			if !regionActive {
				return subnetAddedMsg{networkID: netID, err: fmt.Errorf("region %s did not become ACTIVE in time", region)}
			}
		}

		if openstackID == "" {
			return subnetAddedMsg{networkID: netID, err: fmt.Errorf("could not find OpenStack network ID for region %s", region)}
		}

		// Detect IP version from CIDR
		ipVersion := 4
		if strings.Contains(m.wizard.privNetCIDR, ":") {
			ipVersion = 6
		}

		enableGateway := m.wizard.privNetGatewayMode != 1 // mode 1 = noGateway

		subnetBody := map[string]interface{}{
			"cidr":            m.wizard.privNetCIDR,
			"enableDhcp":      m.wizard.privNetEnableDHCP,
			"enableGatewayIp": enableGateway,
			"ipVersion":       ipVersion,
			"name":            fmt.Sprintf("%s-%s", m.wizard.privNetName, region),
			"allocationPools": []map[string]interface{}{
				{"start": m.wizard.privNetAllocStart, "end": m.wizard.privNetAllocEnd},
			},
		}

		var subnet map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
			m.cloudProject, url.PathEscape(region), url.PathEscape(openstackID))
		if err := httpLib.Client.Post(endpoint, subnetBody, &subnet); err != nil {
			return subnetAddedMsg{networkID: netID, err: err}
		}
		return subnetAddedMsg{networkID: netID}
	}
}
