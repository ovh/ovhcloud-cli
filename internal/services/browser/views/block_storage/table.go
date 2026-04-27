// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package block_storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

type TableView struct {
	views.BaseView
	table        table.Model
	data         []map[string]interface{}
	filterMode   bool
	filterInput  string
	filteredData []map[string]interface{}
}

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
		{Title: "Name", Width: 28},
		{Title: "Status", Width: 12},
		{Title: "Size (GB)", Width: 10},
		{Title: "Type", Width: 18},
		{Title: "Region", Width: 12},
	}

	var rows []table.Row
	for _, volume := range v.filteredData {
		name := getString(volume, "name")
		status := getString(volume, "status")
		size := getSizeStr(volume)
		vType := getString(volume, "type")
		region := getString(volume, "region")

		rows = append(rows, table.Row{name, status, size, vType, region})
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
	s.Header = views.StyleTableHeader
	s.Selected = views.StyleTableSelected
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
				return ShowVolumeDetailMsg{Volume: v.filteredData[idx]}
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
	return " 💾 Block Storage "
}

func (v *TableView) HelpText() string {
	if v.filterMode {
		return "Type to filter • Enter: Confirm • Esc: Cancel"
	}
	return "↑↓: Navigate • /: Filter • Enter: Details • d: Debug • p: Projects • q: Quit"
}

func (v *TableView) GetSelectedVolume() map[string]interface{} {
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

type ShowVolumeDetailMsg struct {
	Volume map[string]interface{}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getSizeStr(volume map[string]interface{}) string {
	switch v := volume["size"].(type) {
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
