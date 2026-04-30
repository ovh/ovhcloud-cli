// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"

	tea "github.com/charmbracelet/bubbletea"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

type volumeBackupCreatedMsg struct {
	name       string
	backupType string 
	err        error
}

// createVolumeBackupOrSnapshot dispatches to snapshot or backup creation.
func (m Model) createVolumeBackupOrSnapshot(volumeID, region, name string, isSnapshot bool) tea.Cmd {
	if isSnapshot {
		return m.createVolumeSnapshot(volumeID, name)
	}
	return m.createVolumeBackup(volumeID, region, name)
}

func (m Model) createVolumeSnapshot(volumeID, name string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return volumeBackupCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/%s/snapshot",
			m.cloudProject, url.PathEscape(volumeID))
		body := map[string]interface{}{"name": name}
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return volumeBackupCreatedMsg{name: name, backupType: "snapshot", err: fmt.Errorf("failed to create snapshot: %w", err)}
		}
		return volumeBackupCreatedMsg{name: name, backupType: "snapshot"}
	}
}

func (m Model) createVolumeBackup(volumeID, region, name string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return volumeBackupCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/volumeBackup",
			m.cloudProject, url.PathEscape(region))
		body := map[string]interface{}{
			"name":     name,
			"volumeId": volumeID,
		}
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return volumeBackupCreatedMsg{name: name, backupType: "backup", err: fmt.Errorf("failed to create backup: %w", err)}
		}
		return volumeBackupCreatedMsg{name: name, backupType: "backup"}
	}
}

// fetchBackupVolumes fetches available block storage volumes to use as backup source.
func (m Model) fetchBackupVolumes() tea.Cmd {
	return func() tea.Msg {
		msg := m.fetchBlockStorageData()
		return backupVolumesLoadedMsg{volumes: msg.data, err: msg.err}
	}
}

type backupVolumesLoadedMsg struct {
	volumes []map[string]interface{}
	err     error
}
