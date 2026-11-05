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
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	ManagedAnalyticsValidEngines            = []string{"clickhouse", "grafana", "kafka", "kafkaConnect", "kafkaMirrorMaker", "opensearch"}
	ManagedAnalyticsDatabaseValidEngines    = []string{"clickhouse"}
	ManagedAnalyticsUserEditValidEngines    = []string{"opensearch"}
	ManagedAnalyticsUserPostValidEngines    = []string{"clickhouse", "kafka", "kafkaConnect", "opensearch"}
	ManagedAnalyticsUserValidEngines        = append(ManagedAnalyticsUserPostValidEngines, "grafana")
	ManagedAnalyticsRoleValidEngines        = []string{"clickhouse"}
	ManagedAnalyticsPermissionValidEngines  = []string{"opensearch"}
	ManagedAnalyticsPatternValidEngines     = []string{"opensearch"}
	ManagedAnalyticsBackupValidEngines      = []string{"clickhouse", "grafana", "opensearch"}
	ManagedAnalyticsCertificateValidEngines = []string{"clickhouse", "kafka"}
	ManagedAnalyticsTopicValidEngines       = []string{"kafka"}

	cloudprojectManagedAnalyticsColumnsToDisplay               = []string{"id", "engine", "currentState.version version", "currentState.description description", "currentState.status status"}
	cloudprojectManagedAnalyticsDatabaseColumnsToDisplay       = []string{"id", "name", "default"}
	cloudprojectManagedAnalyticsUserColumnsToDisplay           = []string{"id", "username", "status", "createdAt"}
	cloudprojectManagedAnalyticsClickhouseUserColumnsToDisplay = []string{"id", "username", "status", "roles", "createdAt"}
	cloudprojectManagedAnalyticsRoleColumnsToDisplay           = []string{"name"}
	cloudprojectManagedAnalyticsPermissionColumnsToDisplay     = []string{"name"}
	cloudprojectManagedAnalyticsPatternColumnsToDisplay        = []string{"id", "maximumIndexCount", "pattern"}
	cloudprojectManagedAnalyticsBackupColumnsToDisplay         = []string{"id", "createdAt", "type", "status", "description"}
	cloudprojectManagedAnalyticsTopicColumnsToDisplay          = []string{"id", "name", "partitions", "replication", "minimumInSyncReplicas", "retentionBytes", "retentionHours"}
	cloudprojectManagedAnalyticsTopicACLColumnsToDisplay       = []string{"id", "username", "topic", "permission"}

	//go:embed templates/cloud_managed_analytics.tmpl
	managedAnalyticsTemplate string
	//go:embed templates/cloud_managed_analytics_database.tmpl
	managedAnalyticsDatabaseTemplate string
	//go:embed templates/cloud_managed_analytics_user.tmpl
	managedAnalyticsUserTemplate string
	//go:embed templates/cloud_managed_analytics_pattern.tmpl
	managedAnalyticsPatternTemplate string
	//go:embed templates/cloud_managed_analytics_certificate.tmpl
	managedAnalyticsCertificateTemplate string
	//go:embed templates/cloud_managed_analytics_backup.tmpl
	managedAnalyticsBackupTemplate string
	//go:embed templates/cloud_managed_analytics_topic.tmpl
	managedAnalyticsTopicTemplate string
	//go:embed templates/cloud_managed_analytics_topic_acl.tmpl
	managedAnalyticsTopicACLTemplate string

	//go:embed parameter-samples/managed-analytics-create.json
	ManagedAnalyticsCreationExample string
	//go:embed parameter-samples/managed-analytics-user-create.json
	ManagedAnalyticsUserCreationExample string
	//go:embed parameter-samples/managed-analytics-pattern-create.json
	ManagedAnalyticsPatternCreationExample string
	//go:embed parameter-samples/managed-analytics-topic-create.json
	ManagedAnalyticsTopicCreationExample string
	ManagedAnalyticsSpec                 struct {
		TargetSpec databaseAnalyticsTargetSpec `json:"targetSpec,omitzero"`

		// Extra fields for CLI only
		Engine            string   `json:"-"`
		CLIDiskSize       int      `json:"-"`
		CLIIPRestrictions []string `json:"-"`
		CLINodesList      []string `json:"-"`
	}

	// ManagedAnalyticsEditSpec holds the parameters of the edit command, which
	// still uses API v1.
	ManagedAnalyticsEditSpec struct {
		Backups struct {
			Regions []string `json:"regions,omitempty"`
			Time    string   `json:"time,omitempty"`
		} `json:"backups,omitzero"`
		DeletionProtection bool                            `json:"deletionProtection,omitempty"`
		Description        string                          `json:"description,omitempty"`
		EnablePrometheus   bool                            `json:"enablePrometheus,omitempty"`
		IPRestrictions     []managedAnalyticsIPRestriction `json:"ipRestrictions,omitempty"`
		MaintenanceTime    string                          `json:"maintenanceTime,omitempty"`
		NodesPattern       struct {
			Flavor string `json:"flavor,omitempty"`
		} `json:"nodesPattern,omitzero"`
		Plan    string `json:"plan,omitempty"`
		Version string `json:"version,omitempty"`

		// Extra fields for CLI only
		CLIIPRestrictions []string `json:"-"`
	}

	ManagedAnalyticsDatabaseSpec struct {
		Name string `json:"name"`
	}

	ManagedAnalyticsUserSpec ManagedAnalyticsUser

	ManagedAnalyticsPatternSpec struct {
		Pattern       string `json:"pattern"`
		MaxIndexCount int    `json:"maximumIndexCount"`
	}

	ManagedAnalyticsTopicSpec struct {
		Name              string `json:"name"`
		MinInsyncReplicas int    `json:"minimumInSyncReplicas,omitempty"`
		Partitions        int    `json:"partitions,omitempty"`
		Replication       int    `json:"replication,omitempty"`
		RetentionBytes    int    `json:"retentionBytes,omitempty"`
		RetentionHours    int    `json:"retentionHours,omitempty"`
	}

	ManagedAnalyticsTopicACLSpec struct {
		Permission string `json:"permission"`
		Topic      string `json:"topic"`
		Username   string `json:"username"`
	}
)

type (
	managedAnalyticsIPRestriction struct {
		Description string `json:"description,omitempty"`
		IP          string `json:"ip,omitempty"`
	}

	ManagedAnalyticsUser struct {
		Name string `json:"name"`

		// OpenSearch specific field
		Acls []managedAnalyticsUserAcl `json:"acls,omitempty"`

		// Extra fields for CLI only
		CLIAcls []string `json:"-"`
	}

	managedAnalyticsUserAcl struct {
		Pattern    string `json:"pattern"`
		Permission string `json:"permission"`
	}
)

// getManagedAnalyticsEngine returns the engine of a managed analytics service,
// checked against the engines that support the feature.
func getManagedAnalyticsEngine(projectID, serviceID string, validEngines []string, feature string) (string, error) {
	engine, _, err := getDatabaseAnalyticsService(projectID, managedAnalyticsCategory, serviceID, ManagedAnalyticsValidEngines)
	if err != nil {
		return "", fmt.Errorf("failed to fetch managed analytics service: %w", err)
	}

	if !slices.Contains(validEngines, engine) {
		return "", fmt.Errorf("engine %s does not support %s, valid engines: %s", engine, feature, strings.Join(validEngines, ", "))
	}

	return engine, nil
}

// managedAnalyticsPermission returns the API v2 value of an ACL permission,
// accepting the lowercase values of API v1 (readwrite for READ_WRITE).
func managedAnalyticsPermission(permission string) string {
	if strings.EqualFold(permission, "readwrite") {
		return "READ_WRITE"
	}
	return strings.ToUpper(permission)
}

func ListManagedAnalytics(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	listDatabaseAnalyticsServices(projectID, managedAnalyticsCategory, ManagedAnalyticsValidEngines, cloudprojectManagedAnalyticsColumnsToDisplay)
}

func GetManagedAnalytics(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	_, analyticsService, err := getDatabaseAnalyticsService(projectID, managedAnalyticsCategory, args[0], ManagedAnalyticsValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	display.OutputObject(analyticsService, args[0], managedAnalyticsTemplate, &flags.OutputFormatConfig)
}

func CreateManagedAnalyticsPreRun(_ *cobra.Command, _ []string) error {
	if !slices.Contains(ManagedAnalyticsValidEngines, ManagedAnalyticsSpec.Engine) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", ManagedAnalyticsSpec.Engine, strings.Join(ManagedAnalyticsValidEngines, ", "))
		return fmt.Errorf("invalid engine %s", ManagedAnalyticsSpec.Engine)
	}
	return nil
}

func CreateManagedAnalytics(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	spec := ManagedAnalyticsSpec
	spec.TargetSpec.Network.AuthorizedIps = newAuthorizedIps(spec.CLIIPRestrictions)
	if spec.CLIDiskSize > 0 {
		spec.TargetSpec.Storage.Size = &databaseAnalyticsUnitAndValue{Unit: "GB", Value: spec.CLIDiskSize}
	}
	if err := spec.TargetSpec.setNodes(spec.CLINodesList); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	analytics, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s", spec.Engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, spec.Engine),
		ManagedAnalyticsCreationExample,
		spec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.version", "targetSpec.plan", "targetSpec.flavor", "targetSpec.region"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create managed analytics service: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, analytics, "✅ Managed analytics service created successfully (id: %s)", analytics["id"])
}

func EditManagedAnalytics(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsValidEngines, ", "))
		return
	}

	// Parse IP restrictions
	for _, restriction := range ManagedAnalyticsEditSpec.CLIIPRestrictions {
		ManagedAnalyticsEditSpec.IPRestrictions = append(ManagedAnalyticsEditSpec.IPRestrictions, managedAnalyticsIPRestriction{IP: restriction})
	}

	// Edit resource
	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		ManagedAnalyticsEditSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteManagedAnalytics(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, _, err := getDatabaseAnalyticsService(projectID, managedAnalyticsCategory, args[0], ManagedAnalyticsValidEngines)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete managed analytics service: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Managed analytics service deleted successfully")
}

func ListManagedAnalyticsDatabases(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsDatabaseValidEngines, "databases")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "database"),
		cloudprojectManagedAnalyticsDatabaseColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsDatabaseValidEngines, "databases")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "database"),
		args[1],
		managedAnalyticsDatabaseTemplate,
	)
}

func CreateManagedAnalyticsDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsDatabaseValidEngines, "databases")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var database map[string]any
	if err := httpLib.Client.Post(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "database"),
		&ManagedAnalyticsDatabaseSpec,
		&database,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database created successfully (id: %s)", database["id"])
}

func DeleteManagedAnalyticsDatabase(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsDatabaseValidEngines, "databases")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "database", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete database: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Database deleted successfully")
}

func ListManagedAnalyticsUsers(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserValidEngines, "users")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	columnsToDisplay := cloudprojectManagedAnalyticsUserColumnsToDisplay
	if engine == "clickhouse" {
		columnsToDisplay = cloudprojectManagedAnalyticsClickhouseUserColumnsToDisplay
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user"),
		columnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserValidEngines, "users")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user"),
		args[1],
		managedAnalyticsUserTemplate,
	)
}

func CreateManagedAnalyticsUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserPostValidEngines, "users")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userSpec, err := ManagedAnalyticsUserSpec.forEngine(engine)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	user, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s/{serviceId}/user", engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user"),
		ManagedAnalyticsUserCreationExample,
		userSpec,
		assets.CloudV2OpenapiSchema,
		nil,
	)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, user, "✅ User created successfully (id: %s)", user["id"])
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚠️ Save this password now, it will not be shown again:\n\nPassword: %s", user["password"])
}

func EditManagedAnalyticsUser(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserEditValidEngines, "user edition")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userSpec, err := ManagedAnalyticsUserSpec.forEngine(engine)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s/{serviceId}/user/{userId}", engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user", args[1]),
		userSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteManagedAnalyticsUser(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserPostValidEngines, "users")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete user: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User deleted successfully")
}

func ResetManagedAnalyticsUserCredentials(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsUserValidEngines, "users")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	user, err := resetDatabaseAnalyticsUserCredentials(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "user", args[1]))
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset user credentials: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User credentials reset successfully")
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚠️ Save this password now, it will not be shown again:\n\nPassword: %s", user["password"])
}

// forEngine returns the user to send for the engine, with its ACLs parsed from
// the command line, or an error when a flag does not apply to the engine.
func (s ManagedAnalyticsUser) forEngine(engine string) (ManagedAnalyticsUser, error) {
	result := ManagedAnalyticsUser{Name: s.Name}

	if len(s.CLIAcls) > 0 && engine != "opensearch" {
		return result, fmt.Errorf("flag --acl is only supported for the opensearch engine")
	}

	for _, acl := range s.CLIAcls {
		parts := strings.SplitN(acl, ":", 2)
		if len(parts) != 2 {
			return result, fmt.Errorf("invalid acl format: %s (expected format: pattern:permission)", acl)
		}
		result.Acls = append(result.Acls, managedAnalyticsUserAcl{
			Pattern:    parts[0],
			Permission: managedAnalyticsPermission(parts[1]),
		})
	}

	return result, nil
}

func ListManagedAnalyticsRoles(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsRoleValidEngines, "roles")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "roles"),
		cloudprojectManagedAnalyticsRoleColumnsToDisplay,
		flags.GenericFilters,
	)
}

func ListManagedAnalyticsPermissions(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsPermissionValidEngines, "permissions")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var permissions []string
	if err := httpLib.Client.Get(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "permissions"), &permissions); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch permissions: %s", err)
		return
	}

	rows := make([]map[string]any, len(permissions))
	for i, name := range permissions {
		rows[i] = map[string]any{"name": name}
	}

	display.RenderTable(rows, cloudprojectManagedAnalyticsPermissionColumnsToDisplay, &flags.OutputFormatConfig)
}

func ListManagedAnalyticsPatterns(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsPatternValidEngines, "patterns")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "pattern"),
		cloudprojectManagedAnalyticsPatternColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsPattern(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsPatternValidEngines, "patterns")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "pattern"),
		args[1],
		managedAnalyticsPatternTemplate,
	)
}

func CreateManagedAnalyticsPattern(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsPatternValidEngines, "patterns")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	pattern, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s/{serviceId}/pattern", engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "pattern"),
		ManagedAnalyticsPatternCreationExample,
		ManagedAnalyticsPatternSpec,
		assets.CloudV2OpenapiSchema,
		nil)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create pattern: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, pattern, "✅ Pattern created successfully (id: %s)", pattern["id"])
}

func DeleteManagedAnalyticsPattern(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsPatternValidEngines, "patterns")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "pattern", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete pattern: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Pattern deleted successfully")
}

func GetManagedAnalyticsCertificates(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsCertificateValidEngines, "certificates")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var certificate map[string]any
	if err := httpLib.Client.Get(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "certificates"), &certificate); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch certificate: %s", err)
		return
	}

	display.OutputObject(certificate, "", managedAnalyticsCertificateTemplate, &flags.OutputFormatConfig)
}

func ListManagedAnalyticsBackups(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if _, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsBackupValidEngines, "backups"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, "backup")+"?serviceId="+url.QueryEscape(args[0]),
		cloudprojectManagedAnalyticsBackupColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsBackup(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if _, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsBackupValidEngines, "backups"); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	getDatabaseAnalyticsBackup(projectID, managedAnalyticsCategory, args[0], args[1], managedAnalyticsBackupTemplate)
}

func ListManagedAnalyticsTopics(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topics")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "topic"),
		cloudprojectManagedAnalyticsTopicColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsTopic(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topics")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "topic"),
		args[1],
		managedAnalyticsTopicTemplate,
	)
}

func CreateManagedAnalyticsTopic(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topics")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	topic, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s/{serviceId}/topic", engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "topic"),
		ManagedAnalyticsTopicCreationExample,
		ManagedAnalyticsTopicSpec,
		assets.CloudV2OpenapiSchema,
		nil)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create topic: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, topic, "✅ Topic created successfully (id: %s)", topic["id"])
}

func EditManagedAnalyticsTopic(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topics")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/publicCloud/project/{projectId}/managedAnalytics/%s/{serviceId}/topic/{topicId}", engine),
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "topic", args[1]),
		ManagedAnalyticsTopicSpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteManagedAnalyticsTopic(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topics")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "topic", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete topic: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Topic deleted successfully")
}

func ListManagedAnalyticsTopicACLs(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topic ACLs")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "acl"),
		cloudprojectManagedAnalyticsTopicACLColumnsToDisplay,
		flags.GenericFilters,
	)
}

func GetManagedAnalyticsTopicACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topic ACLs")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "acl"),
		args[1],
		managedAnalyticsTopicACLTemplate,
	)
}

func CreateManagedAnalyticsTopicACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topic ACLs")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	spec := ManagedAnalyticsTopicACLSpec
	spec.Permission = managedAnalyticsPermission(spec.Permission)

	var topicACL map[string]any
	if err := httpLib.Client.Post(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "acl"), &spec, &topicACL); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create topic ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, topicACL, "✅ Topic ACL created successfully (id: %s)", topicACL["id"])
}

func DeleteManagedAnalyticsTopicACL(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	engine, err := getManagedAnalyticsEngine(projectID, args[0], ManagedAnalyticsTopicValidEngines, "topic ACLs")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(databaseAnalyticsEndpoint(projectID, managedAnalyticsCategory, engine, args[0], "acl", args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete topic ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Topic ACL deleted successfully")
}
