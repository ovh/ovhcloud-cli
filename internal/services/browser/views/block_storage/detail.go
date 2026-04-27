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

// Action indices for volume detail view
const (
	VolumeActionDelete = iota
	VolumeActionRename
	VolumeActionExtend
)

var volumeActionLabels = []string{"Delete", "Rename", "Extend"}

// DetailView displays block storage volume details with actions.
type DetailView struct {
	views.BaseView
	volume         map[string]interface{}
	selectedAction int
	confirmMode    bool
	renameMode  bool
	renameInput string
	extendMode  bool
	extendInput string
}

func NewDetailView(ctx *views.Context, volume map[string]interface{}) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		volume:         volume,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *DetailView) Render(width, height int) string {
	var content strings.Builder

	if v.volume == nil {
		return views.StyleError.Render("No volume data available")
	}

	id := getString(v.volume, "id")
	status := getString(v.volume, "status")
	region := getString(v.volume, "region")
	vType := getString(v.volume, "type")
	createdAt := getString(v.volume, "createdAt")
	description := getString(v.volume, "description")
	size := getSizeStr(v.volume)
	bootable := getBootable(v.volume)

	encryptionLabel := "Aucun"
	encryptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	if strings.HasSuffix(vType, "-luks") {
		encryptionLabel = "Actif"
		encryptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	}

	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Type", vType) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Size", size+" GB") + "\n")
	infoContent.WriteString(views.StyleLabel.Render("Encryption:") + " " + encryptionStyle.Render(encryptionLabel) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Bootable", bootable) + "\n")
	if description != "" {
		infoContent.WriteString(views.RenderKeyValue("Description", description) + "\n")
	}
	infoContent.WriteString(views.RenderKeyValue("Created", createdAt) + "\n")
	content.WriteString(views.RenderBox("Volume Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	attachments := getAttachedTo(v.volume)
	var attachContent strings.Builder
	if len(attachments) > 0 {
		for _, instanceID := range attachments {
			attachContent.WriteString(fmt.Sprintf("  • %s\n", instanceID))
		}
	} else {
		attachContent.WriteString("  Not attached to any instance\n")
	}
	content.WriteString(views.RenderBox(fmt.Sprintf("Attached to (%d)", len(attachments)), attachContent.String(), width-4))
	content.WriteString("\n\n")

	actionsContent := v.renderActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4))

	return content.String()
}

func (v *DetailView) renderActions() string {
	if v.renameMode {
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).
			Width(40)
		return views.StyleStatusWarning.Render("New name:") + "\n" +
			inputStyle.Render(v.renameInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Confirm • Esc: Cancel")
	}

	if v.extendMode {
		currentSize := getSizeStr(v.volume)
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).
			Width(20)
		return views.StyleStatusWarning.Render(fmt.Sprintf("New size in GB (current: %s GB, must be greater):", currentSize)) + "\n" +
			inputStyle.Render(v.extendInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Confirm • Esc: Cancel")
	}

	var parts []string

	for i, label := range volumeActionLabels {
		var style lipgloss.Style
		if i == v.selectedAction {
			style = views.StyleButtonSelected
		} else if label == "Delete" {
			style = views.StyleButtonDanger
		} else {
			style = views.StyleButton
		}
		parts = append(parts, style.Render("["+label+"]"))
	}

	result := strings.Join(parts, " ")

	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render(
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", volumeActionLabels[v.selectedAction]))
	}

	return result
}

func (v *DetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if v.renameMode {
		switch msg.Type {
		case tea.KeyEscape:
			v.renameMode = false
			v.renameInput = ""
		case tea.KeyEnter:
			if v.renameInput != "" {
				name := v.renameInput
				v.renameMode = false
				v.renameInput = ""
				return func() tea.Msg {
					return ExecuteVolumeActionMsg{
						Volume: v.volume,
						Action: VolumeActionRename,
						Param:  name,
					}
				}
			}
		case tea.KeyBackspace:
			if len(v.renameInput) > 0 {
				v.renameInput = v.renameInput[:len(v.renameInput)-1]
			}
		case tea.KeyRunes:
			v.renameInput += string(msg.Runes)
		}
		return nil
	}

	if v.extendMode {
		switch msg.Type {
		case tea.KeyEscape:
			v.extendMode = false
			v.extendInput = ""
		case tea.KeyEnter:
			if v.extendInput != "" {
				newSize := v.extendInput
				currentSizeStr := getSizeStr(v.volume)
				// Validate new size > current size
				var newSizeInt, currentSizeInt int
				fmt.Sscanf(newSize, "%d", &newSizeInt)
				fmt.Sscanf(currentSizeStr, "%d", &currentSizeInt)
				if newSizeInt <= currentSizeInt {
					// Invalid: don't submit, reset input
					v.extendInput = ""
					return nil
				}
				v.extendMode = false
				v.extendInput = ""
				return func() tea.Msg {
					return ExecuteVolumeActionMsg{
						Volume: v.volume,
						Action: VolumeActionExtend,
						Param:  newSize,
					}
				}
			}
		case tea.KeyBackspace:
			if len(v.extendInput) > 0 {
				v.extendInput = v.extendInput[:len(v.extendInput)-1]
			}
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if r >= '0' && r <= '9' {
					v.extendInput += string(r)
				}
			}
		}
		return nil
	}

	switch key {
	case "left":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
		return nil
	case "right":
		if v.selectedAction < len(volumeActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteVolumeActionMsg{
					Volume: v.volume,
					Action: v.selectedAction,
				}
			}
		}
		switch v.selectedAction {
		case VolumeActionDelete:
			v.confirmMode = true
		case VolumeActionRename:
			v.renameInput = getString(v.volume, "name")
			v.renameMode = true
		case VolumeActionExtend:
			v.extendInput = ""
			v.extendMode = true
		}
		return nil
	case "esc":
		if v.confirmMode {
			v.confirmMode = false
			return nil
		}
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *DetailView) Title() string {
	name := getString(v.volume, "name")
	return fmt.Sprintf(" 💾 Block Storage > %s ", name)
}

func (v *DetailView) HelpText() string {
	if v.renameMode || v.extendMode {
		return "Type value • Enter: Confirm • Esc: Cancel"
	}
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • Esc: Back to List • q: Quit"
}

type ExecuteVolumeActionMsg struct {
	Volume map[string]interface{}
	Action int
	Param  string
}

func getBootable(volume map[string]interface{}) string {
	if b, ok := volume["bootable"].(bool); ok {
		if b {
			return "Yes"
		}
		return "No"
	}
	return "-"
}

func getAttachedTo(volume map[string]interface{}) []string {
	var result []string
	if raw, ok := volume["attachedTo"].([]interface{}); ok {
		for _, item := range raw {
			if id, ok := item.(string); ok {
				result = append(result, id)
			}
		}
	}
	return result
}
