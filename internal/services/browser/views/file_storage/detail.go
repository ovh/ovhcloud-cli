// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package file_storage

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// Action indices for file share detail view
const (
	FileShareActionDelete = iota
	FileShareActionRename
	FileShareActionExtend
)

var fileShareActionLabels = []string{"Delete", "Rename", "Extend"}

// DetailView displays file storage share details with actions.
type DetailView struct {
	views.BaseView
	share          map[string]interface{}
	selectedAction int
	confirmMode    bool
	renameMode     bool
	renameInput    string
	extendMode     bool
	extendInput    string
}

// NewDetailView creates a detail view for a file share.
func NewDetailView(ctx *views.Context, share map[string]interface{}) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		share:          share,
		selectedAction: 0,
		confirmMode:    false,
	}
}

// Render displays the full detail panel.
func (v *DetailView) Render(width, height int) string {
	var content strings.Builder

	if v.share == nil {
		return views.StyleError.Render("No file share data available")
	}

	id := getString(v.share, "id")
	name := getString(v.share, "name")
	status := getString(v.share, "status")
	region := getString(v.share, "region")
	shareType := getString(v.share, "type")
	size := getSizeStr(v.share)
	createdAt := getString(v.share, "createdAt")
	if createdAt == "" {
		createdAt = getString(v.share, "creationDate")
	}

	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Name", name) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Type", shareType) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Capacity", size+" GB") + "\n")
	if createdAt != "" {
		infoContent.WriteString(views.RenderKeyValue("Created at", createdAt) + "\n")
	}

	content.WriteString(views.RenderBox("Share information", infoContent.String(), width-4))
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
		return views.StyleStatusWarning.Render("Nouveau nom :") + "\n" +
			inputStyle.Render(v.renameInput+"▌") + "\n\n" +
					views.StyleFooter.Render("Enter: Confirm • Esc: Cancel")
	}

	if v.extendMode {
		currentSize := getSizeStr(v.share)
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF7F")).
			Padding(0, 1).
			Width(20)
		return views.StyleStatusWarning.Render(
				fmt.Sprintf("New size in GB (current: %s GB, must be larger):", currentSize),
			) + "\n" +
				inputStyle.Render(v.extendInput+"▌") + "\n\n" +
				views.StyleFooter.Render("Enter: Confirm • Esc: Cancel")
	}

	var parts []string
	for i, label := range fileShareActionLabels {
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
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Esc to cancel",
				fileShareActionLabels[v.selectedAction]))
	}

	return result
}

// HandleKey processes keyboard input and returns a command.
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
					return ExecuteFileShareActionMsg{
						Share:  v.share,
						Action: FileShareActionRename,
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
				currentSizeStr := getSizeStr(v.share)
				var newSizeInt, currentSizeInt int
				fmt.Sscanf(newSize, "%d", &newSizeInt)
				fmt.Sscanf(currentSizeStr, "%d", &currentSizeInt)
				if newSizeInt <= currentSizeInt {
					v.extendInput = ""
					return nil
				}
				v.extendMode = false
				v.extendInput = ""
				return func() tea.Msg {
					return ExecuteFileShareActionMsg{
						Share:  v.share,
						Action: FileShareActionExtend,
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
		if v.selectedAction < len(fileShareActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteFileShareActionMsg{
					Share:  v.share,
					Action: v.selectedAction,
				}
			}
		}
		switch v.selectedAction {
		case FileShareActionDelete:
			v.confirmMode = true
		case FileShareActionRename:
			v.renameInput = getString(v.share, "name")
			v.renameMode = true
		case FileShareActionExtend:
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

// Title returns the header title.
func (v *DetailView) Title() string {
	name := getString(v.share, "name")
	return fmt.Sprintf(" 📁 File Storage > %s ", name)
}

// HelpText returns the footer help text.
func (v *DetailView) HelpText() string {
	if v.renameMode || v.extendMode {
			return "Type value • Enter: Confirm • Esc: Cancel"
	}
	if v.confirmMode {
		return "Enter: Confirm action • Esc: Cancel"
	}
	return "←→: Select • Enter: Execute • Esc: Back to list • q: Quit"
}

// ExecuteFileShareActionMsg is dispatched when the user confirms an action.
type ExecuteFileShareActionMsg struct {
	Share  map[string]interface{}
	Action int
	Param  string
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getSizeStr(share map[string]interface{}) string {
	switch v := share["size"].(type) {
	case float64:
		return fmt.Sprintf("%d", int(v))
	case int:
		return fmt.Sprintf("%d", v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return fmt.Sprintf("%d", i)
		}
	}
	return "-"
}
