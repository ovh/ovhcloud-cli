// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	ManagedDatabaseValidEngines              = []string{"mongodb", "mysql", "postgresql", "valkey"}
	ManagedDatabaseDatabaseValidEngines      = []string{"mysql", "postgresql"}
	ManagedDatabaseUserEditValidEngines      = []string{"mongodb", "postgresql", "valkey"}
	ManagedDatabaseUserValidEngines          = append(ManagedDatabaseUserEditValidEngines, "mysql")
	ManagedDatabaseRoleValidEngines          = []string{"mongodb", "postgresql"}
	ManagedDatabaseCertificateValidEngines   = []string{"mysql", "postgresql"}
	ManagedDatabaseBackupRestoreValidEngines = []string{"mongodb"}

	cloudprojectManagedDatabaseColumnsToDisplay             = []string{"id", "engine", "currentState.version version", "currentState.description description", "currentState.status status"}
	cloudprojectManagedDatabaseDatabaseColumnsToDisplay     = []string{"id", "name", "default"}
	cloudprojectManagedDatabaseUserWithRoleColumnsToDisplay = []string{"id", "username", "status", "roles", "createdAt"}
	cloudprojectManagedDatabaseUserColumnsToDisplay         = []string{"id", "username", "status", "createdAt"}
	cloudprojectManagedDatabaseValkeyUserColumnsToDisplay   = []string{"id", "username", "status", "categories", "channels", "commands", "keys", "createdAt"}
	cloudprojectManagedDatabaseRoleColumnsToDisplay         = []string{"name"}
	cloudprojectManagedDatabaseBackupColumnsToDisplay       = []string{"id", "createdAt", "type", "status", "description"}

	//go:embed templates/cloud_managed_database.tmpl
	managedDatabaseTemplate string
	//go:embed templates/cloud_managed_database_database.tmpl
	managedDatabaseDatabaseTemplate string
	//go:embed templates/cloud_managed_database_user.tmpl
	managedDatabaseUserTemplate string
	//go:embed templates/cloud_managed_database_certificate.tmpl
	managedDatabaseCertificateTemplate string
	//go:embed templates/cloud_managed_database_backup.tmpl
	managedDatabaseBackupTemplate string

	//go:embed parameter-samples/managed-database-create.json
	ManagedDatabaseCreationExample string
	//go:embed parameter-samples/managed-database-user-create.json
	ManagedDatabaseUserCreationExample string

	ManagedDatabaseSpec struct {
		TargetSpec databaseAnalyticsTargetSpec `json:"targetSpec,omitzero"`

		// Extra fields for CLI only
		Engine            string   `json:"-"`
		CLIDiskSize       int      `json:"-"`
		CLIIPRestrictions []string `json:"-"`
		CLINodesList      []string `json:"-"`
	}

	// ManagedDatabaseEditSpec holds the parameters of the edit command, which
	// still uses API v1.
	ManagedDatabaseEditSpec struct {
		Backups struct {
			Regions []string `json:"regions,omitempty"`
			Time    string   `json:"time,omitempty"`
		} `json:"backups,omitzero"`
		DeletionProtection bool                           `json:"deletionProtection,omitempty"`
		Description        string                         `json:"description,omitempty"`
		EnablePrometheus   bool                           `json:"enablePrometheus,omitempty"`
		IPRestrictions     []managedDatabaseIPRestriction `json:"ipRestrictions,omitempty"`
		MaintenanceTime    string                         `json:"maintenanceTime,omitempty"`
		NodesPattern       struct {
			Flavor string `json:"flavor,omitempty"`
		} `json:"nodesPattern,omitzero"`
		Plan    string `json:"plan,omitempty"`
		Version string `json:"version,omitempty"`

		// Extra fields for CLI only
		CLIIPRestrictions []string `json:"-"`
		CLINetworkID      string   `json:"-"`
		CLISubnetID       string   `json:"-"`
		CLIPublicNetwork  bool     `json:"-"`
	}

	ManagedDatabaseDatabaseSpec struct {
		Name string `json:"name"`
	}

	ManagedDatabaseUserSpec ManagedDatabaseUser

	ManagedDatabaseRoleSpec struct {
		Advanced bool `json:"-"`
	}
)

type (
	managedDatabaseIPRestriction struct {
		Description string `json:"description,omitempty"`
		IP          string `json:"ip,omitempty"`
	}

	ManagedDatabaseUser struct {
		Name string `json:"name"`

		// PostgreSQL and MongoDB specific field
		Roles []string `json:"roles,omitempty"`

		// Valkey specific fields
		Categories []string `json:"categories,omitempty"`
		Channels   []string `json:"channels,omitempty"`
		Commands   []string `json:"commands,omitempty"`
		Keys       []string `json:"keys,omitempty"`
	}
)

// getManagedDatabaseEngine returns the engine of a managed database service,
// checked against the engines that support the command.
func getManagedDatabaseEngine(projectID, serviceID string, validEngines []string) (string, error) {
	engine, _, err := getDatabaseAnalyticsService(projectID, managedDatabaseCategory, serviceID, ManagedDatabaseValidEngines)
	if err != nil {
		return "", fmt.Errorf("failed to fetch database service: %w", err)
	}

	if !slices.Contains(validEngines, engine) {
		return "", fmt.Errorf("invalid engine %s, valid values: %s", engine, strings.Join(validEngines, ", "))
	}

	return engine, nil
}

func ListManagedDatabases(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	listDatabaseAnalyticsServices(projectID, managedDatabaseCategory, ManagedDatabaseValidEngines, cloudprojectManagedDatabaseColumnsToDisplay)
}

func GetManagedDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, databaseService, err := getDatabaseAnalyticsService(projectID, managedDatabaseCategory, args[0], ManagedDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch database service: %s", err)
		return
	}

	display.OutputObject(databaseService, args[0], managedDatabaseTemplate, &flags.OutputFormatConfig)
}

func CreateManagedDatabasePreRun(_ *cobra.Command, _ []string) error {
	if !slices.Contains(ManagedDatabaseValidEngines, ManagedDatabaseSpec.Engine) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", ManagedDatabaseSpec.Engine, strings.Join(ManagedDatabaseValidEngines, ", "))
		return fmt.Errorf("invalid engine %s", ManagedDatabaseSpec.Engine)
	}
	return nil
}

func CreateManagedDatabase(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	spec := ManagedDatabaseSpec
	spec.TargetSpec.Network.AuthorizedIps = newAuthorizedIps(spec.CLIIPRestrictions)
	if spec.CLIDiskSize > 0 {
		spec.TargetSpec.Storage.Size = &databaseAnalyticsUnitAndValue{Unit: "GB", Value: spec.CLIDiskSize}
	}
	if err := spec.TargetSpec.setNodes(spec.CLINodesList); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	managedDatabase, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedDatabase/%s", spec.Engine),
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, spec.Engine),
		ManagedDatabaseCreationExample,
		spec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.version", "targetSpec.plan", "targetSpec.flavor", "targetSpec.region"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create managed database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, managedDatabase, "✅ Managed database created successfully (id: %s)", managedDatabase["id"])
}

func EditManagedDatabase(cmd *cobra.Command, args []string) {
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

	if !slices.Contains(ManagedDatabaseValidEngines, databaseService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", databaseService["engine"].(string), strings.Join(ManagedDatabaseValidEngines, ", "))
		return
	}

	// Parse IP restrictions
	for _, restriction := range ManagedDatabaseEditSpec.CLIIPRestrictions {
		ManagedDatabaseEditSpec.IPRestrictions = append(ManagedDatabaseEditSpec.IPRestrictions, managedDatabaseIPRestriction{IP: restriction})
	}

	// Build extra network fields that bypass the OpenAPI filter
	var networkFields map[string]any
	if ManagedDatabaseEditSpec.CLIPublicNetwork {
		networkFields = map[string]any{
			"networkId": nil,
			"subnetId":  nil,
		}
	} else if ManagedDatabaseEditSpec.CLINetworkID != "" {
		networkFields = map[string]any{
			"networkId": ManagedDatabaseEditSpec.CLINetworkID,
			"subnetId":  ManagedDatabaseEditSpec.CLISubnetID,
		}
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0]))

	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}", url.PathEscape(databaseService["engine"].(string))),
		endpoint,
		&ManagedDatabaseEditSpec,
		assets.CloudOpenapiSchema,
		networkFields,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteManagedDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Managed database deleted successfully")
}

func ListManagedDatabaseDatabases(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "database"),
		cloudprojectManagedDatabaseDatabaseColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedDatabaseDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "database"),
		args[1],
		managedDatabaseDatabaseTemplate,
	)
}

func CreateManagedDatabaseDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var database map[string]any
	if err := httpLib.Client.Post(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "database"),
		&ManagedDatabaseDatabaseSpec,
		&database,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database created successfully (id: %s)", database["id"])
}

func DeleteManagedDatabaseDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseDatabaseValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "database", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database deleted successfully")
}

func ListManagedDatabaseUsers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user"),
		ManagedDatabaseUserSpec.displayByEngine(engine),
		flags.GenericFilters,
	)
}

func GetManagedDatabaseUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user"),
		args[1],
		managedDatabaseUserTemplate,
	)
}

func CreateManagedDatabaseUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := ManagedDatabaseUserSpec.validateForEngine(engine); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	user, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedDatabase/%s/{serviceId}/user", engine),
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user"),
		ManagedDatabaseUserCreationExample,
		ManagedDatabaseUserSpec.byEngine(engine),
		assets.CloudV2OpenapiSchema,
		nil)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, user, "✅ User created successfully (id: %s)", user["id"])
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚠️ Save this password now, it will not be shown again:\n\nPassword: %s", user["password"])
}

func EditManagedDatabaseUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserEditValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := ManagedDatabaseUserSpec.validateForEngine(engine); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userSpec := ManagedDatabaseUserSpec.byEngine(engine)

	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedDatabase/%s/{serviceId}/user/{userId}", engine),
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user", args[1]),
		userSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteManagedDatabaseUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User deleted successfully")
}

func ResetManagedDatabaseUserCredentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseUserValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	user, err := resetDatabaseAnalyticsUserCredentials(databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "user", args[1]))
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset user credentials: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User credentials reset successfully")
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚠️ Save this password now, it will not be shown again:\n\nPassword: %s", user["password"])
}

func (s ManagedDatabaseUser) validateForEngine(engine string) error {
	if len(s.Roles) > 0 && !slices.Contains(ManagedDatabaseRoleValidEngines, engine) {
		return fmt.Errorf("flag --roles is only supported for postgresql and mongodb engines")
	}
	if (len(s.Categories) > 0 || len(s.Channels) > 0 || len(s.Commands) > 0 || len(s.Keys) > 0) && engine != "valkey" {
		return fmt.Errorf("flags --categories, --channels, --commands, --keys are only supported for the valkey engine")
	}
	return nil
}

func (s ManagedDatabaseUser) byEngine(engine string) ManagedDatabaseUser {
	result := ManagedDatabaseUser{Name: s.Name}
	switch engine {
	case "valkey":
		result.Categories = s.Categories
		result.Channels = s.Channels
		result.Commands = s.Commands
		result.Keys = s.Keys
	default:
		if slices.Contains(ManagedDatabaseRoleValidEngines, engine) {
			result.Roles = s.Roles
		}
	}
	return result
}

func (s ManagedDatabaseUser) displayByEngine(engine string) []string {
	result := []string{}
	switch engine {
	case "mysql":
		result = cloudprojectManagedDatabaseUserColumnsToDisplay
	case "valkey":
		result = cloudprojectManagedDatabaseValkeyUserColumnsToDisplay
	default:
		if slices.Contains(ManagedDatabaseRoleValidEngines, engine) {
			result = cloudprojectManagedDatabaseUserWithRoleColumnsToDisplay
		}
	}
	return result
}

func ListManagedDatabaseRoles(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseRoleValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if !ManagedDatabaseRoleSpec.Advanced {
		common.ManageListRequestNoExpand(
			databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "roles"),
			cloudprojectManagedDatabaseRoleColumnsToDisplay,
			flags.GenericFilters,
		)
		return
	}

	if engine != "mongodb" {
		display.OutputError(&flags.OutputFormatConfig, "flag --advanced is only supported for mongodb engines")
		return
	}

	// API v2 does not list the advanced roles
	var rawRoles []string
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/mongodb/%s/roles?advanced=true", projectID, url.PathEscape(args[0])),
		&rawRoles); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch roles: %s", err)
		return
	}

	rows := make([]map[string]any, len(rawRoles))
	for i, role := range rawRoles {
		rows[i] = map[string]any{"name": role}
	}

	filteredRows, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filteredRows, cloudprojectManagedDatabaseRoleColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetManagedDatabaseCertificates(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseCertificateValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var certificate map[string]any
	if err := httpLib.Client.Get(databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, engine, args[0], "certificates"), &certificate); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch certificate: %s", err)
		return
	}

	display.OutputObject(certificate, "", managedDatabaseCertificateTemplate, &flags.OutputFormatConfig)
}

func ListManagedDatabaseBackups(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedDatabaseCategory, "backup")+"?serviceId="+url.QueryEscape(args[0]),
		cloudprojectManagedDatabaseBackupColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedDatabaseBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	getDatabaseAnalyticsBackup(projectID, managedDatabaseCategory, args[0], args[1], managedDatabaseBackupTemplate)
}

func RestoreManagedDatabaseBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedDatabaseEngine(projectID, args[0], ManagedDatabaseBackupRestoreValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// API v2 restores a point in time, not a given backup
	if err := httpLib.Client.Post(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/backup/%s/restore", projectID, url.PathEscape(engine), url.PathEscape(args[0]), url.PathEscape(args[1])),
		nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restore backup: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Backup restore started successfully")
}
