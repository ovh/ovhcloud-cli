// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudQuotaGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/quota",
		httpmock.NewStringResponder(200, `{
			"id": "fakeProjectID",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"currentState": {
				"preventAutomaticQuotaUpgrade": false,
				"regions": [
					{
						"location": {"region": "GRA11"},
						"profile": "default",
						"usage": {
							"compute": {
								"instances": {"limit": 20, "unit": "COUNT", "used": 2},
								"cores": {"limit": 100, "unit": "COUNT", "used": 8},
								"memory": {"limit": 256000, "unit": "MB", "used": 16000}
							}
						}
					}
				]
			},
			"targetSpec": {
				"preventAutomaticQuotaUpgrade": false,
				"regions": [
					{"location": {"region": "GRA11"}, "profile": "default"}
				]
			}
		}`))

	out, err := cmd.Execute("cloud", "quota", "get", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "fakeProjectID",
		"checksum": "abc123",
		"resourceStatus": "READY",
		"currentState": {
			"preventAutomaticQuotaUpgrade": false,
			"regions": [
				{
					"location": {"region": "GRA11"},
					"profile": "default",
					"usage": {
						"compute": {
							"instances": {"limit": 20, "unit": "COUNT", "used": 2},
							"cores": {"limit": 100, "unit": "COUNT", "used": 8},
							"memory": {"limit": 256000, "unit": "MB", "used": 16000}
						}
					}
				}
			]
		},
		"targetSpec": {
			"preventAutomaticQuotaUpgrade": false,
			"regions": [
				{"location": {"region": "GRA11"}, "profile": "default"}
			]
		}
	}`))
}

func (ms *MockSuite) TestCloudQuotaEditFromFileCmd(assert, require *td.T) {
	// GET the current quota (without regions so the file-provided regions are
	// not appended to an existing slice by the merge logic).
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/quota",
		httpmock.NewStringResponder(200, `{
			"id": "fakeProjectID",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"preventAutomaticQuotaUpgrade": false
			}
		}`).Once())

	// The PUT body is the editable subset of the merged resource.
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/quota",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc123",
			"targetSpec": {
				"preventAutomaticQuotaUpgrade": true,
				"regions": [
					{"location": {"region": "GRA11"}, "profile": "large"}
				]
			}
		}`)),
		httpmock.NewStringResponder(200, ``).Once())

	tmpFile, err := os.CreateTemp("", "quota-edit-*.json")
	require.CmpNoError(err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"targetSpec": {
			"preventAutomaticQuotaUpgrade": true,
			"regions": [
				{"location": {"region": "GRA11"}, "profile": "large"}
			]
		}
	}`)
	require.CmpNoError(err)
	tmpFile.Close()

	out, err := cmd.Execute("cloud", "quota", "edit", "--cloud-project", "fakeProjectID", "--from-file", tmpFile.Name())
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}
