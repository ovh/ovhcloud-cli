// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
)

// fullDBServiceResponse is a complete API response for a database service,
// covering all fields accessed by the managed-database template.
const fullDBServiceResponse = `{
	"id": "fakeDatabaseID",
	"createdAt": "2025-01-01T00:00:00Z",
	"resourceStatus": "READY",
	"currentState": {
		"description": "test db",
		"plan": "essential",
		"version": "8",
		"flavor": "db1-4",
		"status": "READY",
		"nodesCount": 1,
		"storage": {"size": {"value": 80, "unit": "GB"}},
		"network": {"authorizedIps": {"rules": []}},
		"endpoints": [{"component": "MYSQL", "domain": "mysql.database.cloud.ovh.net", "port": 20184, "ssl": true, "uri": "mysql://<username>:<password>@mysql.database.cloud.ovh.net:20184/defaultdb"}],
		"nodes": [{"id": "nodeID", "flavor": "db1-4", "location": {"region": "eu-west-gra"}, "status": "READY"}],
		"backup": {"time": "09:10:00", "regions": ["eu-west-gra", "eu-west-gra"], "retentionDays": 2}
	},
	"targetSpec": {"description": "test db", "plan": "essential", "version": "8", "flavor": "db1-4", "deletionProtected": false}
}`

// registerDatabaseAnalyticsProbe mocks the lookup of a service's engine: the route
// of its engine answers the service, the routes of the other engines answer 404.
func registerDatabaseAnalyticsProbe(category string, engines []string, engine, serviceID, service string) {
	for _, e := range engines {
		responder := httpmock.NewStringResponder(http.StatusNotFound,
			fmt.Sprintf(`{"class": "Client::NotFound::ServiceNotFound", "message": "service %s not found"}`, serviceID))
		if e == engine {
			responder = httpmock.NewStringResponder(http.StatusOK, service)
		}
		httpmock.RegisterResponder(http.MethodGet,
			fmt.Sprintf("https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/%s/%s/%s", category, e, serviceID),
			responder,
		)
	}
}

func registerManagedDatabaseProbe(engine string) {
	registerDatabaseAnalyticsProbe("managedDatabase", cloud.ManagedDatabaseValidEngines, engine, "fakeDatabaseID", fullDBServiceResponse)
}

// ── Service ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseListCmd(assert, require *td.T) {
	for _, engine := range cloud.ManagedDatabaseValidEngines {
		body := `[]`
		if engine == "mysql" {
			body = "[" + fullDBServiceResponse + "]"
		}
		httpmock.RegisterResponder(http.MethodGet,
			"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/"+engine,
			httpmock.NewStringResponder(200, body),
		)
	}

	out, err := cmd.Execute("cloud", "managed-database", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────────┬────────┬─────────┬─────────────┬────────┐
│       id       │ engine │ version │ description │ status │
├────────────────┼────────┼─────────┼─────────────┼────────┤
│ fakeDatabaseID │ mysql  │ 8       │ test db     │ READY  │
└────────────────┴────────┴─────────┴─────────────┴────────┘`[1:])
}

func (ms *MockSuite) TestManagedDatabaseGetCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")

	out, err := cmd.Execute("cloud", "managed-database", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(
		td.Contains("fakeDatabaseID"),
		td.Contains("mysql 8"),
		td.Contains("80 GB"),
		td.Contains("mysql.database.cloud.ovh.net:20184"),
		td.Contains("eu-west-gra"),
	))
}

func (ms *MockSuite) TestManagedDatabaseGetNotFoundCmd(assert, require *td.T) {
	registerDatabaseAnalyticsProbe("managedDatabase", cloud.ManagedDatabaseValidEngines, "", "fakeDatabaseID", "")

	out, err := cmd.Execute("cloud", "managed-database", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 failed to fetch database service: service fakeDatabaseID not found"))
}

func (ms *MockSuite) TestManagedDatabaseGetForbiddenCmd(assert, require *td.T) {
	registerDatabaseAnalyticsProbe("managedDatabase", cloud.ManagedDatabaseValidEngines, "", "fakeDatabaseID", "")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID",
		httpmock.NewStringResponder(http.StatusForbidden, `{"class": "Client::Forbidden", "message": "forbidden"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("forbidden"))
}

func (ms *MockSuite) TestManagedDatabaseCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"flavor": "db1-4",
					"nodesCount": 1,
					"plan": "essential",
					"region": "eu-west-gra",
					"version": "8"
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "mysql", "--version", "8", "--plan", "essential", "--nodes-list", "db1-4:eu-west-gra")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed database created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}

func (ms *MockSuite) TestManagedDatabaseCreateWithNodesPatternCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"backup": {"regions": ["eu-west-gra", "eu-west-sbg"], "time": "02:00:00"},
					"description": "my db",
					"flavor": "b3-8",
					"maintenanceWindow": {"startTimeUtc": "03:00:00"},
					"network": {
						"authorizedIps": {"rules": [{"cidr": "1.2.3.4/32"}, {"cidr": "5.6.7.0/24"}]},
						"networkId": "fakeNetworkID",
						"subnetId": "fakeSubnetID"
					},
					"nodesCount": 3,
					"plan": "business",
					"region": "eu-west-gra",
					"storage": {"size": {"unit": "GB", "value": 160}},
					"version": "17"
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "postgresql",
		"--version", "17", "--plan", "business", "--description", "my db",
		"--nodes-pattern.flavor", "b3-8", "--nodes-pattern.number", "3", "--nodes-pattern.region", "eu-west-gra", "--disk-size", "160",
		"--ip-restrictions", "1.2.3.4/32,5.6.7.0/24", "--network-id", "fakeNetworkID", "--subnet-id", "fakeSubnetID",
		"--backups-regions", "eu-west-gra,eu-west-sbg", "--backups-time", "02:00:00", "--maintenance-time", "03:00:00")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed database created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}

func (ms *MockSuite) TestManagedDatabaseCreateHeterogeneousNodesListCmd(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "mysql", "--version", "8", "--plan", "business", "--nodes-list", "db1-4:eu-west-gra,db1-8:eu-west-gra")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid nodes list: all nodes must have the same flavor and region"))
}

func (ms *MockSuite) TestManagedDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	out, err := cmd.Execute("cloud", "managed-database", "create", "--cloud-project", "fakeProjectID", "--engine", "invalid", "--version", "1", "--plan", "essential", "--nodes-list", "db1-4:eu-west-gra")

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
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID",
		httpmock.NewStringResponder(200, fullDBServiceResponse),
	)

	out, err := cmd.Execute("cloud", "managed-database", "delete", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Managed database deleted successfully`)
}

// ── Database ──────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseDatabaseListCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/database",
		httpmock.NewStringResponder(200, `[{"id": "fakeDBID", "name": "mydb", "default": false}]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("mydb"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseGetCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/database/fakeDBID",
		httpmock.NewStringResponder(200, `{"id": "fakeDBID", "name": "mydb", "default": false}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "get", "fakeDatabaseID", "fakeDBID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("mydb"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseCreateCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/database",
		tdhttpmock.JSONBody(td.JSON(`{"name": "mydb"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeDBID", "name": "mydb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: fakeDBID)`)
}

func (ms *MockSuite) TestManagedDatabaseDatabaseCreateInvalidEngineCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mongodb")

	out, err := cmd.Execute("cloud", "managed-database", "database", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "mydb")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine mongodb, valid values: mysql, postgresql"))
}

func (ms *MockSuite) TestManagedDatabaseDatabaseDeleteCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/database/fakeDBID",
		httpmock.NewStringResponder(200, `{"id": "fakeDBID", "name": "mydb"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "database", "delete", "fakeDatabaseID", "fakeDBID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Database deleted successfully`)
}

// ── User ──────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseUserListCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/user",
		httpmock.NewStringResponder(200, `[{"id": "fakeUserID", "username": "myuser", "status": "READY", "roles": ["replication"], "createdAt": "2025-01-01T00:00:00Z"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("myuser"), td.Contains("replication")))
}

func (ms *MockSuite) TestManagedDatabaseUserGetCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "READY", "roles": ["replication"], "createdAt": "2025-01-01T00:00:00Z"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "get", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("myuser"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser"}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateWithRolesCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/user",
		tdhttpmock.JSONBody(td.JSON(`{"name": "myuser", "roles": ["replication"]}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "s3cr3t", "roles": ["replication"]}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser", "--roles", "replication")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: s3cr3t"))
}

func (ms *MockSuite) TestManagedDatabaseUserCreateRolesInvalidForMysqlCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")

	out, err := cmd.Execute("cloud", "managed-database", "user", "create", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--name", "myuser", "--roles", "replication")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 flag --roles is only supported for postgresql and mongodb engines"))
}

func (ms *MockSuite) TestManagedDatabaseUserEditCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "roles": [], "status": "READY"}`),
	)
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/user/fakeUserID",
		tdhttpmock.JSONBody(td.JSON(`{"roles": ["replication"]}`)),
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "roles": ["replication"], "status": "UPDATING"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "edit", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID", "--roles", "replication")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("✅ Resource updated successfully"))
}

func (ms *MockSuite) TestManagedDatabaseUserEditInvalidEngineForMysqlCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")

	out, err := cmd.Execute("cloud", "managed-database", "user", "edit", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID", "--roles", "replication")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine mysql, valid values: mongodb, postgresql, valkey"))
}

func (ms *MockSuite) TestManagedDatabaseUserDeleteCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/user/fakeUserID",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "status": "DELETING"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "delete", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ User deleted successfully`)
}

func (ms *MockSuite) TestManagedDatabaseUserCredentialsResetCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/mysql/fakeDatabaseID/user/fakeUserID/resetPassword",
		httpmock.NewStringResponder(200, `{"id": "fakeUserID", "username": "myuser", "password": "n3wS3cr3t"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "user", "credentials-reset", "fakeDatabaseID", "fakeUserID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Password: n3wS3cr3t"))
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseRoleListCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/roles",
		httpmock.NewStringResponder(200, `[{"id": "cmVwbGljYXRpb24", "name": "replication"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("replication"))
}

func (ms *MockSuite) TestManagedDatabaseRoleListAdvancedCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mongodb")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mongodb/fakeDatabaseID/roles?advanced=true",
		httpmock.NewStringResponder(200, `["clusterMonitor@admin", "readAnyDatabase@admin"]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--advanced")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("clusterMonitor@admin"), td.Contains("readAnyDatabase@admin")))
}

func (ms *MockSuite) TestManagedDatabaseRoleListAdvancedInvalidEngineCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--advanced")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 flag --advanced is only supported for mongodb engines"))
}

func (ms *MockSuite) TestManagedDatabaseRoleListInvalidEngineCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")

	out, err := cmd.Execute("cloud", "managed-database", "role", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine mysql, valid values: mongodb, postgresql"))
}

// ── Certificate ───────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseCertificateGetCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("postgresql")
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/postgresql/fakeDatabaseID/certificates",
		httpmock.NewStringResponder(200, `{"ca": "-----BEGIN CERTIFICATE-----\nfakeCA\n-----END CERTIFICATE-----"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "certificate", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("fakeCA"))
}

func (ms *MockSuite) TestManagedDatabaseCertificateGetInvalidEngineCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mongodb")

	out, err := cmd.Execute("cloud", "managed-database", "certificate", "get", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine mongodb, valid values: mysql, postgresql"))
}

// ── Backup ────────────────────────────────────────────────────────────────────

func (ms *MockSuite) TestManagedDatabaseBackupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/backup?serviceId=fakeDatabaseID",
		httpmock.NewStringResponder(200, `[{"id": "fakeBackupID", "serviceId": "fakeDatabaseID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "daily backup"}]`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "list", "fakeDatabaseID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("fakeBackupID"), td.Contains("daily backup")))
}

func (ms *MockSuite) TestManagedDatabaseBackupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "serviceId": "fakeDatabaseID", "createdAt": "2025-01-01T00:00:00Z", "type": "AUTOMATIC", "status": "READY", "description": "daily backup", "size": {"value": 1, "unit": "GB"}, "regions": ["eu-west-gra"]}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "get", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(td.Contains("fakeBackupID"), td.Contains("eu-west-gra")))
}

func (ms *MockSuite) TestManagedDatabaseBackupGetOfAnotherServiceCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedDatabase/backup/fakeBackupID",
		httpmock.NewStringResponder(200, `{"id": "fakeBackupID", "serviceId": "anotherDatabaseID"}`),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "get", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 backup fakeBackupID not found in service fakeDatabaseID"))
}

func (ms *MockSuite) TestManagedDatabaseBackupRestoreCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mongodb")
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mongodb/fakeDatabaseID/backup/fakeBackupID/restore",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "managed-database", "backup", "restore", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Backup restore started successfully`)
}

func (ms *MockSuite) TestManagedDatabaseBackupRestoreInvalidEngineCmd(assert, require *td.T) {
	registerManagedDatabaseProbe("mysql")

	out, err := cmd.Execute("cloud", "managed-database", "backup", "restore", "fakeDatabaseID", "fakeBackupID", "--cloud-project", "fakeProjectID")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("🛑 invalid engine mysql, valid values: mongodb"))
}
