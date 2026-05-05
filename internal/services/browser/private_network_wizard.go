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
				return privNetCreatedMsg{err: fmt.Errorf("réseau créé mais ID manquant dans la réponse")}
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
						err:     fmt.Errorf("réseau créé mais échec du sous-réseau (CIDR: %s): %w", m.wizard.privNetCIDR, subnetErr),
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
					return privNetCreatedMsg{err: fmt.Errorf("réseau créé mais ID manquant, impossible de créer le sous-réseau")}
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
						err:     fmt.Errorf("réseau créé mais la région '%s' n'est pas devenue active à temps — sous-réseau non créé. Réessayez depuis l'interface OVH.", region),
					}
				}

				noGateway := m.wizard.privNetGatewayMode == 1 // mode 1 = will attach OVH Gateway service

				// Always reserve network+1 for the gateway IP (whether static or OVH Gateway).
				// Not reserving it causes a 409 conflict when the Gateway service tries to claim that IP.
				startIP, endIP, cidrErr := cidrToFirstLast(m.wizard.privNetCIDR, true)
				if cidrErr != nil {
					return privNetCreatedMsg{
						network: network,
						err:     fmt.Errorf("réseau créé mais CIDR invalide ('%s'): %w", m.wizard.privNetCIDR, cidrErr),
					}
				}

				subnetBody := map[string]interface{}{
					"dhcp":      m.wizard.privNetEnableDHCP,
					"network":   m.wizard.privNetCIDR,
					"noGateway": noGateway,
					"region":    region,
					"start":     startIP,
					"end":       endIP,
				}
				var subnet map[string]interface{}
				subnetEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet",
					m.cloudProject, url.PathEscape(netID))
				if err := httpLib.Client.Post(subnetEndpoint, subnetBody, &subnet); err != nil {
					return privNetCreatedMsg{
						network: network,
						err:     fmt.Errorf("réseau créé mais échec du sous-réseau (%s, CIDR: %s): %w", netID, m.wizard.privNetCIDR, err),
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
	content.WriteString(titleStyle.Render("Choisir la localisation du réseau privé :") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Chargement des régions..."))
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
			content.WriteString(typeStyle.Render("  ── Régions (vRack) ──") + "\n")
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
		content.WriteString(dimStyle.Render("  Aucune région disponible.") + "\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0).
		Render("↑↓ Naviguer • Enter : Sélectionner • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Nom du réseau privé :") + "\n\n")

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

	content.WriteString(titleStyle.Render("Option réseau layer 2 – VLAN ID :") + "\n\n")

	// Show already-used VLAN IDs as a hint
	if len(m.wizard.privNetUsedVlanIDs) > 0 {
		var used []string
		for id := range m.wizard.privNetUsedVlanIDs {
			used = append(used, fmt.Sprintf("%d", id))
		}
		sort.Strings(used)
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		content.WriteString(warnStyle.Render("  VLAN déjà utilisés : "+strings.Join(used, ", ")) + "\n\n")
	}

	// Option 0 : auto
	if !m.wizard.privNetDefineVlan {
		content.WriteString(selectedStyle.Render("▶ Pas de VLAN (attribution automatique)") + "\n")
	} else {
		content.WriteString(dimStyle.Render("  Pas de VLAN (attribution automatique)") + "\n")
	}

	// Option 1 : define VLAN
	if m.wizard.privNetDefineVlan {
		content.WriteString(selectedStyle.Render("▶ Définir un VLAN ID") + "\n\n")
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
		content.WriteString(dimStyle.Render("  Définir un VLAN ID") + "\n\n")
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

	content.WriteString(titleStyle.Render("Configurer le sous-réseau :") + "\n\n")

	// Toggle: enable/disable subnet
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	enableLabel := "○ Créer un sous-réseau"
	if m.wizard.privNetEnableSubnet {
		enableLabel = "● Créer un sous-réseau  ✓"
	}
	content.WriteString(selectedStyle.Render(enableLabel) + "\n\n")

	if m.wizard.privNetEnableSubnet {
		// Build example CIDR: 10.{vlanId}.0.0/16, fallback to 10.0.0.0/16
		cidrExample := "10.0.0.0/16"
		if m.wizard.privNetVlanID > 0 {
			cidrExample = fmt.Sprintf("10.%d.0.0/16", m.wizard.privNetVlanID)
		}
		content.WriteString(descStyle.Render("CIDR du sous-réseau (ex : "+cidrExample+") :") + "\n")
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).Width(30)
		cidr := m.wizard.privNetCIDRInput
		if cidr == "" {
			cidr = "(vide)"
		}
		content.WriteString(inputStyle.Render(cidr+"▌") + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  Aucun sous-réseau ne sera créé.") + "\n\n")
	}

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Space : Activer/Désactiver • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardDHCPStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Options de distribution des adresses DHCP :") + "\n\n")

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F"))

	dhcpLabel := "○ DHCP désactivé"
	if m.wizard.privNetEnableDHCP {
		dhcpLabel = "● DHCP activé  ✓"
	}
	content.WriteString(selectedStyle.Render(dhcpLabel) + "\n\n")

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	if m.wizard.privNetEnableDHCP {
		content.WriteString(descStyle.Render("Le DHCP distribuera automatiquement des adresses IP aux instances.") + "\n\n")
	} else {
		content.WriteString(descStyle.Render("Les adresses IP devront être configurées manuellement.") + "\n\n")
	}

	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("Space/←→ : Basculer • Enter : Continuer • ← : Retour • Esc : Annuler"))
	return content.String()
}

func (m Model) renderPrivNetWizardGatewayStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	content.WriteString(titleStyle.Render("Options de passerelle réseau :") + "\n\n")

	// Option 0
	label0 := "Annoncer la première adresse d'un CIDR donné comme passerelle par défaut"
	// Option 1
	label1 := "Assigner une Gateway et connectez-vous au réseau privé"

	if m.wizard.privNetGatewayMode == 0 {
		content.WriteString(selectedStyle.Render("▶ "+label0) + "\n")
		content.WriteString(dimStyle.Render("  "+label1) + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  "+label0) + "\n")
		content.WriteString(selectedStyle.Render("▶ "+label1) + "\n\n")
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		content.WriteString(descStyle.Render("  Adresse IP de la passerelle (vide = première IP du CIDR) :") + "\n")
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
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("Confirmer la création du réseau privé :") + "\n\n")
	content.WriteString(labelStyle.Render("  Région :") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
	content.WriteString(labelStyle.Render("  Nom :") + valueStyle.Render(m.wizard.privNetName) + "\n")

	vlanStr := "automatique"
	if m.wizard.privNetVlanID > 0 {
		vlanStr = fmt.Sprintf("%d", m.wizard.privNetVlanID)
	}
	content.WriteString(labelStyle.Render("  VLAN ID :") + valueStyle.Render(vlanStr) + "\n")

	if m.wizard.privNetEnableSubnet {
		content.WriteString(labelStyle.Render("  Sous-réseau (CIDR) :") + valueStyle.Render(m.wizard.privNetCIDR) + "\n")
		dhcpStr := "désactivé"
		if m.wizard.privNetEnableDHCP {
			dhcpStr = "activé"
		}
		content.WriteString(labelStyle.Render("  DHCP :") + valueStyle.Render(dhcpStr) + "\n")
		var gwStr string
		if m.wizard.privNetGatewayMode == 0 {
			gwStr = "Première IP du CIDR (auto)"
		} else {
			gwStr = "IP assignée"
			if m.wizard.privNetGateway != "" {
				gwStr = m.wizard.privNetGateway
			}
		}
		content.WriteString(labelStyle.Render("  Passerelle :") + valueStyle.Render(gwStr) + "\n")
	} else {
		content.WriteString(labelStyle.Render("  Sous-réseau :") + valueStyle.Render("aucun") + "\n")
	}

	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Création en cours..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Erreur : "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Créer ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Annuler ")
	if m.wizard.privNetConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Créer ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Annuler ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→ : Sélectionner • Enter : Confirmer • ← : Retour • Esc : Annuler"))
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
			m.wizard.step = PrivNetWizardStepName
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
			m.wizard.errorMsg = "Le nom ne peut pas être vide"
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
				m.wizard.errorMsg = fmt.Sprintf("VLAN ID %d est déjà utilisé par un réseau existant", v)
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
				m.wizard.errorMsg = "Le CIDR ne peut pas être vide"
				return m, nil
			}
			m.wizard.privNetCIDR = cidr
		}
		m.wizard.errorMsg = ""
		m.wizard.step = PrivNetWizardStepDHCP
	case "left":
		m.wizard.step = PrivNetWizardStepVlanID
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
	case " ", "h", "l":
		if m.wizard.privNetEnableSubnet {
			m.wizard.privNetEnableDHCP = !m.wizard.privNetEnableDHCP
		}
	case "enter":
		if m.wizard.privNetIsLocalZone {
			m.wizard.step = PrivNetWizardStepConfirm
		} else {
			m.wizard.step = PrivNetWizardStepGateway
		}
	case "left":
		m.wizard.step = PrivNetWizardStepSubnet
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
		m.wizard.step = PrivNetWizardStepDHCP
	case "backspace":
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
				m.wizard.step = PrivNetWizardStepDHCP
			} else {
				m.wizard.step = PrivNetWizardStepGateway
			}
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Création du réseau privé..."
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
