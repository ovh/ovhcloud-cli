// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package instances

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// Action indices for the detail view
const (
	ActionStart = iota
	ActionStop
	ActionSoftReboot
	ActionReboot
	ActionSSH
	ActionDelete
)

var actionLabels = []string{"Start", "Stop", "Soft Reboot", "Reboot", "SSH", "Delete"}

// DetailView displays instance details with actions.
type DetailView struct {
	views.BaseView
	instance       map[string]interface{}
	imageName      string
	floatingIP     string
	selectedAction int
	confirmMode    bool
}

// NewDetailView creates a new instance detail view.
func NewDetailView(ctx *views.Context, instance map[string]interface{}, imageName, floatingIP string) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		instance:       instance,
		imageName:      imageName,
		floatingIP:     floatingIP,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *DetailView) Render(width, height int) string {
	var content strings.Builder

	if v.instance == nil {
		return views.StyleError.Render("No instance data available")
	}

	// Extract instance data
	status := getString(v.instance, "status")
	id := getString(v.instance, "id")
	region := getString(v.instance, "region")
	created := getString(v.instance, "created")
	flavorName := v.getFlavorName()
	imageName := v.imageName
	if imageName == "" {
		imageName = getString(v.instance, "imageId")
	}

	// Get IP addresses
	var publicIPs, privateIPs []string
	if addresses, ok := v.instance["ipAddresses"].([]interface{}); ok {
		for _, addr := range addresses {
			if addrMap, ok := addr.(map[string]interface{}); ok {
				ip := getString(addrMap, "ip")
				version := getString(addrMap, "version")
				ipType := getString(addrMap, "type")
				if ip != "" && version == "4" {
					if ipType == "public" {
						publicIPs = append(publicIPs, ip)
					} else {
						privateIPs = append(privateIPs, ip)
					}
				}
			}
		}
	}

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Flavor", flavorName) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Image", imageName) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Created", created) + "\n")
	content.WriteString(views.RenderBox("Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Network box
	var netContent strings.Builder
	if len(publicIPs) > 0 {
		netContent.WriteString(views.RenderKeyValue("Public IP", strings.Join(publicIPs, ", ")) + "\n")
	}
	if len(privateIPs) > 0 {
		netContent.WriteString(views.RenderKeyValue("Private IP", strings.Join(privateIPs, ", ")) + "\n")
	}
	if v.floatingIP != "" {
		netContent.WriteString(views.RenderKeyValue("Floating IP", v.floatingIP) + "\n")
	}
	if netContent.Len() > 0 {
		content.WriteString(views.RenderBox("Network", netContent.String(), width-4))
		content.WriteString("\n\n")
	}

	// Actions box
	actionsContent := v.renderActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4))

	return content.String()
}

func (v *DetailView) renderActions() string {
	var parts []string

	for i, label := range actionLabels {
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
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actionLabels[v.selectedAction]))
	}

	return result
}

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
		if v.selectedAction < len(actionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		if v.confirmMode {
			// Execute the action
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteInstanceActionMsg{
					Instance: v.instance,
					Action:   v.selectedAction,
				}
			}
		}
		// Ask for confirmation
		v.confirmMode = true
		return nil
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

func (v *DetailView) Title() string {
	name := getString(v.instance, "name")
	return fmt.Sprintf(" 🖥️  Instances > %s ", name)
}

func (v *DetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • d: Debug • Esc: Back to List • q: Quit"
}

func (v *DetailView) getFlavorName() string {
	if flavor, ok := v.instance["flavor"].(map[string]interface{}); ok {
		if name, ok := flavor["name"].(string); ok {
			return name
		}
	}
	return getString(v.instance, "flavorId")
}

// ExecuteInstanceActionMsg signals to execute an action on an instance.
type ExecuteInstanceActionMsg struct {
	Instance map[string]interface{}
	Action   int
}
