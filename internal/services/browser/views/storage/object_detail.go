// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package storage

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// Object action indices
const (
	ObjectActionIdxCopy = iota
	ObjectActionIdxRestore
	ObjectActionIdxPresignedURL
	ObjectActionIdxDelete
)

// ObjectDetailView displays object details with actions
type ObjectDetailView struct {
	views.BaseView
	object         map[string]interface{}
	container      map[string]interface{}
	storageType    StorageType
	selectedAction int
	confirmMode    bool
}

// NewObjectDetailView creates a new object detail view
func NewObjectDetailView(ctx *views.Context, object map[string]interface{}, container map[string]interface{}, storageType StorageType) *ObjectDetailView {
	return &ObjectDetailView{
		BaseView:       views.NewBaseView(ctx),
		object:         object,
		container:      container,
		storageType:    storageType,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *ObjectDetailView) Render(width, height int) string {
	if v.object == nil {
		return views.StyleError.Render("No object data available")
	}

	var content strings.Builder

	key := getString(v.object, "key")
	size := getInt(v.object, "size")
	etag := getString(v.object, "etag")
	lastModified := getString(v.object, "lastModified")
	storageClass := getString(v.object, "storageClass")
	if storageClass == "" {
		storageClass = "STANDARD"
	}
	versionId := getString(v.object, "versionId")
	isLatest := false
	if l, ok := v.object["isLatest"].(bool); ok {
		isLatest = l
	}
	containerName := getString(v.container, "name")
	region := getString(v.container, "region")

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("Key", key) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Container", containerName) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Size", formatSize(int64(size))) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Storage Class", v.renderStorageClass(storageClass)) + "\n")
	if etag != "" {
		infoContent.WriteString(views.RenderKeyValue("ETag", etag) + "\n")
	}
	if lastModified != "" {
		infoContent.WriteString(views.RenderKeyValue("Last Modified", lastModified) + "\n")
	}
	content.WriteString(views.RenderBox("Object Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Version information (if versioning is enabled)
	if versionId != "" {
		var versionContent strings.Builder
		versionContent.WriteString(views.RenderKeyValue("Version ID", versionId) + "\n")
		if isLatest {
			versionContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus("LATEST")) + "\n")
		} else {
			versionContent.WriteString(views.RenderKeyValue("Status", views.StyleMuted.Render("Previous Version")) + "\n")
		}
		content.WriteString(views.RenderBox("Version", versionContent.String(), width-4))
		content.WriteString("\n\n")
	}

	// Restore status (for archived objects)
	if restoreStatus, ok := v.object["restoreStatus"].(map[string]interface{}); ok {
		var restoreContent strings.Builder
		status := getString(restoreStatus, "status")
		restoreContent.WriteString(views.RenderKeyValue("Restore Status", v.renderRestoreStatus(status)) + "\n")
		if expiryDate := getString(restoreStatus, "expiryDate"); expiryDate != "" {
			restoreContent.WriteString(views.RenderKeyValue("Restore Expiry", expiryDate) + "\n")
		}
		content.WriteString(views.RenderBox("Archive Restore", restoreContent.String(), width-4))
		content.WriteString("\n\n")
	}

	// Lock information (if object lock is enabled)
	if lock, ok := v.object["lock"].(map[string]interface{}); ok {
		var lockContent strings.Builder
		mode := getString(lock, "mode")
		retainUntil := getString(lock, "retainUntil")
		if mode != "" {
			lockContent.WriteString(views.RenderKeyValue("Lock Mode", strings.ToUpper(mode)) + "\n")
		}
		if retainUntil != "" {
			lockContent.WriteString(views.RenderKeyValue("Retain Until", retainUntil) + "\n")
		}
		content.WriteString(views.RenderBox("Object Lock", lockContent.String(), width-4))
		content.WriteString("\n\n")
	}

	// Legal hold
	if legalHold := getString(v.object, "legalHold"); legalHold != "" {
		var legalContent strings.Builder
		legalContent.WriteString(views.RenderKeyValue("Legal Hold", strings.ToUpper(legalHold)) + "\n")
		content.WriteString(views.RenderBox("Legal Hold", legalContent.String(), width-4))
		content.WriteString("\n\n")
	}

	// Actions - adjust based on storage class
	actions := v.getAvailableActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(actions), width-4))

	return content.String()
}

func (v *ObjectDetailView) renderStorageClass(class string) string {
	switch class {
	case "HIGH_PERF":
		return views.StyleButtonSuccess.Render("⚡ HIGH_PERF")
	case "STANDARD":
		return views.StyleValue.Render("📦 STANDARD")
	case "STANDARD_IA":
		return views.StyleStatusWarning.Render("🧊 STANDARD_IA (Archive)")
	default:
		return class
	}
}

func (v *ObjectDetailView) renderRestoreStatus(status string) string {
	switch status {
	case "ongoing":
		return views.StyleStatusWarning.Render("🔄 Restore in progress...")
	case "completed":
		return views.StyleButtonSuccess.Render("✅ Restored (temporary)")
	case "not-requested":
		return views.StyleMuted.Render("Not restored")
	default:
		return status
	}
}

func (v *ObjectDetailView) getAvailableActions() []string {
	storageClass := getString(v.object, "storageClass")

	// For archived objects, show restore option
	if storageClass == "STANDARD_IA" {
		// Check if already restoring or restored
		if restoreStatus, ok := v.object["restoreStatus"].(map[string]interface{}); ok {
			status := getString(restoreStatus, "status")
			if status == "ongoing" {
				return []string{"Copy", "Presigned URL", "Delete"} // Hide restore while in progress
			}
		}
		return []string{"Copy", "Restore", "Presigned URL", "Delete"}
	}

	return []string{"Copy", "Presigned URL", "Delete"}
}

func (v *ObjectDetailView) renderActions(actions []string) string {
	var parts []string

	for i, label := range actions {
		var style lipgloss.Style
		if i == v.selectedAction {
			style = views.StyleButtonSelected
		} else if label == "Delete" {
			style = views.StyleButtonDanger
		} else if label == "Restore" {
			style = views.StyleStatusWarning
		} else {
			style = views.StyleButton
		}
		parts = append(parts, style.Render("["+label+"]"))
	}

	result := strings.Join(parts, " ")

	if v.confirmMode {
		actionName := actions[v.selectedAction]
		result += "\n\n" + views.StyleStatusWarning.Render(
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actionName))
	}

	return result
}

func (v *ObjectDetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	actions := v.getAvailableActions()
	actionCount := len(actions)

	switch key {
	case "left":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
		return nil
	case "right":
		if v.selectedAction < actionCount-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		if v.confirmMode {
			// Execute the action
			v.confirmMode = false
			return v.executeAction(actions[v.selectedAction])
		}
		// Ask for confirmation for destructive actions
		if v.isDestructiveAction(actions[v.selectedAction]) {
			v.confirmMode = true
			return nil
		}
		// Non-destructive actions execute immediately
		return v.executeAction(actions[v.selectedAction])

	case "esc", "backspace":
		if v.confirmMode {
			v.confirmMode = false
			return nil
		}
		// Go back to objects list
		return func() tea.Msg {
			return ShowStorageObjectsMsg{
				Container:   v.container,
				StorageType: v.storageType,
			}
		}
	}
	return nil
}

func (v *ObjectDetailView) executeAction(action string) tea.Cmd {
	var actionType StorageObjectActionType
	switch action {
	case "Copy":
		actionType = ObjectActionCopy
	case "Restore":
		actionType = ObjectActionRestore
	case "Delete":
		actionType = ObjectActionDelete
	case "Presigned URL":
		actionType = ObjectActionPresignedURL
	default:
		return nil
	}

	return func() tea.Msg {
		return ExecuteStorageObjectActionMsg{
			Object:      v.object,
			Container:   v.container,
			StorageType: v.storageType,
			Action:      actionType,
		}
	}
}

func (v *ObjectDetailView) isDestructiveAction(action string) bool {
	return action == "Delete" || action == "Restore"
}

func (v *ObjectDetailView) Update(msg tea.Msg) (tea.Cmd, views.View) {
	switch m := msg.(type) {
	case StorageObjectsCopyMsg:
		// Handle copy result
		if m.Err != nil {
			// Could show notification
		}
	case StorageObjectsRestoreMsg:
		// Handle restore result
		if m.Err != nil {
			// Could show notification
		}
	}
	return nil, nil
}

func (v *ObjectDetailView) Title() string {
	key := getString(v.object, "key")
	containerName := getString(v.container, "name")
	// Truncate long keys
	if len(key) > 30 {
		key = key[:27] + "..."
	}
	return fmt.Sprintf(" 📦 %s > %s ", containerName, key)
}

func (v *ObjectDetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • Esc: Back to Objects • q: Quit"
}

// Object returns the current object
func (v *ObjectDetailView) Object() map[string]interface{} {
	return v.object
}

// Container returns the parent container
func (v *ObjectDetailView) Container() map[string]interface{} {
	return v.container
}

// StorageType returns the storage type
func (v *ObjectDetailView) StorageType() StorageType {
	return v.storageType
}
