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
)

// fullAnalyticsServiceResponse is a complete API response for an analytics service,
// covering all fields accessed by the managed-analytics template.
const fullAnalyticsServiceResponse = `{
	"id": "fakeAnalyticsID",
	"engine": "kafka",
	"version": "3",
	"description": "test analytics",
	"status": "READY",
	"category": "analysis",
	"plan": "essential",
	"flavor": "db1-4",
	"createdAt": "2025-01-01T00:00:00Z",
	"networkType": "public",
	"storage": {"size": {"value": 80, "unit": "GB"}},
	"nodes": [{"id": "nodeID", "flavor": "db1-4", "region": "DE", "status": "READY"}]
}`

// ── Service ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service",
		httpmock.NewStringResponder(200, `["fakeAnalyticsID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, fullAnalyticsServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeAnalyticsID"))
}

func (ms *MockSuite) TestManagedAnalyticsGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, fullAnalyticsServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeAnalyticsID"))
}

func (ms *MockSuite) TestManagedAnalyticsCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"nodesList": [
					{
						"flavor": "db1-4",
						"region": "DE"
					}
				],
				"plan": "essential",
				"version": "3"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "create", "--cloud-project", "fakeProjectID", "--engine", "kafka", "--version", "3", "--plan", "essential", "--nodes-list", "db1-4:DE")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed analytics service created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}

func (ms *MockSuite) TestManagedAnalyticsCreateInvalidEngineCmd(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "managed-analytics", "create", "--cloud-project", "fakeProjectID", "--engine", "invalid", "--version", "1", "--plan", "essential", "--nodes-list", "db1-4:DE")

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
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "delete", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed analytics service deleted successfully`)
}

// ── Database ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsDatabaseListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/database",
		httpmock.NewStringResponder(200, `["fakeDbID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDbID"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "get", "fakeAnalyticsID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDbID"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/database",
		tdhttpmock.JSONBody(td.JSON(`{"name": "mydb"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: fakeDbID)`)
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedAnalyticsDatabaseDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/database/fakeDbID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "database", "delete", "fakeAnalyticsID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Database deleted successfully`)
}

// ── User ──────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsUserListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user",
		httpmock.NewStringResponder(200, `["fakeUserID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeUserID"))
}

func (ms *MockSuite) TestManagedAnalyticsUserGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "get", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeUserID"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateWithRolesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	// Use simple responder: CreateResource schema filtering may alter the exact body
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/user",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser", "--roles", "analyst")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedAnalyticsUserCreateInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// kafkaMirrorMaker is not in UserPostValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafkaMirrorMaker"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedAnalyticsUserEditWithAclsCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "acls": [], "status": "READY"}`),
	)
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "edit", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID", "--acls", "logs-.*:readwrite")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedAnalyticsUserEditInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// kafka is not in UserEditValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "edit", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID", "--acls", "logs-.*:readwrite")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedAnalyticsUserDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user/fakeUserID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "delete", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ User deleted successfully`)
}

func (ms *MockSuite) TestManagedAnalyticsUserCredentialsResetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/user/fakeUserID/credentials/reset",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "n3wS3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "user", "credentials-reset", "fakeAnalyticsID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: n3wS3cr3t"))
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsRoleListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/roles",
		httpmock.NewStringResponder(200, `["analyst", "admin"]`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "role", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("analyst"))
}

func (ms *MockSuite) TestManagedAnalyticsRoleListInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "role", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Permissions ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsPermissionListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/permissions",
		httpmock.NewStringResponder(200, `{"names": ["indices:data/read/get", "indices:data/write/index"]}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "permission", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("indices:data/read/get"))
}

func (ms *MockSuite) TestManagedAnalyticsPermissionListInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "permission", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Patterns ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsPatternListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/pattern",
		httpmock.NewStringResponder(200, `["fakePatternID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/pattern/fakePatternID",
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*", "maxIndexCount": 10}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakePatternID"))
}

func (ms *MockSuite) TestManagedAnalyticsPatternGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/pattern/fakePatternID",
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*", "maxIndexCount": 10}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "get", "fakeAnalyticsID", "fakePatternID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakePatternID"))
}

func (ms *MockSuite) TestManagedAnalyticsPatternCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	// Use simple responder: example JSON includes maxIndexCount which affects the POST body
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/pattern",
		httpmock.NewStringResponder(200, `{"id": "fakePatternID", "pattern": "logs-*", "maxIndexCount": 0}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--pattern", "logs-*")

	require.CmpNoError(err)
	assert.String(out, `✅ Pattern created successfully (id: fakePatternID)`)
}

func (ms *MockSuite) TestManagedAnalyticsPatternCreateInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--pattern", "logs-*")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedAnalyticsPatternDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/opensearch/fakeAnalyticsID/pattern/fakePatternID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "pattern", "delete", "fakeAnalyticsID", "fakePatternID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Pattern deleted successfully`)
}

// ── Certificate ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsCertificateGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/certificates",
		httpmock.NewStringResponder(200, `{"ca": "-----BEGIN CERTIFICATE-----\nMIIB..."}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "certificate", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("-----BEGIN CERTIFICATE-----"))
}

func (ms *MockSuite) TestManagedAnalyticsCertificateGetInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// opensearch is not in CertificateValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "certificate", "get", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Backup ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsBackupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/backup",
		httpmock.NewStringResponder(200, `["fakeBackupID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "Daily backup"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeBackupID"))
}

func (ms *MockSuite) TestManagedAnalyticsBackupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/clickhouse/fakeAnalyticsID/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "Daily backup", "size": {"value": 80, "unit": "GB"}, "regions": [{"name": "DE"}]}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "get", "fakeAnalyticsID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeBackupID"))
}

func (ms *MockSuite) TestManagedAnalyticsBackupListInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// kafka is not in BackupValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "backup", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Topics ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsTopicListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic",
		httpmock.NewStringResponder(200, `["fakeTopicID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 1, "minInsyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeTopicID"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 1, "minInsyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "get", "fakeAnalyticsID", "fakeTopicID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeTopicID"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic",
		tdhttpmock.JSONBody(td.JSON(`{"name": "my-topic", "partitions": 3, "replication": 1, "minInsyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 1, "minInsyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "create", "fakeAnalyticsID",
		"--cloud-project", "fakeProjectID",
		"--name", "my-topic",
		"--partitions", "3",
		"--replication", "1",
		"--min-insync-replicas", "1",
		"--retention-bytes", "-1",
		"--retention-hours", "168",
	)

	require.CmpNoError(err)
	assert.String(out, `✅ Topic created successfully (id: fakeTopicID)`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, `{"id": "fakeTopicID", "name": "my-topic", "partitions": 3, "replication": 1, "minInsyncReplicas": 1, "retentionBytes": -1, "retentionHours": 168}`),
	)
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "edit", "fakeAnalyticsID", "fakeTopicID",
		"--cloud-project", "fakeProjectID",
		"--retention-hours", "72",
	)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topic/fakeTopicID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "delete", "fakeAnalyticsID", "fakeTopicID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic deleted successfully`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicListInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// clickhouse is not in TopicValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "clickhouse"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Topic ACLs ────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedAnalyticsTopicACLListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topicAcl",
		httpmock.NewStringResponder(200, `["fakeACLID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topicAcl/fakeACLID",
		httpmock.NewStringResponder(200, `{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "read"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "list", "fakeAnalyticsID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeACLID"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topicAcl/fakeACLID",
		httpmock.NewStringResponder(200, `{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "read"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "get", "fakeAnalyticsID", "fakeACLID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeACLID"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topicAcl",
		tdhttpmock.JSONBody(td.JSON(`{"permission": "read", "topic": "my-topic", "username": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeACLID", "username": "myuser", "topic": "my-topic", "permission": "read"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--permission", "read", "--topic", "my-topic", "--username", "myuser")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic ACL created successfully (id: fakeACLID)`)
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLCreateInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		// opensearch is not in TopicValidEngines
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "opensearch"}`),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "create", "fakeAnalyticsID", "--cloud-project", "fakeProjectID", "--permission", "read", "--topic", "my-topic", "--username", "myuser")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedAnalyticsTopicACLDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeAnalyticsID",
		httpmock.NewStringResponder(200, `{"id": "fakeAnalyticsID", "engine": "kafka"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/kafka/fakeAnalyticsID/topicAcl/fakeACLID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-analytics", "topic-acl", "delete", "fakeAnalyticsID", "fakeACLID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Topic ACL deleted successfully`)
}
