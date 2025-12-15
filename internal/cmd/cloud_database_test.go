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

func (ms *MockSuite) TestCloudDatabaseCreateCmd(assert, require *td.T) {
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

	out, err := cmd.Execute("cloud", "database-service", "create", "--cloud-project", "fakeProjectID", "--engine", "mysql", "--version", "8", "--plan", "essential", "--nodes-list", "db1-4:DE")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}

func (ms *MockSuite) TestCloudDatabaseEditCmd(assert, require *td.T) {
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
			"disk": {
				"type": "high-speed",
				"size": 80
			},
			"storage": {
				"type": "high-speed",
				"size": {
					"unit": "GB",
					"value": 80
				}
			},
			"id": "9d1e7e8a-9abd-11f0-ab87-0050568ce122",
			"engine": "mysql",
			"category": "operational",
			"ipRestrictions": [],
			"status": "READY",
			"nodes": [
				{
					"id": "b282a940-9abd-11f0-a9f2-0050568ce122",
					"createdAt": "2025-09-22T14:16:18.558113+02:00",
					"flavor": "db1-4",
					"name": "mysql-aa3b2t56-aa5f9a639-1.database.cloud.ovh.net",
					"port": 2014,
					"region": "DE",
					"status": "READY"
				}
			],
			"nodeNumber": 1,
			"description": "Default description",
			"version": "6",
			"networkType": "public",
			"flavor": "db1-4",
			"maintenanceTime": "13:16:00",
			"backupTime": "09:10:00",
			"backups": {
				"time": "09:10:00",
				"regions": [
					"DE",
					"GRA"
				],
				"retentionDays": 2,
				"pitr": "2025-09-24T11:10:11+02:00"
			},
			"enablePrometheus": false,
			"deletionProtection": false
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/mysql/fakeDatabaseID",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"backupTime": "09:10:00",
				"backups": {
					"regions": [
						"DE",
						"GRA"
					],
					"time": "09:10:00"
				},
				"deletionProtection": false,
				"description": "Default description",
				"disk": {
					"size": 80
				},
				"enablePrometheus": false,
				"flavor": "db1-4",
				"ipRestrictions": [],
				"maintenanceTime": "13:16:00",
				"nodeNumber": 1,
				"plan": "discovery",
				"storage": {
					"size": {
						"unit": "GB",
						"value": 80
					}
				},
				"version": "8"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "database-service", "edit", "fakeDatabaseID", "--cloud-project", "fakeProjectID", "--version", "8", "--plan", "discovery", "--yaml")

	require.CmpNoError(err)
	assert.String(out, `message: ✅ Resource updated successfully
`)
}
