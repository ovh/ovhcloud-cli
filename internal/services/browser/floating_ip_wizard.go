// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderFIPWizardRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	content.WriteString(titleStyle.Render("Choose a region for the Floating IP:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading regions..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	if len(m.wizard.fipAvailableRegions) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render("No regions available.") + "\n")
	} else {
		maxVisible := 14
		startIdx := 0
		if m.wizard.fipRegionIdx >= maxVisible {
			startIdx = m.wizard.fipRegionIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.wizard.fipAvailableRegions) {
			endIdx = len(m.wizard.fipAvailableRegions)
		}
		for i := startIdx; i < endIdx; i++ {
			r := m.wizard.fipAvailableRegions[i]
			if i == m.wizard.fipRegionIdx {
				content.WriteString(selectedStyle.Render("▶ "+r) + "\n")
			} else {
				content.WriteString(dimStyle.Render("  "+r) + "\n")
			}
		}
		if len(m.wizard.fipAvailableRegions) > maxVisible {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
				Render(fmt.Sprintf("\n  %d / %d régions", m.wizard.fipRegionIdx+1, len(m.wizard.fipAvailableRegions))))
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Naviguer • Enter : Sélectionner • Esc : Annuler"))
	return content.String()
}

func (m Model) renderFIPWizardInstanceStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	content.WriteString(titleStyle.Render("Attach to an instance:") + "\n\n")
	content.WriteString(descStyle.Render(fmt.Sprintf("Region: %s", m.wizard.fipRegion)) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Loading instances..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	// Index 0 = not supported standalone (shown as disabled hint)
	entries := []string{"⚠ No instance (not supported)"}
	for _, inst := range m.wizard.fipInstances {
		name := getStringValue(inst, "name", getStringValue(inst, "id", "unknown"))
		entries = append(entries, name)
	}

	maxVisible := 12
	startIdx := 0
	if m.wizard.fipInstanceIdx >= maxVisible {
		startIdx = m.wizard.fipInstanceIdx - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(entries) {
		endIdx = len(entries)
	}
	for i := startIdx; i < endIdx; i++ {
		if i == m.wizard.fipInstanceIdx {
			content.WriteString(selectedStyle.Render("▶ "+entries[i]) + "\n")
		} else {
			content.WriteString(dimStyle.Render("  "+entries[i]) + "\n")
		}
	}

	content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("↑↓ Navigate • Enter: Select • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) renderFIPWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	lbLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("Confirm Floating IP creation:") + "\n\n")
	content.WriteString(lbLabelStyle.Render("  Region:") + valStyle.Render(m.wizard.fipRegion) + "\n")

	instanceDisplay := "None (standalone)"
	if m.wizard.fipInstanceName != "" {
		instanceDisplay = m.wizard.fipInstanceName
	}
	content.WriteString(lbLabelStyle.Render("  Instance:") + valStyle.Render(instanceDisplay) + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("⏳ Creating..."))
		return content.String()
	}
	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).
			Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	btnCreate := lipgloss.NewStyle().Background(lipgloss.Color("#00FF7F")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Create ")
	btnCancel := lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Cancel ")
	if m.wizard.fipConfirmBtnIdx == 1 {
		btnCreate = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Foreground(lipgloss.Color("#CCCCCC")).Padding(0, 2).Render(" Create ")
		btnCancel = lipgloss.NewStyle().Background(lipgloss.Color("#FF6B6B")).Foreground(lipgloss.Color("#000000")).Bold(true).Padding(0, 2).Render(" Cancel ")
	}
	content.WriteString(btnCreate + "  " + btnCancel + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
		Render("←→: Select • Enter: Confirm • Esc: Cancel"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleFIPWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.fipRegionIdx > 0 {
			m.wizard.fipRegionIdx--
		}
	case "down", "j":
		if m.wizard.fipRegionIdx < len(m.wizard.fipAvailableRegions)-1 {
			m.wizard.fipRegionIdx++
		}
	case "enter":
		if len(m.wizard.fipAvailableRegions) > 0 {
			m.wizard.fipRegion = m.wizard.fipAvailableRegions[m.wizard.fipRegionIdx]
			m.wizard.fipInstanceIdx = 0
			m.wizard.step = FIPWizardStepInstance
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading instances..."
			return m, m.fetchFIPInstances()
		}
	}
	return m, nil
}

func (m Model) handleFIPWizardInstanceKeys(key string) (tea.Model, tea.Cmd) {
	total := 1 + len(m.wizard.fipInstances) // standalone + instances
	switch key {
	case "up", "k":
		if m.wizard.fipInstanceIdx > 0 {
			m.wizard.fipInstanceIdx--
		}
	case "down", "j":
		if m.wizard.fipInstanceIdx < total-1 {
			m.wizard.fipInstanceIdx++
		}
	case "enter":
		if m.wizard.fipInstanceIdx == 0 {
			// Standalone — no instance
			m.wizard.fipInstanceId = ""
			m.wizard.fipInstanceName = ""
		} else {
			inst := m.wizard.fipInstances[m.wizard.fipInstanceIdx-1]
			m.wizard.fipInstanceId = getStringValue(inst, "id", "")
			m.wizard.fipInstanceName = getStringValue(inst, "name", getStringValue(inst, "id", "unknown"))
		}
		m.wizard.fipConfirmBtnIdx = 0
		m.wizard.step = FIPWizardStepConfirm
	case "left":
		m.wizard.step = FIPWizardStepRegion
	}
	return m, nil
}

func (m Model) handleFIPWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.fipConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.fipConfirmBtnIdx = 1
	case "enter":
		if m.wizard.fipConfirmBtnIdx == 1 {
			// Cancel → back to instance selection
			m.wizard.step = FIPWizardStepInstance
			return m, nil
		}
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Creating Floating IP..."
		return m, m.createStandaloneFloatingIP()
	}
	return m, nil
}
