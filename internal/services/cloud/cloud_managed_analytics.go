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
	ManagedAnalyticsValidEngines            = []string{"clickhouse", "grafana", "kafka", "kafkaConnect", "kafkaMirrorMaker", "opensearch"}
	ManagedAnalyticsDatabaseValidEngines    = []string{"clickhouse"}
	ManagedAnalyticsUserEditValidEngines    = []string{"clickhouse", "opensearch"}
	ManagedAnalyticsUserPostValidEngines    = append(ManagedAnalyticsUserEditValidEngines, "kafka", "kafkaConnect")
	ManagedAnalyticsUserValidEngines        = append(ManagedAnalyticsUserPostValidEngines, "grafana")
	ManagedAnalyticsRoleValidEngines        = []string{"clickhouse"}
	ManagedAnalyticsPermissionValidEngines  = []string{"opensearch"}
	ManagedAnalyticsPatternValidEngines     = []string{"opensearch"}
	ManagedAnalyticsBackupValidEngines      = []string{"clickhouse", "grafana", "opensearch"}
	ManagedAnalyticsCertificateValidEngines = []string{"clickhouse", "kafka"}
	ManagedAnalyticsTopicValidEngines       = []string{"kafka"}

	cloudprojectManagedAnalyticsColumnsToDisplay               = []string{"id", "engine", "version", "description", "status"}
	cloudprojectManagedAnalyticsDatabaseColumnsToDisplay       = []string{"id", "name", "default"}
	cloudprojectManagedAnalyticsUserColumnsToDisplay           = []string{"id", "username", "status", "createdAt"}
	cloudprojectManagedAnalyticsClickhouseUserColumnsToDisplay = []string{"id", "username", "status", "roles", "createdAt"}
	cloudprojectManagedAnalyticsRoleColumnsToDisplay           = []string{"role"}
	cloudprojectManagedAnalyticsPermissionColumnsToDisplay     = []string{"name"}
	cloudprojectManagedAnalyticsPatternColumnsToDisplay        = []string{"id", "maxIndexCount", "pattern"}
	cloudprojectManagedAnalyticsBackupColumnsToDisplay         = []string{"id", "createdAt", "type", "status", "description"}
	cloudprojectManagedAnalyticsTopicColumnsToDisplay          = []string{"id", "name", "partitions", "replication", "minInsyncReplicas", "retentionBytes", "retentionHours"}
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
		IPRestrictions  []managedAnalyticsIPRestriction `json:"ipRestrictions,omitempty"`
		MaintenanceTime string                          `json:"maintenanceTime,omitempty"`
		NetworkID       string                          `json:"networkId,omitempty"`
		NodesList       []managedAnalyticsNode          `json:"nodesList,omitempty"`
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

	ManagedAnalyticsDatabaseSpec struct {
		Name string `json:"name"`
	}

	ManagedAnalyticsUserSpec ManagedAnalyticsUser

	ManagedAnalyticsPatternSpec struct {
		Pattern       string `json:"pattern"`
		MaxIndexCount int    `json:"maxIndexCount,omitempty"`
	}

	ManagedAnalyticsTopicSpec struct {
		Name              string `json:"name"`
		MinInsyncReplicas int    `json:"minInsyncReplicas,omitempty"`
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

	managedAnalyticsNode struct {
		Flavor string `json:"flavor,omitempty"`
		Region string `json:"region,omitempty"`
		Role   string `json:"role,omitempty"`
	}

	ManagedAnalyticsUser struct {
		Name string `json:"name"`

		// Clickhouse specific field
		Roles []string `json:"roles,omitempty"`

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

func ListManagedAnalytics(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequest(fmt.Sprintf("/v1/cloud/project/%s/database/service", projectID),
		"",
		cloudprojectManagedAnalyticsColumnsToDisplay,
		append(flags.GenericFilters, `category == "analysis"`))
}

func GetManagedAnalytics(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch managed analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/v1/cloud/project/%s/database/service", projectID),
		args[0],
		managedAnalyticsTemplate)
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

	// Parse IP restrictions
	for _, restriction := range ManagedAnalyticsSpec.CLIIPRestrictions {
		ManagedAnalyticsSpec.IPRestrictions = append(ManagedAnalyticsSpec.IPRestrictions, managedAnalyticsIPRestriction{IP: restriction})
	}

	// Parse nodes list
	for _, node := range ManagedAnalyticsSpec.CLINodesList {
		parts := strings.Split(node, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid node format: %s (expected format: flavor1:region1,flavor2:region2...)", node)
			return
		}
		ManagedAnalyticsSpec.NodesList = append(ManagedAnalyticsSpec.NodesList, managedAnalyticsNode{
			Flavor: parts[0],
			Region: parts[1],
		})
	}

	analytics, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s", url.PathEscape(ManagedAnalyticsSpec.Engine)),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s", projectID, url.PathEscape(ManagedAnalyticsSpec.Engine)),
		ManagedAnalyticsCreationExample,
		ManagedAnalyticsSpec,
		assets.CloudOpenapiSchema,
		[]string{"version", "plan"})
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
	for _, restriction := range ManagedAnalyticsSpec.CLIIPRestrictions {
		ManagedAnalyticsSpec.IPRestrictions = append(ManagedAnalyticsSpec.IPRestrictions, managedAnalyticsIPRestriction{IP: restriction})
	}

	// Edit resource
	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		&ManagedAnalyticsSpec,
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

	// Delete the analytics service
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsDatabaseValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support databases, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsDatabaseValidEngines, ", "))
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/database", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsDatabaseValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support databases, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsDatabaseValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/database", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsDatabaseValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support databases, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsDatabaseValidEngines, ", "))
		return
	}

	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/database",
		projectID,
		url.PathEscape(analyticsService["engine"].(string)),
		url.PathEscape(args[0]),
	)

	var database map[string]any
	if err := httpLib.Client.Post(endpoint, &ManagedAnalyticsDatabaseSpec, &database); err != nil {
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsDatabaseValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support databases, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsDatabaseValidEngines, ", "))
		return
	}

	// Delete the database in the given analytics cluster
	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/database/%s",
		projectID,
		url.PathEscape(analyticsService["engine"].(string)),
		url.PathEscape(args[0]),
		url.PathEscape(args[1]),
	)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserValidEngines, ", "))
		return
	}

	columnsToDisplay := cloudprojectManagedAnalyticsUserColumnsToDisplay
	if analyticsService["engine"].(string) == "clickhouse" {
		columnsToDisplay = cloudprojectManagedAnalyticsClickhouseUserColumnsToDisplay
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/user", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/user", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserPostValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserPostValidEngines, ", "))
		return
	}

	// Parse ACLs
	for _, acl := range ManagedAnalyticsUserSpec.CLIAcls {
		parts := strings.SplitN(acl, ":", 2)
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid acl format: %s (expected format: pattern:permission)", acl)
			return
		}
		ManagedAnalyticsUserSpec.Acls = append(ManagedAnalyticsUserSpec.Acls, managedAnalyticsUserAcl{
			Pattern:    parts[0],
			Permission: parts[1],
		})
	}

	if err := ManagedAnalyticsUserSpec.validateForEngine(analyticsService["engine"].(string)); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userSpec := ManagedAnalyticsUserSpec.forEngine(analyticsService["engine"].(string))

	user, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}/user", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/user", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		ManagedAnalyticsUserCreationExample,
		userSpec,
		assets.CloudOpenapiSchema,
		nil)
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserEditValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserEditValidEngines, ", "))
		return
	}

	// Parse ACLs
	for _, acl := range ManagedAnalyticsUserSpec.CLIAcls {
		parts := strings.SplitN(acl, ":", 2)
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid acl format: %s (expected format: pattern:permission)", acl)
			return
		}
		ManagedAnalyticsUserSpec.Acls = append(ManagedAnalyticsUserSpec.Acls, managedAnalyticsUserAcl{
			Pattern:    parts[0],
			Permission: parts[1],
		})
	}

	if err := ManagedAnalyticsUserSpec.validateForEngine(analyticsService["engine"].(string)); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	userSpec := ManagedAnalyticsUserSpec.forEngine(analyticsService["engine"].(string))

	// Edit resource
	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}/user/{userId}", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/user/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0]), url.PathEscape(args[1])),
		userSpec,
		assets.CloudOpenapiSchema,
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserPostValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserPostValidEngines, ", "))
		return
	}

	// Delete the user in the given analytics cluster
	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/user/%s",
		projectID,
		url.PathEscape(analyticsService["engine"].(string)),
		url.PathEscape(args[0]),
		url.PathEscape(args[1]),
	)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsUserValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support users, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsUserValidEngines, ", "))
		return
	}

	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/user/%s/credentials/reset",
		projectID,
		url.PathEscape(analyticsService["engine"].(string)),
		url.PathEscape(args[0]),
		url.PathEscape(args[1]),
	)

	var user map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &user); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset user credentials: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ User credentials reset successfully")
	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚠️ Save this password now, it will not be shown again:\n\nPassword: %s", user["password"])
}

func (s ManagedAnalyticsUser) validateForEngine(engine string) error {
	if len(s.Roles) > 0 && engine != "clickhouse" {
		return fmt.Errorf("flag --roles is only supported for the clickhouse engine")
	}
	if len(s.Acls) > 0 && engine != "opensearch" {
		return fmt.Errorf("flag --acl is only supported for the opensearch engine")
	}
	return nil
}

func (s ManagedAnalyticsUser) forEngine(engine string) ManagedAnalyticsUser {
	result := ManagedAnalyticsUser{Name: s.Name}
	switch engine {
	case "clickhouse":
		result.Roles = s.Roles
	case "opensearch":
		result.Acls = s.Acls
	}
	return result
}

func ListManagedAnalyticsRoles(_ *cobra.Command, args []string) {
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

	if !slices.Contains(ManagedAnalyticsRoleValidEngines, databaseService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", databaseService["engine"].(string), strings.Join(ManagedAnalyticsRoleValidEngines, ", "))
		return
	}

	var rawRoles []string
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/roles", projectID, url.PathEscape(databaseService["engine"].(string)), url.PathEscape(args[0])),
		&rawRoles); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch roles: %s", err)
		return
	}

	rows := make([]map[string]any, len(rawRoles))
	for i, role := range rawRoles {
		rows[i] = map[string]any{"role": role}
	}

	filteredRows, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filteredRows, cloudprojectManagedAnalyticsRoleColumnsToDisplay, &flags.OutputFormatConfig)
}

func ListManagedAnalyticsPermissions(_ *cobra.Command, args []string) {
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

	if !slices.Contains(ManagedAnalyticsPermissionValidEngines, databaseService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", databaseService["engine"].(string), strings.Join(ManagedAnalyticsPermissionValidEngines, ", "))
		return
	}

	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/permissions",
		projectID,
		url.PathEscape(databaseService["engine"].(string)),
		url.PathEscape(args[0]),
	)

	var result struct {
		Names []string `json:"names"`
	}
	if err := httpLib.Client.Get(endpoint, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch permissions: %s", err)
		return
	}

	rows := make([]map[string]any, len(result.Names))
	for i, name := range result.Names {
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsPatternValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support patterns, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsPatternValidEngines, ", "))
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/pattern", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsPatternValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support patterns, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsPatternValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/pattern", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsPatternValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support patterns, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsPatternValidEngines, ", "))
		return
	}

	pattern, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}/pattern", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/pattern", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		ManagedAnalyticsPatternCreationExample,
		ManagedAnalyticsPatternSpec,
		assets.CloudOpenapiSchema,
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

	// Fetch analytics service to retrieve the engine
	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch managed analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsPatternValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support patterns, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsPatternValidEngines, ", "))
		return
	}

	// Delete the pattern in the given analytics cluster
	endpoint := fmt.Sprintf(
		"/v1/cloud/project/%s/database/%s/%s/pattern/%s",
		projectID,
		url.PathEscape(analyticsService["engine"].(string)),
		url.PathEscape(args[0]),
		url.PathEscape(args[1]),
	)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsCertificateValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support certificates, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsCertificateValidEngines, ", "))
		return
	}

	var certificate map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/certificates", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		&certificate); err != nil {
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsBackupValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsBackupValidEngines, ", "))
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/backup", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsBackupValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "invalid engine %s, valid values: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsBackupValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/backup", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		args[1],
		managedAnalyticsBackupTemplate,
	)
}

func ListManagedAnalyticsTopics(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topics, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topic", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topics, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topic", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topics, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	topic, err := common.CreateResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}/topic", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topic", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		ManagedAnalyticsTopicCreationExample,
		ManagedAnalyticsTopicSpec,
		assets.CloudOpenapiSchema,
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topics, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	// Edit resource
	if err := common.EditResource(
		cmd,
		fmt.Sprintf("/cloud/project/{serviceName}/database/%s/{clusterId}/topic/{topicId}", url.PathEscape(analyticsService["engine"].(string))),
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topic/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0]), url.PathEscape(args[1])),
		&ManagedAnalyticsTopicSpec,
		assets.CloudOpenapiSchema,
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topics, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	if err := httpLib.Client.Delete(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topic/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0]), url.PathEscape(args[1])),
		nil); err != nil {
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topic ACLs, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	common.ManageListRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topicAcl", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		"",
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topic ACLs, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topicAcl", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topic ACLs, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	var topicACL map[string]any
	if err := httpLib.Client.Post(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topicAcl", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0])),
		&ManagedAnalyticsTopicACLSpec,
		&topicACL,
	); err != nil {
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

	var analyticsService map[string]any
	if err := httpLib.Client.Get(
		fmt.Sprintf("/v1/cloud/project/%s/database/service/%s", projectID, url.PathEscape(args[0])), &analyticsService); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch analytics service: %s", err)
		return
	}

	if !slices.Contains(ManagedAnalyticsTopicValidEngines, analyticsService["engine"].(string)) {
		display.OutputError(&flags.OutputFormatConfig, "engine %s does not support topic ACLs, valid engines: %s", analyticsService["engine"].(string), strings.Join(ManagedAnalyticsTopicValidEngines, ", "))
		return
	}

	if err := httpLib.Client.Delete(
		fmt.Sprintf("/v1/cloud/project/%s/database/%s/%s/topicAcl/%s", projectID, url.PathEscape(analyticsService["engine"].(string)), url.PathEscape(args[0]), url.PathEscape(args[1])),
		nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete topic ACL: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Topic ACL deleted successfully")
}
