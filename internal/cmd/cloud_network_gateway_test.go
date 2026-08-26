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

func (ms *MockSuite) TestCloudGatewayV2ListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway",
		httpmock.NewStringResponder(200, `[
			{
				"id": "gw-12345",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-gateway",
					"location": {"region": "GRA11"},
					"status": "ACTIVE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "network", "gateway", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "gw-12345",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-gateway",
				"location": {"region": "GRA11"},
				"status": "ACTIVE"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudGatewayV2GetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway/gw-12345",
		httpmock.NewStringResponder(200, `{
			"id": "gw-12345",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-gateway",
				"location": {"region": "GRA11"},
				"status": "ACTIVE"
			}
		}`))

	out, err := cmd.Execute("cloud", "network", "gateway", "get", "--cloud-project", "fakeProjectID", "gw-12345", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "gw-12345",
		"resourceStatus": "READY",
		"currentState": {
			"name": "my-gateway",
			"location": {"region": "GRA11"},
			"status": "ACTIVE"
		}
	}`))
}

func (ms *MockSuite) TestCloudGatewayV2CreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-gateway",
					"location": {
						"region": "GRA11"
					},
					"externalGateway": {
						"enabled": true,
						"model": "S"
					}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "gw-12345"}`),
	)

	out, err := cmd.Execute("cloud", "network", "gateway", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-gateway", "--region", "GRA11", "--external-gateway-enabled", "--external-gateway-model", "S")
	require.CmpNoError(err)
	assert.String(out, `⚡️ Gateway creation started successfully (id: gw-12345)`)
}

func (ms *MockSuite) TestCloudGatewayV2EditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway/gw-12345",
		httpmock.NewStringResponder(200, `{
			"id": "gw-12345",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-gateway",
				"description": "old description",
				"externalGateway": {
					"enabled": true,
					"model": "S"
				}
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway/gw-12345",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"name": "renamed-gateway",
					"description": "old description",
					"externalGateway": {
						"enabled": true,
						"model": "S"
					}
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "network", "gateway", "edit", "--cloud-project", "fakeProjectID", "gw-12345", "--name", "renamed-gateway")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudGatewayV2DeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/gateway/gw-12345",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "network", "gateway", "delete", "--cloud-project", "fakeProjectID", "gw-12345")
	require.CmpNoError(err)
	assert.String(out, `✅ Gateway gw-12345 is being deleted…`)
}
