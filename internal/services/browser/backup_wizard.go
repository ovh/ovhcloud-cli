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

var backupTypes = []string{"Snapshot", "Backup"}
var backupTypeDescriptions = []string{
	"Volume snapshot — fast, linked to the source volume",
	"Full independent backup — can be restored even if the volume is deleted",
}

// ─── Render functions ─────────────────────────────────────────────────────────

func (m Model) renderBackupWizard(width int) string {
	if m.wizard.isLoading {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Render("⏳ " + m.wizard.loadingMessage)
	}
	switch m.wizard.step {
	case BackupWizardStepVolume:
		return m.renderBackupWizardVolumeStep(width)
	case BackupWizardStepType:
		return m.renderBackupWizardTypeStep(width)
	case BackupWizardStepName:
		return m.renderBackupWizardNameStep(width)
	case BackupWizardStepConfirm:
		return m.renderBackupWizardConfirmStep(width)
	}
	return ""
}

func (m Model) renderBackupWizardVolumeStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("💾 Choose a volume to back up:") + "\n\n")

	if len(m.wizard.backupVolumes) == 0 {
		content.WriteString(dimStyle.Render("  No volume available.") + "\n")
	} else {
		maxVisible := 12
		startIdx := 0
		if m.wizard.backupVolumeIdx >= maxVisible {
			startIdx = m.wizard.backupVolumeIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.wizard.backupVolumes) {
			endIdx = len(m.wizard.backupVolumes)
		}
		if startIdx > 0 {
			content.WriteString(dimStyle.Render(fmt.Sprintf("  (...%d above)", startIdx)) + "\n")
		}
		for i := startIdx; i < endIdx; i++ {
			v := m.wizard.backupVolumes[i]
			name := getStringValue(v, "name", "-")
			region := getStringValue(v, "region", "-")
			size := 0
			if s, ok := v["size"].(float64); ok {
				size = int(s)
			}
			label := fmt.Sprintf("%-30s  %s  %d GB", name, region, size)
			if i == m.wizard.backupVolumeIdx {
				content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", label)) + "\n")
			} else {
				content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", label)) + "\n")
			}
		}
		if endIdx < len(m.wizard.backupVolumes) {
			content.WriteString(dimStyle.Render(fmt.Sprintf("  (...%d below)", len(m.wizard.backupVolumes)-endIdx)) + "\n")
		}
	}
	content.WriteString("\n" + hintStyle.Render("↑↓: Navigate • Enter: Next • Esc: Cancel"))
	return content.String()
}

func (m Model) renderBackupWizardTypeStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).MarginLeft(4)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	volName := ""
	if m.wizard.backupVolumeIdx < len(m.wizard.backupVolumes) {
		volName = getStringValue(m.wizard.backupVolumes[m.wizard.backupVolumeIdx], "name", "-")
	}
	content.WriteString(titleStyle.Render(fmt.Sprintf("Volume: %s", volName)) + "\n\n")
	content.WriteString(titleStyle.Render("Choose a backup type:") + "\n\n")

	for i, t := range backupTypes {
		if i == m.wizard.backupTypeIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", t)) + "\n")
			content.WriteString(descStyle.Render(backupTypeDescriptions[i]) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", t)) + "\n")
		}
	}
	content.WriteString("\n" + hintStyle.Render("↑↓: Navigate • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderBackupWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(40)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	backupTypeName := backupTypes[m.wizard.backupTypeIdx]
	content.WriteString(titleStyle.Render(fmt.Sprintf("Name your %s:", strings.ToLower(backupTypeName))) + "\n\n")
	content.WriteString(inputStyle.Render(m.wizard.backupNameInput+"▌") + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(errStyle.Render("  ❌ "+m.wizard.errorMsg) + "\n\n")
	}
	content.WriteString(hintStyle.Render("Type a name • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderBackupWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	vol := m.wizard.backupVolumes[m.wizard.backupVolumeIdx]
	volName := getStringValue(vol, "name", "-")
	volRegion := getStringValue(vol, "region", "-")
	backupTypeName := backupTypes[m.wizard.backupTypeIdx]

	content.WriteString(titleStyle.Render("Confirm creation:") + "\n\n")
	content.WriteString(labelStyle.Render("  Volume:") + valueStyle.Render(volName) + "\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(volRegion) + "\n")
	content.WriteString(labelStyle.Render("  Type:") + valueStyle.Render(backupTypeName) + "\n")
	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.backupName) + "\n\n")

	baseBtn := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2).Bold(true)
	inactiveBtn := baseBtn.
		Foreground(lipgloss.Color("#888888")).
		BorderForeground(lipgloss.Color("#444444"))

	var createBtn, cancelBtn string
	if m.wizard.backupConfirmBtnIdx == 0 {
		createBtn = baseBtn.Foreground(lipgloss.Color("#00FF7F")).BorderForeground(lipgloss.Color("#00FF7F")).Render("✓ Create")
		cancelBtn = inactiveBtn.Render("✗ Cancel")
	} else {
		createBtn = inactiveBtn.Render("✓ Create")
		cancelBtn = baseBtn.Foreground(lipgloss.Color("#FF6B6B")).BorderForeground(lipgloss.Color("#FF6B6B")).Render("✗ Cancel")
	}
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, createBtn, "  ", cancelBtn))
	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("←→: Select • Enter: Confirm • Esc: Back"))
	return content.String()
}

// ─── Key handlers ─────────────────────────────────────────────────────────────

func (m Model) handleBackupWizardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch m.wizard.step {
	case BackupWizardStepVolume:
		return m.handleBackupWizardVolumeKeys(key)
	case BackupWizardStepType:
		return m.handleBackupWizardTypeKeys(key)
	case BackupWizardStepName:
		return m.handleBackupWizardNameKeys(msg)
	case BackupWizardStepConfirm:
		return m.handleBackupWizardConfirmKeys(key)
	}
	return m, nil
}

func (m Model) handleBackupWizardVolumeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.backupVolumeIdx > 0 {
			m.wizard.backupVolumeIdx--
		}
	case "down", "j":
		if m.wizard.backupVolumeIdx < len(m.wizard.backupVolumes)-1 {
			m.wizard.backupVolumeIdx++
		}
	case "enter":
		if len(m.wizard.backupVolumes) > 0 {
			m.wizard.step = BackupWizardStepType
		}
	case "esc":
		m.mode = TableView
		m.wizard = WizardData{}
	}
	return m, nil
}

func (m Model) handleBackupWizardTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.backupTypeIdx > 0 {
			m.wizard.backupTypeIdx--
		}
	case "down", "j":
		if m.wizard.backupTypeIdx < len(backupTypes)-1 {
			m.wizard.backupTypeIdx++
		}
	case "enter":
		m.wizard.step = BackupWizardStepName
		m.wizard.backupNameInput = ""
		m.wizard.errorMsg = ""
	case "esc":
		m.wizard.step = BackupWizardStepVolume
	}
	return m, nil
}

func (m Model) handleBackupWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.wizard.backupNameInput)
		if name == "" {
			m.wizard.errorMsg = "Name cannot be empty."
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.backupName = name
		m.wizard.step = BackupWizardStepConfirm
		m.wizard.backupConfirmBtnIdx = 0
	case "backspace":
		if len(m.wizard.backupNameInput) > 0 {
			runes := []rune(m.wizard.backupNameInput)
			m.wizard.backupNameInput = string(runes[:len(runes)-1])
		}
	case "esc":
		m.wizard.step = BackupWizardStepType
	default:
		if len(msg.Runes) > 0 {
			m.wizard.backupNameInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleBackupWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.backupConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.backupConfirmBtnIdx = 1
	case "enter":
		if m.wizard.backupConfirmBtnIdx == 1 {
			// Cancel
			m.mode = TableView
			m.wizard = WizardData{}
			return m, nil
		}
		// Create
		vol := m.wizard.backupVolumes[m.wizard.backupVolumeIdx]
		volumeID := getStringValue(vol, "id", "")
		region := getStringValue(vol, "region", "")
		name := m.wizard.backupName
		isSnapshot := m.wizard.backupTypeIdx == 0
		m.wizard.isLoading = true
		if isSnapshot {
			m.wizard.loadingMessage = fmt.Sprintf("Creating snapshot '%s'...", name)
		} else {
			m.wizard.loadingMessage = fmt.Sprintf("Creating backup '%s'...", name)
		}
		return m, m.createVolumeBackupOrSnapshot(volumeID, region, name, isSnapshot)
	case "esc":
		m.wizard.step = BackupWizardStepName
	}
	return m, nil
}
