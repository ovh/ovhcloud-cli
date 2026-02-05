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

// ObjectsTableView displays a list of objects within a storage container
type ObjectsTableView struct {
	views.BaseView
	table        table.Model
	data         []map[string]interface{}
	filteredData []map[string]interface{}
	container    map[string]interface{}
	storageType  StorageType
	filterMode   bool
	filterInput  string
	isLoading    bool
}

// NewObjectsTableView creates a new objects table view
func NewObjectsTableView(ctx *views.Context, container map[string]interface{}, objects []map[string]interface{}, storageType StorageType) *ObjectsTableView {
	v := &ObjectsTableView{
		BaseView:     views.NewBaseView(ctx),
		data:         objects,
		filteredData: objects,
		container:    container,
		storageType:  storageType,
		isLoading:    objects == nil,
	}
	v.table = v.createTable()
	return v
}

// createTable builds the bubbles table from the data
func (v *ObjectsTableView) createTable() table.Model {
	columns := []table.Column{
		{Title: "Key", Width: 40},
		{Title: "Size", Width: 12},
		{Title: "Storage Class", Width: 15},
		{Title: "Last Modified", Width: 20},
	}

	var rows []table.Row
	for _, item := range v.filteredData {
		key := getString(item, "key")
		size := getInt(item, "size")
		storageClass := getString(item, "storageClass")
		if storageClass == "" {
			storageClass = "STANDARD"
		}
		lastModified := getString(item, "lastModified")
		if len(lastModified) > 19 {
			lastModified = lastModified[:19]
		}

		// Check for restore status
		restoreStatus := ""
		if rs, ok := item["restoreStatus"].(map[string]interface{}); ok {
			if status := getString(rs, "status"); status != "" {
				restoreStatus = " [" + status + "]"
			}
		}

		rows = append(rows, table.Row{
			key,
			formatSize(int64(size)),
			storageClass + restoreStatus,
			lastModified,
		})
	}

	ctx := v.Context()
	height := ctx.Height - 18
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

func (v *ObjectsTableView) Render(width, height int) string {
	var content strings.Builder

	// Container info header
	containerName := getString(v.container, "name")
	region := getString(v.container, "region")
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(views.ColorPrimary)
	content.WriteString(headerStyle.Render(fmt.Sprintf("📦 %s (%s)", containerName, region)))
	content.WriteString("\n\n")

	// Filter input
	if v.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(views.ColorPrimary)
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s▌", v.filterInput)) + "\n\n")
	} else if v.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(views.ColorDimmed)
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s (press / to edit)", v.filterInput)) + "\n\n")
	}

	// Loading state
	if v.isLoading {
		loadingStyle := lipgloss.NewStyle().Foreground(views.ColorMuted)
		content.WriteString(loadingStyle.Render("Loading objects..."))
		return content.String()
	}

	// Table or empty state
	if len(v.filteredData) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(views.ColorMuted)
		if v.filterInput != "" {
			content.WriteString(emptyStyle.Render("No objects match filter: " + v.filterInput))
		} else {
			content.WriteString(emptyStyle.Render("No objects in this container."))
		}
	} else {
		content.WriteString(v.table.View())
	}

	return content.String()
}

func (v *ObjectsTableView) HandleKey(msg tea.KeyMsg) tea.Cmd {
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
		// Return selected object for detail view
		idx := v.table.Cursor()
		if idx >= 0 && idx < len(v.filteredData) {
			return func() tea.Msg {
				return ShowStorageObjectDetailMsg{
					Object:      v.filteredData[idx],
					Container:   v.container,
					StorageType: v.storageType,
				}
			}
		}

	case "up", "down", "j", "k":
		var cmd tea.Cmd
		v.table, cmd = v.table.Update(msg)
		return cmd

	case "esc", "backspace":
		// Go back to container detail
		if v.filterInput != "" {
			v.filterInput = ""
			v.applyFilter()
			return nil
		}
		// Return to container detail view
		return func() tea.Msg {
			return ShowStorageDetailMsg{
				Item:        v.container,
				StorageType: v.storageType,
			}
		}

	case "r":
		// Refresh objects list
		return func() tea.Msg {
			return ShowStorageObjectsMsg{
				Container:   v.container,
				StorageType: v.storageType,
			}
		}
	}
	return nil
}

func (v *ObjectsTableView) applyFilter() {
	if v.filterInput == "" {
		v.filteredData = v.data
	} else {
		filter := strings.ToLower(v.filterInput)
		v.filteredData = nil
		for _, item := range v.data {
			key := strings.ToLower(getString(item, "key"))
			storageClass := strings.ToLower(getString(item, "storageClass"))
			if strings.Contains(key, filter) || strings.Contains(storageClass, filter) {
				v.filteredData = append(v.filteredData, item)
			}
		}
	}
	v.table = v.createTable()
}

func (v *ObjectsTableView) Update(msg tea.Msg) (tea.Cmd, views.View) {
	switch m := msg.(type) {
	case StorageObjectsLoadedMsg:
		if m.Err == nil && v.container != nil {
			containerName := getString(v.container, "name")
			msgContainerName := getString(m.Container, "name")
			if containerName == msgContainerName {
				v.data = m.Objects
				v.isLoading = false
				v.applyFilter()
			}
		}
	}
	return nil, nil
}

func (v *ObjectsTableView) Title() string {
	containerName := getString(v.container, "name")
	return fmt.Sprintf(" 📦 %s > Objects ", containerName)
}

func (v *ObjectsTableView) HelpText() string {
	return "Enter: Details • /: Filter • r: Refresh • Esc: Back • q: Quit"
}

// UpdateData updates the table with new object data
func (v *ObjectsTableView) UpdateData(objects []map[string]interface{}) {
	v.data = objects
	v.isLoading = false
	v.applyFilter()
}

// Container returns the container this view is for
func (v *ObjectsTableView) Container() map[string]interface{} {
	return v.container
}

// StorageType returns the storage type
func (v *ObjectsTableView) StorageType() StorageType {
	return v.storageType
}
