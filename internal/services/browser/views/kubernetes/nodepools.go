// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package kubernetes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// NodePoolsView displays a list of node pools for a Kubernetes cluster.
type NodePoolsView struct {
	views.BaseView
	table       table.Model
	cluster     map[string]interface{}
	nodePools   []map[string]interface{}
	filterInput string
	filtering   bool
}

// NewNodePoolsView creates a new node pools list view.
func NewNodePoolsView(ctx *views.Context, cluster map[string]interface{}, nodePools []map[string]interface{}) *NodePoolsView {
	v := &NodePoolsView{
		BaseView:  views.NewBaseView(ctx),
		cluster:   cluster,
		nodePools: nodePools,
	}
	v.createTable()
	return v
}

func (v *NodePoolsView) createTable() {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Status", Width: 12},
		{Title: "Flavor", Width: 15},
		{Title: "Nodes", Width: 10},
		{Title: "Range", Width: 12},
		{Title: "Autoscale", Width: 10},
	}

	rows := make([]table.Row, 0, len(v.nodePools))
	for _, pool := range v.nodePools {
		name := getString(pool, "name")
		status := getString(pool, "status")
		flavor := getString(pool, "flavor")
		current := getIntValue(pool, "currentNodes")
		desired := getIntValue(pool, "desiredNodes")
		minNodes := getIntValue(pool, "minNodes")
		maxNodes := getIntValue(pool, "maxNodes")
		autoscale := getBool(pool, "autoscale")

		nodesStr := fmt.Sprintf("%d/%d", current, desired)
		rangeStr := fmt.Sprintf("%d-%d", minNodes, maxNodes)
		autoscaleStr := "No"
		if autoscale {
			autoscaleStr = "Yes"
		}

		rows = append(rows, table.Row{name, status, flavor, nodesStr, rangeStr, autoscaleStr})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = views.StyleTableHeader
	s.Selected = views.StyleTableSelected
	t.SetStyles(s)

	v.table = t
}

func (v *NodePoolsView) Render(width, height int) string {
	var content strings.Builder

	// Show filter if active
	if v.filtering {
		content.WriteString(views.StyleFilter.Render(" Filter: " + v.filterInput + "█ "))
		content.WriteString("\n\n")
	}

	content.WriteString(v.table.View())

	if len(v.nodePools) == 0 {
		content.WriteString("\n\n")
		content.WriteString(views.StyleSubtle.Render("  No node pools found. Press 'n' to create a new node pool."))
	}

	return content.String()
}

func (v *NodePoolsView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	// Handle filter mode
	if v.filtering {
		switch key {
		case "esc":
			v.filtering = false
			v.filterInput = ""
			v.applyFilter()
			return nil
		case "enter":
			v.filtering = false
			return nil
		case "backspace":
			if len(v.filterInput) > 0 {
				v.filterInput = v.filterInput[:len(v.filterInput)-1]
				v.applyFilter()
			}
			return nil
		default:
			if len(key) == 1 {
				v.filterInput += key
				v.applyFilter()
			}
			return nil
		}
	}

	switch key {
	case "enter":
		row := v.table.SelectedRow()
		if len(row) > 0 {
			poolName := row[0]
			return func() tea.Msg {
				return ShowNodePoolDetailMsg{
					Cluster:  v.cluster,
					PoolName: poolName,
				}
			}
		}
	case "/":
		v.filtering = true
		v.filterInput = ""
		return nil
	case "n":
		return func() tea.Msg {
			return CreateNodePoolMsg{Cluster: v.cluster}
		}
	case "esc":
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	case "r":
		return func() tea.Msg {
			return RefreshNodePoolsMsg{Cluster: v.cluster}
		}
	case "up", "k":
		v.table.MoveUp(1)
		return nil
	case "down", "j":
		v.table.MoveDown(1)
		return nil
	}

	return nil
}

func (v *NodePoolsView) applyFilter() {
	if v.filterInput == "" {
		v.createTable()
		return
	}

	filter := strings.ToLower(v.filterInput)
	filtered := make([]map[string]interface{}, 0)
	for _, pool := range v.nodePools {
		name := strings.ToLower(getString(pool, "name"))
		flavor := strings.ToLower(getString(pool, "flavor"))
		if strings.Contains(name, filter) || strings.Contains(flavor, filter) {
			filtered = append(filtered, pool)
		}
	}

	// Temporarily replace and rebuild
	original := v.nodePools
	v.nodePools = filtered
	v.createTable()
	v.nodePools = original
}

func (v *NodePoolsView) Title() string {
	clusterName := getString(v.cluster, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s > Node Pools ", clusterName)
}

func (v *NodePoolsView) HelpText() string {
	if v.filtering {
		return "Type to filter • Enter: Apply • Esc: Clear"
	}
	return "↑↓: Navigate • Enter: View Details • n: New Pool • r: Refresh • /: Filter • Esc: Back"
}

// ShowNodePoolDetailMsg signals to show node pool details.
type ShowNodePoolDetailMsg struct {
	Cluster  map[string]interface{}
	PoolName string
}

// CreateNodePoolMsg signals to create a new node pool.
type CreateNodePoolMsg struct {
	Cluster map[string]interface{}
}

// RefreshNodePoolsMsg signals to refresh the node pools list.
type RefreshNodePoolsMsg struct {
	Cluster map[string]interface{}
}

// NodePoolDetailView displays details for a single node pool.
type NodePoolDetailView struct {
	views.BaseView
	cluster        map[string]interface{}
	nodePool       map[string]interface{}
	selectedAction int
	confirmMode    bool
}

// Action indices for node pool detail view
const (
	NodePoolActionScale = iota
	NodePoolActionDelete
)

var nodePoolActionLabels = []string{"Scale", "Delete"}

// NewNodePoolDetailView creates a new node pool detail view.
func NewNodePoolDetailView(ctx *views.Context, cluster, nodePool map[string]interface{}) *NodePoolDetailView {
	return &NodePoolDetailView{
		BaseView:       views.NewBaseView(ctx),
		cluster:        cluster,
		nodePool:       nodePool,
		selectedAction: 0,
		confirmMode:    false,
	}
}

func (v *NodePoolDetailView) Render(width, height int) string {
	var content strings.Builder

	if v.nodePool == nil {
		return views.StyleError.Render("No node pool data available")
	}

	name := getString(v.nodePool, "name")
	status := getString(v.nodePool, "status")
	id := getString(v.nodePool, "id")
	flavor := getString(v.nodePool, "flavor")
	current := getIntValue(v.nodePool, "currentNodes")
	desired := getIntValue(v.nodePool, "desiredNodes")
	minNodes := getIntValue(v.nodePool, "minNodes")
	maxNodes := getIntValue(v.nodePool, "maxNodes")
	autoscale := getBool(v.nodePool, "autoscale")
	monthlyBilled := getBool(v.nodePool, "monthlyBilled")
	antiAffinity := getBool(v.nodePool, "antiAffinity")
	createdAt := getString(v.nodePool, "createdAt")

	// Information box
	var infoContent strings.Builder
	infoContent.WriteString(views.RenderKeyValue("ID", id) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Status", views.RenderStatus(status)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Flavor", flavor) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Current Nodes", fmt.Sprintf("%d", current)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Desired Nodes", fmt.Sprintf("%d", desired)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Min/Max", fmt.Sprintf("%d / %d", minNodes, maxNodes)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Autoscale", boolToStr(autoscale)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Monthly Billed", boolToStr(monthlyBilled)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Anti-Affinity", boolToStr(antiAffinity)) + "\n")
	infoContent.WriteString(views.RenderKeyValue("Created", createdAt) + "\n")
	content.WriteString(views.RenderBox(fmt.Sprintf("Node Pool: %s", name), infoContent.String(), width-4))
	content.WriteString("\n\n")

	// Actions
	actionsContent := v.renderActions()
	content.WriteString(views.RenderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4))

	return content.String()
}

func (v *NodePoolDetailView) renderActions() string {
	var parts []string

	for i, label := range nodePoolActionLabels {
		var style = views.StyleButton
		if i == v.selectedAction {
			style = views.StyleButtonSelected
		} else if label == "Delete" {
			style = views.StyleButtonDanger
		}
		parts = append(parts, style.Render("["+label+"]"))
	}

	result := strings.Join(parts, " ")

	if v.confirmMode {
		result += "\n\n" + views.StyleStatusWarning.Render(
			fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", nodePoolActionLabels[v.selectedAction]))
	}

	return result
}

func (v *NodePoolDetailView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "left":
		if v.selectedAction > 0 {
			v.selectedAction--
			v.confirmMode = false
		}
		return nil
	case "right":
		if v.selectedAction < len(nodePoolActionLabels)-1 {
			v.selectedAction++
			v.confirmMode = false
		}
		return nil
	case "enter":
		// Scale action goes directly to scale view
		if v.selectedAction == NodePoolActionScale {
			return func() tea.Msg {
				return ExecuteNodePoolActionMsg{
					Cluster:  v.cluster,
					NodePool: v.nodePool,
					Action:   v.selectedAction,
				}
			}
		}
		if v.confirmMode {
			v.confirmMode = false
			return func() tea.Msg {
				return ExecuteNodePoolActionMsg{
					Cluster:  v.cluster,
					NodePool: v.nodePool,
					Action:   v.selectedAction,
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

func (v *NodePoolDetailView) Title() string {
	clusterName := getString(v.cluster, "name")
	poolName := getString(v.nodePool, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s > %s ", clusterName, poolName)
}

func (v *NodePoolDetailView) HelpText() string {
	if v.confirmMode {
		return "Enter: Confirm Action • Esc: Cancel"
	}
	return "←→: Select Action • Enter: Execute • d: Debug • Esc: Back • q: Quit"
}

// ExecuteNodePoolActionMsg signals to execute an action on a node pool.
type ExecuteNodePoolActionMsg struct {
	Cluster  map[string]interface{}
	NodePool map[string]interface{}
	Action   int
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func boolToStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// NodePoolScaleView displays the scale form for a node pool.
type NodePoolScaleView struct {
	views.BaseView
	cluster       map[string]interface{}
	nodePool      map[string]interface{}
	fields        []ScaleField
	selectedField int
	errorMsg      string
}

// ScaleField represents an editable field in the scale view.
type ScaleField struct {
	Label string
	Value int
	Min   int
	Max   int
}

// NewNodePoolScaleView creates a new node pool scale view.
func NewNodePoolScaleView(ctx *views.Context, cluster, nodePool map[string]interface{}) *NodePoolScaleView {
	desired := getIntValue(nodePool, "desiredNodes")
	minNodes := getIntValue(nodePool, "minNodes")
	maxNodes := getIntValue(nodePool, "maxNodes")

	return &NodePoolScaleView{
		BaseView:      views.NewBaseView(ctx),
		cluster:       cluster,
		nodePool:      nodePool,
		selectedField: 0,
		fields: []ScaleField{
			{Label: "Desired Nodes", Value: desired, Min: 0, Max: 100},
			{Label: "Min Nodes", Value: minNodes, Min: 0, Max: 100},
			{Label: "Max Nodes", Value: maxNodes, Min: 0, Max: 100},
		},
	}
}

func (v *NodePoolScaleView) Render(width, height int) string {
	var content strings.Builder

	poolName := getString(v.nodePool, "name")

	content.WriteString(views.StyleInfo.Render(fmt.Sprintf("Scaling node pool: %s", poolName)))
	content.WriteString("\n\n")

	for i, field := range v.fields {
		prefix := "  "
		if i == v.selectedField {
			prefix = "▶ "
		}

		valueStyle := views.StyleSubtle
		if i == v.selectedField {
			valueStyle = views.StyleHighlight
		}

		content.WriteString(fmt.Sprintf("%s%s: %s\n",
			prefix,
			field.Label,
			valueStyle.Render(fmt.Sprintf("◀ %d ▶", field.Value))))
	}

	content.WriteString("\n")
	content.WriteString(views.StyleSubtle.Render("Use ↑/↓ to select field, ←/→ to adjust value"))
	content.WriteString("\n")
	content.WriteString(views.StyleSubtle.Render("Press Enter to apply, Escape to cancel"))

	if v.errorMsg != "" {
		content.WriteString("\n\n")
		content.WriteString(views.StyleError.Render("⚠️  " + v.errorMsg))
	}

	return content.String()
}

func (v *NodePoolScaleView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "up", "k":
		if v.selectedField > 0 {
			v.selectedField--
		}
		return nil
	case "down", "j":
		if v.selectedField < len(v.fields)-1 {
			v.selectedField++
		}
		return nil
	case "left", "h":
		if v.fields[v.selectedField].Value > v.fields[v.selectedField].Min {
			v.fields[v.selectedField].Value--
		}
		return nil
	case "right", "l":
		if v.fields[v.selectedField].Value < v.fields[v.selectedField].Max {
			v.fields[v.selectedField].Value++
		}
		return nil
	case "enter":
		// Validate
		desired := v.fields[0].Value
		minNodes := v.fields[1].Value
		maxNodes := v.fields[2].Value

		if minNodes > maxNodes {
			v.errorMsg = "Min nodes cannot be greater than max nodes"
			return nil
		}
		if desired < minNodes || desired > maxNodes {
			v.errorMsg = "Desired nodes must be between min and max"
			return nil
		}

		return func() tea.Msg {
			return SubmitNodePoolScaleMsg{
				Cluster:      v.cluster,
				NodePool:     v.nodePool,
				DesiredNodes: desired,
				MinNodes:     minNodes,
				MaxNodes:     maxNodes,
			}
		}
	case "esc":
		if v.errorMsg != "" {
			v.errorMsg = ""
			return nil
		}
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *NodePoolScaleView) Title() string {
	clusterName := getString(v.cluster, "name")
	poolName := getString(v.nodePool, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s > %s > Scale ", clusterName, poolName)
}

func (v *NodePoolScaleView) HelpText() string {
	return "↑↓: Select Field • ←→: Adjust Value • Enter: Apply • Esc: Cancel"
}

// SubmitNodePoolScaleMsg signals to submit the node pool scale.
type SubmitNodePoolScaleMsg struct {
	Cluster      map[string]interface{}
	NodePool     map[string]interface{}
	DesiredNodes int
	MinNodes     int
	MaxNodes     int
}

// UpdatePolicyView displays the update policy selection.
type UpdatePolicyView struct {
	views.BaseView
	cluster        map[string]interface{}
	policies       []string
	selectedPolicy int
}

var updatePolicies = []string{"ALWAYS_UPDATE", "MINIMAL_DOWNTIME", "NEVER_UPDATE"}

// NewUpdatePolicyView creates a new update policy view.
func NewUpdatePolicyView(ctx *views.Context, cluster map[string]interface{}) *UpdatePolicyView {
	currentPolicy := getString(cluster, "updatePolicy")
	selected := 0
	for i, p := range updatePolicies {
		if p == currentPolicy {
			selected = i
			break
		}
	}

	return &UpdatePolicyView{
		BaseView:       views.NewBaseView(ctx),
		cluster:        cluster,
		policies:       updatePolicies,
		selectedPolicy: selected,
	}
}

func (v *UpdatePolicyView) Render(width, height int) string {
	var content strings.Builder

	clusterName := getString(v.cluster, "name")
	content.WriteString(views.StyleInfo.Render(fmt.Sprintf("Update Policy for: %s", clusterName)))
	content.WriteString("\n\n")

	for i, policy := range v.policies {
		prefix := "  "
		if i == v.selectedPolicy {
			prefix = "▶ "
		}

		style := views.StyleSubtle
		if i == v.selectedPolicy {
			style = views.StyleHighlight
		}

		description := ""
		switch policy {
		case "ALWAYS_UPDATE":
			description = "Automatically update to latest patch version"
		case "MINIMAL_DOWNTIME":
			description = "Update with minimal service disruption"
		case "NEVER_UPDATE":
			description = "Manual updates only"
		}

		content.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(policy)))
		content.WriteString(fmt.Sprintf("    %s\n\n", views.StyleSubtle.Render(description)))
	}

	return content.String()
}

func (v *UpdatePolicyView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "up", "k":
		if v.selectedPolicy > 0 {
			v.selectedPolicy--
		}
		return nil
	case "down", "j":
		if v.selectedPolicy < len(v.policies)-1 {
			v.selectedPolicy++
		}
		return nil
	case "enter":
		return func() tea.Msg {
			return SubmitUpdatePolicyMsg{
				Cluster: v.cluster,
				Policy:  v.policies[v.selectedPolicy],
			}
		}
	case "esc":
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *UpdatePolicyView) Title() string {
	clusterName := getString(v.cluster, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s > Update Policy ", clusterName)
}

func (v *UpdatePolicyView) HelpText() string {
	return "↑↓: Select Policy • Enter: Apply • Esc: Cancel"
}

// SubmitUpdatePolicyMsg signals to submit the update policy change.
type SubmitUpdatePolicyMsg struct {
	Cluster map[string]interface{}
	Policy  string
}

// UpgradeView displays the version upgrade selection.
type UpgradeView struct {
	views.BaseView
	cluster         map[string]interface{}
	versions        []string
	selectedVersion int
}

// NewUpgradeView creates a new upgrade view.
func NewUpgradeView(ctx *views.Context, cluster map[string]interface{}, versions []string) *UpgradeView {
	return &UpgradeView{
		BaseView:        views.NewBaseView(ctx),
		cluster:         cluster,
		versions:        versions,
		selectedVersion: 0,
	}
}

func (v *UpgradeView) Render(width, height int) string {
	var content strings.Builder

	clusterName := getString(v.cluster, "name")
	currentVersion := getString(v.cluster, "version")

	content.WriteString(views.StyleInfo.Render(fmt.Sprintf("Upgrade cluster: %s", clusterName)))
	content.WriteString("\n")
	content.WriteString(views.StyleSubtle.Render(fmt.Sprintf("Current version: %s", currentVersion)))
	content.WriteString("\n\n")

	if len(v.versions) == 0 {
		content.WriteString(views.StyleStatusReady.Render("✓ Cluster is already on the latest version"))
	} else {
		content.WriteString("Available versions:\n\n")
		for i, ver := range v.versions {
			prefix := "  "
			if i == v.selectedVersion {
				prefix = "▶ "
			}

			style := views.StyleSubtle
			if i == v.selectedVersion {
				style = views.StyleHighlight
			}

			content.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(ver)))
		}
	}

	return content.String()
}

func (v *UpgradeView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if len(v.versions) == 0 {
		if key == "esc" || key == "enter" {
			return func() tea.Msg {
				return views.GoBackMsg{}
			}
		}
		return nil
	}

	switch key {
	case "up", "k":
		if v.selectedVersion > 0 {
			v.selectedVersion--
		}
		return nil
	case "down", "j":
		if v.selectedVersion < len(v.versions)-1 {
			v.selectedVersion++
		}
		return nil
	case "enter":
		return func() tea.Msg {
			return SubmitUpgradeMsg{
				Cluster: v.cluster,
				Version: v.versions[v.selectedVersion],
			}
		}
	case "esc":
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *UpgradeView) Title() string {
	clusterName := getString(v.cluster, "name")
	return fmt.Sprintf(" ☸️  Kubernetes > %s > Upgrade ", clusterName)
}

func (v *UpgradeView) HelpText() string {
	if len(v.versions) == 0 {
		return "Enter/Esc: Back"
	}
	return "↑↓: Select Version • Enter: Upgrade • Esc: Cancel"
}

// SubmitUpgradeMsg signals to submit the cluster upgrade.
type SubmitUpgradeMsg struct {
	Cluster map[string]interface{}
	Version string
}

// DeleteConfirmView displays a delete confirmation dialog.
type DeleteConfirmView struct {
	views.BaseView
	cluster   map[string]interface{}
	nodePool  map[string]interface{} // nil for cluster delete
	confirmed bool
}

// NewDeleteConfirmView creates a new delete confirmation view.
func NewDeleteConfirmView(ctx *views.Context, cluster, nodePool map[string]interface{}) *DeleteConfirmView {
	return &DeleteConfirmView{
		BaseView:  views.NewBaseView(ctx),
		cluster:   cluster,
		nodePool:  nodePool,
		confirmed: false,
	}
}

func (v *DeleteConfirmView) Render(width, height int) string {
	var content strings.Builder

	var target, name string
	if v.nodePool != nil {
		target = "node pool"
		name = getString(v.nodePool, "name")
	} else {
		target = "cluster"
		name = getString(v.cluster, "name")
	}

	content.WriteString(views.StyleStatusError.Render("⚠️  DELETE CONFIRMATION"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("You are about to delete the %s:\n\n", target))
	content.WriteString(views.StyleHighlight.Render(fmt.Sprintf("  %s", name)))
	content.WriteString("\n\n")
	content.WriteString(views.StyleStatusWarning.Render("This action cannot be undone!"))
	content.WriteString("\n\n")

	if v.confirmed {
		content.WriteString(views.StyleButtonSelected.Render("[Confirm Delete]"))
		content.WriteString(" ")
		content.WriteString(views.StyleButton.Render("[Cancel]"))
	} else {
		content.WriteString(views.StyleButton.Render("[Confirm Delete]"))
		content.WriteString(" ")
		content.WriteString(views.StyleButtonSelected.Render("[Cancel]"))
	}

	return content.String()
}

func (v *DeleteConfirmView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "left", "right", "tab":
		v.confirmed = !v.confirmed
		return nil
	case "enter":
		if v.confirmed {
			return func() tea.Msg {
				return ConfirmDeleteMsg{
					Cluster:  v.cluster,
					NodePool: v.nodePool,
				}
			}
		}
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	case "esc":
		return func() tea.Msg {
			return views.GoBackMsg{}
		}
	}
	return nil
}

func (v *DeleteConfirmView) Title() string {
	if v.nodePool != nil {
		return " ⚠️  Delete Node Pool "
	}
	return " ⚠️  Delete Cluster "
}

func (v *DeleteConfirmView) HelpText() string {
	return "←→: Toggle Selection • Enter: Confirm • Esc: Cancel"
}

// ConfirmDeleteMsg signals deletion was confirmed.
type ConfirmDeleteMsg struct {
	Cluster  map[string]interface{}
	NodePool map[string]interface{}
}
