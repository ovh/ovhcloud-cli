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
	SnapshotActionDelete       = iota
	SnapshotActionCreateVolume
)

var snapshotActionLabels = []string{"Delete", "Create Volume"}

const (
	snapshotSubNone = 0
	snapshotSubName = 1
	snapshotSubSize = 2
)

// ExecuteSnapshotActionMsg is dispatched when snapshot action is confirmed.
type ExecuteSnapshotActionMsg struct {
	Snapshot   map[string]interface{}
	Action     int
	VolumeName string
	VolumeSize string // GB as string
}

// SnapshotDetailView displays a volume snapshot with Delete and Create Volume actions.
type SnapshotDetailView struct {
	views.BaseView
	snapshot       map[string]interface{}
	selectedAction int
	confirmMode    bool
	subMenu        int
	nameInput      string
	sizeInput      string
}

func NewSnapshotDetailView(ctx *views.Context, snapshot map[string]interface{}) *SnapshotDetailView {
	return &SnapshotDetailView{
		BaseView: views.NewBaseView(ctx),
		snapshot: snapshot,
	}
}

func (v *SnapshotDetailView) Render(width, height int) string {
	var content strings.Builder

	id := getString(v.snapshot, "id")
	name := getString(v.snapshot, "name")
	status := getString(v.snapshot, "status")
	region := getString(v.snapshot, "region")
	volumeId := getString(v.snapshot, "volumeId")
	created := getString(v.snapshot, "creationDate")
	if len(created) > 19 {
		created = created[:19]
	}
	size := getSizeStr(v.snapshot)

	var info strings.Builder
	info.WriteString(views.RenderKeyValue("ID", id) + "\n")
	info.WriteString(views.RenderKeyValue("Name", name) + "\n")
	info.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	info.WriteString(views.RenderKeyValue("Region", region) + "\n")
	info.WriteString(views.RenderKeyValue("Size", size+" GB") + "\n")
	info.WriteString(views.RenderKeyValue("Source Volume", volumeId) + "\n")
	info.WriteString(views.RenderKeyValue("Created", created) + "\n")
	content.WriteString(views.RenderBox("Snapshot Information", info.String(), width-4))
	content.WriteString("\n\n")

	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(size), width-4))
	return content.String()
}

func (v *SnapshotDetailView) renderActions(currentSize string) string {
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(40)

	switch v.subMenu {
	case snapshotSubName:
		return views.StyleStatusWarning.Render("New volume name:") + "\n" +
			inputStyle.Render(v.nameInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Next • Esc: Cancel")
	case snapshotSubSize:
		return views.StyleStatusWarning.Render(fmt.Sprintf("Size in GB (current: %s GB, must be ≥):", currentSize)) + "\n" +
			inputStyle.Render(v.sizeInput+"▌") + "\n\n" +
			views.StyleFooter.Render("Enter: Create • Esc: Back")
	}

	var parts []string
	for i, label := range snapshotActionLabels {
		var style lipgloss.Style
		// Disable "Create Volume" if snapshot is not available
		if i == SnapshotActionCreateVolume && getString(v.snapshot, "status") != "available" {
			style = views.StyleButtonDisabled
		} else if i == v.selectedAction {
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
	if getString(v.snapshot, "status") != "available" {
		result += "\n\n" + views.StyleStatusWarning.Render(
			fmt.Sprintf("⚠️  Snapshot status: %s — Create Volume requires status: available", getString(v.snapshot, "status")))
	}
	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render("⚠️  Press Enter to confirm Delete, Escape to cancel")
	}
	return result
}

func (v *SnapshotDetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Name input step
	if v.subMenu == snapshotSubName {
		switch msg.Type {
		case tea.KeyEscape:
			v.subMenu = snapshotSubNone
			v.nameInput = ""
		case tea.KeyEnter:
			if strings.TrimSpace(v.nameInput) != "" {
				v.subMenu = snapshotSubSize
				if v.sizeInput == "" {
					v.sizeInput = getSizeStr(v.snapshot)
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

	// Size input step
	if v.subMenu == snapshotSubSize {
		switch msg.Type {
		case tea.KeyEscape:
			v.subMenu = snapshotSubName
		case tea.KeyEnter:
			if v.sizeInput != "" {
				snap := v.snapshot
				name := v.nameInput
				size := v.sizeInput
				v.subMenu = snapshotSubNone
				v.nameInput = ""
				v.sizeInput = ""
				return func() tea.Msg {
					return ExecuteSnapshotActionMsg{
						Snapshot:   snap,
						Action:     SnapshotActionCreateVolume,
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

	switch key {
	case "left", "h":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
	case "right", "l":
		if v.selectedAction < len(snapshotActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			snap := v.snapshot
			return func() tea.Msg {
				return ExecuteSnapshotActionMsg{Snapshot: snap, Action: SnapshotActionDelete}
			}
		}
		switch v.selectedAction {
		case SnapshotActionDelete:
			v.confirmMode = true
		case SnapshotActionCreateVolume:
			if getString(v.snapshot, "status") != "available" {
				return nil // disabled
			}
			v.subMenu = snapshotSubName
			v.nameInput = getString(v.snapshot, "name")
			v.sizeInput = getSizeStr(v.snapshot)
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

func (v *SnapshotDetailView) Title() string {
	return fmt.Sprintf(" 📸 Snapshot > %s ", getString(v.snapshot, "name"))
}

func (v *SnapshotDetailView) HelpText() string {
	if v.subMenu != snapshotSubNone {
		return "Type value • Enter: Confirm • Esc: Cancel"
	}
	if v.confirmMode {
		return "Enter: Confirm Delete • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • Esc: Back • q: Quit"
}
