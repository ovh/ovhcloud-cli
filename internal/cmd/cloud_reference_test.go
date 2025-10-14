// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudReferenceRancherVersionsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/rancher/version",
		httpmock.NewStringResponder(200, `[
			{
				"cause": "END_OF_SUPPORT",
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.9.4",
				"message": "This Rancher version is no more supported, creations and updates to this version have been disabled.",
				"name": "2.9.4",
				"status": "UNAVAILABLE"
			},
			{
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.10.4",
				"name": "2.10.4",
				"status": "AVAILABLE"
			},
			{
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.11.3",
				"name": "2.11.3",
				"status": "AVAILABLE"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-versions", "--json", "--cloud-project", "fakeProjectID", "--filter", `status=="AVAILABLE"`)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.10.4",
			"name": "2.10.4",
			"status": "AVAILABLE"
		},
		{
			"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.11.3",
			"name": "2.11.3",
			"status": "AVAILABLE"
		}
	]`))
}

func (ms *MockSuite) TestCloudReferenceRancherPlansListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/rancher/plan",
		httpmock.NewStringResponder(200, `[
			{
				"name": "OVHCLOUD_EDITION",
				"status": "AVAILABLE"
			},
			{
				"name": "STANDARD",
				"status": "AVAILABLE"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-plans", "--cloud-project", "fakeProjectID", "--format", "name")

	require.CmpNoError(err)
	assert.String(out, `"OVHCLOUD_EDITION"
"STANDARD"
`)
}

func (ms *MockSuite) TestCloudReferenceRancherPlansListCmdWithNil(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/rancher/plan",
		httpmock.NewStringResponder(200, `[
			{
				"name": "OVHCLOUD_EDITION",
				"status": "AVAILABLE"
			},
			{
				"name": "STANDARD",
				"status": "AVAILABLE"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-plans", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────┬───────────┬─────────┐
│       name       │  status   │ message │
├──────────────────┼───────────┼─────────┤
│ OVHCLOUD_EDITION │ AVAILABLE │         │
│ STANDARD         │ AVAILABLE │         │
└──────────────────┴───────────┴─────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesPlansListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/database/capabilities",
		httpmock.NewStringResponder(200, `{
			"plans": [
				{
					"lifecycle": {
						"status": "STABLE",
						"startDate": "2023-12-07"
					},
					"name": "production",
					"description": "Production grade plan",
					"backupRetention": "P14D",
					"order": 4,
					"tags": []
				},
				{
					"lifecycle": {
						"status": "STABLE",
						"startDate": "2021-07-01"
					},
					"name": "enterprise",
					"description": "Enterprise plan",
					"backupRetention": "P30D",
					"order": 5,
					"tags": []
				},
				{
					"lifecycle": {
						"status": "STABLE",
						"startDate": "2023-12-07"
					},
					"name": "advanced",
					"description": "Advanced grade plan",
					"backupRetention": "P30D",
					"order": 6,
					"tags": []
				}
			]
		}`).Once())

	out, err := cmd.Execute("cloud", "reference", "database", "list-plans", "--cloud-project", "fakeProjectID", "--filter", `lifecycle.startDate>"2022-01-01"`)

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬───────────────────────┬────────┬─────────────────┐
│    name    │      description      │ status │ backupRetention │
├────────────┼───────────────────────┼────────┼─────────────────┤
│ production │ Production grade plan │ STABLE │ P14D            │
│ advanced   │ Advanced grade plan   │ STABLE │ P30D            │
└────────────┴───────────────────────┴────────┴─────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesFlavorsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/database/capabilities",
		httpmock.NewStringResponder(200, `{
			"flavors": [
				{
					"lifecycle": {
						"status": "STABLE",
						"startDate": "2023-12-07"
					},
					"name": "db2-free",
					"core": 0,
					"memory": 0,
					"storage": 512,
					"specifications": {
						"core": 0,
						"memory": {
							"unit": "MB",
							"value": 0
						},
						"storage": {
							"unit": "MB",
							"value": 512
						}
					},
					"order": 0,
					"tags": []
				},
				{
					"lifecycle": {
						"status": "STABLE",
						"startDate": "2023-12-07"
					},
					"name": "db2-2",
					"core": 1,
					"memory": 2,
					"storage": 10,
					"specifications": {
						"core": 1,
						"memory": {
							"unit": "GB",
							"value": 2
						},
						"storage": {
							"unit": "GB",
							"value": 10
						}
					},
					"order": 3,
					"tags": []
				}
			]
		}`).Once())

	out, err := cmd.Execute("cloud", "reference", "database", "list-node-flavors", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────┬──────┬────────┬─────────┐
│   name   │ core │ memory │ storage │
├──────────┼──────┼────────┼─────────┤
│ db2-free │ 0    │ 0 MB   │ 512 MB  │
│ db2-2    │ 1    │ 2 GB   │ 10 GB   │
└──────────┴──────┴────────┴─────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesEnginesListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/database/capabilities",
		httpmock.NewStringResponder(200, `{
			"engines": [
				{
					"name": "postgresql",
					"storage": "replicated",
					"versions": [
						"13",
						"14",
						"15",
						"16",
						"17"
					],
					"defaultVersion": "17",
					"description": "object-relational database management system",
					"sslModes": [
						"require"
					],
					"category": "operational"
				},
				{
					"name": "mysql",
					"storage": "replicated",
					"versions": [
						"8"
					],
					"defaultVersion": "8",
					"description": "relational database management system",
					"sslModes": [
						"REQUIRED"
					],
					"category": "operational"
				},
				{
					"name": "mongodb",
					"storage": "replicated",
					"versions": [
						"4.4",
						"5.0",
						"6.0",
						"7.0",
						"8.0"
					],
					"defaultVersion": "8.0",
					"description": "document-based database management system",
					"sslModes": [
						"required"
					],
					"category": "operational"
				}
			]
		}`).Once())

	out, err := cmd.Execute("cloud", "reference", "database", "list-engines", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬──────────────────────────────────────────────┬─────────────┬─────────────────────────────┬────────────────┐
│    name    │                 description                  │  category   │          versions           │ defaultVersion │
├────────────┼──────────────────────────────────────────────┼─────────────┼─────────────────────────────┼────────────────┤
│ postgresql │ Object-Relational Database Management System │ operational │ 13 | 14 | 15 | 16 | 17      │ 17             │
│ mysql      │ Relational Database Management System        │ operational │ 8                           │ 8              │
│ mongodb    │ Document-Based Database Management System    │ operational │ 4.4 | 5.0 | 6.0 | 7.0 | 8.0 │ 8.0            │
└────────────┴──────────────────────────────────────────────┴─────────────┴─────────────────────────────┴────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}
