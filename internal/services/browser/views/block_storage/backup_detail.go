// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package block_storage

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

const (
	BackupActionDelete       = iota
	BackupActionRestore      // restore onto existing volume
	BackupActionCreateVolume // create new volume then restore
)

var backupActionLabels = []string{"Delete", "Restore to Volume", "Create Volume"}

const (
	backupSubNone   = 0
	backupSubPicker = 1 // existing-volume picker for Restore
	backupSubName   = 2 // name input for Create Volume
	backupSubSize   = 3 // size input for Create Volume
)

// ExecuteBackupActionMsg is dispatched when a backup action is confirmed.
type ExecuteBackupActionMsg struct {
	Backup     map[string]interface{}
	Action     int
	VolumeID   string // for Restore action
	VolumeName string // for CreateVolume action
	VolumeSize string // for CreateVolume action
}

// BackupVolumesLoadedMsg is sent by the manager after loading volumes for restore picker.
type BackupVolumesLoadedMsg struct {
	Volumes []map[string]interface{}
}

// BackupDetailView displays a volume backup with Delete, Restore and Create Volume actions.
type BackupDetailView struct {
	views.BaseView
	backup         map[string]interface{}
	selectedAction int
	confirmMode    bool
	subMenu        int
	restoreVolumes []map[string]interface{}
	restoreIdx     int
	nameInput      string
	sizeInput      string
}

func NewBackupDetailView(ctx *views.Context, backup map[string]interface{}) *BackupDetailView {
	return &BackupDetailView{
		BaseView: views.NewBaseView(ctx),
		backup:   backup,
	}
}

// SetRestoreVolumes is called by the manager after volumes are loaded.
func (v *BackupDetailView) SetRestoreVolumes(volumes []map[string]interface{}) {
	v.restoreVolumes = volumes
	v.restoreIdx = 0
}

func (v *BackupDetailView) Render(width, height int) string {
	var content strings.Builder

	id := getString(v.backup, "id")
	name := getString(v.backup, "name")
	status := getString(v.backup, "status")
	region := getString(v.backup, "region")
	volumeId := getString(v.backup, "volumeId")
	created := getString(v.backup, "creationDate")
	if len(created) > 19 {
		created = created[:19]
	}
	size := getSizeStr(v.backup)

	var info strings.Builder
	info.WriteString(views.RenderKeyValue("ID", id) + "\n")
	info.WriteString(views.RenderKeyValue("Name", name) + "\n")
	info.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	info.WriteString(views.RenderKeyValue("Region", region) + "\n")
	info.WriteString(views.RenderKeyValue("Size", size+" GB") + "\n")
	info.WriteString(views.RenderKeyValue("Source Volume", volumeId) + "\n")
	info.WriteString(views.RenderKeyValue("Created", created) + "\n")
	content.WriteString(views.RenderBox("Backup Information", info.String(), width-4))
	content.WriteString("\n\n")

	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(), width-4))
	return content.String()
}

func (v *BackupDetailView) renderActions() string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(40)

	if v.subMenu == backupSubName {
		return views.StyleStatusWarning.Render("Nom du nouveau volume :") + "\n" +
			inputStyle.Render(v.nameInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Suivant • Esc: Annuler")
	}

	if v.subMenu == backupSubSize {
		currentSize := getSizeStr(v.backup)
		return views.StyleStatusWarning.Render(fmt.Sprintf("Taille en GB (backup: %s GB, doit être ≥) :", currentSize)) + "\n" +
			inputStyle.Render(v.sizeInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Créer • Esc: Retour")
	}

	if v.subMenu == backupSubPicker {
		if len(v.restoreVolumes) == 0 {
			return views.StyleStatusWarning.Render("⏳ Chargement des volumes...") + "\n\n" +
				views.StyleFooter.Render("Esc: Annuler")
		}
		selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
		var sb strings.Builder
		sb.WriteString(views.StyleStatusWarning.Render("⚠️  Choisissez le volume cible (les données seront écrasées) :") + "\n\n")
		maxVisible := 8
		startIdx := 0
		if v.restoreIdx >= maxVisible {
			startIdx = v.restoreIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(v.restoreVolumes) {
			endIdx = len(v.restoreVolumes)
		}
		if startIdx > 0 {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  (...%d au-dessus)", startIdx)) + "\n")
		}
		for i := startIdx; i < endIdx; i++ {
			vol := v.restoreVolumes[i]
			label := fmt.Sprintf("%-28s  %s  %s GB", getString(vol, "name"), getString(vol, "region"), getSizeStr(vol))
			if i == v.restoreIdx {
				sb.WriteString(selectedStyle.Render("  ▶ "+label) + "\n")
			} else {
				sb.WriteString(itemStyle.Render("    "+label) + "\n")
			}
		}
		if endIdx < len(v.restoreVolumes) {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  (...%d en-dessous)", len(v.restoreVolumes)-endIdx)) + "\n")
		}
		sb.WriteString("\n" + views.StyleFooter.Render("↑↓: Navigate • Enter: Confirm Restore • Esc: Annuler"))
		return sb.String()
	}

	var parts []string
	for i, label := range backupActionLabels {
		var style lipgloss.Style
		if i == v.selectedAction {
			if label == "Delete" {
				style = views.StyleButtonDangerSelected
			} else {
				style = views.StyleButtonSelected
			}
		} else if label == "Delete" {
			style = views.StyleButtonDanger
		} else {
			style = views.StyleButton
		}
		parts = append(parts, style.Render("["+label+"]"))
	}
	result := strings.Join(parts, " ")
	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render("⚠️  Press Enter to confirm Delete, Escape to cancel")
	}
	return result
}

func (v *BackupDetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Name input for Create Volume
	if v.subMenu == backupSubName {
		switch msg.Type {
		case tea.KeyEscape:
			v.subMenu = backupSubNone
			v.nameInput = ""
		case tea.KeyEnter:
			if strings.TrimSpace(v.nameInput) != "" {
				v.subMenu = backupSubSize
				if v.sizeInput == "" {
					v.sizeInput = getSizeStr(v.backup)
				}
			}
		case tea.KeyBackspace:
			if len(v.nameInput) > 0 {
				runes := []rune(v.nameInput)
				v.nameInput = string(runes[:len(runes)-1])
			}
		case tea.KeyRunes:
			v.nameInput += string(msg.Runes)
		}
		return nil
	}

	// Size input for Create Volume
	if v.subMenu == backupSubSize {
		switch msg.Type {
		case tea.KeyEscape:
			v.subMenu = backupSubName
		case tea.KeyEnter:
			if v.sizeInput != "" {
				backup := v.backup
				name := v.nameInput
				size := v.sizeInput
				v.subMenu = backupSubNone
				v.nameInput = ""
				v.sizeInput = ""
				return func() tea.Msg {
					return ExecuteBackupActionMsg{
						Backup:     backup,
						Action:     BackupActionCreateVolume,
						VolumeName: name,
						VolumeSize: size,
					}
				}
			}
		case tea.KeyBackspace:
			if len(v.sizeInput) > 0 {
				v.sizeInput = v.sizeInput[:len(v.sizeInput)-1]
			}
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' {
					v.sizeInput += string(r)
				}
			}
		}
		return nil
	}

	if v.subMenu == backupSubPicker {
		switch key {
		case "up", "k":
			if v.restoreIdx > 0 {
				v.restoreIdx--
			}
		case "down", "j":
			if v.restoreIdx < len(v.restoreVolumes)-1 {
				v.restoreIdx++
			}
		case "enter":
			if len(v.restoreVolumes) > 0 {
				backup := v.backup
				volID := getString(v.restoreVolumes[v.restoreIdx], "id")
				v.subMenu = backupSubNone
				return func() tea.Msg {
					return ExecuteBackupActionMsg{
						Backup:   backup,
						Action:   BackupActionRestore,
						VolumeID: volID,
					}
				}
			}
		case "esc":
			v.subMenu = backupSubNone
		}
		return nil
	}

	switch key {
	case "left", "h":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
	case "right", "l":
		if v.selectedAction < len(backupActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			backup := v.backup
			return func() tea.Msg {
				return ExecuteBackupActionMsg{Backup: backup, Action: BackupActionDelete}
			}
		}
		switch v.selectedAction {
		case BackupActionDelete:
			v.confirmMode = true
		case BackupActionRestore:
			// Signal manager to load volumes, show picker (empty, loading state)
			v.subMenu = backupSubPicker
			v.restoreVolumes = nil
			v.restoreIdx = 0
			backup := v.backup
			return func() tea.Msg {
				return LoadBackupRestoreVolumesMsg{Backup: backup}
			}
		case BackupActionCreateVolume:
			v.subMenu = backupSubName
			v.nameInput = getString(v.backup, "name")
			v.sizeInput = getSizeStr(v.backup)
		}
	case "esc":
		if v.confirmMode {
			v.confirmMode = false
			return nil
		}
		return func() tea.Msg { return views.GoBackMsg{} }
	}
	return nil
}

// LoadBackupRestoreVolumesMsg asks the manager to fetch volumes for the restore picker.
type LoadBackupRestoreVolumesMsg struct {
	Backup map[string]interface{}
}

func (v *BackupDetailView) Title() string {
	return fmt.Sprintf(" 💾 Backup > %s ", getString(v.backup, "name"))
}

func (v *BackupDetailView) HelpText() string {
	switch v.subMenu {
	case backupSubPicker:
		return "↑↓: Navigate • Enter: Restore • Esc: Cancel"
	case backupSubName:
		return "Type name • Enter: Suivant • Esc: Cancel"
	case backupSubSize:
		return "Type size in GB • Enter: Créer • Esc: Retour"
	}
	if v.confirmMode {
		return "Enter: Confirm Delete • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • Esc: Back • q: Quit"
}
