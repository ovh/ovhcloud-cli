// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudContainerRegistryListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry",
		httpmock.NewStringResponder(200, `[
			{
				"createdAt": "2025-08-22T09:24:18.953364Z",
				"deliveredAt": "2025-08-22T09:26:54.540629Z",
				"iamEnabled": false,
				"id": "0b1b2dc2-952b-11f0-afd9-0050568ce122",
				"name": "ZuperRegistry",
				"region": "EU-WEST-PAR",
				"size": 0,
				"status": "READY",
				"updatedAt": "2025-08-22T09:28:41.468178Z",
				"url": "https://fake.url.bhs5.container-registry.ovh.net",
				"version": "2.12.2"
			}
		]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA", "EU-WEST-PAR"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA",
		httpmock.NewStringResponder(200, `{
			"name": "GRA",
			"type": "region",
			"status": "UP",
			"services": [],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "GRA"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/EU-WEST-PAR",
		httpmock.NewStringResponder(200, `{
			"name": "EU-WEST-PAR",
			"type": "region-3-az",
			"status": "UP",
			"services": [],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "EU",
			"availabilityZones": [],
			"datacenterLocation": "EU-WEST-PAR"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/containerRegistry/0b1b2dc2-952b-11f0-afd9-0050568ce122/plan",
		httpmock.NewStringResponder(200, `{
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
		}`).Once())

	out, err := cmd.Execute("cloud", "container-registry", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────────────────────────────────┬───────────────┬─────────────┬───────┬────────────────┬─────────┬────────┐
│                  id                  │     name      │   region    │ plan  │ deploymentMode │ version │ status │
├──────────────────────────────────────┼───────────────┼─────────────┼───────┼────────────────┼─────────┼────────┤
│ 0b1b2dc2-952b-11f0-afd9-0050568ce122 │ ZuperRegistry │ EU-WEST-PAR │ SMALL │ 3-AZ           │ 2.12.2  │ READY  │
└──────────────────────────────────────┴───────────────┴─────────────┴───────┴────────────────┴─────────┴────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}
