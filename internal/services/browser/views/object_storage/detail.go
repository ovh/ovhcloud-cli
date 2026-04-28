// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package object_storage

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// Action indices for object storage container detail view
const (
	ContainerActionDelete = iota
)

var containerActionLabels = []string{"Delete"}

// DetailView displays object storage container details with actions.
type DetailView struct {
	views.BaseView
	container      map[string]interface{}
	selectedAction int
	confirmMode    bool
}

// NewDetailView creates a detail view for an S3 container.
func NewDetailView(ctx *views.Context, container map[string]interface{}) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		container:      container,
		selectedAction: 0,
		confirmMode:    false,
	}
}

// Render displays the full detail panel.
func (v *DetailView) Render(width, height int) string {
	var content strings.Builder

	if v.container == nil {
		return views.StyleError.Render("No container data available")
	}

	name := getString(v.container, "name")
	region := getString(v.container, "region")
	createdAt := getString(v.container, "createdAt")
	if createdAt == "" {
		createdAt = "-"
	}

	// Versioning
	versioningStatus := "-"
	if vers, ok := v.container["versioning"].(map[string]interface{}); ok {
		if s, ok := vers["status"].(string); ok {
			versioningStatus = s
		}
	}

	// Encryption
	encryptionStatus := "Disabled"
	encryptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	if enc, ok := v.container["encryption"].(map[string]interface{}); ok {
		if alg, _ := enc["sseAlgorithm"].(string); alg != "" {
			encryptionStatus = "Active (" + alg + ")"
			encryptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
		}
	}

	// Object lock
	objectLockStatus := "-"
	if ol, ok := v.container["objectLock"].(map[string]interface{}); ok {
		if s, _ := ol["status"].(string); s != "" {
			objectLockStatus = s
		}
	}

	// Objects count and size
	objectsCount := "-"
	if v, ok := v.container["objectsCount"]; ok {
		switch n := v.(type) {
		case float64:
			objectsCount = fmt.Sprintf("%d", int(n))
		}
	}

	sizeStr := "-"
	if sz, ok := v.container["objectsSize"]; ok {
		switch n := sz.(type) {
		case float64:
			if n < 1024 {
				sizeStr = fmt.Sprintf("%.0f B", n)
			} else if n < 1024*1024 {
				sizeStr = fmt.Sprintf("%.1f KB", n/1024)
			} else if n < 1024*1024*1024 {
				sizeStr = fmt.Sprintf("%.1f MB", n/1024/1024)
			} else {
				sizeStr = fmt.Sprintf("%.2f GB", n/1024/1024/1024)
			}
		}
	}

	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("Name", name) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Created at", createdAt) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Objects", objectsCount) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Total size", sizeStr) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Versioning", versioningStatus) + "\n")
	infoContent.WriteString(views.StyleLabel.Render("Encryption:") + " " + encryptionStyle.Render(encryptionStatus) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Object Lock", objectLockStatus) + "\n")

	content.WriteString(views.RenderBox("Container information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	actionsContent := v.renderActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4))

	return content.String()
}

func (v *DetailView) renderActions() string {
	var parts []string
	for i, label := range containerActionLabels {
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
			"⚠️  Press Enter to confirm deletion, Esc to cancel")
	}

	return result
}

// HandleKey processes keyboard input and returns a command.
func (v *DetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "left":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
		return nil
	case "right":
		if v.selectedAction < len(containerActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteContainerActionMsg{
					Container: v.container,
					Action:    v.selectedAction,
				}
			}
		}
		switch v.selectedAction {
		case ContainerActionDelete:
			v.confirmMode = true
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

// Title returns the header title.
func (v *DetailView) Title() string {
	name := getString(v.container, "name")
	return fmt.Sprintf(" 🪣 Object Storage > %s ", name)
}

// HelpText returns the footer help text.
func (v *DetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm deletion • Esc: Cancel"
	}
	return "←→: Select • Enter: Execute • Esc: Back to list • q: Quit"
}

// ExecuteContainerActionMsg is dispatched when the user confirms an action.
type ExecuteContainerActionMsg struct {
	Container map[string]interface{}
	Action    int
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
