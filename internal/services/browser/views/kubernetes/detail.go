// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package kubernetes

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// Action indices for cluster detail view
const (
	ActionKubeconfig = iota
	ActionK9s
	ActionPools
	ActionUpgrade
	ActionPolicy
	ActionDelete
)

var clusterActionLabels = []string{"Kubeconfig", "K9s", "Pools", "Upgrade", "Policy", "Delete"}

// DetailView displays Kubernetes cluster details with actions.
type DetailView struct {
	views.BaseView
	cluster        map[string]interface{}
	nodePools      []map[string]interface{}
	selectedAction int
	confirmMode    bool
}

// NewDetailView creates a new cluster detail view.
func NewDetailView(ctx *views.Context, cluster map[string]interface{}, nodePools []map[string]interface{}) *DetailView {
	return &DetailView{
		BaseView:       views.NewBaseView(ctx),
		cluster:        cluster,
		nodePools:      nodePools,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *DetailView) Render(width, height int) string {
	var content strings.Builder

	if v.cluster == nil {
		return views.StyleError.Render("No cluster data available")
	}

	// Extract cluster data
	status := getString(v.cluster, "status")
	id := getString(v.cluster, "id")
	region := getString(v.cluster, "region")
	version := getString(v.cluster, "version")
	updatePolicy := getString(v.cluster, "updatePolicy")
	createdAt := getString(v.cluster, "createdAt")
	url := getString(v.cluster, "url")

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Region", region) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Version", version) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Update Policy", updatePolicy) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Created", createdAt) + "\n")
	if url != "" {
		infoContent.WriteString(views.RenderKeyValue("API URL", url) + "\n")
	}
	content.WriteString(views.RenderBox("Cluster Information", infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Node pools box
	var poolsContent strings.Builder
	if len(v.nodePools) > 0 {
		for _, pool := range v.nodePools {
			poolName := getString(pool, "name")
			poolStatus := getString(pool, "status")
			flavor := getString(pool, "flavor")

			// Get node counts
			desired := getIntValue(pool, "desiredNodes")
			current := getIntValue(pool, "currentNodes")

			poolsContent.WriteString(fmt.Sprintf("  • %s (%s) - %s - %d/%d nodes\n",
				poolName, flavor, views.RenderStatus(poolStatus), current, desired))
		}
	} else {
		poolsContent.WriteString("  No node pools configured\n")
	}
	content.WriteString(views.RenderBox(fmt.Sprintf("Node Pools (%d)", len(v.nodePools)), poolsContent.String(), width-4))
	content.WriteString("\n\n")

	// Actions box
	actionsContent := v.renderActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4))

	return content.String()
}

func (v *DetailView) renderActions() string {
	var parts []string

	for i, label := range clusterActionLabels {
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
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", clusterActionLabels[v.selectedAction]))
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
		if v.selectedAction < len(clusterActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		// Pools action doesn't need confirmation
		if v.selectedAction == ActionPools {
			return func() tea.Msg {
				return ExecuteClusterActionMsg{
					Cluster: v.cluster,
					Action:  v.selectedAction,
				}
			}
		}
		if v.confirmMode {
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteClusterActionMsg{
					Cluster: v.cluster,
					Action:  v.selectedAction,
				}
			}
		}
		v.confirmMode = true
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
	name := getString(v.cluster, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s ", name)
}

func (v *DetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • d: Debug • Esc: Back to List • q: Quit"
}

// UpdateNodePools updates the node pools list.
func (v *DetailView) UpdateNodePools(nodePools []map[string]interface{}) {
	v.nodePools = nodePools
}

// ExecuteClusterActionMsg signals to execute an action on a cluster.
type ExecuteClusterActionMsg struct {
	Cluster map[string]interface{}
	Action  int
}

func getIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case json.Number:
			i, _ := val.Int64()
			return int(i)
		}
	}
	return 0
}
