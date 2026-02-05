// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package storage

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
)

// TableView displays a list of storage items with type tabs
type TableView struct {
	views.BaseView
	table        table.Model
	data         []map[string]interface{}
	storageType  StorageType
	filterMode   bool
	filterInput  string
	filteredData []map[string]interface{}
}

// NewTableView creates a new storage table view
func NewTableView(ctx *views.Context, data []map[string]interface{}, storageType StorageType) *TableView {
	v := &TableView{
		BaseView:     views.NewBaseView(ctx),
		data:         data,
		storageType:  storageType,
		filteredData: data,
	}
	v.table = v.createTable()
	return v
}

// createTable builds the bubbles table from the data
func (v *TableView) createTable() table.Model {
	var columns []table.Column
	var rows []table.Row

	switch v.storageType {
	case StorageTypeS3:
		columns = []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Region", Width: 12},
			{Title: "Objects", Width: 10},
			{Title: "Size", Width: 15},
			{Title: "Created", Width: 20},
		}
		for _, item := range v.filteredData {
			name := getString(item, "name")
			region := getString(item, "region")
			objectsCount := getInt(item, "objectsCount")
			objectsSize := getInt(item, "objectsSize")
			created := getString(item, "createdAt")
			if len(created) > 19 {
				created = created[:19]
			}
			rows = append(rows, table.Row{
				name,
				region,
				fmt.Sprintf("%d", objectsCount),
				formatSize(int64(objectsSize)),
				created,
			})
		}

	case StorageTypeSwift:
		columns = []table.Column{
			{Title: "Name", Width: 30},
			{Title: "Region", Width: 12},
			{Title: "Type", Width: 12},
			{Title: "Objects", Width: 10},
			{Title: "Size", Width: 15},
		}
		for _, item := range v.filteredData {
			name := getString(item, "name")
			region := getString(item, "region")
			containerType := getString(item, "containerType")
			if containerType == "" {
				containerType = "private"
			}
			storedObjects := getInt(item, "storedObjects")
			storedBytes := getInt(item, "storedBytes")
			rows = append(rows, table.Row{
				name,
				region,
				containerType,
				fmt.Sprintf("%d", storedObjects),
				formatSize(int64(storedBytes)),
			})
		}

	case StorageTypeBlock:
		columns = []table.Column{
			{Title: "Name", Width: 25},
			{Title: "Region", Width: 12},
			{Title: "Size (GB)", Width: 10},
			{Title: "Type", Width: 15},
			{Title: "Status", Width: 12},
			{Title: "Attached To", Width: 20},
		}
		for _, item := range v.filteredData {
			name := getString(item, "name")
			region := getString(item, "region")
			size := getInt(item, "size")
			volType := getString(item, "type")
			status := getString(item, "status")
			attachedTo := v.getAttachedInstance(item)
			rows = append(rows, table.Row{
				name,
				region,
				fmt.Sprintf("%d", size),
				volType,
				status,
				attachedTo,
			})
		}
	}

	ctx := v.Context()
	height := ctx.Height - 18 // Account for tabs, header, footer
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

func (v *TableView) getAttachedInstance(volume map[string]interface{}) string {
	if attachedTo, ok := volume["attachedTo"].([]interface{}); ok && len(attachedTo) > 0 {
		if first, ok := attachedTo[0].(string); ok {
			// Truncate long instance IDs
			if len(first) > 18 {
				return first[:15] + "..."
			}
			return first
		}
	}
	return "-"
}

func (v *TableView) Render(width, height int) string {
	var content strings.Builder

	// Storage type tabs
	content.WriteString(v.renderTabs(width))
	content.WriteString("\n\n")

	// Filter input
	if v.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(views.ColorPrimary)
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s▌", v.filterInput)) + "\n\n")
	} else if v.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(views.ColorDimmed)
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s (press / to edit)", v.filterInput)) + "\n\n")
	}

	// Table
	if len(v.filteredData) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(views.ColorMuted)
		if v.filterInput != "" {
			content.WriteString(emptyStyle.Render("No results match filter: " + v.filterInput))
		} else {
			content.WriteString(emptyStyle.Render(fmt.Sprintf("No %s containers found. Press 'c' to create one.", v.storageType.String())))
		}
	} else {
		content.WriteString(v.table.View())
	}

	return content.String()
}

func (v *TableView) renderTabs(width int) string {
	types := []StorageType{StorageTypeS3, StorageTypeSwift, StorageTypeBlock}
	var tabs []string

	for _, t := range types {
		label := fmt.Sprintf(" %s %s ", t.Icon(), t.String())
		if t == v.storageType {
			style := lipgloss.NewStyle().
				Background(views.ColorPrimary).
				Foreground(views.ColorWhite).
				Bold(true).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		} else {
			style := lipgloss.NewStyle().
				Foreground(views.ColorMuted).
				Padding(0, 1)
			tabs = append(tabs, style.Render(label))
		}
	}

	tabBar := strings.Join(tabs, " ")
	hint := lipgloss.NewStyle().Foreground(views.ColorDimmed).Render("  (Tab/1-2-3 to switch)")

	return tabBar + hint
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

	case "tab":
		// Cycle through storage types
		nextType := (v.storageType + 1) % 3
		return func() tea.Msg {
			return SwitchStorageTypeMsg{StorageType: nextType}
		}

	case "1":
		if v.storageType != StorageTypeS3 {
			return func() tea.Msg {
				return SwitchStorageTypeMsg{StorageType: StorageTypeS3}
			}
		}
	case "2":
		if v.storageType != StorageTypeSwift {
			return func() tea.Msg {
				return SwitchStorageTypeMsg{StorageType: StorageTypeSwift}
			}
		}
	case "3":
		if v.storageType != StorageTypeBlock {
			return func() tea.Msg {
				return SwitchStorageTypeMsg{StorageType: StorageTypeBlock}
			}
		}

	case "enter":
		// Return selected item for detail view
		idx := v.table.Cursor()
		if idx >= 0 && idx < len(v.filteredData) {
			return func() tea.Msg {
				return ShowStorageDetailMsg{
					Item:        v.filteredData[idx],
					StorageType: v.storageType,
				}
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
			region := strings.ToLower(getString(item, "region"))
			if strings.Contains(name, filter) || strings.Contains(region, filter) {
				v.filteredData = append(v.filteredData, item)
			}
		}
	}
	v.table = v.createTable()
}

func (v *TableView) Update(msg tea.Msg) (tea.Cmd, views.View) {
	return nil, nil
}

func (v *TableView) Title() string {
	return fmt.Sprintf(" 💾 Storage > %s ", v.storageType.String())
}

func (v *TableView) HelpText() string {
	return "Tab/1-2-3: Switch Type • Enter: Details • /: Filter • c: Create • d: Debug • q: Quit"
}

// StorageType returns the current storage type
func (v *TableView) StorageType() StorageType {
	return v.storageType
}

// UpdateData updates the table with new data
func (v *TableView) UpdateData(data []map[string]interface{}) {
	v.data = data
	v.applyFilter()
}

// formatSize formats bytes into human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
