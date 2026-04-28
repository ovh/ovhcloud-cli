// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// ─── File Storage (NFS Share) API ─────────────────────────────────────────────

// fetchFileShareRegions probes each region concurrently for file storage support.
func (m Model) fetchFileShareRegions() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return fileShareRegionsLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		var regionNames []string
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
		if err := httpLib.Client.Get(endpoint, &regionNames); err != nil {
			return fileShareRegionsLoadedMsg{err: fmt.Errorf("failed to fetch regions: %w", err)}
		}

		type probeResult struct {
			region    string
			supported bool
		}
		ch := make(chan probeResult, len(regionNames))
		for _, name := range regionNames {
			go func(regionName string) {
				probe := fmt.Sprintf("/v1/cloud/project/%s/region/%s/share",
					m.cloudProject, url.PathEscape(regionName))
				var result []map[string]interface{}
				err := httpLib.Client.Get(probe, &result)
				ch <- probeResult{region: regionName, supported: err == nil}
			}(name)
		}

		var supported []string
		for range regionNames {
			r := <-ch
			if r.supported {
				supported = append(supported, r.region)
			}
		}
		sort.Strings(supported)

		if len(supported) == 0 {
			return fileShareRegionsLoadedMsg{err: fmt.Errorf("no regions support file storage in this project")}
		}
		return fileShareRegionsLoadedMsg{regions: supported}
	}
}

// fetchFileShareNetworks fetches private networks available for the selected region,
// pre-extracting the OpenStack UUID needed by the file share API.
func (m Model) fetchFileShareNetworks() tea.Cmd {
	region := m.wizard.selectedRegion
	return func() tea.Msg {
		if m.cloudProject == "" {
			return fileShareNetworksLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		var networks []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private", m.cloudProject)
		if err := httpLib.Client.Get(endpoint, &networks); err != nil {
			return fileShareNetworksLoadedMsg{err: fmt.Errorf("failed to fetch networks: %w", err)}
		}
		for i, net := range networks {
			openstackId := ""
			if regions, ok := net["regions"].([]interface{}); ok {
				for _, r := range regions {
					if rm, ok := r.(map[string]interface{}); ok {
						if rm["region"] == region {
							openstackId, _ = rm["openstackId"].(string)
							break
						}
					}
				}
			}
			if openstackId != "" {
				networks[i]["_openstackId"] = openstackId
			}
		}
		var filtered []map[string]interface{}
		for _, net := range networks {
			if _, ok := net["_openstackId"]; ok {
				filtered = append(filtered, net)
			}
		}
		sort.Slice(filtered, func(i, j int) bool {
			iName, _ := filtered[i]["name"].(string)
			jName, _ := filtered[j]["name"].(string)
			return iName < jName
		})
		return fileShareNetworksLoadedMsg{networks: filtered}
	}
}

// fetchFileShareSubnets fetches subnets for a private network.
func (m Model) fetchFileShareSubnets(networkID string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return fileShareSubnetsLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		var subnets []map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet", m.cloudProject, networkID)
		if err := httpLib.Client.Get(endpoint, &subnets); err != nil {
			return fileShareSubnetsLoadedMsg{err: fmt.Errorf("failed to fetch subnets: %w", err)}
		}
		sort.Slice(subnets, func(i, j int) bool {
			iCIDR, _ := subnets[i]["cidr"].(string)
			jCIDR, _ := subnets[j]["cidr"].(string)
			return iCIDR < jCIDR
		})
		return fileShareSubnetsLoadedMsg{subnets: subnets}
	}
}

// createFileShare creates a new NFS file share via the OVH API.
func (m Model) createFileShare() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return fileShareCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		body := map[string]interface{}{
			"name":      m.wizard.fileShareName,
			"type":      m.wizard.fileShareType,
			"size":      m.wizard.fileShareSize,
			"networkId": m.wizard.fileShareNetworkId,
			"subnetId":  m.wizard.fileShareSubnetId,
		}
		var share map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/share",
			m.cloudProject, url.PathEscape(m.wizard.selectedRegion))
		if err := httpLib.Client.Post(endpoint, body, &share); err != nil {
			return fileShareCreatedMsg{err: fmt.Errorf("failed to create file share: %w", err)}
		}
		return fileShareCreatedMsg{share: share}
	}
}

// fetchFileStorageData returns all NFS shares across every region.
func (m Model) fetchFileStorageData() dataLoadedMsg {
	if m.cloudProject == "" {
		return dataLoadedMsg{err: fmt.Errorf("no cloud project selected")}
	}
	var regionNames []string
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject)
	if err := httpLib.Client.Get(endpoint, &regionNames); err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to fetch regions: %w", err)}
	}

	type regionResult struct{ shares []map[string]interface{} }
	ch := make(chan regionResult, len(regionNames))
	for _, name := range regionNames {
		go func(regionName string) {
			probe := fmt.Sprintf("/v1/cloud/project/%s/region/%s/share",
				m.cloudProject, url.PathEscape(regionName))
			var shares []map[string]interface{}
			if err := httpLib.Client.Get(probe, &shares); err != nil {
				ch <- regionResult{}
				return
			}
			for i := range shares {
				shares[i]["region"] = regionName
			}
			ch <- regionResult{shares: shares}
		}(name)
	}

	var allShares []map[string]interface{}
	for range regionNames {
		r := <-ch
		allShares = append(allShares, r.shares...)
	}
	sort.Slice(allShares, func(i, j int) bool {
		iName, _ := allShares[i]["name"].(string)
		jName, _ := allShares[j]["name"].(string)
		return iName < jName
	})
	return dataLoadedMsg{data: allShares}
}

// createFileStorageTable builds the table model for file shares.
func createFileStorageTable(data []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Nom", Width: 25},
		{Title: "ID", Width: 20},
		{Title: "Région", Width: 12},
		{Title: "Type", Width: 16},
		{Title: "Capacité", Width: 12},
		{Title: "Statut", Width: 12},
	}

	var rows []table.Row
	for _, share := range data {
		name := getString(share, "name")
		id := getString(share, "id")
		region := getString(share, "region")
		shareType := getString(share, "type")
		status := getString(share, "status")

		sizeStr := "-"
		if sz, ok := share["size"]; ok {
			switch v := sz.(type) {
			case float64:
				sizeStr = fmt.Sprintf("%d GB", int(v))
			case json.Number:
				if f, err := v.Float64(); err == nil {
					sizeStr = fmt.Sprintf("%d GB", int(f))
				}
			case int:
				sizeStr = fmt.Sprintf("%d GB", v)
			}
		}
		rows = append(rows, table.Row{name, id, region, shareType, sizeStr, status})
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

// ─── File Storage message handlers ────────────────────────────────────────────

func (m Model) handleFileShareRegionsLoaded(msg fileShareRegionsLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	m.wizard.fileShareRegions = msg.regions
	if len(msg.regions) == 0 {
		m.wizard.errorMsg = "No regions support file storage in this project"
		return m, nil
	}
	m.wizard.selectedIndex = 0
	return m, nil
}

func (m Model) handleFileShareNetworksLoaded(msg fileShareNetworksLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	m.wizard.fileShareNetworks = msg.networks
	m.wizard.selectedIndex = 0
	return m, nil
}

func (m Model) handleFileShareSubnetsLoaded(msg fileShareSubnetsLoadedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	m.wizard.fileShareSubnets = msg.subnets
	m.wizard.selectedIndex = 0
	return m, nil
}

func (m Model) handleFileShareCreated(msg fileShareCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	name := getString(msg.share, "name")
	if name == "" {
		name = "file share"
	}
	m.notification = fmt.Sprintf("✅ File share '%s' created successfully!", name)
	m.notificationExpiry = time.Now().Add(5 * time.Second)
	m.wizard = WizardData{}
	m.mode = LoadingView
	return m, tea.Batch(
		m.fetchDataForPath("/storage/file"),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		}),
	)
}
