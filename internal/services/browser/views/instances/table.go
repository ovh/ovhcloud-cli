// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package instances

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// TableView displays a list of instances in a table.
type TableView struct {
	views.BaseView
	table         table.Model
	data          []map[string]interface{}
	imageMap      map[string]string // imageId -> imageName
	floatingIPMap map[string]string // instanceId -> floatingIP
	filterMode    bool
	filterInput   string
	filteredData  []map[string]interface{}
}

// NewTableView creates a new instance table view.
func NewTableView(ctx *views.Context, data []map[string]interface{}, imageMap, floatingIPMap map[string]string) *TableView {
	v := &TableView{
		BaseView:      views.NewBaseView(ctx),
		data:          data,
		imageMap:      imageMap,
		floatingIPMap: floatingIPMap,
		filteredData:  data,
	}
	v.table = v.createTable()
	return v
}

// createTable builds the bubbles table from the data.
func (v *TableView) createTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Status", Width: 12},
		{Title: "Flavor", Width: 15},
		{Title: "Image", Width: 20},
		{Title: "Region", Width: 12},
		{Title: "IP Address", Width: 16},
	}

	var rows []table.Row
	for _, instance := range v.filteredData {
		name := getString(instance, "name")
		status := getString(instance, "status")
		region := getString(instance, "region")
		flavorID := getString(instance, "flavorId")
		imageID := getString(instance, "imageId")
		id := getString(instance, "id")

		// Get image name from map
		imageName := imageID
		if v.imageMap != nil {
			if name, ok := v.imageMap[imageID]; ok {
				imageName = name
			}
		}

		// Get floating IP if available
		ip := getFirstIP(instance)
		if v.floatingIPMap != nil {
			if fip, ok := v.floatingIPMap[id]; ok && fip != "" {
				ip = fip + " (floating)"
			}
		}

		rows = append(rows, table.Row{name, status, flavorID, imageName, region, ip})
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

	// Filter input
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

	// Filter mode handling
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
		// Return selected instance for detail view
		idx := v.table.Cursor()
		if idx >= 0 && idx < len(v.filteredData) {
			return func() tea.Msg {
				return ShowInstanceDetailMsg{Instance: v.filteredData[idx]}
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
	return " 🖥️  Instances "
}

func (v *TableView) HelpText() string {
	if v.filterMode {
		return "Type to filter • Enter: Confirm • Esc: Cancel"
	}
	return "↑↓: Navigate • /: Filter • Enter: Details • c: Create • d: Debug • p: Projects • q: Quit"
}

// GetSelectedInstance returns the currently selected instance.
func (v *TableView) GetSelectedInstance() map[string]interface{} {
	idx := v.table.Cursor()
	if idx >= 0 && idx < len(v.filteredData) {
		return v.filteredData[idx]
	}
	return nil
}

// UpdateData updates the table with new data.
func (v *TableView) UpdateData(data []map[string]interface{}, imageMap, floatingIPMap map[string]string) {
	cursor := v.table.Cursor()
	v.data = data
	v.imageMap = imageMap
	v.floatingIPMap = floatingIPMap
	v.applyFilter()
	if cursor >= 0 && cursor < len(v.filteredData) {
		v.table.SetCursor(cursor)
	}
}

// ShowInstanceDetailMsg signals to show instance detail.
type ShowInstanceDetailMsg struct {
	Instance map[string]interface{}
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFirstIP(instance map[string]interface{}) string {
	if addresses, ok := instance["ipAddresses"].([]interface{}); ok {
		for _, addr := range addresses {
			if addrMap, ok := addr.(map[string]interface{}); ok {
				if ip, ok := addrMap["ip"].(string); ok {
					return ip
				}
			}
		}
	}
	return ""
}
