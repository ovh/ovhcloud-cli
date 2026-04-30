// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	block_storage "github.com/ovh/ovhcloud-cli/internal/services/browser/views/block_storage"
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

// ─── Snapshot actions ─────────────────────────────────────────────────────────

type snapshotActionDoneMsg struct {
	action int
	name   string
	err    error
}

func (m Model) executeSnapshotAction(msg block_storage.ExecuteSnapshotActionMsg) tea.Cmd {
	switch msg.Action {
	case block_storage.SnapshotActionDelete:
		return m.deleteSnapshot(msg.Snapshot)
	case block_storage.SnapshotActionCreateVolume:
		return m.createVolumeFromSnapshot(msg.Snapshot, msg.VolumeName, msg.VolumeSize)
	}
	return nil
}

func (m Model) deleteSnapshot(snapshot map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return snapshotActionDoneMsg{action: block_storage.SnapshotActionDelete, err: fmt.Errorf("no cloud project selected")}
		}
		id := fmt.Sprintf("%v", snapshot["id"])
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/snapshot/%s", m.cloudProject, url.PathEscape(id))
		if err := httpLib.Client.Delete(endpoint, nil); err != nil {
			return snapshotActionDoneMsg{action: block_storage.SnapshotActionDelete, err: err}
		}
		name := fmt.Sprintf("%v", snapshot["name"])
		return snapshotActionDoneMsg{action: block_storage.SnapshotActionDelete, name: name}
	}
}

func (m Model) createVolumeFromSnapshot(snapshot map[string]interface{}, volName, volSize string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return snapshotActionDoneMsg{action: block_storage.SnapshotActionCreateVolume, err: fmt.Errorf("no cloud project selected")}
		}
		snapshotID := fmt.Sprintf("%v", snapshot["id"])
		sourceVolumeID := fmt.Sprintf("%v", snapshot["volumeId"])
		region := fmt.Sprintf("%v", snapshot["region"])
		sizeInt, _ := strconv.Atoi(volSize)
		if sizeInt <= 0 {
			sizeInt = 10
		}
		// Fetch source volume type so we use the correct type
		volType := "classic"
		if sourceVolumeID != "" && sourceVolumeID != "<nil>" {
			var sourceVol map[string]interface{}
			srcEndpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/%s", m.cloudProject, url.PathEscape(sourceVolumeID))
			if err := httpLib.Client.Get(srcEndpoint, &sourceVol); err == nil {
				if t, ok := sourceVol["type"].(string); ok && t != "" {
					volType = t
				}
			}
		}
		body := map[string]interface{}{
			"name":       volName,
			"region":     region,
			"size":       sizeInt,
			"type":       volType,
			"snapshotId": snapshotID,
		}
		var result map[string]interface{}
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/volume", m.cloudProject)
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return snapshotActionDoneMsg{action: block_storage.SnapshotActionCreateVolume, err: err}
		}
		return snapshotActionDoneMsg{action: block_storage.SnapshotActionCreateVolume, name: volName}
	}
}

// ─── Backup actions ───────────────────────────────────────────────────────────

type backupActionDoneMsg struct {
	action int
	name   string
	err    error
}

func (m Model) executeBackupAction(msg block_storage.ExecuteBackupActionMsg) tea.Cmd {
	switch msg.Action {
	case block_storage.BackupActionDelete:
		return m.deleteBackup(msg.Backup)
	case block_storage.BackupActionRestore:
		return m.restoreBackup(msg.Backup, msg.VolumeID)
	case block_storage.BackupActionCreateVolume:
		return m.createVolumeFromBackup(msg.Backup, msg.VolumeName, msg.VolumeSize)
	}
	return nil
}

func (m Model) deleteBackup(backup map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return backupActionDoneMsg{action: block_storage.BackupActionDelete, err: fmt.Errorf("no cloud project selected")}
		}
		id := fmt.Sprintf("%v", backup["id"])
		region := fmt.Sprintf("%v", backup["region"])
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/volumeBackup/%s",
			m.cloudProject, url.PathEscape(region), url.PathEscape(id))
		if err := httpLib.Client.Delete(endpoint, nil); err != nil {
			return backupActionDoneMsg{action: block_storage.BackupActionDelete, err: err}
		}
		name := fmt.Sprintf("%v", backup["name"])
		return backupActionDoneMsg{action: block_storage.BackupActionDelete, name: name}
	}
}

func (m Model) restoreBackup(backup map[string]interface{}, volumeID string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return backupActionDoneMsg{action: block_storage.BackupActionRestore, err: fmt.Errorf("no cloud project selected")}
		}
		id := fmt.Sprintf("%v", backup["id"])
		region := fmt.Sprintf("%v", backup["region"])
		endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/volumeBackup/%s/restore",
			m.cloudProject, url.PathEscape(region), url.PathEscape(id))
		body := map[string]interface{}{"volumeId": volumeID}
		var result map[string]interface{}
		if err := httpLib.Client.Post(endpoint, body, &result); err != nil {
			return backupActionDoneMsg{action: block_storage.BackupActionRestore, err: err}
		}
		name := fmt.Sprintf("%v", backup["name"])
		return backupActionDoneMsg{action: block_storage.BackupActionRestore, name: name}
	}
}

func (m Model) createVolumeFromBackup(backup map[string]interface{}, volName, volSize string) tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return backupActionDoneMsg{action: block_storage.BackupActionCreateVolume, err: fmt.Errorf("no cloud project selected")}
		}
		id := fmt.Sprintf("%v", backup["id"])
		region := fmt.Sprintf("%v", backup["region"])
		sourceVolumeID := fmt.Sprintf("%v", backup["volumeId"])
		sizeInt, _ := strconv.Atoi(volSize)
		if sizeInt <= 0 {
			sizeInt = 10
		}
		// Fetch source volume type to preserve it
		volType := "classic"
		if sourceVolumeID != "" && sourceVolumeID != "<nil>" {
			var sourceVol map[string]interface{}
			srcEndpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/%s", m.cloudProject, url.PathEscape(sourceVolumeID))
			if err := httpLib.Client.Get(srcEndpoint, &sourceVol); err == nil {
				if t, ok := sourceVol["type"].(string); ok && t != "" {
					volType = t
				}
			}
		}
		// Step 1: create a new empty volume
		volBody := map[string]interface{}{
			"name":   volName,
			"region": region,
			"size":   sizeInt,
			"type":   volType,
		}
		var newVol map[string]interface{}
		volEndpoint := fmt.Sprintf("/v1/cloud/project/%s/volume", m.cloudProject)
		if err := httpLib.Client.Post(volEndpoint, volBody, &newVol); err != nil {
			return backupActionDoneMsg{action: block_storage.BackupActionCreateVolume, err: fmt.Errorf("failed to create volume: %w", err)}
		}
		newVolID := fmt.Sprintf("%v", newVol["id"])
		// Wait for the new volume to become available before restoring
		volDetailEndpoint := fmt.Sprintf("/v1/cloud/project/%s/volume/%s", m.cloudProject, url.PathEscape(newVolID))
		for i := 0; i < 60; i++ {
			time.Sleep(3 * time.Second)
			var volStatus map[string]interface{}
			if err := httpLib.Client.Get(volDetailEndpoint, &volStatus); err == nil {
				if status, _ := volStatus["status"].(string); status == "available" {
					break
				}
			}
		}
		// Step 2: restore backup onto the new volume
		restoreEndpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/volumeBackup/%s/restore",
			m.cloudProject, url.PathEscape(region), url.PathEscape(id))
		restoreBody := map[string]interface{}{"volumeId": newVolID}
		var restoreResult map[string]interface{}
		if err := httpLib.Client.Post(restoreEndpoint, restoreBody, &restoreResult); err != nil {
			return backupActionDoneMsg{action: block_storage.BackupActionCreateVolume, err: fmt.Errorf("volume created but restore failed: %w", err)}
		}
		return backupActionDoneMsg{action: block_storage.BackupActionCreateVolume, name: volName}
	}
}

// fetchVolumesForRegion fetches volumes in a given region for the restore picker.
func (m Model) fetchVolumesForRegion(region string) tea.Cmd {
	return func() tea.Msg {
		msg := m.fetchBlockStorageData()
		if msg.err != nil {
			return block_storage.BackupVolumesLoadedMsg{Volumes: nil}
		}
		// Filter to same region
		var filtered []map[string]interface{}
		for _, v := range msg.data {
			if fmt.Sprintf("%v", v["region"]) == region {
				filtered = append(filtered, v)
			}
		}
		return block_storage.BackupVolumesLoadedMsg{Volumes: filtered}
	}
}
