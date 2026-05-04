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

// Action identifiers for object storage container detail view.
const (
	ContainerActionDelete     = iota 
	ContainerActionChangeType        
	ContainerActionAddPolicy       
)

const (
	subMenuNone       = 0
	subMenuChangeType = 1
	subMenuAddPolicy  = 2 
	subMenuPickRole   = 3 
)

var swiftTypeOptions = []string{"Private", "Public", "Static"}
var policyRoles = []string{"readWrite", "readOnly", "admin", "deny"}

// containerAction pairs an action ID with a button label.
type containerAction struct {
	id    int
	label string
}

// DetailView displays object storage container details with actions.
type DetailView struct {
	views.BaseView
	container      map[string]interface{}
	users          []map[string]interface{} // cloud users
	selectedAction int
	confirmMode    bool
	subMenu        int
	subMenuIdx     int
	policyUserIdx  int
}

// NewDetailView creates a detail view for a container.
func NewDetailView(ctx *views.Context, container map[string]interface{}, users []map[string]interface{}) *DetailView {
	return &DetailView{
		BaseView:  views.NewBaseView(ctx),
		container: container,
		users:     users,
	}
}

func (v *DetailView) category() string {
	if t, ok := v.container["_type"].(string); ok {
		return t
	}
	return ""
}

// getActions returns the available actions for this container.
func (v *DetailView) getActions() []containerAction {
	acts := []containerAction{{id: ContainerActionDelete, label: "Delete"}}
	if v.category() == "Swift" {
		acts = append(acts, containerAction{id: ContainerActionChangeType, label: "Change Type"})
	} else if v.category() == "S3" {
		acts = append(acts, containerAction{id: ContainerActionAddPolicy, label: "Add User"})
	}
	return acts
}

// ─── Render ───────────────────────────────────────────────────────────────────

func (v *DetailView) Render(width, height int) string {
	if v.container == nil {
		return views.StyleError.Render("No container data available")
	}
	var content strings.Builder
	content.WriteString(views.RenderBox("Container information", v.renderInfo(), width-4))
	content.WriteString("\n\n")
	switch v.subMenu {
	case subMenuChangeType:
		content.WriteString(v.renderChangeTypeMenu(width))
	case subMenuAddPolicy:
		content.WriteString(v.renderPickUserMenu(width))
	case subMenuPickRole:
		content.WriteString(v.renderPickRoleMenu(width))
	default:
		content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(), width-4))
	}
	return content.String()
}

func (v *DetailView) renderInfo() string {
	var info strings.Builder
	name := getString(v.container, "name")
	region := getString(v.container, "region")
	createdAt := getString(v.container, "createdAt")
	if createdAt == "" {
		createdAt = "-"
	}

	info.WriteString(views.RenderKeyValue("Name", name) + "\n")
	info.WriteString(views.RenderKeyValue("Region", region) + "\n")
	info.WriteString(views.RenderKeyValue("Created at", createdAt) + "\n")
	info.WriteString(views.RenderKeyValue("API", v.category()) + "\n")

	if v.category() == "Swift" {
		swiftType := getString(v.container, "containerType")
		if swiftType == "" {
			swiftType = "-"
		}
		info.WriteString(views.RenderKeyValue("Type", swiftType) + "\n")
		info.WriteString(views.RenderKeyValue("Objects", getCountStr(v.container, "storedObjects", "objectsCount")) + "\n")
		info.WriteString(views.RenderKeyValue("Total size", getSizeStr(v.container, "storedBytes", "objectsSize")) + "\n")
	} else {
		// Versioning
		versioningStatus := "-"
		if vers, ok := v.container["versioning"].(map[string]interface{}); ok {
			if s, ok := vers["status"].(string); ok {
				versioningStatus = s
			}
		}

		// Encryption
		encryptionStatus := "No encryption"
		encryptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		if enc, ok := v.container["encryption"].(map[string]interface{}); ok {
			if alg, _ := enc["sseAlgorithm"].(string); alg != "" {
				encryptionStatus = "SSE-OMK (OVHcloud-managed keys)"
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

		// Owner: resolve ownerId → username if possible
		ownerDisplay := "-"
		if rawID, ok := v.container["ownerId"]; ok {
			ownerIDStr := fmt.Sprintf("%v", rawID)
			ownerDisplay = ownerIDStr
			for _, u := range v.users {
				if fmt.Sprintf("%v", u["_userId"]) == ownerIDStr {
					if uname := getString(u, "_username"); uname != "" {
						ownerDisplay = uname
					}
					break
				}
			}
		}

		info.WriteString(views.RenderKeyValue("Objects", getCountStr(v.container, "objectsCount", "storedObjects")) + "\n")
		info.WriteString(views.RenderKeyValue("Total size", getSizeStr(v.container, "objectsSize", "storedBytes")) + "\n")
		info.WriteString(views.RenderKeyValue("Owner", ownerDisplay) + "\n")
		info.WriteString(views.RenderKeyValue("Versioning", versioningStatus) + "\n")
		info.WriteString(views.StyleLabel.Render("Encryption:") + " " + encryptionStyle.Render(encryptionStatus) + "\n")
		info.WriteString(views.RenderKeyValue("Object Lock", objectLockStatus) + "\n")
	}
	return info.String()
}

func (v *DetailView) renderActions() string {
	actions := v.getActions()
	var parts []string
	for i, act := range actions {
		var style lipgloss.Style
		if i == v.selectedAction {
			if act.label == "Delete" {
				style = views.StyleButtonDangerSelected
			} else {
				style = views.StyleButtonSelected
			}
		} else if act.label == "Delete" {
			style = views.StyleButtonDanger
		} else {
			style = views.StyleButton
		}
		parts = append(parts, style.Render("["+act.label+"]"))
	}
	result := strings.Join(parts, " ")
	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render("⚠️  Press Enter to confirm deletion, Esc to cancel")
	}
	return result
}

func (v *DetailView) renderChangeTypeMenu(width int) string {
	var content strings.Builder
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	current := getString(v.container, "containerType")
	if current != "" {
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
		content.WriteString(descStyle.Render(fmt.Sprintf("Current type: %s", current)) + "\n\n")
	}
	for i, opt := range swiftTypeOptions {
		if i == v.subMenuIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", opt)) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", opt)) + "\n")
		}
	}
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("↑↓: Select • Enter: Confirm • Esc: Cancel"))
	return views.RenderBox("Change container type", content.String(), width-4)
}

func (v *DetailView) policyUserCandidates() []map[string]interface{} {
	seen := map[string]bool{}
	var result []map[string]interface{}
	for _, u := range v.users {
		uid := fmt.Sprintf("%v", u["_userId"])
		if uid == "" || uid == "0" || uid == "<nil>" {
			continue
		}
		if seen[uid] {
			continue
		}
		seen[uid] = true
		result = append(result, u)
	}
	return result
}

func (v *DetailView) renderPickUserMenu(width int) string {
	var content strings.Builder
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	candidates := v.policyUserCandidates()
	if len(candidates) == 0 {
		content.WriteString(dimStyle.Render("No cloud user available. Please create an S3 user first.") + "\n")
		content.WriteString("\n" + hintStyle.Render("Esc: Cancel"))
		return views.RenderBox("Add user access", content.String(), width-4)
	}
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("Select the user to grant access:") + "\n\n")
	for i, u := range candidates {
		name := fmt.Sprintf("%v", u["_username"])
		if name == "" || name == "<nil>" {
			name = fmt.Sprintf("%v", u["internalName"])
		}
		if i == v.subMenuIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", name)) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %s", name)) + "\n")
		}
	}
	content.WriteString("\n" + hintStyle.Render("↑↓: Select • Enter: Next • Esc: Cancel"))
	return views.RenderBox("Add user access (1/2: User)", content.String(), width-4)
}

func (v *DetailView) renderPickRoleMenu(width int) string {
	var content strings.Builder
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	roleDescriptions := map[string]string{
		"readWrite": "Read + write (recommended)",
		"readOnly":  "Read only",
		"admin":     "Full access (admin)",
		"deny":      "Deny all access",
	}
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("Select the role:") + "\n\n")
	for i, role := range policyRoles {
		desc := roleDescriptions[role]
		if i == v.subMenuIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %-12s  %s", role, desc)) + "\n")
		} else {
			content.WriteString(itemStyle.Render(fmt.Sprintf("    %-12s  %s", role, desc)) + "\n")
		}
	}
	content.WriteString("\n" + hintStyle.Render("↑↓: Select • Enter: Confirm • Esc: Back"))
	return views.RenderBox("Add user access (2/2: Role)", content.String(), width-4)
}

// ─── Key handling ─────────────────────────────────────────────────────────────

func (v *DetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch v.subMenu {
	case subMenuChangeType:
		return v.handleSubMenuKey(key)
	case subMenuAddPolicy:
		return v.handlePickUserKey(key)
	case subMenuPickRole:
		return v.handlePickRoleKey(key)
	}

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
			container := v.container
			return func() tea.Msg {
				return ExecuteContainerActionMsg{Container: container, Action: ContainerActionDelete}
			}
		}
		if v.selectedAction >= len(actions) {
			return nil
		}
		switch actions[v.selectedAction].id {
		case ContainerActionDelete:
			v.confirmMode = true
		case ContainerActionChangeType:
			v.subMenu = subMenuChangeType
			v.subMenuIdx = 0
		case ContainerActionAddPolicy:
			v.subMenu = subMenuAddPolicy
			v.subMenuIdx = 0
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

func (v *DetailView) handlePickUserKey(key string) tea.Cmd {
	candidates := v.policyUserCandidates()
	switch key {
	case "up":
		if v.subMenuIdx > 0 {
			v.subMenuIdx--
		}
	case "down":
		if v.subMenuIdx < len(candidates)-1 {
			v.subMenuIdx++
		}
	case "enter":
		if len(candidates) > 0 {
			v.policyUserIdx = v.subMenuIdx
			v.subMenu = subMenuPickRole
			v.subMenuIdx = 0
		}
	case "esc":
		v.subMenu = subMenuNone
	}
	return nil
}

func (v *DetailView) handlePickRoleKey(key string) tea.Cmd {
	switch key {
	case "up":
		if v.subMenuIdx > 0 {
			v.subMenuIdx--
		}
	case "down":
		if v.subMenuIdx < len(policyRoles)-1 {
			v.subMenuIdx++
		}
	case "enter":
		v.subMenu = subMenuNone
		candidates := v.policyUserCandidates()
		selectedUser := candidates[v.policyUserIdx]
		roleName := policyRoles[v.subMenuIdx]
		container := v.container
		return func() tea.Msg {
			return ExecuteContainerActionMsg{
				Container: container,
				Action:    ContainerActionAddPolicy,
				ExtraData: map[string]interface{}{
					"userId":   selectedUser["_userId"],
					"roleName": roleName,
				},
			}
		}
	case "esc":
		v.subMenu = subMenuAddPolicy
		v.subMenuIdx = v.policyUserIdx
	}
	return nil
}

func (v *DetailView) handleSubMenuKey(key string) tea.Cmd {
	switch key {
	case "up":
		if v.subMenuIdx > 0 {
			v.subMenuIdx--
		}
	case "down":
		if v.subMenuIdx < len(swiftTypeOptions)-1 {
			v.subMenuIdx++
		}
	case "enter":
		v.subMenu = subMenuNone
		newType := strings.ToLower(swiftTypeOptions[v.subMenuIdx])
		container := v.container
		return func() tea.Msg {
			return ExecuteContainerActionMsg{
				Container: container,
				Action:    ContainerActionChangeType,
				ExtraData: map[string]interface{}{"containerType": newType},
			}
		}
	case "esc":
		v.subMenu = subMenuNone
	}
	return nil
}


// ─── Metadata ────────────────────────────────────────────────────────────────

func (v *DetailView) Title() string {
	name := getString(v.container, "name")
	return fmt.Sprintf(" Object Storage > %s ", name)
}

func (v *DetailView) HelpText() string {
	switch v.subMenu {
	case subMenuChangeType:
		return "↑↓: Select • Enter: Confirm • Esc: Cancel"
	case subMenuAddPolicy:
		return "↑↓: Select user • Enter: Next • Esc: Cancel"
	case subMenuPickRole:
		return "↑↓: Select role • Enter: Confirm • Esc: Back"
	}
	if v.confirmMode {
		return "Enter: Confirm deletion • Esc: Cancel"
	}
	return "←→: Select • Enter: Execute • Esc: Back to list • q: Quit"
}

// ExecuteContainerActionMsg is dispatched when the user confirms an action.
type ExecuteContainerActionMsg struct {
	Container map[string]interface{}
	Action    int
	ExtraData map[string]interface{}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getCountStr(m map[string]interface{}, fields ...string) string {
	for _, f := range fields {
		if v, ok := m[f]; ok {
			if n, ok := v.(float64); ok {
				return fmt.Sprintf("%d", int(n))
			}
		}
	}
	return "-"
}

func getSizeStr(m map[string]interface{}, fields ...string) string {
	for _, f := range fields {
		if v, ok := m[f]; ok {
			if n, ok := v.(float64); ok {
				if n < 1024 {
					return fmt.Sprintf("%.0f B", n)
				} else if n < 1024*1024 {
					return fmt.Sprintf("%.1f KB", n/1024)
				} else if n < 1024*1024*1024 {
					return fmt.Sprintf("%.1f MB", n/1024/1024)
				}
				return fmt.Sprintf("%.2f GB", n/1024/1024/1024)
			}
		}
	}
	return "-"
}

