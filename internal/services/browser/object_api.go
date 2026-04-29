// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// fetchSwiftRegions loads available regions for Swift containers.
func (m Model) fetchSwiftRegions() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return objectStorageInitDataLoadedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		var regionNames []string
		if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/region", m.cloudProject), &regionNames); err != nil {
			return swiftRegionsLoadedMsg{regions: nil}
		}
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
						if n, _ := sm["name"].(string); n == "storage" || n == "storage-object" {
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
		return swiftRegionsLoadedMsg{regions: supportedRegions}
	}
}

func (m Model) createObjectContainer() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return objectContainerCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		if m.wizard.objectTypeIdx == 1 {
			body := map[string]interface{}{
				"containerName": m.wizard.objectName,
				"region":        m.wizard.objectSwiftRegion,
			}
			body["archive"] = (m.wizard.objectSwiftTypeIdx == 1)
			endpoint := fmt.Sprintf("/v1/cloud/project/%s/storage", m.cloudProject)
			var container map[string]interface{}
			if err := httpLib.Client.Post(endpoint, body, &container); err != nil {
				return objectContainerCreatedMsg{err: fmt.Errorf("failed to create Swift container: %w", err)}
			}
			return objectContainerCreatedMsg{container: container}
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

// createS3User creates a new cloud user with objectstore access, then creates S3 credentials for it.
func (m Model) createS3User() tea.Cmd {
	return func() tea.Msg {
		if m.cloudProject == "" {
			return s3UserCreatedMsg{err: fmt.Errorf("no cloud project selected")}
		}

		body := map[string]interface{}{
			"description": m.wizard.s3UserDesc,
			"role":        "objectstore_operator",
		}
		var user map[string]interface{}
		userEndpoint := fmt.Sprintf("/v1/cloud/project/%s/user", m.cloudProject)
		if err := httpLib.Client.Post(userEndpoint, body, &user); err != nil {
			return s3UserCreatedMsg{err: fmt.Errorf("failed to create user: %w", err)}
		}
		var userId int64
		switch v := user["id"].(type) {
		case float64:
			userId = int64(v)
		case int64:
			userId = v
		case int:
			userId = int64(v)
		case json.Number:
			userId, _ = v.Int64()
		default:
			idStr := fmt.Sprintf("%v", v)
			if _, err := fmt.Sscanf(idStr, "%d", &userId); err != nil {
				return s3UserCreatedMsg{user: user, err: fmt.Errorf("could not parse user id %q: %w", idStr, err)}
			}
		}

		// Poll until user status is "ok" (max ~30s)
		userGetEndpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%d", m.cloudProject, userId)
		for i := 0; i < 30; i++ {
			var u map[string]interface{}
			if err := httpLib.Client.Get(userGetEndpoint, &u); err == nil {
				if status, _ := u["status"].(string); status == "ok" {
					user = u
					break
				}
			}
			time.Sleep(1 * time.Second)
		}

		// Create S3 credentials for the user
		var credentials map[string]interface{}
		credsEndpoint := fmt.Sprintf("/v1/cloud/project/%s/user/%d/s3Credentials", m.cloudProject, userId)
		if err := httpLib.Client.Post(credsEndpoint, nil, &credentials); err != nil {
			return s3UserCreatedMsg{user: user, err: fmt.Errorf("user created but S3 credentials failed: %w", err)}
		}

		return s3UserCreatedMsg{user: user, credentials: credentials}
	}
}

// saveAWSCredentials appends the S3 credentials to ~/.aws/credentials under a new profile.
func saveAWSCredentials(accessKey, secretKey, username string) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return s3CredentialsSavedMsg{err: fmt.Errorf("could not find home directory: %w", err)}
		}

		awsDir := filepath.Join(homeDir, ".aws")
		if err := os.MkdirAll(awsDir, 0700); err != nil {
			return s3CredentialsSavedMsg{err: fmt.Errorf("could not create ~/.aws directory: %w", err)}
		}

		credPath := filepath.Join(awsDir, "credentials")

		profileName := "ovhcloud"
		if username != "" {
			profileName = "ovhcloud-" + username
		}

		existingContent := ""
		if data, err := os.ReadFile(credPath); err == nil {
			existingContent = string(data)
		}

		if strings.Contains(existingContent, "["+profileName+"]") {
			for i := 2; i < 100; i++ {
				candidate := fmt.Sprintf("%s-%d", profileName, i)
				if !strings.Contains(existingContent, "["+candidate+"]") {
					profileName = candidate
					break
				}
			}
		}

		newEntry := fmt.Sprintf("\n[%s]\naws_access_key_id = %s\naws_secret_access_key = %s\n",
			profileName, accessKey, secretKey)

		flags := os.O_WRONLY | os.O_APPEND
		if existingContent == "" {
			flags = os.O_WRONLY | os.O_CREATE
		}
		f, err := os.OpenFile(credPath, flags, 0600)
		if err != nil {
			return s3CredentialsSavedMsg{err: fmt.Errorf("could not open credentials file: %w", err)}
		}
		defer f.Close()

		if _, err := f.WriteString(newEntry); err != nil {
			return s3CredentialsSavedMsg{err: fmt.Errorf("could not write credentials: %w", err)}
		}

		return s3CredentialsSavedMsg{filePath: credPath, profileName: profileName}
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
		m.notification = fmt.Sprintf("❌ Action failed: %s", msg.err.Error())
		m.notificationExpiry = time.Now().Add(8 * time.Second)
		return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg {
			return clearNotificationMsg{}
		})
	}
	m.notification = "✅ Container deleted successfully!"
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

func (m Model) handleS3UserCreated(msg s3UserCreatedMsg) (tea.Model, tea.Cmd) {
	m.wizard.isLoading = false
	m.wizard.loadingMessage = ""
	if msg.err != nil {
		m.wizard.errorMsg = msg.err.Error()
		return m, nil
	}
	m.s3CreatedUser = msg.user
	m.s3CreatedCredentials = msg.credentials
	m.s3CredentialsSavedPath = ""
	m.s3CredentialsSaveError = ""
	m.wizard = WizardData{}
	m.mode = S3CredentialsView
	return m, nil
}

func (m Model) handleS3CredentialsSaved(msg s3CredentialsSavedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.s3CredentialsSaveError = msg.err.Error()
	} else {
		m.s3CredentialsSavedPath = fmt.Sprintf("%s  (profile: [%s])", msg.filePath, msg.profileName)
	}
	return m, nil
}

// createObjectStorageTable builds the table for S3 containers.
func createObjectStorageTable(data []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 28},
		{Title: "Location", Width: 12},
		{Title: "Deployment mode", Width: 22},
		{Title: "Offer", Width: 16},
		{Title: "Number of objects", Width: 25},
		{Title: "Space Used", Width: 14},
		{Title: "Type", Width: 10},
	}

	var rows []table.Row
	for _, c := range data {
		name := getString(c, "name")
		region := getString(c, "region")

		deployMode := getString(c, "_deployMode")
		if deployMode == "" {
			deployMode = "-"
		}

		category := getString(c, "_type")

		offer := "-"
		if category == "Swift" {
			offer = "Swift"
		} else if category == "S3" {
			s3Offer := getString(c, "_offer")
			if s3Offer != "" {
				offer = s3Offer
			} else {
				offer = "S3 Compatible"
			}
		}

		// Type: only Swift has a meaningful type (Private/Public/Static); '-' for S3
		containerType := "-"
		if category == "Swift" {
			swiftType := getString(c, "containerType")
			switch swiftType {
			case "private":
				containerType = "Private"
			case "public":
				containerType = "Public"
			case "static":
				containerType = "Static"
			default:
				if swiftType != "" {
					containerType = swiftType
				}
			}
		}

		objectsCount := "-"
		for _, field := range []string{"objectsCount", "storedObjects"} {
			if v, ok := c[field]; ok {
				if n, ok := v.(float64); ok {
					objectsCount = fmt.Sprintf("%d", int(n))
					break
				}
			}
		}
		sizeStr := "-"
		for _, field := range []string{"objectsSize", "storedBytes"} {
			if v, ok := c[field]; ok {
				if n, ok := v.(float64); ok {
					if n < 1024 {
						sizeStr = fmt.Sprintf("%.0f B", n)
					} else if n < 1024*1024 {
						sizeStr = fmt.Sprintf("%.1f KB", n/1024)
					} else if n < 1024*1024*1024 {
						sizeStr = fmt.Sprintf("%.1f MB", n/1024/1024)
					} else {
						sizeStr = fmt.Sprintf("%.2f GB", n/1024/1024/1024)
					}
					break
				}
			}
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

// createObjectStorageUsersTable creates a table to display S3 users/credentials.
func createObjectStorageUsersTable(users []map[string]interface{}, width, height int) table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Description", Width: 40},
		{Title: "Access Key", Width: 32},
		{Title: "User S3", Width: 20},
	}

	var rows []table.Row
	for _, s3Cred := range users {
		// Name: internalName from credentials, fallback to username
		name := getString(s3Cred, "internalName")
		if name == "" {
			name = getString(s3Cred, "_username")
		}

		// Description from user info
		description := getString(s3Cred, "_userDescription")
		if description == "" {
			description = "-"
		}

		// Access key
		accessKey := getString(s3Cred, "access")
		if accessKey == "" {
			accessKey = "No credentials"
		}

		// S3 User: enabled status
		s3User := "Enabled"
		if accessKey == "No credentials" {
			s3User = "Disabled"
		}

		rows = append(rows, table.Row{name, description, accessKey, s3User})
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
