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

// ─── Object Storage wizard render functions ───────────────────────────────────

var objectContainerTypes = []string{"Standard", "High Performance"}

func (m Model) renderObjectWizardNameStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7B68EE")).
		Padding(0, 1).
		Width(40)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("🪣 Container name:") + "\n\n")
	content.WriteString(inputStyle.Render(m.wizard.objectNameInput+"▌") + "\n\n")
	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
		content.WriteString(errStyle.Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	} else {
		content.WriteString(hintStyle.Render("Lettres minuscules, chiffres et tirets uniquement (3-63 car).") + "\n\n")
	}
	content.WriteString(hintStyle.Render("Enter: Next • Esc: Cancel"))
	return content.String()
}

func (m Model) renderObjectWizardTypeStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("📦 Container type:") + "\n\n")
	for i, t := range objectContainerTypes {
		if i == m.wizard.objectTypeIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", t)) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", t)) + "\n")
		}
	}
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("↑↓: Select • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderObjectWizardRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("🌍 Region:") + "\n\n")
	if len(m.wizard.objectRegions) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render("Aucune région disponible.") + "\n")
	} else {
		maxVisible := 10
		startIdx := 0
		if m.wizard.selectedIndex >= maxVisible {
			startIdx = m.wizard.selectedIndex - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.wizard.objectRegions) {
			endIdx = len(m.wizard.objectRegions)
		}
		if startIdx > 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ↑ more") + "\n")
		}
		for i := startIdx; i < endIdx; i++ {
			r := m.wizard.objectRegions[i]
			if i == m.wizard.selectedIndex {
				content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", r)) + "\n")
			} else {
				content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", r)) + "\n")
			}
		}
		if endIdx < len(m.wizard.objectRegions) {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ↓ more") + "\n")
		}
	}
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("↑↓: Select • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderObjectWizardReplicationStep(width int) string {
	return renderObjectToggleStep("🔄 Offsite Replication:",
		"Automatically replicate objects to another geographic zone.",
		m.wizard.objectReplication)
}

func (m Model) renderObjectWizardVersioningStep(width int) string {
	return renderObjectToggleStep("📂 Versioning:",
		"Keep multiple versions of each object (required for Object Lock).",
		m.wizard.objectVersioning)
}

func (m Model) renderObjectWizardObjectLockStep(width int) string {
	return renderObjectToggleStep("🔒 Object Lock (WORM):",
		"Prevent deletion or modification of objects for a defined period.",
		m.wizard.objectLock)
}

func (m Model) renderObjectWizardEncryptionStep(width int) string {
	return renderObjectToggleStep("🔐 Server-side Encryption (AES-256):",
		"Automatically encrypt all objects stored in this container.",
		m.wizard.objectEncryption)
}

func renderObjectToggleStep(title, description string, enabled bool) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render(title) + "\n\n")
	content.WriteString(descStyle.Render(description) + "\n\n")

	onStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Padding(0, 2)
	offStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 2)
	selectedBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	normalBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#444444")).Padding(0, 1)

	var onBtn, offBtn string
	if enabled {
		onBtn = selectedBorder.Render(onStyle.Render("✓ Enabled"))
		offBtn = normalBorder.Render(offStyle.Render("  Disabled"))
	} else {
		onBtn = normalBorder.Render(offStyle.Render("  Enabled"))
		offBtn = selectedBorder.Render(onStyle.Render("✗ Disabled"))
	}
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, onBtn, "  ", offBtn))
	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("←→ or y/n: Toggle • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderObjectWizardUserStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("👤 Owner user:") + "\n\n")

	// First option: no user (project-level)
	if m.wizard.objectUserIdx == 0 {
		content.WriteString(selectedStyle.Render("  ▶ [Aucun utilisateur spécifique]") + "\n")
	} else {
		content.WriteString(itemStyle.Render("    [Aucun utilisateur spécifique]") + "\n")
	}

	for i, user := range m.wizard.objectUsers {
		username, _ := user["username"].(string)
		if username == "" {
			if desc, ok := user["description"].(string); ok {
				username = desc
			} else {
				username = fmt.Sprintf("user-%d", i+1)
			}
		}
		idx := i + 1
		if idx == m.wizard.objectUserIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", username)) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", username)) + "\n")
		}
	}

	if len(m.wizard.objectUsers) == 0 {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("    (aucun utilisateur trouvé)") + "\n")
	}

	content.WriteString("\n")
	content.WriteString(hintStyle.Render("↑↓: Select • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderObjectWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render("✅ Container summary:") + "\n\n")
	content.WriteString(labelStyle.Render("  Name:          ") + valueStyle.Render(m.wizard.objectName) + "\n")

	typeName := "Standard"
	if m.wizard.objectTypeIdx == 1 {
		typeName = "High Performance"
	}
	content.WriteString(labelStyle.Render("  Type:          ") + valueStyle.Render(typeName) + "\n")

	region := m.wizard.selectedRegion
	if region == "" && len(m.wizard.objectRegions) > 0 {
		region = m.wizard.objectRegions[0]
	}
	content.WriteString(labelStyle.Render("  Region:        ") + valueStyle.Render(region) + "\n")
	content.WriteString(labelStyle.Render("  Replication:   ") + valueStyle.Render(boolToEnglish(m.wizard.objectReplication)) + "\n")
	content.WriteString(labelStyle.Render("  Versioning:    ") + valueStyle.Render(boolToEnglish(m.wizard.objectVersioning)) + "\n")
	content.WriteString(labelStyle.Render("  Object Lock:   ") + valueStyle.Render(boolToEnglish(m.wizard.objectLock)) + "\n")
	content.WriteString(labelStyle.Render("  Encryption:    ") + valueStyle.Render(boolToEnglish(m.wizard.objectEncryption)) + "\n")

	if m.wizard.objectUserIdx > 0 && m.wizard.objectUserIdx <= len(m.wizard.objectUsers) {
		user := m.wizard.objectUsers[m.wizard.objectUserIdx-1]
		username, _ := user["username"].(string)
		content.WriteString(labelStyle.Render("  User:          ") + valueStyle.Render(username) + "\n")
	} else {
		content.WriteString(labelStyle.Render("  User:          ") + valueStyle.Render("(none)") + "\n")
	}

	content.WriteString("\n")

	createStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#00AA55")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2)
	cancelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	var createBtn, cancelBtn string
	if m.wizard.objectConfirmBtnIdx == 0 {
		createBtn = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00FF7F")).Render(createStyle.Render("  ▶ [Create]"))
		cancelBtn = lipgloss.NewStyle().Padding(1).Render(cancelStyle.Render("    [Cancel]"))
	} else {
		createBtn = lipgloss.NewStyle().Padding(1).Render(createStyle.Render("    [Create]"))
		cancelBtn = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF6B6B")).Render(cancelStyle.Render("  ▶ [Cancel]"))
	}
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, createBtn, "  ", cancelBtn))
	content.WriteString("\n\n")
	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
		content.WriteString(errStyle.Render("⚠ Error: "+m.wizard.errorMsg) + "\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("←→: Toggle • Enter: Retry • N: Change name • Esc: Back"))
	} else {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("←→: Toggle • Enter: Confirm • Esc: Back"))
	}
	return content.String()
}

func boolToEnglish(v bool) string {
	if v {
		return "Enabled"
	}
	return "Disabled"
}

// ─── Object Storage wizard key handlers ──────────────────────────────────────

func (m Model) handleObjectWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := strings.TrimSpace(m.wizard.objectNameInput)
		if name == "" {
			return m, nil
		}
		// Enforce S3 bucket naming rules
		if len(name) < 3 || len(name) > 63 {
			m.wizard.errorMsg = "Le nom doit contenir entre 3 et 63 caractères."
			return m, nil
		}
		for _, ch := range name {
			if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
				m.wizard.errorMsg = "Uniquement lettres minuscules, chiffres et tirets (-). Pas de majuscules ni espaces."
				return m, nil
			}
		}
		if name[0] == '-' || name[len(name)-1] == '-' {
			m.wizard.errorMsg = "Le nom ne peut pas commencer ou finir par un tiret."
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.objectName = name
		m.wizard.step = ObjectWizardStepType
		m.wizard.selectedIndex = m.wizard.objectTypeIdx
		return m, nil
	case tea.KeyEscape:
		m.mode = TableView
		m.wizard = WizardData{}
		return m, nil
	case tea.KeyBackspace:
		if len(m.wizard.objectNameInput) > 0 {
			m.wizard.objectNameInput = m.wizard.objectNameInput[:len(m.wizard.objectNameInput)-1]
		}
	case tea.KeyRunes:
		m.wizard.objectNameInput += string(msg.Runes)
	}
	return m, nil
}

func (m Model) handleObjectWizardTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.objectTypeIdx > 0 {
			m.wizard.objectTypeIdx--
		}
	case "down", "j":
		if m.wizard.objectTypeIdx < len(objectContainerTypes)-1 {
			m.wizard.objectTypeIdx++
		}
	case "enter":
		m.wizard.step = ObjectWizardStepRegion
		m.wizard.selectedIndex = 0
	case "esc":
		m.wizard.step = ObjectWizardStepName
	}
	return m, nil
}

func (m Model) handleObjectWizardRegionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(m.wizard.objectRegions)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(m.wizard.objectRegions) > 0 {
			m.wizard.selectedRegion = m.wizard.objectRegions[m.wizard.selectedIndex]
		}
		m.wizard.step = ObjectWizardStepReplication
	case "esc":
		m.wizard.step = ObjectWizardStepType
	}
	return m, nil
}

// handleObjectWizardToggleKeys handles yes/no toggle steps.
// nextStep is the step to advance to on Enter.
// field is a pointer to the bool to toggle.
func (m Model) handleObjectWizardToggleKeys(key string, nextStep WizardStep, field *bool) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h", "y":
		*field = true
	case "right", "l", "n":
		*field = false
	case "enter":
		m.wizard.step = nextStep
	case "esc":
		m.wizard.step = prevObjectStep(m.wizard.step)
	}
	return m, nil
}

func prevObjectStep(step WizardStep) WizardStep {
	switch step {
	case ObjectWizardStepReplication:
		return ObjectWizardStepRegion
	case ObjectWizardStepVersioning:
		return ObjectWizardStepReplication
	case ObjectWizardStepObjectLock:
		return ObjectWizardStepVersioning
	case ObjectWizardStepUser:
		return ObjectWizardStepObjectLock
	case ObjectWizardStepEncryption:
		return ObjectWizardStepUser
	default:
		return ObjectWizardStepReplication
	}
}

func (m Model) handleObjectWizardUserKeys(key string) (tea.Model, tea.Cmd) {
	totalItems := 1 + len(m.wizard.objectUsers) // 0 = no user, 1..N = users
	switch key {
	case "up", "k":
		if m.wizard.objectUserIdx > 0 {
			m.wizard.objectUserIdx--
		}
	case "down", "j":
		if m.wizard.objectUserIdx < totalItems-1 {
			m.wizard.objectUserIdx++
		}
	case "enter":
		m.wizard.step = ObjectWizardStepEncryption
	case "esc":
		m.wizard.step = ObjectWizardStepObjectLock
	}
	return m, nil
}

func (m Model) handleObjectWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.objectConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.objectConfirmBtnIdx = 1
	case "n":
		// Shortcut to go back to name step (useful when there's a name conflict error)
		m.wizard.errorMsg = ""
		m.wizard.step = ObjectWizardStepName
		return m, nil
	case "enter":
		if m.wizard.objectConfirmBtnIdx == 1 {
			// Cancel
			m.mode = TableView
			m.wizard = WizardData{}
			return m, nil
		}
		// Clear previous error before retrying
		m.wizard.errorMsg = ""
		// Create
		m.wizard.isLoading = true
		m.wizard.loadingMessage = fmt.Sprintf("Creating container '%s'...", m.wizard.objectName)
		return m, m.createObjectContainer()
	case "esc":
		m.wizard.step = ObjectWizardStepEncryption
	}
	return m, nil
}
