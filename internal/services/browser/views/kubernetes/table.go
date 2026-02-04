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
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// TableView displays a list of Kubernetes clusters in a table.
type TableView struct {
	views.BaseView
	table        table.Model
	data         []map[string]interface{}
	filterMode   bool
	filterInput  string
	filteredData []map[string]interface{}
}

// NewTableView creates a new Kubernetes table view.
func NewTableView(ctx *views.Context, data []map[string]interface{}) *TableView {
	v := &TableView{
		BaseView:     views.NewBaseView(ctx),
		data:         data,
		filteredData: data,
	}
	v.table = v.createTable()
	return v
}

func (v *TableView) createTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 12},
		{Title: "Version", Width: 10},
		{Title: "Region", Width: 12},
		{Title: "Nodes", Width: 8},
	}

	var rows []table.Row
	for _, cluster := range v.filteredData {
		name := getString(cluster, "name")
		status := getString(cluster, "status")
		version := getString(cluster, "version")
		region := getString(cluster, "region")
		nodes := getNodeCount(cluster)

		rows = append(rows, table.Row{name, status, version, region, nodes})
	}

	ctx := v.Context()
	height := ctx.Height - 15
	if height < 5 {
		height = 5
	}
	if height > 20 {
		height = 20
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func (v *TableView) Render(width, height int) string {
	var content strings.Builder

	if v.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s▌", v.filterInput)) + "\n\n")
	} else if v.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s (press / to edit)", v.filterInput)) + "\n\n")
	}

	content.WriteString(v.table.View())

	return content.String()
}

func (v *TableView) HandleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	if v.filterMode {
		switch msg.Type {
		case tea.KeyEscape:
			v.filterMode = false
			return nil
		case tea.KeyEnter:
			v.filterMode = false
			v.applyFilter()
			return nil
		case tea.KeyBackspace:
			if len(v.filterInput) > 0 {
				v.filterInput = v.filterInput[:len(v.filterInput)-1]
			}
			v.applyFilter()
			return nil
		case tea.KeyRunes:
			v.filterInput += string(msg.Runes)
			v.applyFilter()
			return nil
		}
		return nil
	}

	switch key {
	case "/":
		v.filterMode = true
		return nil
	case "enter":
		idx := v.table.Cursor()
		if idx >= 0 && idx < len(v.filteredData) {
			return func() tea.Msg {
				return ShowClusterDetailMsg{Cluster: v.filteredData[idx]}
			}
		}
	case "up", "down", "j", "k":
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return cmd
	case "esc":
		if v.filterInput != "" {
			v.filterInput = ""
			v.applyFilter()
			return nil
		}
	}
	return nil
}

func (v *TableView) applyFilter() {
	if v.filterInput == "" {
		v.filteredData = v.data
	} else {
		filter := strings.ToLower(v.filterInput)
		v.filteredData = nil
		for _, item := range v.data {
			name := strings.ToLower(getString(item, "name"))
			if strings.Contains(name, filter) {
				v.filteredData = append(v.filteredData, item)
			}
		}
	}
	v.table = v.createTable()
}

func (v *TableView) Title() string {
	return " ☸️  Kubernetes "
}

func (v *TableView) HelpText() string {
	if v.filterMode {
		return "Type to filter • Enter: Confirm • Esc: Cancel"
	}
	return "↑↓: Navigate • /: Filter • Enter: Details • c: Create • d: Debug • p: Projects • q: Quit"
}

// GetSelectedCluster returns the currently selected cluster.
func (v *TableView) GetSelectedCluster() map[string]interface{} {
	idx := v.table.Cursor()
	if idx >= 0 && idx < len(v.filteredData) {
		return v.filteredData[idx]
	}
	return nil
}

// UpdateData updates the table with new data.
func (v *TableView) UpdateData(data []map[string]interface{}) {
	cursor := v.table.Cursor()
	v.data = data
	v.applyFilter()
	if cursor >= 0 && cursor < len(v.filteredData) {
		v.table.SetCursor(cursor)
	}
}

// ShowClusterDetailMsg signals to show cluster detail.
type ShowClusterDetailMsg struct {
	Cluster map[string]interface{}
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getNodeCount(cluster map[string]interface{}) string {
	// Try different possible fields
	if nodes, ok := cluster["nodesCount"]; ok {
		return fmt.Sprintf("%v", nodes)
	}
	return "-"
}
