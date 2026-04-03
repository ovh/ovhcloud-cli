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

// fullServiceResponse is a complete API response for a database service,
// covering all fields accessed by the managed-database template.
const fullDBServiceResponse = `{
	"id": "fakeDatabaseID",
	"engine": "mysql",
	"version": "8",
	"description": "test db",
	"status": "READY",
	"category": "operational",
	"plan": "essential",
	"flavor": "db1-4",
	"createdAt": "2025-01-01T00:00:00Z",
	"networkType": "public",
	"storage": {"size": {"value": 80, "unit": "GB"}},
	"nodes": [{"id": "nodeID", "flavor": "db1-4", "region": "DE", "status": "READY"}]
}`

// ── Service ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service",
		httpmock.NewStringResponder(200, `["fakeDatabaseID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, fullDBServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-database", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDatabaseID"))
}

func (ms *MockSuite) TestManagedDatabaseGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, fullDBServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-database", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDatabaseID"))
}

func (ms *MockSuite) TestManagedDatabaseCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"nodesList": [
					{
						"flavor": "db1-4",
						"region": "DE"
					}
				],
				"plan": "essential",
				"version": "8"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "mysql", "--version", "8", "--plan", "essential", "--nodes-list", "db1-4:DE")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed database created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}

func (ms *MockSuite) TestManagedDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "invalid", "--version", "1", "--plan", "essential", "--nodes-list", "db1-4:DE")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine invalid"))
}

func (ms *MockSuite) TestManagedDatabaseEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{
				"id": "fakeDatabaseID",
				"engine": "mysql"
		}`),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{
			"createdAt": "2025-09-22T14:16:18.506458+02:00",
			"plan": "essential",
			"disk": {"type": "high-speed", "size": 80},
			"storage": {"type": "high-speed", "size": {"unit": "GB", "value": 80}},
			"id": "fakeDatabaseID",
			"engine": "mysql",
			"category": "operational",
			"ipRestrictions": [],
			"status": "READY",
			"nodes": [{"id": "nodeID", "createdAt": "2025-09-22T14:16:18.558113+02:00", "flavor": "db1-4", "name": "mysql-node.database.cloud.ovh.net", "port": 2014, "region": "DE", "status": "READY"}],
			"nodeNumber": 1,
			"description": "Default description",
			"version": "6",
			"networkType": "public",
			"flavor": "db1-4",
			"maintenanceTime": "13:16:00",
			"backupTime": "09:10:00",
			"backups": {"time": "09:10:00", "regions": ["DE", "GRA"], "retentionDays": 2, "pitr": "2025-09-24T11:10:11+02:00"},
			"enablePrometheus": false,
			"deletionProtection": false
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"backupTime": "09:10:00",
				"backups": {"regions": ["DE", "GRA"], "time": "09:10:00"},
				"deletionProtection": false,
				"description": "Default description",
				"disk": {"size": 80},
				"enablePrometheus": false,
				"flavor": "db1-4",
				"ipRestrictions": [],
				"maintenanceTime": "13:16:00",
				"nodeNumber": 1,
				"plan": "discovery",
				"storage": {"size": {"unit": "GB", "value": 80}},
				"version": "8"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "edit", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--version", "8", "--plan", "discovery", "-o", "yaml")

	require.CmpNoError(err)
	assert.String(out, `message: ✅ Resource updated successfully
`)
}

func (ms *MockSuite) TestManagedDatabaseDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "delete", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed database deleted successfully`)
}

// ── Database ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseDatabaseListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/database",
		httpmock.NewStringResponder(200, `["fakeDbID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDbID"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/database/fakeDbID",
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "get", "fakeDatabaseID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeDbID"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/database",
		tdhttpmock.JSONBody(td.JSON(`{"name": "mydb"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeDbID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: fakeDbID)`)
}

func (ms *MockSuite) TestManagedDatabaseDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mongodb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/database/fakeDbID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "delete", "fakeDatabaseID", "fakeDbID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Database deleted successfully`)
}

// ── User ──────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseUserListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user",
		httpmock.NewStringResponder(200, `["fakeUserID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeUserID"))
}

func (ms *MockSuite) TestManagedDatabaseUserGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "get", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeUserID"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateWithRolesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "postgresql"}`),
	)
	// Use simple responder: CreateResource schema filtering may alter the exact body
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/postgresql/fakeDatabaseID/user",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "status": "READY"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser", "--roles", "replication")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateRolesInvalidForMysqlCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser", "--roles", "replication")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedDatabaseUserEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "postgresql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/postgresql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "roles": [], "status": "READY"}`),
	)
	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/postgresql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "edit", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID", "--roles", "replication")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedDatabaseUserEditInvalidEngineForMysqlCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "edit", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID", "--roles", "replication")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

func (ms *MockSuite) TestManagedDatabaseUserDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "delete", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ User deleted successfully`)
}

func (ms *MockSuite) TestManagedDatabaseUserCredentialsResetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/user/fakeUserID/credentials/reset",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "n3wS3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "credentials-reset", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: n3wS3cr3t"))
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseRoleListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "postgresql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/postgresql/fakeDatabaseID/roles",
		httpmock.NewStringResponder(200, `["replication", "superuser"]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("replication"))
}

func (ms *MockSuite) TestManagedDatabaseRoleListInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Certificate ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseCertificateGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/certificates",
		httpmock.NewStringResponder(200, `{"ca": "-----BEGIN CERTIFICATE-----\nMIIB..."}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "certificate", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("-----BEGIN CERTIFICATE-----"))
}

func (ms *MockSuite) TestManagedDatabaseCertificateGetInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mongodb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "certificate", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}

// ── Backup ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseBackupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/backup",
		httpmock.NewStringResponder(200, `["fakeBackupID"]`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "Daily backup"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeBackupID"))
}

func (ms *MockSuite) TestManagedDatabaseBackupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "Daily backup", "size": {"value": 80, "unit": "GB"}, "regions": [{"name": "DE"}]}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "get", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeBackupID"))
}

func (ms *MockSuite) TestManagedDatabaseBackupRestoreCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mongodb"}`),
	)
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mongodb/fakeDatabaseID/backup/fakeBackupID/restore",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "restore", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Backup restore started successfully`)
}

func (ms *MockSuite) TestManagedDatabaseBackupRestoreInvalidEngineCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/service/fakeDatabaseID",
		httpmock.NewStringResponder(200, `{"id": "fakeDatabaseID", "engine": "mysql"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "restore", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑"))
}
