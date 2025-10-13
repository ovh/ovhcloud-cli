// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudStorageS3BulkDeletePrefixCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["BHS"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS",
		httpmock.NewStringResponder(200, `{
			"name": "BHS",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "storage",
					"status": "UP"
				},
				{
					"name": "storage-s3-high-perf",
					"status": "UP"
				},
				{
					"name": "storage-s3-standard",
					"status": "UP"
				}
			],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer",
		httpmock.NewStringResponder(200, `{
			"name": "fakeContainer",
			"virtualHost": "https://fakeContainer.test.ovh.net/",
			"ownerId": 0,
			"objectsCount": 15,
			"objectsSize": 4147089,
			"objects": [
				{"key": "logs/log1.txt"},
				{"key": "logs/log2.txt"},
				{"key": "images/img1.png"}
			],
			"region": "BHS",
			"createdAt": "2025-02-10T14:24:12Z"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object?prefix=logs%2F",
		httpmock.NewStringResponder(200, `[
			{"key": "logs/log1.txt"},
			{"key": "logs/log2.txt"}
		]`).Then(httpmock.NewStringResponder(200, `[]`)),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"objects": [
					{"key": "logs/log1.txt"},
					{"key": "logs/log2.txt"}
				]
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "storage-s3", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--prefix", "logs/", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Objects deleted successfully"}`))
}

func (ms *MockSuite) TestCloudStorageS3BulkDeleteAllCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["BHS"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS",
		httpmock.NewStringResponder(200, `{
			"name": "BHS",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "storage",
					"status": "UP"
				},
				{
					"name": "storage-s3-high-perf",
					"status": "UP"
				},
				{
					"name": "storage-s3-standard",
					"status": "UP"
				}
			],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer",
		httpmock.NewStringResponder(200, `{
			"name": "fakeContainer",
			"virtualHost": "https://fakeContainer.test.ovh.net/",
			"ownerId": 0,
			"objectsCount": 15,
			"objectsSize": 4147089,
			"objects": [
				{"key": "logs/log1.txt"},
				{"key": "logs/log2.txt"},
				{"key": "images/img1.png"}
			],
			"region": "BHS",
			"createdAt": "2025-02-10T14:24:12Z"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/object",
		httpmock.NewStringResponder(200, `[
			{"key": "logs/log1.txt"},
			{"key": "logs/log2.txt"},
			{"key": "images/img1.png"}
		]`).Then(httpmock.NewStringResponder(200, `[]`)),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/region/BHS/storage/fakeContainer/bulkDeleteObjects",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"objects": [
					{"key": "logs/log1.txt"},
					{"key": "logs/log2.txt"},
					{"key": "images/img1.png"}
				]
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "storage-s3", "bulk-delete", "fakeContainer", "--cloud-project", "fakeProjectID", "--all", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Objects deleted successfully"}`))
}
