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

// Action identifiers for user detail view.
const (
	UserActionShowSecret = iota
	UserActionDisable
	UserActionEnable
	UserActionDeleteUser
)

// UserDetailView displays S3 user details with activate/deactivate and secret key actions.
type UserDetailView struct {
	views.BaseView
	user           map[string]interface{}
	selectedAction int
	confirmMode    bool
	secretKey      string // populated after Show Secret action
	showSecret     bool
}

// NewUserDetailView creates a detail view for an S3 user.
func NewUserDetailView(ctx *views.Context, user map[string]interface{}) *UserDetailView {
	return &UserDetailView{
		BaseView: views.NewBaseView(ctx),
		user:     user,
	}
}

func (v *UserDetailView) hasCredentials() bool {
	acc := fmt.Sprintf("%v", v.user["access"])
	return acc != "" && acc != "<nil>" && acc != "No credentials"
}

func (v *UserDetailView) getActions() []containerAction {
	var acts []containerAction
	if v.hasCredentials() {
		acts = append(acts, containerAction{id: UserActionShowSecret, label: "Show Secret"})
		acts = append(acts, containerAction{id: UserActionDisable, label: "Disable"})
	} else {
		acts = append(acts, containerAction{id: UserActionEnable, label: "Enable"})
	}
	acts = append(acts, containerAction{id: UserActionDeleteUser, label: "Delete User"})
	return acts
}

// ─── Render ───────────────────────────────────────────────────────────────────

func (v *UserDetailView) Render(width, height int) string {
	if v.user == nil {
		return views.StyleError.Render("No user data available")
	}
	var content strings.Builder
	content.WriteString(views.RenderBox("User information", v.renderInfo(), width-4))
	content.WriteString("\n\n")

	if v.showSecret && v.secretKey != "" {
		content.WriteString(v.renderSecretBox(width))
		content.WriteString("\n\n")
	}

	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(), width-4))
	return content.String()
}

func (v *UserDetailView) renderInfo() string {
	var info strings.Builder

	name := fmt.Sprintf("%v", v.user["_username"])
	if name == "" || name == "<nil>" {
		name = getString(v.user, "internalName")
	}
	description := getString(v.user, "_userDescription")
	if description == "" {
		description = "-"
	}
	userID := fmt.Sprintf("%v", v.user["_userId"])

	statusLabel := "Disabled"
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	if v.hasCredentials() {
		statusLabel = "Active"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	}

	info.WriteString(views.RenderKeyValue("Username", name) + "\n")
	info.WriteString(views.RenderKeyValue("Description", description) + "\n")
	info.WriteString(views.RenderKeyValue("User ID", userID) + "\n")
	info.WriteString(views.StyleLabel.Render("Status:") + " " + statusStyle.Render(statusLabel) + "\n")

	if v.hasCredentials() {
		access := fmt.Sprintf("%v", v.user["access"])
		info.WriteString(views.RenderKeyValue("Access Key", access) + "\n")
	}

	return info.String()
}

func (v *UserDetailView) renderSecretBox(width int) string {
	var content strings.Builder
	secretStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(secretStyle.Render(v.secretKey) + "\n\n")
	content.WriteString(hintStyle.Render("⚠️  Note this key, it will not be shown again."))
	return views.RenderBox("Secret Key", content.String(), width-4)
}

func (v *UserDetailView) renderActions() string {
	actions := v.getActions()
	var parts []string
	for i, act := range actions {
		var style lipgloss.Style
		if i == v.selectedAction {
			switch act.label {
			case "Delete User", "Disable":
				style = views.StyleButtonDangerSelected
			default:
				style = views.StyleButtonSelected
			}
		} else {
			switch act.label {
			case "Delete User", "Disable":
				style = views.StyleButtonDanger
			default:
				style = views.StyleButton
			}
		}
		parts = append(parts, style.Render("["+act.label+"]"))
	}
	result := strings.Join(parts, " ")
	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render("⚠️  Press Enter to confirm, Esc to cancel")
	}
	return result
}

// ─── Key handling ─────────────────────────────────────────────────────────────

func (v *UserDetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	actions := v.getActions()

	switch key {
	case "left":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
	case "right":
		if v.selectedAction < len(actions)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
	case "enter":
		if v.confirmMode {
			v.confirmMode = false
			user := v.user
			action := actions[v.selectedAction].id
			return func() tea.Msg {
				return ExecuteUserActionMsg{User: user, Action: action}
			}
		}
		if v.selectedAction >= len(actions) {
			return nil
		}
		act := actions[v.selectedAction]
		switch act.id {
		case UserActionShowSecret, UserActionEnable:
			user := v.user
			return func() tea.Msg {
				return ExecuteUserActionMsg{User: user, Action: act.id}
			}
		case UserActionDisable, UserActionDeleteUser:
			v.confirmMode = true
		}
	case "esc":
		if v.confirmMode {
			v.confirmMode = false
			return nil
		}
		if v.showSecret {
			v.showSecret = false
			v.secretKey = ""
			return nil
		}
		return func() tea.Msg { return views.GoBackMsg{} }
	}
	return nil
}

// SetSecret populates the revealed secret key and shows it.
func (v *UserDetailView) SetSecret(secret string) {
	v.secretKey = secret
	v.showSecret = true
}

// ─── Metadata ────────────────────────────────────────────────────────────────

func (v *UserDetailView) Title() string {
	name := fmt.Sprintf("%v", v.user["_username"])
	if name == "" || name == "<nil>" {
		name = getString(v.user, "internalName")
	}
	return fmt.Sprintf(" 👤 Object Storage > Users > %s ", name)
}

func (v *UserDetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm • Esc: Cancel"
	}
	if v.showSecret {
		return "Esc: Close secret key"
	}
	return "←→: Select • Enter: Execute • Esc: Back to list • q: Quit"
}

// ExecuteUserActionMsg is dispatched when the user confirms an action.
type ExecuteUserActionMsg struct {
	User   map[string]interface{}
	Action int
}
