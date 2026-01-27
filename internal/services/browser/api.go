// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// fetchDataForPath initiates an API call based on the path
func (m Model) fetchDataForPath(path string) tea.Cmd {
	switch path {
	case "/projects":
		return m.fetchProjects
	case "/instances":
		return m.fetchInstances
	case "/kubernetes":
		return m.fetchKubernetes
	case "/databases":
		return m.fetchDatabases
	case "/storage/s3":
		return m.fetchS3Storage
	case "/storage/swift":
		return m.fetchSwiftStorage
	case "/storage/block":
		return m.fetchBlockStorage
	case "/networks/private":
		return m.fetchPrivateNetworks
	case "/networks/public":
		return m.fetchPublicNetworks
	case "/networks/loadbalancer":
		return m.fetchLoadBalancers
	default:
		return nil
	}
}

// fetchProjects fetches the list of cloud projects
func (m Model) fetchProjects() tea.Msg {
	// First, get the list of project IDs (the API returns an array of strings)
	var projectIDs []string
	err := httpLib.Client.Get("/v1/cloud/project", &projectIDs)
	if err != nil {
		return projectsLoadedMsg{
			projects: nil,
			err:      err,
		}
	}

	// Now fetch details for each project
	var projects []map[string]interface{}
	for _, id := range projectIDs {
		var project map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s", id)
		if err := httpLib.Client.Get(endpoint, &project); err == nil {
			projects = append(projects, project)
		}
	}

	return projectsLoadedMsg{
		projects: projects,
		err:      nil,
	}
}

// fetchInstances fetches the list of instances
func (m Model) fetchInstances() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected. Please configure a default project"),
		}
	}

	var instances []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &instances)

	return dataLoadedMsg{
		data: instances,
		err:  err,
	}
}

// fetchKubernetes fetches the list of Kubernetes clusters
func (m Model) fetchKubernetes() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// First, get the list of cluster IDs (the API returns an array of strings)
	var clusterIDs []string
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &clusterIDs)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Now fetch details for each cluster
	var clusters []map[string]interface{}
	for _, id := range clusterIDs {
		var cluster map[string]interface{}
		detailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s", m.cloudProject, id)
		if err := httpLib.Client.Get(detailEndpoint, &cluster); err == nil {
			clusters = append(clusters, cluster)
		}
	}

	return dataLoadedMsg{
		data: clusters,
		err:  nil,
	}
}

// fetchDatabases fetches the list of database services
func (m Model) fetchDatabases() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// First, get the list of database service IDs (the API returns an array of strings)
	var serviceIDs []string
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/service", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &serviceIDs)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Now fetch details for each database service
	var databases []map[string]interface{}
	for _, id := range serviceIDs {
		var db map[string]interface{}
		detailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", m.cloudProject, id)
		if err := httpLib.Client.Get(detailEndpoint, &db); err == nil {
			databases = append(databases, db)
		}
	}

	return dataLoadedMsg{
		data: databases,
		err:  nil,
	}
}

// fetchS3Storage fetches S3 storage containers across all regions
func (m Model) fetchS3Storage() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// First, get regions with S3 storage feature available
	var regions []map[string]interface{}
	regionsEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
	err := httpLib.Client.Get(regionsEndpoint, &regions)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Filter regions with storage-s3 features and fetch containers
	var allContainers []map[string]interface{}
	for _, region := range regions {
		regionName, ok := region["name"].(string)
		if !ok {
			continue
		}

		// Check if region has S3 storage feature
		services, ok := region["services"].([]interface{})
		if !ok {
			continue
		}

		hasS3 := false
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if name, ok := svcMap["name"].(string); ok {
					if name == "storage-s3-high-perf" || name == "storage-s3-standard" {
						hasS3 = true
						break
					}
				}
			}
		}

		if !hasS3 {
			continue
		}

		// Fetch containers for this region
		var containers []map[string]interface{}
		storageEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage", m.cloudProject, regionName)
		if err := httpLib.Client.Get(storageEndpoint, &containers); err == nil {
			allContainers = append(allContainers, containers...)
		}
	}

	return dataLoadedMsg{
		data: allContainers,
		err:  nil,
	}
}

// fetchSwiftStorage fetches Swift storage containers
func (m Model) fetchSwiftStorage() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	var containers []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/storage", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &containers)

	return dataLoadedMsg{
		data: containers,
		err:  err,
	}
}

// fetchBlockStorage fetches block storage volumes
func (m Model) fetchBlockStorage() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	var volumes []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/volume", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &volumes)

	return dataLoadedMsg{
		data: volumes,
		err:  err,
	}
}

// fetchPrivateNetworks fetches private networks
func (m Model) fetchPrivateNetworks() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	var networks []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &networks)

	return dataLoadedMsg{
		data: networks,
		err:  err,
	}
}

// fetchPublicNetworks fetches public networks
func (m Model) fetchPublicNetworks() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	var networks []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/public", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &networks)

	return dataLoadedMsg{
		data: networks,
		err:  err,
	}
}

// fetchLoadBalancers fetches load balancers
func (m Model) fetchLoadBalancers() tea.Msg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	var loadbalancers []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &loadbalancers)

	return dataLoadedMsg{
		data: loadbalancers,
		err:  err,
	}
}

// handleProjectsLoaded processes the loaded projects data
func (m Model) handleProjectsLoaded(msg projectsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.mode = ErrorView
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	if len(msg.projects) == 0 {
		m.mode = ErrorView
		m.errorMsg = "No projects found"
		return m, nil
	}

	// Create table from projects
	m.table = createProjectsTable(msg.projects, m.width, m.height)
	m.currentData = msg.projects // Store raw data for detail viewing
	m.mode = TableView

	return m, nil
}

// handleInstancesLoaded processes the loaded instances data
func (m Model) handleInstancesLoaded(msg instancesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.mode = ErrorView
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	// Create table from instances
	m.table = createInstancesTable(msg.instances, m.width, m.height)
	m.currentData = msg.instances // Store raw data for detail viewing
	m.mode = TableView

	return m, nil
}

// handleDataLoaded processes generic data loaded messages
func (m Model) handleDataLoaded(msg dataLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.mode = ErrorView
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	if len(msg.data) == 0 {
		m.mode = ErrorView
		m.errorMsg = "No data found"
		return m, nil
	}

	// Create generic table
	m.table = createGenericTable(msg.data, m.width, m.height)
	m.currentData = msg.data // Store raw data for detail viewing
	m.mode = TableView

	return m, nil
}

// createProjectsTable creates a table for displaying projects
func createProjectsTable(projects []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Project ID", Width: 40},
		{Title: "Name", Width: 30},
		{Title: "Status", Width: 15},
		{Title: "Description", Width: 40},
	}

	var rows []table.Row
	for _, project := range projects {
		row := table.Row{
			getString(project, "project_id"),
			getString(project, "projectName"),
			getString(project, "status"),
			getString(project, "description"),
		}
		rows = append(rows, row)
	}

	// Calculate table height: leave room for header(2) + nav(3) + title(3) + footer(3) + borders(4)
	tableHeight := height - 15
	if tableHeight < 5 {
		tableHeight = 5
	}
	if tableHeight > 20 {
		tableHeight = 20
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
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

// createInstancesTable creates a table for displaying instances
func createInstancesTable(instances []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 40},
		{Title: "Name", Width: 25},
		{Title: "Status", Width: 15},
		{Title: "Region", Width: 15},
		{Title: "Flavor", Width: 20},
	}

	var rows []table.Row
	for _, instance := range instances {
		row := table.Row{
			getString(instance, "id"),
			getString(instance, "name"),
			getString(instance, "status"),
			getString(instance, "region"),
			getString(instance, "flavorId"),
		}
		rows = append(rows, row)
	}

	// Calculate table height: leave room for header(2) + nav(3) + title(3) + footer(3) + borders(4)
	tableHeight := height - 15
	if tableHeight < 5 {
		tableHeight = 5
	}
	if tableHeight > 20 {
		tableHeight = 20
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
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

// createGenericTable creates a generic table for any data
func createGenericTable(data []map[string]interface{}, width, height int) table.Model {
	if len(data) == 0 {
		return table.Model{}
	}

	// Get all keys from first item to create columns
	var keys []string
	for key := range data[0] {
		keys = append(keys, key)
	}

	columns := make([]table.Column, 0, len(keys))
	colWidth := width / len(keys)
	if colWidth < 15 {
		colWidth = 15
	}
	if colWidth > 40 {
		colWidth = 40
	}

	for _, key := range keys {
		columns = append(columns, table.Column{
			Title: key,
			Width: colWidth,
		})
	}

	var rows []table.Row
	for _, item := range data {
		row := make(table.Row, len(keys))
		for i, key := range keys {
			row[i] = getString(item, key)
		}
		rows = append(rows, row)
	}

	// Calculate table height: leave room for header(2) + nav(3) + title(3) + footer(3) + borders(4)
	tableHeight := height - 15
	if tableHeight < 5 {
		tableHeight = 5
	}
	if tableHeight > 20 {
		tableHeight = 20
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
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

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
