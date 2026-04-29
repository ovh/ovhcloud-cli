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

var objectContainerTypes = []string{
	"S3 (API compatible S3)",
	"Swift (Swift API)",
}

var objectContainerTypeDescriptions = []string{
	"Un large éventail de fonctionnalités compatibles avec S3.\nDisponible en 1-AZ, 3-AZ et Local Zones (Standard ou High Performance selon la région)",
	"Solution basique pour le stockage sans besoin particulier en matière de performance.\nStockage objet natif d'OpenStack, avec les API Swift",
}

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
			displayText := fmt.Sprintf("  ▶ %s", t)
			if i == 0 {
				badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
				displayText += " " + badgeStyle.Render("[Recommandée]")
			}
			content.WriteString(selectedStyle.Render(displayText) + "\n")
			if i < len(objectContainerTypeDescriptions) {
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).MarginLeft(4)
				content.WriteString(descStyle.Render(objectContainerTypeDescriptions[i]) + "\n")
			}
		} else {
			displayText := fmt.Sprintf("    %s", t)
			if i == 0 {
				displayText += " [Recommandée]"
			}
			content.WriteString(itemStyle.Render(displayText) + "\n")
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

	typeName := objectContainerTypes[m.wizard.objectTypeIdx]
	content.WriteString(labelStyle.Render("  Type:          ") + valueStyle.Render(typeName) + "\n")

	if m.wizard.objectTypeIdx == 1 {
		swiftType := swiftContainerTypes[m.wizard.objectSwiftTypeIdx]
		content.WriteString(labelStyle.Render("  Swift Type:    ") + valueStyle.Render(swiftType) + "\n")
		content.WriteString(labelStyle.Render("  Region:        ") + valueStyle.Render(m.wizard.objectSwiftRegion) + "\n")
		content.WriteString("\n")
	} else {
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
	}

	baseBtn := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2).Bold(true)
	inactiveBtn := baseBtn.
		Foreground(lipgloss.Color("#888888")).
		BorderForeground(lipgloss.Color("#444444"))

	var createBtn, cancelBtn string
	if m.wizard.objectConfirmBtnIdx == 0 {
		createBtn = baseBtn.Foreground(lipgloss.Color("#00FF7F")).BorderForeground(lipgloss.Color("#00FF7F")).Render("✓ Create")
		cancelBtn = inactiveBtn.Render("✗ Cancel")
	} else {
		createBtn = inactiveBtn.Render("✓ Create")
		cancelBtn = baseBtn.Foreground(lipgloss.Color("#FF6B6B")).BorderForeground(lipgloss.Color("#FF6B6B")).Render("✗ Cancel")
	}
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, createBtn, "  ", cancelBtn))
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
		if m.wizard.objectTypeIdx == 1 {
			m.wizard.step = ObjectWizardStepSwiftType
			m.wizard.selectedIndex = 0
		} else {
			m.wizard.step = ObjectWizardStepRegion
			m.wizard.selectedIndex = 0
		}
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

// ─── S3 User wizard render functions ─────────────────────────────────────────

func (m Model) renderS3UserWizardDescStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7B68EE")).
		Padding(0, 1).
		Width(40)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)

	content.WriteString(titleStyle.Render("👤 Create S3 User") + "\n\n")
	content.WriteString(dimStyle.Render("Enter a description for this user (e.g. \"my-app-user\"):") + "\n\n")
	content.WriteString(inputStyle.Render(m.wizard.s3UserDescInput+"▌") + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(errStyle.Render("⚠ "+m.wizard.errorMsg) + "\n\n")
	}

	return content.String()
}

func (m Model) renderS3UserWizardConfirmStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)

	content.WriteString(titleStyle.Render("👤 Confirm S3 User Creation") + "\n\n")
	content.WriteString(labelStyle.Render("  Description:") + valueStyle.Render(m.wizard.s3UserDesc) + "\n")
	content.WriteString(labelStyle.Render("  Role:") + valueStyle.Render("objectstore_operator") + "\n\n")
	content.WriteString(infoStyle.Render("  ℹ  An S3 access key + secret key will be generated.\n     The secret key will only be shown once.") + "\n\n")

	// Buttons
	baseBtn := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2).Bold(true)
	inactiveBtn := baseBtn.
		Foreground(lipgloss.Color("#888888")).
		BorderForeground(lipgloss.Color("#444444"))

	var createBtn, cancelBtn string
	if m.wizard.s3UserConfirmBtnIdx == 0 {
		createBtn = baseBtn.Foreground(lipgloss.Color("#00FF7F")).BorderForeground(lipgloss.Color("#00FF7F")).Render("✓ Create")
		cancelBtn = inactiveBtn.Render("✗ Cancel")
	} else {
		createBtn = inactiveBtn.Render("✓ Create")
		cancelBtn = baseBtn.Foreground(lipgloss.Color("#FF6B6B")).BorderForeground(lipgloss.Color("#FF6B6B")).Render("✗ Cancel")
	}
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, createBtn, "  ", cancelBtn) + "\n\n")

	if m.wizard.errorMsg != "" {
		content.WriteString(errStyle.Render("⚠ Error: "+m.wizard.errorMsg) + "\n\n")
	}

	return content.String()
}

// renderS3CredentialsView renders the post-creation credentials display.
func (m Model) renderS3CredentialsView(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("✅ S3 User created successfully!") + "\n\n")
	content.WriteString(warningStyle.Render("⚠  Save these credentials now — the secret key will never be shown again.") + "\n\n")

	username := getStringValue(m.s3CreatedUser, "username", "")
	accessKey := ""
	secretKey := ""
	if m.s3CreatedCredentials != nil {
		accessKey = getStringValue(m.s3CreatedCredentials, "access", "")
		secretKey = getStringValue(m.s3CreatedCredentials, "secret", "")
	}

	content.WriteString(labelStyle.Render("  Username:") + valueStyle.Render(username) + "\n")
	content.WriteString(labelStyle.Render("  Access Key:") + valueStyle.Render(accessKey) + "\n")
	content.WriteString(labelStyle.Render("  Secret Key:") + valueStyle.Render(secretKey) + "\n\n")

	if m.s3CredentialsSavedPath != "" {
		savedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
		content.WriteString(savedStyle.Render("✅ Credentials saved to: "+m.s3CredentialsSavedPath) + "\n\n")
	} else if m.s3CredentialsSaveError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("❌ Save error: "+m.s3CredentialsSaveError) + "\n\n")
	} else {
		content.WriteString(dimStyle.Render("  Press [s] to save to ~/.aws/credentials") + "\n\n")
	}

	content.WriteString(dimStyle.Render("  Press [Enter] or [Esc] to return to the users list") + "\n")

	return content.String()
}

// ─── S3 User wizard key handlers ─────────────────────────────────────────────

func (m Model) handleS3UserWizardDescKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		desc := strings.TrimSpace(m.wizard.s3UserDescInput)
		if desc == "" {
			m.wizard.errorMsg = "Description cannot be empty."
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.s3UserDesc = desc
		m.wizard.step = S3UserWizardStepConfirm
	case "esc":
		m.mode = TableView
		m.wizard = WizardData{}
	case "backspace":
		if len(m.wizard.s3UserDescInput) > 0 {
			m.wizard.s3UserDescInput = m.wizard.s3UserDescInput[:len(m.wizard.s3UserDescInput)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.wizard.s3UserDescInput += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) handleS3UserWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h":
		m.wizard.s3UserConfirmBtnIdx = 0
	case "right", "l":
		m.wizard.s3UserConfirmBtnIdx = 1
	case "enter":
		if m.wizard.s3UserConfirmBtnIdx == 1 {
			m.mode = TableView
			m.wizard = WizardData{}
			return m, nil
		}
		m.wizard.errorMsg = ""
		m.wizard.isLoading = true
		m.wizard.loadingMessage = fmt.Sprintf("Creating user '%s'...", m.wizard.s3UserDesc)
		return m, m.createS3User()
	case "esc":
		m.wizard.step = S3UserWizardStepDescription
	}
	return m, nil
}

// ─── Swift wizard render functions ───────────────────────────────────────────

var swiftContainerTypes = []string{
	"Static hosting",
	"Private",
	"Public",
}

var swiftContainerTypeDescriptions = []string{
	"Hébergement statique - Accès rapide et performant pour vos sites. Liez vos domaines et déposez vos fichiers",
	"Privé - Facturation, informations légales, logs. Archivez simplement et selon vos usages",
	"Public - Multimédia, binaires, e-commerce. Stockez une infinité de données",
}

func (m Model) renderObjectWizardSwiftTypeStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("📦 Type de conteneur Swift:") + "\n\n")
	for i, t := range swiftContainerTypes {
		if i == m.wizard.objectSwiftTypeIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", t)) + "\n")
			if i < len(swiftContainerTypeDescriptions) {
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).MarginLeft(4)
				content.WriteString(descStyle.Render(swiftContainerTypeDescriptions[i]) + "\n")
			}
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", t)) + "\n")
		}
	}
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("↑↓: Select • Enter: Next • Esc: Back"))
	return content.String()
}

func (m Model) renderObjectWizardSwiftRegionStep(width int) string {
	var content strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	content.WriteString(titleStyle.Render("🌍 Région:") + "\n\n")

	if len(m.wizard.objectSwiftRegions) == 0 {
		content.WriteString(itemStyle.Render("  (aucune région disponible)") + "\n\n")
	} else {
		maxVisible := 10
		startIdx := 0
		if m.wizard.selectedIndex >= maxVisible {
			startIdx = m.wizard.selectedIndex - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.wizard.objectSwiftRegions) {
			endIdx = len(m.wizard.objectSwiftRegions)
		}

		if startIdx > 0 {
			content.WriteString(itemStyle.Render(fmt.Sprintf("  (...%d more above)", startIdx)) + "\n")
		}

		for i := startIdx; i < endIdx; i++ {
			r := m.wizard.objectSwiftRegions[i]
			if i == m.wizard.selectedIndex {
				content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", r)) + "\n")
			} else {
				content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", r)) + "\n")
			}
		}

		if endIdx < len(m.wizard.objectSwiftRegions) {
			content.WriteString(itemStyle.Render(fmt.Sprintf("  (...%d more below)", len(m.wizard.objectSwiftRegions)-endIdx)) + "\n")
		}
		content.WriteString("\n")
	}

	content.WriteString(hintStyle.Render("↑↓: Navigate • Enter: Next • Esc: Back"))
	return content.String()
}

// ─── Swift wizard key handlers ───────────────────────────────────────────────

func (m Model) handleObjectWizardSwiftTypeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.objectSwiftTypeIdx > 0 {
			m.wizard.objectSwiftTypeIdx--
		}
	case "down", "j":
		if m.wizard.objectSwiftTypeIdx < len(swiftContainerTypes)-1 {
			m.wizard.objectSwiftTypeIdx++
		}
	case "enter":
		m.wizard.step = ObjectWizardStepSwiftRegion
		m.wizard.selectedIndex = 0
		// Load Swift regions if not loaded
		if len(m.wizard.objectSwiftRegions) == 0 {
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading regions..."
			return m, m.fetchSwiftRegions()
		}
	case "esc":
		m.wizard.step = ObjectWizardStepType
	}
	return m, nil
}

func (m Model) handleObjectWizardSwiftRegionKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(m.wizard.objectSwiftRegions)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(m.wizard.objectSwiftRegions) > 0 {
			m.wizard.objectSwiftRegion = m.wizard.objectSwiftRegions[m.wizard.selectedIndex]
			m.wizard.step = ObjectWizardStepConfirm
		}
	case "esc":
		m.wizard.step = ObjectWizardStepSwiftType
	}
	return m, nil
}

func (m Model) handleS3CredentialsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		if m.s3CreatedCredentials != nil && m.s3CredentialsSavedPath == "" && m.s3CredentialsSaveError == "" {
			accessKey := getStringValue(m.s3CreatedCredentials, "access", "")
			secretKey := getStringValue(m.s3CreatedCredentials, "secret", "")
			username := getStringValue(m.s3CreatedUser, "username", "")
			return m, saveAWSCredentials(accessKey, secretKey, username)
		}
	case "enter", "esc":
		m.mode = LoadingView
		m.s3CreatedUser = nil
		m.s3CreatedCredentials = nil
		m.s3CredentialsSavedPath = ""
		m.s3CredentialsSaveError = ""
		return m, m.fetchDataForPath("/storage/object")
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

