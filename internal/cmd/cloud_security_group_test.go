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

func (ms *MockSuite) TestCloudSecurityGroupListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup",
		httpmock.NewStringResponder(200, `[
			{
				"id": "sg-12345",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-sg",
					"location": {
						"region": "GRA11"
					}
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "network", "security-group", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "sg-12345",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-sg",
				"location": {"region": "GRA11"}
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudSecurityGroupGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup/sg-12345",
		httpmock.NewStringResponder(200, `{
			"id": "sg-12345",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-sg",
				"location": {"region": "GRA11"}
			}
		}`))

	out, err := cmd.Execute("cloud", "network", "security-group", "get", "--cloud-project", "fakeProjectID", "sg-12345", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "sg-12345",
		"resourceStatus": "READY",
		"currentState": {
			"name": "my-sg",
			"location": {"region": "GRA11"}
		}
	}`))
}

func (ms *MockSuite) TestCloudSecurityGroupCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-sg",
					"description": "test group",
					"location": {
						"region": "GRA11"
					}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "sg-12345"}`),
	)

	out, err := cmd.Execute("cloud", "network", "security-group", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-sg", "--description", "test group", "--region", "GRA11")
	require.CmpNoError(err)
	assert.String(out, `✅ Security group created successfully (id: sg-12345)`)
}

func (ms *MockSuite) TestCloudSecurityGroupEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup/sg-12345",
		httpmock.NewStringResponder(200, `{
			"id": "sg-12345",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-sg",
				"description": "old description"
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup/sg-12345",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"name": "renamed-sg",
					"description": "old description"
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "network", "security-group", "edit", "--cloud-project", "fakeProjectID", "sg-12345", "--name", "renamed-sg")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudSecurityGroupDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/securityGroup/sg-12345",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "network", "security-group", "delete", "--cloud-project", "fakeProjectID", "sg-12345")
	require.CmpNoError(err)
	assert.String(out, `✅ Security group sg-12345 is being deleted…`)
}
