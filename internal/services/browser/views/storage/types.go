// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package storage

import "github.com/ovh/ovhcloud-cli/internal/services/browser/views"

// StorageType represents the type of storage being viewed
type StorageType int

const (
	StorageTypeS3 StorageType = iota
	StorageTypeSwift
	StorageTypeBlock
)

// String returns the display name for the storage type
func (s StorageType) String() string {
	switch s {
	case StorageTypeS3:
		return "S3"
	case StorageTypeSwift:
		return "Swift"
	case StorageTypeBlock:
		return "Block Storage"
	default:
		return "Unknown"
	}
}

// Path returns the API path for the storage type
func (s StorageType) Path() string {
	switch s {
	case StorageTypeS3:
		return "/storage/s3"
	case StorageTypeSwift:
		return "/storage/swift"
	case StorageTypeBlock:
		return "/storage/block"
	default:
		return "/storage/s3"
	}
}

// Icon returns the icon for the storage type
func (s StorageType) Icon() string {
	switch s {
	case StorageTypeS3:
		return "📦"
	case StorageTypeSwift:
		return "🌀"
	case StorageTypeBlock:
		return "💿"
	default:
		return "💾"
	}
}

// Messages for storage view communication

// ShowStorageDetailMsg signals to show detail view for a storage item
type ShowStorageDetailMsg struct {
	Item        map[string]interface{}
	StorageType StorageType
}

// SwitchStorageTypeMsg signals to switch to a different storage type
type SwitchStorageTypeMsg struct {
	StorageType StorageType
}

// ExecuteStorageActionMsg signals to execute an action on a storage item
type ExecuteStorageActionMsg struct {
	Item        map[string]interface{}
	StorageType StorageType
	Action      int
}

// RefreshStorageMsg signals to refresh the current storage view
type RefreshStorageMsg struct{}

// Object-level messages for browsing and actions

// ShowStorageObjectsMsg signals to show objects view for a container
type ShowStorageObjectsMsg struct {
	Container   map[string]interface{}
	StorageType StorageType
}

// ShowStorageObjectDetailMsg signals to show detail view for an object
type ShowStorageObjectDetailMsg struct {
	Object      map[string]interface{}
	Container   map[string]interface{}
	StorageType StorageType
}

// StorageObjectActionType represents object-level actions
type StorageObjectActionType int

const (
	ObjectActionCopy StorageObjectActionType = iota
	ObjectActionRestore
	ObjectActionDelete
	ObjectActionPresignedURL
	ObjectActionViewVersions
)

// ExecuteStorageObjectActionMsg signals to execute an action on an object
type ExecuteStorageObjectActionMsg struct {
	Object      map[string]interface{}
	Container   map[string]interface{}
	StorageType StorageType
	Action      StorageObjectActionType
}

// StorageObjectsCopyMsg is sent when an object copy operation completes
type StorageObjectsCopyMsg struct {
	Object map[string]interface{}
	Err    error
}

// StorageObjectsRestoreMsg is sent when an object restore operation completes
type StorageObjectsRestoreMsg struct {
	Object map[string]interface{}
	Err    error
}

// StorageObjectsLoadedMsg is sent when objects are loaded for a container
type StorageObjectsLoadedMsg struct {
	Container map[string]interface{}
	Objects   []map[string]interface{}
	Err       error
}

// Helper functions

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getInt safely extracts an int from a map (handles float64 from JSON)
func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return views.StyleValue.Render(string(rune(bytes)) + " B")
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return views.StyleValue.Render(string(rune(bytes/div)) + " " + "KMGTPE"[exp:exp+1] + "iB")
}
