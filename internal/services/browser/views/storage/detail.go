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

// Action indices for storage items
const (
	ActionViewObjects = iota
	ActionEdit
	ActionDelete
)

// S3 specific actions
const (
	S3ActionViewObjects = iota
	S3ActionPresignedURL
	S3ActionEdit
	S3ActionDelete
)

// Block storage specific actions
const (
	BlockActionAttach = iota
	BlockActionDetach
	BlockActionResize
	BlockActionSnapshot
	BlockActionDelete
)

// DetailView displays storage item details with actions
type DetailView struct {
	views.BaseView
	item           map[string]interface{}
	storageType    StorageType
	selectedAction int
	confirmMode    bool
}

// NewDetailView creates a new storage detail view
func NewDetailView(ctx *views.Context, item map[string]interface{}, storageType StorageType) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		item:           item,
		storageType:    storageType,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *DetailView) Render(width, height int) string {
	if v.item == nil {
		return views.StyleError.Render("No storage data available")
	}

	switch v.storageType {
	case StorageTypeS3:
		return v.renderS3Detail(width)
	case StorageTypeSwift:
		return v.renderSwiftDetail(width)
	case StorageTypeBlock:
		return v.renderBlockDetail(width)
	default:
		return v.renderGenericDetail(width)
	}
}

func (v *DetailView) renderS3Detail(width int) string {
	var content strings.Builder

	name := getString(v.item, "name")
	region := getString(v.item, "region")
	objectsCount := getInt(v.item, "objectsCount")
	objectsSize := getInt(v.item, "objectsSize")
	created := getString(v.item, "createdAt")
	virtualHost := getString(v.item, "virtualHost")

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("Name", name) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Created", created) + "\n")
	if virtualHost != "" {
		infoContent.WriteString(views.RenderKeyValue("Virtual Host", virtualHost) + "\n")
	}
	content.WriteString(views.RenderBox("Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Storage stats box
	var statsContent strings.Builder
	statsContent.WriteString(views.RenderKeyValue("Objects Count", fmt.Sprintf("%d", objectsCount)) + "\n")
	statsContent.WriteString(views.RenderKeyValue("Total Size", formatSize(int64(objectsSize))) + "\n")

	// Versioning status
	if versioning, ok := v.item["versioning"].(map[string]interface{}); ok {
		status := getString(versioning, "status")
		if status != "" {
			statsContent.WriteString(views.RenderKeyValue("Versioning", views.RenderStatus(strings.ToUpper(status))) + "\n")
		}
	}

	// Encryption
	if encryption, ok := v.item["encryption"].(map[string]interface{}); ok {
		algo := getString(encryption, "sseAlgorithm")
		if algo != "" {
			statsContent.WriteString(views.RenderKeyValue("Encryption", algo) + "\n")
		}
	}

	content.WriteString(views.RenderBox("Storage", statsContent.String(), width-4))
	content.WriteString("\n\n")

	// Actions
	actions := []string{"View Objects", "Presigned URL", "Edit", "Delete"}
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(actions), width-4))

	return content.String()
}

func (v *DetailView) renderSwiftDetail(width int) string {
	var content strings.Builder

	name := getString(v.item, "name")
	id := getString(v.item, "id")
	region := getString(v.item, "region")
	containerType := getString(v.item, "containerType")
	if containerType == "" {
		containerType = "private"
	}
	storedObjects := getInt(v.item, "storedObjects")
	storedBytes := getInt(v.item, "storedBytes")
	staticURL := getString(v.item, "staticUrl")
	archive := false
	if a, ok := v.item["archive"].(bool); ok {
		archive = a
	}

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("Name", name) + "\n")
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Type", containerType) + "\n")
	if archive {
		infoContent.WriteString(views.RenderKeyValue("Archive", "Yes (Cold Storage)") + "\n")
	}
	if staticURL != "" {
		infoContent.WriteString(views.RenderKeyValue("Static URL", staticURL) + "\n")
	}
	content.WriteString(views.RenderBox("Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Storage stats box
	var statsContent strings.Builder
	statsContent.WriteString(views.RenderKeyValue("Objects Count", fmt.Sprintf("%d", storedObjects)) + "\n")
	statsContent.WriteString(views.RenderKeyValue("Total Size", formatSize(int64(storedBytes))) + "\n")
	content.WriteString(views.RenderBox("Storage", statsContent.String(), width-4))
	content.WriteString("\n\n")

	// Actions
	actions := []string{"View Objects", "Edit", "Delete"}
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(actions), width-4))

	return content.String()
}

func (v *DetailView) renderBlockDetail(width int) string {
	var content strings.Builder

	name := getString(v.item, "name")
	id := getString(v.item, "id")
	region := getString(v.item, "region")
	size := getInt(v.item, "size")
	volType := getString(v.item, "type")
	status := getString(v.item, "status")
	description := getString(v.item, "description")
	created := getString(v.item, "createdAt")
	bootable := false
	if b, ok := v.item["bootable"].(bool); ok {
		bootable = b
	}

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("Name", name) + "\n")
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(strings.ToUpper(status))) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Size", fmt.Sprintf("%d GB", size)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Type", volType) + "\n")
	if bootable {
		infoContent.WriteString(views.RenderKeyValue("Bootable", "Yes") + "\n")
	}
	if description != "" {
		infoContent.WriteString(views.RenderKeyValue("Description", description) + "\n")
	}
	if created != "" {
		infoContent.WriteString(views.RenderKeyValue("Created", created) + "\n")
	}
	content.WriteString(views.RenderBox("Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Attachment box
	var attachContent strings.Builder
	if attachedTo, ok := v.item["attachedTo"].([]interface{}); ok && len(attachedTo) > 0 {
		attachContent.WriteString(views.RenderKeyValue("Attached To", fmt.Sprintf("%d instance(s)", len(attachedTo))) + "\n")
		for i, inst := range attachedTo {
			if instID, ok := inst.(string); ok {
				attachContent.WriteString(views.RenderKeyValue(fmt.Sprintf("  Instance %d", i+1), instID) + "\n")
			}
		}
	} else {
		attachContent.WriteString(views.StyleMuted.Render("Not attached to any instance") + "\n")
	}
	content.WriteString(views.RenderBox("Attachment", attachContent.String(), width-4))
	content.WriteString("\n\n")

	// Actions
	actions := []string{"Attach", "Detach", "Resize", "Snapshot", "Delete"}
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", v.renderActions(actions), width-4))

	return content.String()
}

func (v *DetailView) renderGenericDetail(width int) string {
	var content strings.Builder
	content.WriteString(views.StyleSubheader.Render("Storage Item Details") + "\n\n")

	for key, value := range v.item {
		content.WriteString(views.RenderKeyValue(key, fmt.Sprintf("%v", value)) + "\n")
	}

	return content.String()
}

func (v *DetailView) renderActions(actions []string) string {
	var parts []string

	for i, label := range actions {
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
		actionName := actions[v.selectedAction]
		result += "\n\n" + views.StyleStatusWarning.Render(
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actionName))
	}

	return result
}

func (v *DetailView) getActionCount() int {
	switch v.storageType {
	case StorageTypeS3:
		return 4 // View Objects, Presigned URL, Edit, Delete
	case StorageTypeSwift:
		return 3 // View Objects, Edit, Delete
	case StorageTypeBlock:
		return 5 // Attach, Detach, Resize, Snapshot, Delete
	default:
		return 3
	}
}

func (v *DetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	actionCount := v.getActionCount()

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
			return func() tea.Msg {
				return ExecuteStorageActionMsg{
					Item:        v.item,
					StorageType: v.storageType,
					Action:      v.selectedAction,
				}
			}
		}
		// Ask for confirmation for destructive actions
		if v.isDestructiveAction() {
			v.confirmMode = true
			return nil
		}
		// Non-destructive actions execute immediately
		return func() tea.Msg {
			return ExecuteStorageActionMsg{
				Item:        v.item,
				StorageType: v.storageType,
				Action:      v.selectedAction,
			}
		}
	case "esc":
		if v.confirmMode {
			v.confirmMode = false
			return nil
		}
		// Go back to table
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *DetailView) isDestructiveAction() bool {
	switch v.storageType {
	case StorageTypeS3:
		return v.selectedAction == S3ActionDelete
	case StorageTypeSwift:
		return v.selectedAction == ActionDelete
	case StorageTypeBlock:
		return v.selectedAction == BlockActionDelete || v.selectedAction == BlockActionDetach
	default:
		return v.selectedAction == ActionDelete
	}
}

func (v *DetailView) Update(msg tea.Msg) (tea.Cmd, views.View) {
	return nil, nil
}

func (v *DetailView) Title() string {
	name := getString(v.item, "name")
	return fmt.Sprintf(" %s %s > %s ", v.storageType.Icon(), v.storageType.String(), name)
}

func (v *DetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • d: Debug • Esc: Back to List • q: Quit"
}
