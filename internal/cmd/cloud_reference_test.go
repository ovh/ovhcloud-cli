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

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-versions", "-o", "json", "--cloud-project", "fakeProjectID", "--filter", `status=="AVAILABLE"`)

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

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-plans", "--cloud-project", "fakeProjectID", "-o", "name")

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
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesPlansListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/capabilities",
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
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesFlavorsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/capabilities",
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
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceDatabasesEnginesListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/database/capabilities",
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
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceContainerRegistryPlansListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/capabilities/containerRegistry",
		httpmock.NewStringResponder(200, `[
			{
				"regionName": "GRA",
				"regionType": "REGION-1-AZ",
				"plans": [
					{
						"code": "registry.s-plan-equivalent.hour.consumption",
						"createdAt": "2019-09-13T15:53:33.599585Z",
						"updatedAt": "2021-03-29T10:09:03.960847Z",
						"name": "SMALL",
						"id": "9f728ba5-998b-4401-ab0f-497cd8bc6a89",
						"registryLimits": {
							"imageStorage": 214748364800,
							"parallelRequest": 15
						},
						"features": {
							"vulnerability": false
						}
					},
					{
						"code": "registry.m-plan-equivalent.hour.consumption",
						"createdAt": "2019-09-13T15:53:33.601794Z",
						"updatedAt": "2023-12-04T11:03:43.109685Z",
						"name": "MEDIUM",
						"id": "c5ddc763-be75-48f7-b7ec-e923ca040bee",
						"registryLimits": {
							"imageStorage": 644245094400,
							"parallelRequest": 45
						},
						"features": {
							"vulnerability": true
						}
					}
				]
			},
			{
				"regionName": "DE",
				"regionType": "REGION-1-AZ",
				"plans": [
					{
						"code": "registry.s-plan-equivalent.hour.consumption",
						"createdAt": "2019-09-13T15:53:33.599585Z",
						"updatedAt": "2021-03-29T10:09:03.960847Z",
						"name": "SMALL",
						"id": "9f728ba5-998b-4401-ab0f-497cd8bc6a89",
						"registryLimits": {
							"imageStorage": 214748364800,
							"parallelRequest": 15
						},
						"features": {
							"vulnerability": false
						}
					},
					{
						"code": "registry.m-plan-equivalent.hour.consumption",
						"createdAt": "2019-09-13T15:53:33.601794Z",
						"updatedAt": "2023-12-04T11:03:43.109685Z",
						"name": "MEDIUM",
						"id": "c5ddc763-be75-48f7-b7ec-e923ca040bee",
						"registryLimits": {
							"imageStorage": 644245094400,
							"parallelRequest": 45
						},
						"features": {
							"vulnerability": true
						}
					}
				]
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "container-registry", "list-plans", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────────────────────────┬────────┬───────────────┬──────────────┬─────────────────┐
│                  id                  │  name  │ vulnerability │ imageStorage │ parallelRequest │
├──────────────────────────────────────┼────────┼───────────────┼──────────────┼─────────────────┤
│ 9f728ba5-998b-4401-ab0f-497cd8bc6a89 │ SMALL  │ false         │ 200G         │ 15              │
│ c5ddc763-be75-48f7-b7ec-e923ca040bee │ MEDIUM │ true          │ 600G         │ 45              │
└──────────────────────────────────────┴────────┴───────────────┴──────────────┴─────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceContainerRegistryPlansListCmdWithFilter(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/capabilities/containerRegistry",
		httpmock.NewStringResponder(200, `[
			{
				"regionName": "GRA",
				"regionType": "REGION-1-AZ",
				"plans": [
					{
						"code": "registry.s-plan-equivalent.hour.consumption",
						"name": "SMALL",
						"id": "9f728ba5-998b-4401-ab0f-497cd8bc6a89",
						"registryLimits": {
							"imageStorage": 214748364800,
							"parallelRequest": 15
						},
						"features": {
							"vulnerability": false
						}
					},
					{
						"code": "registry.m-plan-equivalent.hour.consumption",
						"name": "MEDIUM",
						"id": "c5ddc763-be75-48f7-b7ec-e923ca040bee",
						"registryLimits": {
							"imageStorage": 644245094400,
							"parallelRequest": 45
						},
						"features": {
							"vulnerability": true
						}
					}
				]
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "container-registry", "list-plans", "--cloud-project", "fakeProjectID", "--filter", `vulnerability==true`)

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────────────────────────┬────────┬───────────────┬──────────────┬─────────────────┐
│                  id                  │  name  │ vulnerability │ imageStorage │ parallelRequest │
├──────────────────────────────────────┼────────┼───────────────┼──────────────┼─────────────────┤
│ c5ddc763-be75-48f7-b7ec-e923ca040bee │ MEDIUM │ true          │ 600G         │ 45              │
└──────────────────────────────────────┴────────┴───────────────┴──────────────┴─────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestCloudReferenceContainerRegistryRegionsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/capabilities/containerRegistry",
		httpmock.NewStringResponder(200, `[
			{
				"regionName": "GRA",
				"regionType": "REGION-1-AZ",
				"plans": []
			},
			{
				"regionName": "DE",
				"regionType": "REGION-1-AZ",
				"plans": []
			},
			{
				"regionName": "EU-WEST-PAR",
				"regionType": "REGION-3-AZ",
				"plans": []
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "container-registry", "list-regions", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌─────────────┬──────┐
│    name     │ type │
├─────────────┼──────┤
│ GRA         │ 1-AZ │
│ DE          │ 1-AZ │
│ EU-WEST-PAR │ 3-AZ │
└─────────────┴──────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}
