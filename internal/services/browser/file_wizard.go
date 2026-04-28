// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── File Storage wizard render functions ─────────────────────────────────────

func (m Model) renderFileWizardNameStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter a name for the file share:") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(40)
	content.WriteString(inputStyle.Render(m.wizard.fileShareNameInput+"▌") + "\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("Type to enter • Enter: Continue • Esc: Cancel"))
	return content.String()
}

func (m Model) renderFileWizardRegionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select region for the file share:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading available regions..."))
		return content.String()
	}

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, region := range m.wizard.fileShareRegions {
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + region))
		} else {
			content.WriteString(listStyle.Render("  " + region))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderFileWizardTypeStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select performance mode:") + "\n\n")

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	content.WriteString(descStyle.Render("Choose the performance tier for the NFS share.") + "\n\n")

	types := []string{"standard-1az"}
	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, t := range types {
		if i == m.wizard.fileShareTypeIdx {
			content.WriteString(selectedStyle.Render("▶ " + t))
		} else {
			content.WriteString(listStyle.Render("  " + t))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderFileWizardSizeStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter file share size (GB):") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(20)
	content.WriteString(inputStyle.Render(m.wizard.fileShareSizeInput+"▌") + "\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("Type size in GB • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderFileWizardNetworkStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	if m.wizard.isLoading {
		if m.wizard.fileShareNetworkMenuIdx == 0 {
			content.WriteString(titleStyle.Render("Select private network:") + "\n\n")
			content.WriteString(loadingStyle.Render("Loading networks..."))
		} else {
			content.WriteString(titleStyle.Render("Select subnet:") + "\n\n")
			content.WriteString(loadingStyle.Render("Loading subnets..."))
		}
		return content.String()
	}

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	if m.wizard.fileShareNetworkMenuIdx == 0 {
		content.WriteString(titleStyle.Render("Select private network for the file share:") + "\n\n")
		for i, network := range m.wizard.fileShareNetworks {
			name, _ := network["name"].(string)
			if i == m.wizard.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ " + name))
			} else {
				content.WriteString(listStyle.Render("  " + name))
			}
			content.WriteString("\n")
		}
		if len(m.wizard.fileShareNetworks) == 0 {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
			content.WriteString(dimStyle.Render("  No private networks available in this region.") + "\n")
		}
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
		content.WriteString(helpStyle.Render("↑↓ Navigate • Enter: Select • ← Back • Esc: Cancel"))
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("Select subnet (network: %s):", m.wizard.fileShareNetworkName)) + "\n\n")
		for i, subnet := range m.wizard.fileShareSubnets {
			cidr, _ := subnet["cidr"].(string)
			if i == m.wizard.selectedIndex {
				content.WriteString(selectedStyle.Render("▶ " + cidr))
			} else {
				content.WriteString(listStyle.Render("  " + cidr))
			}
			content.WriteString("\n")
		}
		if len(m.wizard.fileShareSubnets) == 0 {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
			content.WriteString(dimStyle.Render("  No subnets available for this network.") + "\n")
		}
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
		content.WriteString(helpStyle.Render("↑↓ Navigate • Enter: Select • ← Back to networks • Esc: Cancel"))
	}
	return content.String()
}

func (m Model) renderFileWizardConfirmStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	regionStr := m.wizard.selectedRegion

	content.WriteString(titleStyle.Render("Confirm file share creation:") + "\n\n")
	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.fileShareName) + "\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(regionStr) + "\n")
	content.WriteString(labelStyle.Render("  Type:") + valueStyle.Render(m.wizard.fileShareType) + "\n")
	content.WriteString(labelStyle.Render("  Size:") + valueStyle.Render(fmt.Sprintf("%d GB", m.wizard.fileShareSize)) + "\n")
	content.WriteString(labelStyle.Render("  Network:") + valueStyle.Render(m.wizard.fileShareNetworkName) + "\n")
	content.WriteString(labelStyle.Render("  Subnet:") + valueStyle.Render(m.wizard.fileShareSubnetCIDR) + "\n")
	content.WriteString("\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render(m.wizard.loadingMessage))
		return content.String()
	}

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	cancelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	if m.wizard.fileShareConfirmBtnIdx == 0 {
		content.WriteString(createStyle.Render("  ▶ [Create File Share]") + "    ")
		content.WriteString(dimStyle.Render("[Cancel]") + "\n")
	} else {
		content.WriteString(dimStyle.Render("    [Create File Share]") + "    ")
		content.WriteString(cancelStyle.Render("▶ [Cancel]") + "\n")
	}

	return content.String()
}

// ─── File Storage wizard key handler functions ────────────────────────────────

func (m Model) handleFileWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := strings.TrimSpace(m.wizard.fileShareNameInput)
		if name == "" {
			m.wizard.errorMsg = "File share name cannot be empty"
			return m, nil
		}
		m.wizard.fileShareName = name
		m.wizard.errorMsg = ""
		m.wizard.step = FileWizardStepRegion
		m.wizard.selectedIndex = 0
		if len(m.wizard.fileShareRegions) > 0 {
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading available regions..."
		return m, m.fetchFileShareRegions()
	case tea.KeyBackspace:
		if len(m.wizard.fileShareNameInput) > 0 {
			m.wizard.fileShareNameInput = m.wizard.fileShareNameInput[:len(m.wizard.fileShareNameInput)-1]
		}
	case tea.KeyRunes:
		m.wizard.fileShareNameInput += string(msg.Runes)
	}
	return m, nil
}

func (m Model) handleFileWizardRegionKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	regions := m.wizard.fileShareRegions
	switch key {
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(regions)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(regions) == 0 {
			return m, nil
		}
		m.wizard.selectedRegion = regions[m.wizard.selectedIndex]
		m.wizard.errorMsg = ""
		m.wizard.step = FileWizardStepType
		m.wizard.fileShareTypeIdx = 0
		m.wizard.fileShareType = "standard-1az"
	case "left":
		m.wizard.step = FileWizardStepName
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

func (m Model) handleFileWizardTypeKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	types := []string{"standard-1az"}
	switch key {
	case "up", "k":
		if m.wizard.fileShareTypeIdx > 0 {
			m.wizard.fileShareTypeIdx--
		}
	case "down", "j":
		if m.wizard.fileShareTypeIdx < len(types)-1 {
			m.wizard.fileShareTypeIdx++
		}
	case "enter":
		m.wizard.fileShareType = types[m.wizard.fileShareTypeIdx]
		m.wizard.errorMsg = ""
		m.wizard.step = FileWizardStepSize
		if m.wizard.fileShareSize > 0 {
			m.wizard.fileShareSizeInput = fmt.Sprintf("%d", m.wizard.fileShareSize)
		} else {
			m.wizard.fileShareSizeInput = ""
		}
	case "left":
		m.wizard.step = FileWizardStepRegion
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

func (m Model) handleFileWizardSizeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		sizeStr := strings.TrimSpace(m.wizard.fileShareSizeInput)
		size, err := strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			m.wizard.errorMsg = "Size must be a positive integer (GB)"
			return m, nil
		}
		m.wizard.fileShareSize = size
		m.wizard.errorMsg = ""
		m.wizard.step = FileWizardStepNetwork
		m.wizard.fileShareNetworkMenuIdx = 0
		m.wizard.selectedIndex = 0
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading private networks..."
		return m, m.fetchFileShareNetworks()
	case tea.KeyBackspace:
		if len(m.wizard.fileShareSizeInput) > 0 {
			m.wizard.fileShareSizeInput = m.wizard.fileShareSizeInput[:len(m.wizard.fileShareSizeInput)-1]
		}
	case tea.KeyLeft:
		m.wizard.step = FileWizardStepType
		m.wizard.fileShareTypeIdx = 0
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.wizard.fileShareSizeInput += string(r)
			}
		}
	}
	return m, nil
}

func (m Model) handleFileWizardNetworkKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	if m.wizard.fileShareNetworkMenuIdx == 0 {
		networks := m.wizard.fileShareNetworks
		switch key {
		case "up", "k":
			if m.wizard.selectedIndex > 0 {
				m.wizard.selectedIndex--
			}
		case "down", "j":
			if m.wizard.selectedIndex < len(networks)-1 {
				m.wizard.selectedIndex++
			}
		case "enter":
			if len(networks) == 0 {
				return m, nil
			}
			network := networks[m.wizard.selectedIndex]
			networkPnId, _ := network["id"].(string)
			name, _ := network["name"].(string)
			openstackId, _ := network["_openstackId"].(string)
			if openstackId == "" {
				openstackId = networkPnId
			}
			m.wizard.fileShareNetworkId = openstackId
			m.wizard.fileShareNetworkName = name
			m.wizard.errorMsg = ""
			m.wizard.fileShareNetworkMenuIdx = 1
			m.wizard.selectedIndex = 0
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading subnets..."
			return m, m.fetchFileShareSubnets(networkPnId)
		case "left":
			m.wizard.step = FileWizardStepSize
			m.wizard.fileShareNetworkMenuIdx = 0
		}
	} else {
		subnets := m.wizard.fileShareSubnets
		switch key {
		case "up", "k":
			if m.wizard.selectedIndex > 0 {
				m.wizard.selectedIndex--
			}
		case "down", "j":
			if m.wizard.selectedIndex < len(subnets)-1 {
				m.wizard.selectedIndex++
			}
		case "enter":
			if len(subnets) == 0 {
				return m, nil
			}
			subnet := subnets[m.wizard.selectedIndex]
			subnetId, _ := subnet["id"].(string)
			subnetCIDR, _ := subnet["cidr"].(string)
			m.wizard.fileShareSubnetId = subnetId
			m.wizard.fileShareSubnetCIDR = subnetCIDR
			m.wizard.errorMsg = ""
			m.wizard.step = FileWizardStepConfirm
			m.wizard.fileShareConfirmBtnIdx = 0
		case "left":
			m.wizard.fileShareNetworkMenuIdx = 0
			m.wizard.selectedIndex = 0
		}
	}
	return m, nil
}

func (m Model) handleFileWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	switch key {
	case "right", "tab":
		if m.wizard.fileShareConfirmBtnIdx == 0 {
			m.wizard.fileShareConfirmBtnIdx = 1
		} else {
			m.wizard.fileShareConfirmBtnIdx = 0
		}
	case "enter":
		if m.wizard.fileShareConfirmBtnIdx == 0 {
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating file share..."
			return m, m.createFileShare()
		}
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, m.fetchDataForPath("/storage/file")
	case "left", "esc":
		m.wizard.step = FileWizardStepNetwork
		m.wizard.fileShareNetworkMenuIdx = 1
	}
	return m, nil
}
