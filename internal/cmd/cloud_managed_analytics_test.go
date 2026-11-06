// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
)

// fullAnalyticsServiceResponse is a complete API response for an analytics service,
// covering all fields accessed by the managed-analytics template.
const fullAnalyticsServiceResponse = `{
	"id": "fakeAnalyticsID",
	"createdAt": "2025-01-01T00:00:00Z",
	"resourceStatus": "READY",
	"currentState": {
		"description": "test analytics",
		"plan": "essential",
		"version": "3",
		"flavor": "db1-4",
		"status": "READY",
		"nodesCount": 1,
		"storage": {"size": {"value": 80, "unit": "GB"}},
		"network": {"networkId": "fakeNetworkID", "subnetId": "fakeSubnetID"},
		"endpoints": [{"component": "KAFKA", "domain": "kafka.database.cloud.ovh.net", "port": 20185, "ssl": true, "uri": "kafka.database.cloud.ovh.net:20185"}],
		"nodes": [{"id": "nodeID", "flavor": "db1-4", "location": {"region": "eu-west-gra"}, "status": "READY"}]
	},
	"targetSpec": {"description": "test analytics", "plan": "essential", "version": "3", "flavor": "db1-4", "deletionProtected": false}
}`

func registerManagedAnalyticsProbe(engine string) {
	registerDatabaseAnalyticsProbe("managedAnalytics", cloud.ManagedAnalyticsValidEngines, engine, "fakeAnalyticsID", fullAnalyticsServiceResponse)
}

// ── Service ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsListCmd(assert, require *td.T) {
	for _, engine := range cloud.ManagedAnalyticsValidEngines {
		body := `[]`
		if engine == "kafka" {
			body = "[" + fullAnalyticsServiceResponse + "]"
		}
		httpmock.RegisterResponder(http.MethodGet,
			"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/"+engine,
			httpmock.NewStringResponder(200, body),
		)
	}

	out, err := cmd.Execute("cloud", "managed-analytics", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("fakeAnalyticsID"), td.Contains("kafka"), td.Contains("test analytics")))
}

func (ms *MockSuite) TestManagedAnalyticsGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("fakeAnalyticsID"), td.Contains("kafka 3"), td.Contains("private")))
}

func (ms *MockSuite) TestManagedAnalyticsCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"flavor": "db1-4",
					"nodesCount": 1,
					"plan": "essential",
					"region": "eu-west-gra",
					"version": "3"
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "create", "--cloud-project", "fakeProjectID", "--engine", "kafka", "--version", "3", "--plan", "essential", "--nodes-list", "db1-4:eu-west-gra")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed analytics service created successfully (id: fakeAnalyticsID)`)
}

func (ms *MockSuite) TestManagedAnalyticsCreateInvalidEngineCmd(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "managed-analytics", "create", "--cloud-project", "fakeProjectID", "--engine", "invalid", "--version", "1", "--plan", "essential", "--nodes-list", "db1-4:eu-west-gra")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine invalid"))
}

func (ms *MockSuite) TestManagedAnalyticsEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{
				"id": "fakeAnalyticsID",
				"engine": "kafka"
		}`),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{
			"id": "fakeAnalyticsID",
			"engine": "kafka",
			"category": "analysis",
			"plan": "essential",
			"disk": {"size": 80},
			"storage": {"size": {"unit": "GB", "value": 80}},
			"ipRestrictions": [],
			"status": "READY",
			"nodes": [{"id": "nodeID", "flavor": "db1-4", "region": "DE", "status": "READY"}],
			"description": "Default description",
			"version": "3",
			"networkType": "public",
			"flavor": "db1-4",
			"maintenanceTime": "13:16:00",
			"enablePrometheus": false,
			"deletionProtection": false
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"deletionProtection": false,
				"description": "Default description",
				"disk": {"size": 80},
				"enablePrometheus": false,
				"flavor": "db1-4",
				"ipRestrictions": [],
				"maintenanceTime": "13:16:00",
				"plan": "discovery",
				"storage": {"size": {"unit": "GB", "value": 80}},
				"version": "4"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "edit", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--version", "4", "--plan", "discovery", "-o", "yaml")

	require.CmpNoError(err)
	assert.String(out, `message: ✅ Resource updated successfully
`)
}

func (ms *MockSuite) TestManagedAnalyticsDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID",
		httpmock.NewStringResponder(200, fullAnalyticsServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "delete", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed analytics service deleted successfully`)
}

// ── Database ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsDatabaseListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/database",
		httpmock.NewStringResponder(200, `[{"id": "fakeDbID", "name": "mydb", "default": false}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("mydb"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "get", "fakeAnalyticsID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("mydb"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseCreateCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/database",
		tdhttpmock.JSONBody(td.JSON(`{"name": "mydb"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: fakeDbID)`)
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support databases, valid engines: clickhouse"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "delete", "fakeAnalyticsID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Database deleted successfully`)
}

// ── User ──────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsUserListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/user",
		httpmock.NewStringResponder(200, `[{"id": "fakeUserID", "username": "myuser", "status": "READY", "createdAt": "2025-01-01T00:00:00Z"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("myuser"))
}

func (ms *MockSuite) TestManagedAnalyticsUserGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY", "createdAt": "2025-01-01T00:00:00Z"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "get", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("myuser"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateClickhouseCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "roles": []}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafkaMirrorMaker")

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafkaMirrorMaker does not support users"))
}

func (ms *MockSuite) TestManagedAnalyticsUserEditWithAclsCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY", "acls": [{"pattern": "metrics-.*", "permission": "READ"}]}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/user/fakeUserID",
		tdhttpmock.JSONBody(td.JSON(`{"acls": [{"pattern": "metrics-.*", "permission": "READ"}, {"pattern": "logs-.*", "permission": "READ_WRITE"}, {"pattern": "audit-.*", "permission": "DENY"}]}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "UPDATING"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "edit", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID", "--acls", "logs-.*:readwrite", "--acls", "audit-.*:DENY")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedAnalyticsUserEditClickhouseCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "edit", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID", "--editor")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine clickhouse does not support user edition, valid engines: opensearch"))
}

func (ms *MockSuite) TestManagedAnalyticsUserEditInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "edit", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID", "--acls", "logs-.*:readwrite")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support user edition, valid engines: opensearch"))
}

func (ms *MockSuite) TestManagedAnalyticsUserDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "DELETING"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "delete", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ User deleted successfully`)
}

func (ms *MockSuite) TestManagedAnalyticsUserCredentialsResetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/user/fakeUserID/resetPassword",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "n3wS3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "credentials-reset", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: n3wS3cr3t"))
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsRoleListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/clickhouse/fakeAnalyticsID/roles",
		httpmock.NewStringResponder(200, `[{"id": "YW5hbHlzdA", "name": "analyst"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "role", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("analyst"))
}

func (ms *MockSuite) TestManagedAnalyticsRoleListInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "role", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support roles, valid engines: clickhouse"))
}

// ── Permissions ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsPermissionListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/permissions",
		httpmock.NewStringResponder(200, `["admin", "read", "readwrite"]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "permission", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("admin"), td.Contains("readwrite")))
}

func (ms *MockSuite) TestManagedAnalyticsPermissionListInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "permission", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support permissions, valid engines: opensearch"))
}

// ── Patterns ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsPatternListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/pattern",
		httpmock.NewStringResponder(200, `[{"id": "fakePatternID", "pattern": "logs-*", "maximumIndexCount": 5}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("logs-*"), td.Contains("5")))
}

func (ms *MockSuite) TestManagedAnalyticsPatternGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/pattern/fakePatternID",
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*", "maximumIndexCount": 5}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "get", "fakeAnalyticsID", "fakePatternID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("logs-*"), td.Contains("5")))
}

func (ms *MockSuite) TestManagedAnalyticsPatternCreateCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/pattern",
		tdhttpmock.JSONBody(td.JSON(`{"pattern": "logs-*", "maximumIndexCount": 0}`)),
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*", "maximumIndexCount": 0}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--pattern", "logs-*")

	require.CmpNoError(err)
	assert.String(out, `✅ Pattern created successfully (id: fakePatternID)`)
}

func (ms *MockSuite) TestManagedAnalyticsPatternCreateInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--pattern", "logs-*")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support patterns, valid engines: opensearch"))
}

func (ms *MockSuite) TestManagedAnalyticsPatternDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/opensearch/fakeAnalyticsID/pattern/fakePatternID",
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "delete", "fakeAnalyticsID", "fakePatternID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Pattern deleted successfully`)
}

// ── Certificate ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsCertificateGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/certificates",
		httpmock.NewStringResponder(200, `{"ca": "-----BEGIN CERTIFICATE-----\nfakeCA\n-----END CERTIFICATE-----"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "certificate", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeCA"))
}

func (ms *MockSuite) TestManagedAnalyticsCertificateGetInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")

	out, err := cmd.Execute("cloud", "managed-analytics", "certificate", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine opensearch does not support certificates, valid engines: clickhouse, kafka"))
}

// ── Backup ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsBackupListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/backup?serviceId=fakeAnalyticsID",
		httpmock.NewStringResponder(200, `[{"id": "fakeBackupID", "serviceId": "fakeAnalyticsID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "daily backup"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeBackupID"))
}

func (ms *MockSuite) TestManagedAnalyticsBackupGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "serviceId": "fakeAnalyticsID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "daily backup", "size": {"value": 1, "unit": "GB"}, "regions": ["eu-west-gra"]}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "get", "fakeAnalyticsID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("fakeBackupID"), td.Contains("eu-west-gra")))
}

func (ms *MockSuite) TestManagedAnalyticsBackupListInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine kafka does not support backups, valid engines: clickhouse, grafana, opensearch"))
}

// ── Topics ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsTopicListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic",
		httpmock.NewStringResponder(200, `[{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 2, "minimumInSyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("my-topic"), td.Contains("168")))
}

func (ms *MockSuite) TestManagedAnalyticsTopicGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 2, "minimumInSyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "get", "fakeAnalyticsID", "fakeTopicID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("my-topic"), td.Contains("168")))
}

func (ms *MockSuite) TestManagedAnalyticsTopicCreateCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"name": "my-topic",
				"minimumInSyncReplicas": 1,
				"partitions": 3,
				"replication": 2,
				"retentionBytes": -1,
				"retentionHours": 168
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "create", "fakeAnalyticsID",
		"--cloud-project", "fakeProjectID", "--name", "my-topic", "--min-insync-replicas", "1",
		"--partitions", "3", "--replication", "2", "--retention-bytes", "-1", "--retention-hours", "168")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic created successfully (id: fakeTopicID)`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicEditCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 2, "minimumInSyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic/fakeTopicID",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"minimumInSyncReplicas": 1,
				"partitions": 6,
				"replication": 2,
				"retentionBytes": -1,
				"retentionHours": 72
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "edit", "fakeAnalyticsID", "fakeTopicID",
		"--cloud-project", "fakeProjectID", "--partitions", "6", "--retention-hours", "72")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "delete", "fakeAnalyticsID", "fakeTopicID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic deleted successfully`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicListInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("clickhouse")

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine clickhouse does not support topics, valid engines: kafka"))
}

// ── Topic ACLs ────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsTopicACLListCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/acl",
		httpmock.NewStringResponder(200, `[{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "READ"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("myuser"), td.Contains("READ")))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLGetCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/acl/fakeACLID",
		httpmock.NewStringResponder(200, `{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "READ"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "get", "fakeAnalyticsID", "fakeACLID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-topic"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLCreateCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/acl",
		tdhttpmock.JSONBody(td.JSON(`{"permission": "READ_WRITE", "topic": "my-topic", "username": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "READ_WRITE"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--permission", "readwrite", "--topic", "my-topic", "--username", "myuser")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic ACL created successfully (id: fakeACLID)`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLCreateInvalidEngineCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("opensearch")

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--permission", "read", "--topic", "my-topic", "--username", "myuser")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 engine opensearch does not support topic ACLs, valid engines: kafka"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLDeleteCmd(assert, require *td.T) {
	registerManagedAnalyticsProbe("kafka")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedAnalytics/kafka/fakeAnalyticsID/acl/fakeACLID",
		httpmock.NewStringResponder(200, `{"id": "fakeACLID"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "delete", "fakeAnalyticsID", "fakeACLID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic ACL deleted successfully`)
}
