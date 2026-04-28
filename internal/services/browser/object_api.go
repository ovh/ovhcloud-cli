// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	object_storage "github.com/ovh/ovhcloud-cli/internal/services/browser/views/object_storage"
)

// ─── Object Storage API ───────────────────────────────────────────────────────

// fetchObjectStorageInitData loads S3-capable regions and cloud users concurrently.
func (m Model) fetchObjectStorageInitData() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return objectStorageInitDataLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		// Fetch region names
		var regionNames []string
		if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject), &regionNames); err != nil {
			return objectStorageInitDataLoadedMsg{err: fmt.Errorf("failed to fetch regions: %w", err)}
		}

		// Probe regions concurrently for S3 support
		type probeResult struct {
			region    string
			supported bool
		}
		ch := make(chan probeResult, len(regionNames))
		for _, name := range regionNames {
			go func(r string) {
				var region map[string]interface{}
				ep := fmt.Sprintf("/v1/cloud/project/%s/region/%s", m.cloudProject, url.PathEscape(r))
				if err := httpLib.Client.Get(ep, &region); err != nil {
					ch <- probeResult{region: r, supported: false}
					return
				}
				services, _ := region["services"].([]interface{})
				for _, svc := range services {
					if sm, ok := svc.(map[string]interface{}); ok {
						if n, _ := sm["name"].(string); n == "storage-s3-high-perf" || n == "storage-s3-standard" {
							ch <- probeResult{region: r, supported: true}
							return
						}
					}
				}
				ch <- probeResult{region: r, supported: false}
			}(name)
		}

		var supportedRegions []string
		for range regionNames {
			r := <-ch
			if r.supported {
				supportedRegions = append(supportedRegions, r.region)
			}
		}
		sort.Strings(supportedRegions)

		// Fetch cloud users
		var users []map[string]interface{}
		userEndpoint := fmt.Sprintf("/v1/cloud/project/%s/user", m.cloudProject)
		if err := httpLib.Client.Get(userEndpoint, &users); err != nil {
			// Non-fatal: continue without users
			users = nil
		}
		sort.Slice(users, func(i, j int) bool {
			iName, _ := users[i]["username"].(string)
			jName, _ := users[j]["username"].(string)
			return iName < jName
		})

		if len(supportedRegions) == 0 {
			return objectStorageInitDataLoadedMsg{err: fmt.Errorf("no regions support object storage in this project")}
		}
		return objectStorageInitDataLoadedMsg{regions: supportedRegions, users: users}
	}
}

// createObjectContainer creates a new S3 container via the OVH API.
func (m Model) createObjectContainer() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return objectContainerCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		body := map[string]interface{}{
			"name": m.wizard.objectName,
		}

		// Encryption
		if m.wizard.objectEncryption {
			body["encryption"] = map[string]interface{}{
				"sseAlgorithm": "AES256",
			}
		}

		// ObjectLock requires versioning; enable both if lock is requested.
		enableVersioning := m.wizard.objectVersioning || m.wizard.objectLock

		// Versioning
		if enableVersioning {
			body["versioning"] = map[string]interface{}{
				"status": "enabled",
			}
		}

		// Object Lock (requires versioning)
		if m.wizard.objectLock {
			body["objectLock"] = map[string]interface{}{
				"status": "enabled",
			}
		}

		// Owner (user)
		if m.wizard.objectUserIdx > 0 && m.wizard.objectUserIdx <= len(m.wizard.objectUsers) {
			user := m.wizard.objectUsers[m.wizard.objectUserIdx-1]
			if ownerId, ok := user["id"]; ok {
				switch v := ownerId.(type) {
				case float64:
					body["ownerId"] = int(v)
				case int:
					body["ownerId"] = v
				}
			}
		}

		// Container type (storageClass in name? No — it's a separate field per region).
		// Type is encoded in the region selection (High Perf vs Standard regions).

		region := m.wizard.selectedRegion
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage",
			m.cloudProject, url.PathEscape(region))

		var container map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &container); err != nil {
			return objectContainerCreatedMsg{err: fmt.Errorf("failed to create container: %w", err)}
		}
		return objectContainerCreatedMsg{container: container}
	}
}

// deleteObjectContainer deletes an S3 container.
func (m Model) deleteObjectContainer(containerName, region string) tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/storage/%s",
			m.cloudProject, url.PathEscape(region), url.PathEscape(containerName))
		err := httpLib.Client.Delete(endpoint, nil)
		return objectContainerActionDoneMsg{action: object_storage.ContainerActionDelete, err: err}
	}
}

// ─── Object Storage message handlers ─────────────────────────────────────────

func (m Model) handleObjectStorageInitDataLoaded(msg objectStorageInitDataLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	m.wizard.objectRegions = msg.regions
	m.wizard.objectUsers = msg.users
	m.wizard.selectedIndex = 0
	return m, nil
}

func (m Model) handleObjectContainerCreated(msg objectContainerCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	name := getString(msg.container, "name")
	if name == "" {
		name = "container"
	}
	m.notification = fmt.Sprintf("✅ Container '%s' created successfully!", name)
	m.notificationExpiry = time.Now().Add(5 * time.Second)
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/storage/object"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

func (m Model) handleObjectContainerActionDone(msg objectContainerActionDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notification = fmt.Sprintf("❌ Action échouée: %s", msg.err.Error())
		m.notificationExpiry = time.Now().Add(8 * time.Second)
		return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		})
	}
	m.notification = "✅ Container supprimé avec succès!"
	m.notificationExpiry = time.Now().Add(5 * time.Second)
	m.objectDetailView = nil
	m.detailData = nil
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/storage/object"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}

// createObjectStorageTable builds the table for S3 containers.
func createObjectStorageTable(data []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Nom", Width: 28},
		{Title: "Localisation", Width: 12},
		{Title: "Mode de déploiement", Width: 22},
		{Title: "Offre", Width: 16},
		{Title: "Nbr objets", Width: 11},
		{Title: "Espace utilisé", Width: 14},
		{Title: "Type", Width: 10},
	}

	var rows []table.Row
	for _, c := range data {
		name := getString(c, "name")
		region := getString(c, "region")

		// Mode de déploiement: derived from virtualHost presence
		virtualHost := getString(c, "virtualHost")
		deployMode := "Multi-AZ"
		if virtualHost == "" {
			deployMode = "Single-AZ"
		}

		// Offre: injected by fetchS3StorageData from the region service name
		offer := getString(c, "_offer")
		if offer == "" {
			offer = "Standard"
		}

		objectsCount := "-"
		if v, ok := c["objectsCount"]; ok {
			switch n := v.(type) {
			case float64:
				objectsCount = fmt.Sprintf("%d", int(n))
			}
		}

		sizeStr := "-"
		if v, ok := c["objectsSize"]; ok {
			switch n := v.(type) {
			case float64:
				if n < 1024 {
					sizeStr = fmt.Sprintf("%.0f B", n)
				} else if n < 1024*1024 {
					sizeStr = fmt.Sprintf("%.1f KB", n/1024)
				} else if n < 1024*1024*1024 {
					sizeStr = fmt.Sprintf("%.1f MB", n/1024/1024)
				} else {
					sizeStr = fmt.Sprintf("%.2f GB", n/1024/1024/1024)
				}
			}
		}

		// Type: from the container's own type field if present
		containerType := getString(c, "type")
		if containerType == "" {
			containerType = "-"
		}

		rows = append(rows, table.Row{name, region, deployMode, offer, objectsCount, sizeStr, containerType})
	}

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
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}
