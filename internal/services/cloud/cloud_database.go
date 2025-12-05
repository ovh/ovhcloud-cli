// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectDatabaseColumnsToDisplay = []string{"id", "engine", "version", "description", "status"}

	//go:embed templates/cloud_database.tmpl
	cloudDatabaseTemplate string

	//go:embed parameter-samples/database-create.json
	DatabaseCreationExample string

	DatabaseSpec struct {
		Backups struct {
			Regions []string `json:"regions,omitempty"`
			Time    string   `json:"time,omitempty"`
		} `json:"backups,omitzero"`
		DeletionProtection bool   `json:"deletionProtection,omitempty"`
		Description        string `json:"description,omitempty"`
		Disk               struct {
			Size int `json:"size,omitempty"`
		} `json:"disk,omitzero"`
		EnablePrometheus bool `json:"enablePrometheus,omitempty"`
		ForkFrom         struct {
			BackupID    string `json:"backupId,omitempty"`
			PointInTime string `json:"pointInTime,omitempty"`
			ServiceID   string `json:"serviceId,omitempty"`
		} `json:"forkFrom,omitzero"`
		IPRestrictions  []databaseIPRestriction `json:"ipRestrictions,omitempty"`
		MaintenanceTime string                  `json:"maintenanceTime,omitempty"`
		NetworkID       string                  `json:"networkId,omitempty"`
		NodesList       []databaseNode          `json:"nodesList,omitempty"`
		NodesPattern    struct {
			Flavor string `json:"flavor,omitempty"`
			Number int    `json:"number,omitempty"`
			Region string `json:"region,omitempty"`
		} `json:"nodesPattern,omitzero"`
		Plan     string `json:"plan,omitempty"`
		SubnetID string `json:"subnetId,omitempty"`
		Version  string `json:"version,omitempty"`

		// Extra fields for CLI only
		Engine            string   `json:"-"`
		CLIIPRestrictions []string `json:"-"`
		CLINodesList      []string `json:"-"`
	}

	DatabaseDatabaseSpec struct {
		Name string `json:"name"`
	}
)

type (
	databaseIPRestriction struct {
		Description string `json:"description,omitempty"`
		IP          string `json:"ip,omitempty"`
	}

	databaseNode struct {
		Flavor string `json:"flavor,omitempty"`
		Region string `json:"region,omitempty"`
		Role   string `json:"role,omitempty"`
	}
)

func ListCloudDatabases(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequest(fmt.Sprintf("/v1/cloud/project/%s/database/service", projectID), "", cloudprojectDatabaseColumnsToDisplay, flags.GenericFilters)
}

func GetCloudDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/database/service", projectID), args[0], cloudDatabaseTemplate)
}

func CreateDatabase(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Parse IP restrictions
	for _, restriction := range DatabaseSpec.CLIIPRestrictions {
		DatabaseSpec.IPRestrictions = append(DatabaseSpec.IPRestrictions, databaseIPRestriction{IP: restriction})
	}

	// Parse nodes list
	for _, node := range DatabaseSpec.CLINodesList {
		parts := strings.Split(node, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid node format: %s (expected format: flavor1:region1,flavor2:region2...)", node)
			return
		}
		DatabaseSpec.NodesList = append(DatabaseSpec.NodesList, databaseNode{
			Flavor: parts[0],
			Region: parts[1],
		})
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s", projectID, url.PathEscape(DatabaseSpec.Engine))
	database, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/database/"+url.PathEscape(DatabaseSpec.Engine),
		endpoint,
		DatabaseCreationExample,
		DatabaseSpec,
		assets.CloudOpenapiSchema,
		[]string{"version", "plan"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, database, "✅ Database created successfully (id: %s)", database["id"])
}

func DeleteDatabase(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	// Delete the database
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database deleted successfully")
}

func EditDatabase(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	// Parse IP restrictions
	for _, restriction := range DatabaseSpec.CLIIPRestrictions {
		DatabaseSpec.IPRestrictions = append(DatabaseSpec.IPRestrictions, databaseIPRestriction{IP: restriction})
	}

	// Edit resource
	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/database/"+url.PathEscape(databaseService["engine"].(string))+"/{clusterId}",
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0])),
		DatabaseSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateDatabaseInDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/database",
		projectID,
		url.PathEscape(databaseService["engine"].(string)),
		url.PathEscape(args[0]),
	)

	var responseBody map[string]any
	if err := httpLib.Client.Post(endpoint, &DatabaseDatabaseSpec, &responseBody); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, responseBody, "✅ Database created successfully (id: %s)", responseBody["id"])
}

func DeleteDatabaseInDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	// Delete the database in the given database cluster
	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/database/%s",
		projectID,
		url.PathEscape(databaseService["engine"].(string)),
		url.PathEscape(args[0]),
		url.PathEscape(args[1]),
	)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database deleted successfully")
}

func ListDatabasesInDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/database", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0])),
		"",
		[]string{"id", "name", "default"},
		flags.GenericFilters,
	)
}

func GetDatabaseInDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch database service to retrieve the engine
	var databaseService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &databaseService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/database", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0])),
		args[1],
		"",
	)
}
