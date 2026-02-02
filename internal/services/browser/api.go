// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// fetchDataForPath initiates an API call based on the path
// It captures the current product to tag the response message
func (m Model) fetchDataForPath(path string) tea.Cmd {
	product := m.currentProduct // Capture current product for the response

	switch path {
	case "/projects":
		return func() tea.Msg {
			msg := m.fetchProjectsData()
			msg.forProduct = product
			return msg
		}
	case "/instances":
		return func() tea.Msg {
			msg := m.fetchInstancesData()
			msg.forProduct = product
			return msg
		}
	case "/kubernetes":
		return func() tea.Msg {
			msg := m.fetchKubernetesData()
			msg.forProduct = product
			return msg
		}
	case "/databases":
		return func() tea.Msg {
			msg := m.fetchDatabasesData()
			msg.forProduct = product
			return msg
		}
	case "/storage/s3":
		return func() tea.Msg {
			msg := m.fetchS3StorageData()
			msg.forProduct = product
			return msg
		}
	case "/storage/swift":
		return func() tea.Msg {
			msg := m.fetchSwiftStorageData()
			msg.forProduct = product
			return msg
		}
	case "/storage/block":
		return func() tea.Msg {
			msg := m.fetchBlockStorageData()
			msg.forProduct = product
			return msg
		}
	case "/networks/private":
		return func() tea.Msg {
			msg := m.fetchPrivateNetworksData()
			msg.forProduct = product
			return msg
		}
	case "/networks/public":
		return func() tea.Msg {
			msg := m.fetchPublicNetworksData()
			msg.forProduct = product
			return msg
		}
	case "/networks/loadbalancer":
		return func() tea.Msg {
			msg := m.fetchLoadBalancersData()
			msg.forProduct = product
			return msg
		}
	default:
		return nil
	}
}

// fetchProjectsData fetches the list of cloud projects (returns data, not tea.Msg)
func (m Model) fetchProjectsData() projectsLoadedMsg {
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

// fetchInstancesData fetches the list of instances immediately and returns
func (m Model) fetchInstancesData() instancesLoadedMsg {
	if m.cloudProject == "" {
		return instancesLoadedMsg{
			err: fmt.Errorf("no cloud project selected. Please configure a default project"),
		}
	}

	var instances []map[string]interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &instances)
	if err != nil {
		return instancesLoadedMsg{err: err}
	}

	// Return instances immediately without waiting for images and floating IPs
	return instancesLoadedMsg{
		instances:     instances,
		imageMap:      make(map[string]string),
		floatingIPMap: make(map[string]string),
		err:           err,
	}
}

// fetchInstancesEnrichedData fetches images and floating IPs in parallel
func (m Model) fetchInstancesEnrichedData(instances []map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		// Fetch images to build imageId -> imageName map
		imageMap := make(map[string]string)
		var images []map[string]interface{}
		imagesEndpoint := fmt.Sprintf("/v1/cloud/project/%s/image", m.cloudProject)
		if imgErr := httpLib.Client.Get(imagesEndpoint, &images); imgErr == nil {
			for _, img := range images {
				if id, ok := img["id"].(string); ok {
					if name, ok := img["name"].(string); ok {
						imageMap[id] = name
					}
				}
			}
		}

		// Fetch floating IPs from all regions to build instanceId -> floatingIP map
		floatingIPMap := make(map[string]string)
		// Get unique regions from instances
		regionSet := make(map[string]bool)
		for _, inst := range instances {
			if region, ok := inst["region"].(string); ok && region != "" {
				regionSet[region] = true
			}
		}
		// Fetch floating IPs for each region
		for region := range regionSet {
			var floatingIPs []map[string]interface{}
			fipEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip", m.cloudProject, region)
			if fipErr := httpLib.Client.Get(fipEndpoint, &floatingIPs); fipErr == nil {
				for _, fip := range floatingIPs {
					// Check if floating IP is associated to an instance
					if associatedEntity, ok := fip["associatedEntity"].(map[string]interface{}); ok {
						if instanceId, ok := associatedEntity["id"].(string); ok && instanceId != "" {
							if ip, ok := fip["ip"].(string); ok {
								floatingIPMap[instanceId] = ip
							}
						}
					}
				}
			}
		}

		return instancesEnrichedMsg{
			imageMap:      imageMap,
			floatingIPMap: floatingIPMap,
		}
	}
}

// fetchKubernetesData fetches the list of Kubernetes clusters
func (m Model) fetchKubernetesData() dataLoadedMsg {
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

// fetchDatabasesData fetches the list of database services
func (m Model) fetchDatabasesData() dataLoadedMsg {
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

// fetchS3StorageData fetches S3 storage containers across all regions
func (m Model) fetchS3StorageData() dataLoadedMsg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// First, get region names (API returns array of strings)
	var regionNames []string
	regionsEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
	err := httpLib.Client.Get(regionsEndpoint, &regionNames)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Fetch details for each region to check if it has S3 storage feature
	var allContainers []map[string]interface{}
	for _, regionName := range regionNames {
		// Get region details to check for S3 feature
		var region map[string]interface{}
		regionDetailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s", m.cloudProject, regionName)
		if err := httpLib.Client.Get(regionDetailEndpoint, &region); err != nil {
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

		// Fetch containers for this region - API may return array of strings or objects
		var rawResponse []interface{}
		storageEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage", m.cloudProject, regionName)
		if err := httpLib.Client.Get(storageEndpoint, &rawResponse); err == nil {
			for _, item := range rawResponse {
				if containerName, ok := item.(string); ok {
					// It's a container name, fetch details
					var container map[string]interface{}
					detailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage/%s", m.cloudProject, regionName, containerName)
					if err := httpLib.Client.Get(detailEndpoint, &container); err == nil {
						allContainers = append(allContainers, container)
					}
				} else if containerObj, ok := item.(map[string]interface{}); ok {
					// It's already a full object
					allContainers = append(allContainers, containerObj)
				}
			}
		}
	}

	return dataLoadedMsg{
		data: allContainers,
		err:  nil,
	}
}

// fetchSwiftStorageData fetches Swift storage containers
func (m Model) fetchSwiftStorageData() dataLoadedMsg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// Try to fetch as array of interfaces first (could be strings or objects)
	var rawResponse []interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/storage", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &rawResponse)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Check if response contains strings (IDs) or objects
	var containers []map[string]interface{}
	if len(rawResponse) > 0 {
		if _, ok := rawResponse[0].(string); ok {
			// Response contains string IDs, fetch details for each
			for _, item := range rawResponse {
				if containerID, ok := item.(string); ok {
					var container map[string]interface{}
					detailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/storage/%s", m.cloudProject, containerID)
					if err := httpLib.Client.Get(detailEndpoint, &container); err == nil {
						containers = append(containers, container)
					}
				}
			}
		} else {
			// Response contains full objects
			for _, item := range rawResponse {
				if obj, ok := item.(map[string]interface{}); ok {
					containers = append(containers, obj)
				}
			}
		}
	}

	return dataLoadedMsg{
		data: containers,
		err:  nil,
	}
}

// fetchBlockStorageData fetches block storage volumes
func (m Model) fetchBlockStorageData() dataLoadedMsg {
	if m.cloudProject == "" {
		return dataLoadedMsg{
			err: fmt.Errorf("no cloud project selected"),
		}
	}

	// Try to fetch as array of interfaces first (could be strings or objects)
	var rawResponse []interface{}
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/volume", m.cloudProject)
	err := httpLib.Client.Get(endpoint, &rawResponse)
	if err != nil {
		return dataLoadedMsg{
			data: nil,
			err:  err,
		}
	}

	// Check if response contains strings (IDs) or objects
	var volumes []map[string]interface{}
	if len(rawResponse) > 0 {
		if _, ok := rawResponse[0].(string); ok {
			// Response contains string IDs, fetch details for each
			for _, item := range rawResponse {
				if volumeID, ok := item.(string); ok {
					var volume map[string]interface{}
					detailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/%s", m.cloudProject, volumeID)
					if err := httpLib.Client.Get(detailEndpoint, &volume); err == nil {
						volumes = append(volumes, volume)
					}
				}
			}
		} else {
			// Response contains full objects
			for _, item := range rawResponse {
				if obj, ok := item.(map[string]interface{}); ok {
					volumes = append(volumes, obj)
				}
			}
		}
	}

	return dataLoadedMsg{
		data: volumes,
		err:  nil,
	}
}

// fetchPrivateNetworksData fetches private networks
func (m Model) fetchPrivateNetworksData() dataLoadedMsg {
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

// fetchPublicNetworksData fetches public networks
func (m Model) fetchPublicNetworksData() dataLoadedMsg {
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

// fetchLoadBalancersData fetches load balancers
func (m Model) fetchLoadBalancersData() dataLoadedMsg {
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
	// Ignore stale response if user switched to a different product
	if msg.forProduct != m.currentProduct {
		return m, nil
	}

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

	// Store projects list for later use
	m.projectsList = msg.projects
	// Create table from projects
	m.table = createProjectsTable(msg.projects, m.width, m.height)
	m.currentData = msg.projects // Store raw data for selection
	m.mode = ProjectSelectView   // Show project selection view

	return m, nil
}

// handleInstancesLoaded processes the loaded instances data
func (m Model) handleInstancesLoaded(msg instancesLoadedMsg) (tea.Model, tea.Cmd) {
	// Ignore stale response if user switched to a different product
	if msg.forProduct != m.currentProduct {
		return m, nil
	}

	if msg.err != nil {
		m.mode = ErrorView
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	// Debug: dump instances to file
	if len(msg.instances) > 0 {
		debugData, _ := json.MarshalIndent(msg.instances[0], "", "  ")
		os.WriteFile("/tmp/instance_debug.json", debugData, 0644)
	}

	// Preserve table cursor position during refresh
	currentCursor := m.table.Cursor()

	// Store maps in model for later use (filtering, etc.)
	m.imageMap = msg.imageMap
	m.floatingIPMap = msg.floatingIPMap

	// Create table from instances immediately with empty maps
	m.table = createInstancesTable(msg.instances, msg.imageMap, msg.floatingIPMap, m.width, m.height)
	m.currentData = msg.instances // Store raw data for detail viewing
	m.mode = TableView

	// Restore cursor position if valid
	if currentCursor >= 0 && currentCursor < len(msg.instances) {
		m.table.SetCursor(currentCursor)
	}

	// Fetch images and floating IPs in parallel to enrich the display
	return m, m.fetchInstancesEnrichedData(msg.instances)
}

// handleInstancesEnriched processes the enriched instances data (images and floating IPs)
func (m Model) handleInstancesEnriched(msg instancesEnrichedMsg) (tea.Model, tea.Cmd) {
	// Preserve cursor position before recreating table
	currentCursor := m.table.Cursor()

	// Update the maps with enriched data
	m.imageMap = msg.imageMap
	m.floatingIPMap = msg.floatingIPMap

	// Recreate the table with enriched data
	m.table = createInstancesTable(m.currentData, msg.imageMap, msg.floatingIPMap, m.width, m.height)

	// Restore cursor position if valid
	if currentCursor >= 0 && currentCursor < len(m.currentData) {
		m.table.SetCursor(currentCursor)
	}

	// Note: Don't schedule refresh here - it's handled by refreshTickMsg to avoid accumulation
	return m, nil
}

// scheduleRefresh schedules an automatic refresh of the data
func (m Model) scheduleRefresh() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

// handleDataLoaded processes generic data loaded messages
func (m Model) handleDataLoaded(msg dataLoadedMsg) (tea.Model, tea.Cmd) {
	// Ignore stale response if user switched to a different product
	if msg.forProduct != m.currentProduct {
		return m, nil
	}

	if msg.err != nil {
		m.mode = ErrorView
		m.errorMsg = msg.err.Error()
		return m, nil
	}

	if len(msg.data) == 0 {
		// Show empty view with creation prompt
		m.mode = EmptyView
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
		{Title: "Name", Width: 40},
		{Title: "Status", Width: 15},
	}

	var rows []table.Row
	for _, project := range projects {
		row := table.Row{
			getString(project, "project_id"),
			getString(project, "description"),
			getString(project, "status"),
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

// createInstancesTable creates a table for displaying instances (like OVHcloud web UI)
func createInstancesTable(instances []map[string]interface{}, imageMap map[string]string, floatingIPMap map[string]string, width, height int) table.Model {
	// Sort instances by name for stable ordering
	sort.Slice(instances, func(i, j int) bool {
		nameI := getString(instances[i], "name")
		nameJ := getString(instances[j], "name")
		return nameI < nameJ
	})

	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Location", Width: 20},
		{Title: "Model", Width: 12},
		{Title: "Image", Width: 18},
		{Title: "Public IP", Width: 16},
		{Title: "Status", Width: 10},
	}

	var rows []table.Row
	for _, instance := range instances {
		// Extract public IP from ipAddresses array
		publicIP := ""
		if addresses, ok := instance["ipAddresses"].([]interface{}); ok {
			for _, addr := range addresses {
				if addrMap, ok := addr.(map[string]interface{}); ok {
					ipType := getString(addrMap, "type")
					if ipType == "public" {
						version := getNumericValue(addrMap, "version")
						if version == 4 { // Prefer IPv4
							publicIP = getString(addrMap, "ip")
							break
						} else if publicIP == "" {
							publicIP = getString(addrMap, "ip")
						}
					}
				}
			}
		}

		// If no public IP found, check for floating IP
		if publicIP == "" && floatingIPMap != nil {
			instanceId := getString(instance, "id")
			if fip, ok := floatingIPMap[instanceId]; ok {
				publicIP = fip + " (FIP)"
			}
		}

		// Extract image name from imageMap or use imageId
		imageName := ""
		imageId := getString(instance, "imageId")
		if imageMap != nil {
			if name, ok := imageMap[imageId]; ok {
				imageName = name
			}
		}
		if imageName == "" {
			// No image name found, show shortened UUID
			if len(imageId) > 8 {
				imageName = imageId[:8] + "..."
			} else {
				imageName = imageId
			}
		}

		// Extract flavor name from planCode (e.g., "b2-7.consumption" -> "b2-7")
		flavorName := ""
		if flavor, ok := instance["flavor"].(map[string]interface{}); ok {
			flavorName = getString(flavor, "name")
		}
		if flavorName == "" {
			// Try to extract from planCode
			planCode := getString(instance, "planCode")
			if planCode != "" {
				// Remove .consumption, .monthly.postpaid, etc.
				for _, suffix := range []string{".consumption", ".monthly.postpaid", ".monthly"} {
					if idx := len(planCode) - len(suffix); idx > 0 && planCode[idx:] == suffix {
						flavorName = planCode[:idx]
						break
					}
				}
				if flavorName == "" {
					flavorName = planCode
				}
			} else {
				// Fallback to shortened flavorId
				flavorId := getString(instance, "flavorId")
				if len(flavorId) > 8 {
					flavorName = flavorId[:8] + "..."
				} else {
					flavorName = flavorId
				}
			}
		}

		row := table.Row{
			getString(instance, "name"),
			getString(instance, "region"),
			flavorName,
			ansi.Truncate(imageName, 18, "..."),
			publicIP,
			getString(instance, "status"),
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

// ============================================
// Wizard API functions for instance creation
// ============================================

// fetchRegions fetches available regions by querying all images and extracting unique regions
func (m Model) fetchRegions() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return regionsLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Fetch all images to determine available regions
		var allImages []map[string]interface{}
		imageEndpoint := fmt.Sprintf("/v1/cloud/project/%s/image", m.cloudProject)
		err := httpLib.Client.Get(imageEndpoint, &allImages)
		if err != nil {
			return regionsLoadedMsg{err: err}
		}

		// Extract unique regions from images
		regionMap := make(map[string]bool)
		var instanceRegions []map[string]interface{}
		for _, image := range allImages {
			regionStr := getString(image, "region")
			if regionStr != "" && !regionMap[regionStr] {
				regionMap[regionStr] = true
				region := map[string]interface{}{
					"name": regionStr,
					"id":   regionStr,
				}
				instanceRegions = append(instanceRegions, region)
			}
		}

		return regionsLoadedMsg{
			regions: instanceRegions,
			images:  allImages, // Cache all images for later use
			err:     nil,
		}
	}
}

// fetchFlavors fetches available flavors for a specific region
func (m Model) fetchFlavors(region string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return flavorsLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var flavors []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/flavor?region=%s", m.cloudProject, region)
		err := httpLib.Client.Get(endpoint, &flavors)

		// Filter out unavailable flavors
		if err == nil {
			var availableFlavors []map[string]interface{}
			for _, flavor := range flavors {
				available, _ := flavor["available"].(bool)
				if available {
					availableFlavors = append(availableFlavors, flavor)
				}
			}
			flavors = availableFlavors
		}

		return flavorsLoadedMsg{
			flavors: flavors,
			err:     err,
		}
	}
}

// fetchImages fetches available images for a specific region (from cache or API)
func (m Model) fetchImages(region string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return imagesLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var images []map[string]interface{}

		// Check if we have cached images from region fetch
		if len(m.wizard.images) > 0 {
			// Filter cached images by the selected region and public/active status
			for _, image := range m.wizard.images {
				imageRegion := getString(image, "region")
				visibility := getString(image, "visibility")
				status := getString(image, "status")
				if imageRegion == region && visibility == "public" && status == "active" {
					images = append(images, image)
				}
			}
		} else {
			// Fallback to API if no cached images (shouldn't happen in normal flow)
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/image?region=%s", m.cloudProject, region)
			err := httpLib.Client.Get(endpoint, &images)
			if err != nil {
				return imagesLoadedMsg{
					images: images,
					err:    err,
				}
			}

			// Filter for common/usable images (exclude snapshots, etc.)
			var publicImages []map[string]interface{}
			for _, image := range images {
				visibility, _ := image["visibility"].(string)
				status, _ := image["status"].(string)
				if visibility == "public" && status == "active" {
					publicImages = append(publicImages, image)
				}
			}
			images = publicImages
		}

		return imagesLoadedMsg{
			images: images,
			err:    nil,
		}
	}
}

// deleteInstance deletes an instance by its ID
func (m Model) deleteInstance(instanceId string) tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s", m.cloudProject, instanceId)
		err := httpLib.Client.Delete(endpoint, nil)
		if err != nil {
			return instanceDeletedMsg{err: fmt.Errorf("failed to delete instance: %w", err)}
		}
		return instanceDeletedMsg{success: true, instanceId: instanceId}
	}
}

// createInstance creates a new instance with the wizard data
func (m Model) createInstance() tea.Cmd {
	// Always use v1 API for instance creation
	// Floating IP will be attached after instance creation if needed
	return m.createInstanceWithNetworking()
}

// createInstanceWithNetworking creates the instance with the configured networking (v1 API)
func (m Model) createInstanceWithNetworking() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return instanceCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Build request body for v1 API
		requestBody := map[string]interface{}{
			"flavorId": m.wizard.selectedFlavor,
			"imageId":  m.wizard.selectedImage,
			"name":     m.wizard.instanceName,
			"region":   m.wizard.selectedRegion,
		}

		// Add SSH key if selected
		if m.wizard.selectedSSHKey != "" {
			requestBody["sshKeyId"] = m.wizard.selectedSSHKey
		}

		// Configure networks
		// Note: For public network only, don't specify any networks (API default is public)
		// For private network, specify only the private network
		// For both, specify only the private network (public is added automatically when no networks specified,
		// but when networks are specified, we need to handle it differently)

		if m.wizard.selectedPrivateNetwork != "" {
			// Private network selected - add it to networks array
			networks := []map[string]interface{}{
				{"networkId": m.wizard.selectedPrivateNetwork},
			}
			requestBody["networks"] = networks
		}
		// If only public network (no private network selected), don't add "networks" key
		// The API will use public network by default

		var instance map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance", m.cloudProject)
		err := httpLib.Client.Post(endpoint, requestBody, &instance)

		return instanceCreatedMsg{
			instance: instance,
			err:      err,
		}
	}
}

// handleRegionsLoaded processes the loaded regions data for the wizard
func (m Model) handleRegionsLoaded(msg regionsLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	// Sort regions alphabetically by name
	sort.Slice(msg.regions, func(i, j int) bool {
		namei := getString(msg.regions[i], "name")
		namej := getString(msg.regions[j], "name")
		return namei < namej
	})

	m.wizard.regions = msg.regions
	m.wizard.images = msg.images // Cache all images for image selection step
	m.wizard.selectedIndex = 0
	return m, nil
}

// handleFlavorsLoaded processes the loaded flavors data for the wizard
func (m Model) handleFlavorsLoaded(msg flavorsLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	m.wizard.flavors = msg.flavors
	m.wizard.selectedIndex = 0
	return m, nil
}

// handleImagesLoaded processes the loaded images data for the wizard
func (m Model) handleImagesLoaded(msg imagesLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	m.wizard.images = msg.images
	m.wizard.selectedIndex = 0
	return m, nil
}

// fetchSSHKeys fetches SSH keys for the selected region
func (m Model) fetchSSHKeys() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return sshKeysLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var sshKeys []map[string]interface{}
		// Query SSH keys filtered by the selected region
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/sshkey?region=%s", m.cloudProject, m.wizard.selectedRegion)
		err := httpLib.Client.Get(endpoint, &sshKeys)

		return sshKeysLoadedMsg{
			sshKeys: sshKeys,
			err:     err,
		}
	}
}

// handleSSHKeysLoaded processes the loaded SSH keys data for the wizard
func (m Model) handleSSHKeysLoaded(msg sshKeysLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	// Store SSH keys directly (create and no-key options are handled in rendering)
	m.wizard.sshKeys = msg.sshKeys
	m.wizard.selectedIndex = 0
	return m, nil
}

// createSSHKey creates a new SSH key in the cloud project
func (m Model) createSSHKey() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return sshKeyCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		requestBody := map[string]interface{}{
			"name":      m.wizard.newSSHKeyName,
			"publicKey": m.wizard.newSSHKeyPublicKey,
		}

		var sshKey map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/sshkey", m.cloudProject)
		err := httpLib.Client.Post(endpoint, requestBody, &sshKey)

		return sshKeyCreatedMsg{
			sshKey: sshKey,
			err:    err,
		}
	}
}

// handleSSHKeyCreated processes the result of SSH key creation
func (m Model) handleSSHKeyCreated(msg sshKeyCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = fmt.Sprintf("Failed to create SSH key: %s", msg.err)
		return m, nil
	}

	// Extract the created SSH key ID and name
	sshKeyId := getString(msg.sshKey, "id")
	sshKeyName := getString(msg.sshKey, "name")

	// Store the SSH key ID for instance creation
	m.wizard.selectedSSHKey = sshKeyId
	m.wizard.selectedSSHKeyName = sshKeyName + " (new)"
	m.wizard.createdSSHKeyId = sshKeyId // Track for cleanup if needed
	m.wizard.creatingSSHKey = false

	// Add notification
	m.notification = fmt.Sprintf("✅ SSH key '%s' created successfully!", sshKeyName)
	m.notificationExpiry = time.Now().Add(3 * time.Second)

	// Go to network step
	m.wizard.step = WizardStepNetwork
	m.wizard.selectedIndex = 0
	m.wizard.filterInput = ""
	m.wizard.isLoading = true
	m.wizard.loadingMessage = "Loading networks..."

	return m, tea.Batch(
		m.fetchPrivateNetworks(),
		tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// fetchPrivateNetworks fetches private networks for the project and region
func (m Model) fetchPrivateNetworks() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return privateNetworksLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var networks []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private", m.cloudProject)
		err := httpLib.Client.Get(endpoint, &networks)

		if err != nil {
			return privateNetworksLoadedMsg{
				networks: networks,
				err:      err,
			}
		}

		// Subnets are not needed for the wizard display - just networks

		return privateNetworksLoadedMsg{
			networks: networks,
			err:      err,
		}
	}
}

// handlePrivateNetworksLoaded processes the loaded private networks data for the wizard
func (m Model) handlePrivateNetworksLoaded(msg privateNetworksLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	// Filter networks available in the selected region
	var availableNetworks []map[string]interface{}
	for _, network := range msg.networks {
		// Check if network has regions that include selected region
		if regions, ok := network["regions"].([]interface{}); ok {
			for _, r := range regions {
				if regionMap, ok := r.(map[string]interface{}); ok {
					regionName := getString(regionMap, "region")
					if regionName == m.wizard.selectedRegion {
						availableNetworks = append(availableNetworks, network)
						break
					}
				}
			}
		}
	}

	// Build the list: No Network, Create New, then existing networks
	noNetworkOption := map[string]interface{}{
		"id":   "",
		"name": "(No Private Network)",
	}
	createNetworkOption := map[string]interface{}{
		"id":   "__create_new__",
		"name": "+ Create new private network",
	}
	m.wizard.privateNetworks = []map[string]interface{}{noNetworkOption, createNetworkOption}
	m.wizard.privateNetworks = append(m.wizard.privateNetworks, availableNetworks...)
	m.wizard.selectedIndex = 0
	m.wizard.usePublicNetwork = true // Default to public network enabled
	m.wizard.networkMenuIndex = 0
	// Reset network creation state
	m.wizard.creatingNetwork = false
	m.wizard.newNetworkName = ""
	m.wizard.newNetworkVlanId = rand.Intn(4094) + 1 // Random VLAN ID 1-4094
	m.wizard.newNetworkCIDR = "10.0.0.0/24"
	m.wizard.newNetworkDHCP = true
	m.wizard.networkCreateField = 0
	return m, nil
}

// handleInstanceCreated processes the result of instance creation
func (m Model) handleInstanceCreated(msg instanceCreatedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	instanceId := getString(msg.instance, "id")
	instanceName := getString(msg.instance, "name")
	if instanceName == "" {
		instanceName = instanceId
	}

	// Store instance info for potential floating IP attachment
	m.wizard.createdInstanceId = instanceId
	m.wizard.createdInstanceName = instanceName

	// Check if we need to attach a floating IP after instance creation
	needsFloatingIP := !m.wizard.usePublicNetwork && m.wizard.selectedPrivateNetwork != "" && m.wizard.selectedFloatingIP != ""

	if needsFloatingIP {
		// Need to wait for instance to get a private IP, then attach floating IP
		m.wizard.loadingMessage = "Waiting for instance private IP..."
		m.notification = fmt.Sprintf("⏳ Instance '%s' created, attaching Floating IP...", instanceName)
		m.notificationExpiry = time.Now().Add(30 * time.Second)

		// Poll for the instance's private IP
		return m, m.waitForInstanceIP(instanceId, instanceName)
	}

	// No floating IP needed - show success and finish
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	m.notification = fmt.Sprintf("✅ Instance '%s' created successfully!", instanceName)
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// Reset wizard and go back to instances view
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

func (m Model) handleInstanceDeleted(msg instanceDeletedMsg) (tea.Model, tea.Cmd) {
	m.deleteTarget = nil
	m.deleteConfirmInput = ""

	if msg.err != nil {
		m.notification = fmt.Sprintf("❌ %s", msg.err.Error())
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = TableView
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		})
	}

	m.notification = "✅ Instance deleted successfully!"
	m.notificationExpiry = time.Now().Add(5 * time.Second)
	m.mode = LoadingView

	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// waitForInstanceIP polls the instance until it has a private IP
func (m Model) waitForInstanceIP(instanceId, instanceName string) tea.Cmd {
	return func() tea.Msg {
		// Poll for up to 60 seconds
		maxAttempts := 12
		for attempt := 0; attempt < maxAttempts; attempt++ {
			var instance map[string]interface{}
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", m.cloudProject, instanceId)
			err := httpLib.Client.Get(endpoint, &instance)
			if err != nil {
				return instanceIPReadyMsg{
					instanceId:   instanceId,
					instanceName: instanceName,
					err:          fmt.Errorf("failed to get instance: %w", err),
				}
			}

			// Look for private IP in ipAddresses array
			if ipAddresses, ok := instance["ipAddresses"].([]interface{}); ok {
				for _, ipAddr := range ipAddresses {
					if ipMap, ok := ipAddr.(map[string]interface{}); ok {
						ipType := getString(ipMap, "type")
						ip := getString(ipMap, "ip")
						if ipType == "private" && ip != "" {
							return instanceIPReadyMsg{
								instanceId:   instanceId,
								instanceName: instanceName,
								privateIP:    ip,
							}
						}
					}
				}
			}

			// Wait 5 seconds before next attempt
			time.Sleep(5 * time.Second)
		}

		return instanceIPReadyMsg{
			instanceId:   instanceId,
			instanceName: instanceName,
			err:          fmt.Errorf("timeout waiting for instance private IP"),
		}
	}
}

// handleInstanceIPReady processes when instance has a private IP ready
func (m Model) handleInstanceIPReady(msg instanceIPReadyMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		// Offer cleanup if we have created resources
		if m.hasCreatedResources() {
			m.wizard.cleanupPending = true
			m.wizard.cleanupError = fmt.Sprintf("Instance created but floating IP attachment failed: %s", msg.err)
			return m, nil
		}
		m.notification = fmt.Sprintf("⚠️ Instance created but floating IP attachment failed: %s", msg.err)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return clearNotificationMsg{}
			}),
		)
	}

	// Now attach the floating IP (this also creates a gateway)
	m.wizard.loadingMessage = "Creating gateway and attaching Floating IP..."
	m.notification = fmt.Sprintf("⏳ Attaching Floating IP to '%s'...", msg.instanceName)
	m.notificationExpiry = time.Now().Add(30 * time.Second)

	return m, m.attachFloatingIP(msg.instanceId, msg.instanceName, msg.privateIP)
}

// attachFloatingIP creates and attaches a floating IP to the instance
func (m Model) attachFloatingIP(instanceId, instanceName, privateIP string) tea.Cmd {
	return func() tea.Msg {
		// Build request body
		requestBody := map[string]interface{}{
			"ip": privateIP,
		}

		// Add gateway creation if needed (small gateway)
		requestBody["gateway"] = map[string]interface{}{
			"model": "s",
			"name":  fmt.Sprintf("gw-%s", instanceName),
		}

		var result map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/floatingIp",
			m.cloudProject, m.wizard.selectedRegion, instanceId)
		err := httpLib.Client.Post(endpoint, requestBody, &result)

		if err != nil {
			return floatingIPAttachedMsg{
				instanceName: instanceName,
				err:          fmt.Errorf("failed to attach floating IP: %w", err),
			}
		}

		return floatingIPAttachedMsg{
			instanceName: instanceName,
		}
	}
}

// handleFloatingIPAttached processes the result of floating IP attachment
func (m Model) handleFloatingIPAttached(msg floatingIPAttachedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		// Offer cleanup if we have created resources
		if m.hasCreatedResources() {
			m.wizard.cleanupPending = true
			m.wizard.cleanupError = msg.err.Error()
			return m, nil
		}
		m.notification = fmt.Sprintf("⚠️ Instance created but: %s", msg.err)
	} else {
		m.notification = fmt.Sprintf("✅ Instance '%s' created with Floating IP!", msg.instanceName)
	}
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// Reset wizard and go back to instances view
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// hasCreatedResources checks if any resources were created during the wizard
func (m Model) hasCreatedResources() bool {
	return m.wizard.createdInstanceId != "" ||
		m.wizard.createdNetworkId != "" ||
		m.wizard.createdSubnetId != "" ||
		m.wizard.createdGatewayId != "" ||
		m.wizard.createdFloatingIPId != ""
}

// cleanupCreatedResources deletes all resources created during the wizard
func (m Model) cleanupCreatedResources() tea.Cmd {
	return func() tea.Msg {
		var deletedResources []string
		var errors []string

		// Delete in reverse order of creation: floating IP -> gateway -> instance -> subnet -> network -> SSH key

		// Delete floating IP if created
		if m.wizard.createdFloatingIPId != "" {
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip/%s",
				m.cloudProject, m.wizard.selectedRegion, m.wizard.createdFloatingIPId)
			if err := httpLib.Client.Delete(endpoint, nil); err != nil {
				errors = append(errors, fmt.Sprintf("Floating IP: %s", err))
			} else {
				deletedResources = append(deletedResources, "Floating IP")
			}
		}

		// Delete gateway if created
		if m.wizard.createdGatewayId != "" {
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway/%s",
				m.cloudProject, m.wizard.selectedRegion, m.wizard.createdGatewayId)
			if err := httpLib.Client.Delete(endpoint, nil); err != nil {
				errors = append(errors, fmt.Sprintf("Gateway: %s", err))
			} else {
				deletedResources = append(deletedResources, "Gateway")
			}
		}

		// Delete instance if created
		if m.wizard.createdInstanceId != "" {
			endpoint := fmt.Sprintf("/cloud/project/%s/instance/%s", m.cloudProject, m.wizard.createdInstanceId)
			if err := httpLib.Client.Delete(endpoint, nil); err != nil {
				errors = append(errors, fmt.Sprintf("Instance: %s", err))
			} else {
				deletedResources = append(deletedResources, "Instance")
			}
		}

		// Delete network if created (this will also delete the subnet)
		if m.wizard.createdNetworkId != "" {
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s",
				m.cloudProject, m.wizard.createdNetworkId)
			if err := httpLib.Client.Delete(endpoint, nil); err != nil {
				errors = append(errors, fmt.Sprintf("Network: %s", err))
			} else {
				deletedResources = append(deletedResources, "Network")
			}
		}

		// Delete SSH key if created
		if m.wizard.createdSSHKeyId != "" {
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/sshkey/%s",
				m.cloudProject, m.wizard.createdSSHKeyId)
			if err := httpLib.Client.Delete(endpoint, nil); err != nil {
				errors = append(errors, fmt.Sprintf("SSH Key: %s", err))
			} else {
				deletedResources = append(deletedResources, "SSH Key")
			}
		}

		return cleanupCompletedMsg{
			deletedResources: deletedResources,
			errors:           errors,
		}
	}
}

// handleCleanupCompleted processes the result of cleanup
func (m Model) handleCleanupCompleted(msg cleanupCompletedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if len(msg.deletedResources) > 0 {
		m.notification = fmt.Sprintf("🗑️ Cleaned up: %s", strings.Join(msg.deletedResources, ", "))
	}
	if len(msg.errors) > 0 {
		m.notification = fmt.Sprintf("⚠️ Cleanup partial - errors: %s", strings.Join(msg.errors, "; "))
	}
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// Reset wizard and go back to instances view
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// ============================================
// Floating IP Functions
// ============================================

// fetchFloatingIPs fetches available floating IPs for the selected region
func (m Model) fetchFloatingIPs() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return floatingIPsLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var floatingIPs []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip", m.cloudProject, m.wizard.selectedRegion)
		err := httpLib.Client.Get(endpoint, &floatingIPs)

		return floatingIPsLoadedMsg{
			floatingIPs: floatingIPs,
			err:         err,
		}
	}
}

// handleFloatingIPsLoaded processes the loaded floating IPs data
func (m Model) handleFloatingIPsLoaded(msg floatingIPsLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		// If region doesn't support floating IPs, just show create option
		m.wizard.floatingIPs = []map[string]interface{}{
			{"id": "__none__", "name": "(No Floating IP - no external access)"},
			{"id": "__create_new__", "name": "+ Create new Floating IP"},
		}
		m.wizard.selectedIndex = 0
		return m, nil
	}

	// Filter floating IPs that are not associated (available)
	var availableFloatingIPs []map[string]interface{}
	for _, fip := range msg.floatingIPs {
		// Check if floating IP is not associated to an instance
		associatedEntity := getString(fip, "associatedEntity")
		if associatedEntity == "" {
			availableFloatingIPs = append(availableFloatingIPs, fip)
		}
	}

	// Build the list: No floating IP, Create New, then available IPs
	noFIPOption := map[string]interface{}{
		"id":   "__none__",
		"name": "(No Floating IP - no external access)",
	}
	createFIPOption := map[string]interface{}{
		"id":   "__create_new__",
		"name": "+ Create new Floating IP",
	}
	m.wizard.floatingIPs = []map[string]interface{}{noFIPOption, createFIPOption}
	m.wizard.floatingIPs = append(m.wizard.floatingIPs, availableFloatingIPs...)
	m.wizard.selectedIndex = 0
	return m, nil
}

// createGatewayIfNeeded checks if the subnet has a gateway, and creates one if not
func (m Model) createGatewayIfNeeded() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return gatewayCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// First, check if there's already a gateway in the region
		var gateways []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway", m.cloudProject, m.wizard.selectedRegion)
		err := httpLib.Client.Get(endpoint, &gateways)
		if err == nil && len(gateways) > 0 {
			// Gateway exists, check if it's attached to our network
			for _, gw := range gateways {
				if interfaces, ok := gw["interfaces"].([]interface{}); ok {
					for _, iface := range interfaces {
						if ifaceMap, ok := iface.(map[string]interface{}); ok {
							networkId := getString(ifaceMap, "networkId")
							if networkId == m.wizard.selectedPrivateNetwork {
								// Gateway already exists for this network
								return gatewayCreatedMsg{gateway: gw, err: nil}
							}
						}
					}
				}
			}
		}

		// Need to create a gateway - create S size by default
		gatewayBody := map[string]interface{}{
			"name":  fmt.Sprintf("gw-%s", m.wizard.instanceName),
			"model": "s",
			"network": map[string]interface{}{
				"id": m.wizard.selectedPrivateNetwork,
			},
		}

		bodyBytes, _ := json.Marshal(gatewayBody)
		var gateway map[string]interface{}
		err = httpLib.Client.Post(endpoint, string(bodyBytes), &gateway)

		return gatewayCreatedMsg{
			gateway: gateway,
			err:     err,
		}
	}
}

// handleGatewayCreated processes the result of gateway creation
func (m Model) handleGatewayCreated(msg gatewayCreatedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		// Gateway creation failed, but we can continue without it
		// The floating IP just won't work
		m.wizard.errorMsg = fmt.Sprintf("Warning: Gateway creation failed: %s", msg.err.Error())
	}

	// Continue with instance creation
	return m, m.createInstanceWithNetworking()
}

// createFloatingIP creates a new floating IP in the selected region
func (m Model) createFloatingIP(instanceId string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return floatingIPCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Create the floating IP
		fipBody := map[string]interface{}{
			"description": fmt.Sprintf("FIP for %s", m.wizard.instanceName),
		}

		bodyBytes, _ := json.Marshal(fipBody)
		var floatingIP map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip", m.cloudProject, m.wizard.selectedRegion)
		err := httpLib.Client.Post(endpoint, string(bodyBytes), &floatingIP)

		if err != nil {
			return floatingIPCreatedMsg{err: err}
		}

		// Now associate it with the instance
		fipId := getString(floatingIP, "id")
		if fipId != "" && instanceId != "" {
			associateBody := map[string]interface{}{
				"instanceId": instanceId,
			}
			bodyBytes, _ = json.Marshal(associateBody)
			associateEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/floatingip/%s/attach", m.cloudProject, m.wizard.selectedRegion, fipId)
			httpLib.Client.Post(associateEndpoint, string(bodyBytes), nil)
		}

		return floatingIPCreatedMsg{
			floatingIP: floatingIP,
			err:        nil,
		}
	}
}

// handleFloatingIPCreated processes the result of floating IP creation
func (m Model) handleFloatingIPCreated(msg floatingIPCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""

	if msg.err != nil {
		// Show error but instance was still created
		m.notification = fmt.Sprintf("⚠️ Instance created but floating IP failed: %s", msg.err.Error())
		m.notificationExpiry = time.Now().Add(8 * time.Second)
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances"),
			tea.Tick(8*time.Second, func(t time.Time) tea.Msg {
				return clearNotificationMsg{}
			}),
		)
	}

	// Success! The instance is created and floating IP is attached
	fipIP := getString(msg.floatingIP, "ip")
	m.notification = fmt.Sprintf("✅ Instance created with Floating IP %s!", fipIP)
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// Reset wizard and go back to instances view
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// ensureGatewayAndCreateFloatingIP ensures a gateway exists for the subnet and creates a floating IP
func (m Model) ensureGatewayAndCreateFloatingIP(instanceId string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return floatingIPCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Step 1: Check if gateway exists for this subnet
		gwCheckEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway?subnetId=%s",
			m.cloudProject, m.wizard.selectedRegion, m.wizard.selectedSubnetId)
		var gateways []map[string]interface{}
		err := httpLib.Client.Get(gwCheckEndpoint, &gateways)

		hasGateway := err == nil && len(gateways) > 0

		// Step 2: Wait for instance to have a private IP
		var privateIP string
		for retry := 0; retry < 20; retry++ {
			var instance map[string]interface{}
			instanceEndpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s", m.cloudProject, instanceId)
			if err := httpLib.Client.Get(instanceEndpoint, &instance); err == nil {
				// Look for private IP in ipAddresses
				if ipAddresses, ok := instance["ipAddresses"].([]interface{}); ok {
					for _, ipAddr := range ipAddresses {
						if ipMap, ok := ipAddr.(map[string]interface{}); ok {
							ipType := getString(ipMap, "type")
							if ipType == "private" {
								privateIP = getString(ipMap, "ip")
								break
							}
						}
					}
				}
			}
			if privateIP != "" {
				break
			}
			time.Sleep(3 * time.Second)
		}

		if privateIP == "" {
			return floatingIPCreatedMsg{err: fmt.Errorf("could not get instance private IP after waiting (instance may still be starting)")}
		}

		// Step 3: Create floating IP with gateway info if needed
		fipBody := map[string]interface{}{
			"ip": privateIP,
		}

		// If no gateway exists, include gateway creation parameters
		if !hasGateway {
			fipBody["gateway"] = map[string]interface{}{
				"model": "s",
				"name":  fmt.Sprintf("gw-%s", m.wizard.instanceName),
			}
		}

		bodyBytes, _ := json.Marshal(fipBody)
		var floatingIPResult map[string]interface{}
		fipEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/instance/%s/floatingIp",
			m.cloudProject, m.wizard.selectedRegion, instanceId)
		err = httpLib.Client.Post(fipEndpoint, string(bodyBytes), &floatingIPResult)

		if err != nil {
			return floatingIPCreatedMsg{err: fmt.Errorf("floating IP creation failed: %w", err)}
		}

		return floatingIPCreatedMsg{
			floatingIP: floatingIPResult,
			err:        nil,
		}
	}
}

// waitForOperation waits for a cloud operation to complete
func waitForOperation(projectID, operationID string, timeout time.Duration) error {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/operation/%s", projectID, operationID)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		var operation map[string]interface{}
		if err := httpLib.Client.Get(endpoint, &operation); err != nil {
			return fmt.Errorf("error fetching operation: %w", err)
		}

		status := getString(operation, "status")
		switch status {
		case "in-error":
			return fmt.Errorf("operation ended in error")
		case "completed":
			return nil
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("timeout waiting for operation")
}

// ============================================
// Instance Actions (Reboot, Stop, Rescue, etc.)
// ============================================

// executeInstanceAction executes an action on the current instance
func (m Model) executeInstanceAction(actionIndex int) tea.Cmd {
	return func() tea.Msg {
		if m.detailData == nil {
			return instanceActionMsg{err: fmt.Errorf("no instance selected")}
		}

		instanceId := getString(m.detailData, "id")
		if instanceId == "" {
			return instanceActionMsg{err: fmt.Errorf("instance ID not found")}
		}

		actions := []string{"ssh", "reboot", "rescue", "stop_or_start", "vnc", "reinstall"}
		if actionIndex < 0 || actionIndex >= len(actions) {
			return instanceActionMsg{err: fmt.Errorf("invalid action index")}
		}

		action := actions[actionIndex]
		var err error

		switch action {
		case "ssh":
			// Get public IP from instance
			publicIP := ""
			if addresses, ok := m.detailData["ipAddresses"].([]interface{}); ok {
				for _, addr := range addresses {
					if addrMap, ok := addr.(map[string]interface{}); ok {
						ipType := getString(addrMap, "type")
						if ipType == "public" {
							version := getNumericValue(addrMap, "version")
							if version == 4 {
								publicIP = getString(addrMap, "ip")
								break
							} else if publicIP == "" {
								publicIP = getString(addrMap, "ip")
							}
						}
					}
				}
			}
			// If no public IP, check for floating IP
			if publicIP == "" && m.floatingIPMap != nil {
				if fip, ok := m.floatingIPMap[instanceId]; ok {
					publicIP = fip
				}
			}
			if publicIP == "" {
				return instanceActionMsg{err: fmt.Errorf("no public IP or floating IP found for SSH connection")}
			}
			// Return special SSH message with the IP
			// Try to detect user from image name
			user := "ubuntu" // Default
			imageId := getString(m.detailData, "imageId")
			imageName := ""
			// First try to get image name from imageMap
			if m.imageMap != nil {
				if name, ok := m.imageMap[imageId]; ok {
					imageName = strings.ToLower(name)
				}
			}
			// Fallback to imageId if no name found
			if imageName == "" {
				imageName = strings.ToLower(imageId)
			}
			if strings.Contains(imageName, "debian") {
				user = "debian"
			} else if strings.Contains(imageName, "centos") {
				user = "centos"
			} else if strings.Contains(imageName, "fedora") {
				user = "fedora"
			} else if strings.Contains(imageName, "arch") {
				user = "arch"
			} else if strings.Contains(imageName, "rocky") {
				user = "rocky"
			} else if strings.Contains(imageName, "almalinux") {
				user = "almalinux"
			}
			return sshConnectionMsg{ip: publicIP, user: user}

		case "reboot":
			// POST /cloud/project/{serviceName}/instance/{instanceId}/reboot
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/reboot", m.cloudProject, instanceId)
			body := map[string]string{"type": "soft"}
			err = httpLib.Client.Post(endpoint, body, nil)

		case "rescue":
			// Check instance status to determine if we should rescue or unrescue
			status := strings.ToUpper(getString(m.detailData, "status"))
			if status == "RESCUE" {
				// Exit rescue mode
				action = "unrescue"
				endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/rescueMode", m.cloudProject, instanceId)
				body := map[string]bool{"rescue": false}
				err = httpLib.Client.Post(endpoint, body, nil)
			} else {
				// Enter rescue mode
				endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/rescueMode", m.cloudProject, instanceId)
				body := map[string]bool{"rescue": true}
				err = httpLib.Client.Post(endpoint, body, nil)
			}

		case "stop_or_start":
			// Check instance status to determine if we should start or stop
			status := strings.ToUpper(getString(m.detailData, "status"))
			if status == "SHUTOFF" {
				// Start the instance
				action = "start"
				endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/start", m.cloudProject, instanceId)
				err = httpLib.Client.Post(endpoint, nil, nil)
			} else {
				// Stop the instance
				action = "stop"
				endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/stop", m.cloudProject, instanceId)
				err = httpLib.Client.Post(endpoint, nil, nil)
			}

		case "vnc":
			// POST /cloud/project/{serviceName}/instance/{instanceId}/vnc
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/vnc", m.cloudProject, instanceId)
			var result map[string]interface{}
			err = httpLib.Client.Post(endpoint, nil, &result)
			if err == nil {
				// Get the VNC URL from response
				if vncUrl, ok := result["url"].(string); ok {
					return instanceActionMsg{action: "vnc", instanceId: instanceId, err: fmt.Errorf("VNC URL: %s", vncUrl)}
				}
			}

		case "reinstall":
			// POST /cloud/project/{serviceName}/instance/{instanceId}/reinstall
			// Note: This requires imageId, which we can get from the current instance
			imageId := getString(m.detailData, "imageId")
			if imageId == "" {
				return instanceActionMsg{err: fmt.Errorf("cannot reinstall: no image ID found")}
			}
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/instance/%s/reinstall", m.cloudProject, instanceId)
			body := map[string]string{"imageId": imageId}
			err = httpLib.Client.Post(endpoint, body, nil)
		}

		return instanceActionMsg{
			action:     action,
			instanceId: instanceId,
			err:        err,
		}
	}
}

// handleInstanceAction processes the result of an instance action
func (m Model) handleInstanceAction(msg instanceActionMsg) (tea.Model, tea.Cmd) {
	actionNames := map[string]string{
		"ssh":       "SSH",
		"reboot":    "Reboot",
		"rescue":    "Rescue Mode",
		"unrescue":  "Exit Rescue",
		"stop":      "Stop",
		"start":     "Start",
		"vnc":       "Console",
		"reinstall": "Reinstall",
	}

	actionName := actionNames[msg.action]
	if actionName == "" {
		actionName = msg.action
	}

	if msg.err != nil {
		// Special case for VNC - the "error" contains the URL
		if msg.action == "vnc" && strings.HasPrefix(msg.err.Error(), "VNC URL:") {
			m.notification = fmt.Sprintf("🖥️ %s", msg.err.Error())
		} else if msg.action == "ssh" && strings.Contains(msg.err.Error(), "255") {
			// SSH exit code 255 = connection error
			m.notification = "❌ SSH failed: connection error (check security group, port 22, or SSH key)"
		} else {
			m.notification = fmt.Sprintf("❌ %s failed: %s", actionName, msg.err)
		}
	} else {
		if msg.action == "ssh" {
			m.notification = "✅ SSH session ended"
		} else {
			m.notification = fmt.Sprintf("✅ %s initiated successfully!", actionName)
		}
	}
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// For SSH, stay on detail view - don't refresh the list
	if msg.action == "ssh" {
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		})
	}

	// Refresh the instances list to see updated status
	return m, tea.Batch(
		m.fetchDataForPath("/instances"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// createPrivateNetwork starts the network creation process (Step 1: Create network)
func (m Model) createPrivateNetwork() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return networkCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Step 1: Create the private network
		networkBody := map[string]interface{}{
			"name":    m.wizard.newNetworkName,
			"vlanId":  m.wizard.newNetworkVlanId,
			"regions": []string{m.wizard.selectedRegion},
		}

		var network map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private", m.cloudProject)
		err := httpLib.Client.Post(endpoint, networkBody, &network)
		if err != nil {
			return networkCreatedMsg{err: fmt.Errorf("failed to create network: %w", err)}
		}

		networkId := getString(network, "id")
		if networkId == "" {
			return networkCreatedMsg{err: fmt.Errorf("network created but ID not returned")}
		}

		// Return step message to continue with subnet creation
		return networkStepMsg{
			step:      "network_created",
			networkId: networkId,
			network:   network,
		}
	}
}

// handleNetworkStep handles network creation steps
func (m Model) handleNetworkStep(msg networkStepMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	switch msg.step {
	case "network_created":
		// Network created, now create subnet
		m.wizard.loadingMessage = "Creating subnet..."
		return m, m.createSubnet(msg.networkId, msg.network)

	case "subnet_created":
		// Subnet created, network is ready
		m.wizard.loadingMessage = ""
		return m.handleNetworkCreated(networkCreatedMsg{network: msg.network, err: nil})
	}

	return m, nil
}

// createSubnet creates a subnet for the network (Step 2)
func (m Model) createSubnet(networkId string, network map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		cidr := m.wizard.newNetworkCIDR
		if cidr == "" {
			cidr = "10.0.0.0/24"
		}

		// Calculate gateway (first usable IP) and DHCP range
		// For 10.0.0.0/24: gateway=10.0.0.1, start=10.0.0.2, end=10.0.0.254
		parts := strings.Split(cidr, "/")
		ipParts := strings.Split(parts[0], ".")
		if len(ipParts) != 4 {
			return networkStepMsg{
				step: "subnet_created",
				err:  fmt.Errorf("invalid CIDR format"),
			}
		}

		baseIP := ipParts[0] + "." + ipParts[1] + "." + ipParts[2]
		gateway := baseIP + ".1"
		dhcpStart := baseIP + ".2"
		dhcpEnd := baseIP + ".254"

		subnetBody := map[string]interface{}{
			"region":    m.wizard.selectedRegion,
			"network":   cidr,
			"noGateway": false,
			"dhcp":      m.wizard.newNetworkDHCP,
		}

		// Only add IP pool if DHCP is enabled
		if m.wizard.newNetworkDHCP {
			subnetBody["start"] = dhcpStart
			subnetBody["end"] = dhcpEnd
		}

		subnetEndpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet", m.cloudProject, networkId)

		// Retry creating subnet with exponential backoff (network needs to activate)
		var subnet map[string]interface{}
		var subnetErr error
		maxRetries := 10
		for retry := 0; retry < maxRetries; retry++ {
			subnetErr = httpLib.Client.Post(subnetEndpoint, subnetBody, &subnet)
			if subnetErr == nil {
				break
			}
			// Check if it's a "wait for activation" error
			if strings.Contains(subnetErr.Error(), "region activation") ||
				strings.Contains(subnetErr.Error(), "Please wait") {
				// Wait before retrying (2s, 3s, 4s, ...)
				time.Sleep(time.Duration(2+retry) * time.Second)
				continue
			}
			// Other error, don't retry
			break
		}

		if subnetErr != nil {
			return networkStepMsg{
				step:      "subnet_created",
				networkId: networkId,
				network:   network,
				err:       fmt.Errorf("network created but subnet failed: %w", subnetErr),
			}
		}

		// Add gateway info to network for display
		network["gateway"] = gateway
		network["subnet"] = cidr

		return networkStepMsg{
			step:      "subnet_created",
			networkId: networkId,
			network:   network,
		}
	}
}

// handleNetworkCreated processes the result of network creation
func (m Model) handleNetworkCreated(msg networkCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	m.wizard.creatingNetwork = false

	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}

	// Network created successfully - select it and add to list
	networkId := getString(msg.network, "id")
	networkName := getString(msg.network, "name")

	// Store created network ID for potential cleanup
	m.wizard.createdNetworkId = networkId

	m.wizard.selectedPrivateNetwork = networkId
	m.wizard.selectedPrivateNetworkName = networkName

	// Add the new network to the list (after "No Network" and "Create New")
	newNetworkEntry := map[string]interface{}{
		"id":   networkId,
		"name": networkName + " (new)",
	}

	// Insert after the first two options
	if len(m.wizard.privateNetworks) >= 2 {
		m.wizard.privateNetworks = append(
			m.wizard.privateNetworks[:2],
			append([]map[string]interface{}{newNetworkEntry}, m.wizard.privateNetworks[2:]...)...,
		)
		m.wizard.selectedIndex = 2 // Select the newly created network
	} else {
		m.wizard.privateNetworks = append(m.wizard.privateNetworks, newNetworkEntry)
		m.wizard.selectedIndex = len(m.wizard.privateNetworks) - 1
	}

	m.notification = fmt.Sprintf("✅ Network '%s' created successfully!", networkName)
	m.notificationExpiry = time.Now().Add(5 * time.Second)

	// Reset creation fields
	m.wizard.newNetworkName = ""
	m.wizard.newNetworkCIDR = "10.0.0.0/24"
	m.wizard.newNetworkDHCP = true
	m.wizard.networkCreateField = 0

	return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}
