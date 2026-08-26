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

// ---------------------------------------------------------------------------
// All public IPs – list (API v2)
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudPublicIPListAllCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp",
		httpmock.NewStringResponder(200, `[
			{"ip": "1.2.3.4", "type": "FLOATING_IP"},
			{"ip": "5.6.7.8", "type": "EXT_NET_IP"}
		]`))

	out, err := cmd.Execute("cloud", "ip", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"ip": "1.2.3.4", "type": "FLOATING_IP"},
		{"ip": "5.6.7.8", "type": "EXT_NET_IP"}
	]`))
}

// ---------------------------------------------------------------------------
// Floating IP – list / get / create / edit / delete (API v2)
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudFloatingIPListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating",
		httpmock.NewStringResponder(200, `[
			{
				"id": "1.2.3.4",
				"resourceStatus": "READY",
				"currentState": {"ip": "1.2.3.4", "status": "ACTIVE", "location": {"region": "GRA11"}}
			}
		]`))

	out, err := cmd.Execute("cloud", "ip", "floating", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "1.2.3.4",
			"resourceStatus": "READY",
			"currentState": {"ip": "1.2.3.4", "status": "ACTIVE", "location": {"region": "GRA11"}}
		}
	]`))
}

func (ms *MockSuite) TestCloudFloatingIPGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating/1.2.3.4",
		httpmock.NewStringResponder(200, `{
			"id": "1.2.3.4",
			"resourceStatus": "READY",
			"currentState": {"ip": "1.2.3.4", "status": "ACTIVE"}
		}`))

	out, err := cmd.Execute("cloud", "ip", "floating", "get", "1.2.3.4", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "1.2.3.4",
		"resourceStatus": "READY",
		"currentState": {"ip": "1.2.3.4", "status": "ACTIVE"}
	}`))
}

func (ms *MockSuite) TestCloudFloatingIPCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating",
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"description": "My floating IP",
				"location": {"region": "GRA11"}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "1.2.3.4"}`))

	out, err := cmd.Execute("cloud", "ip", "floating", "create", "--cloud-project", "fakeProjectID", "--region", "GRA11", "--description", "My floating IP")
	require.CmpNoError(err)
	assert.String(out, `✅ Floating IP creation started successfully (id: 1.2.3.4)`)
}

func (ms *MockSuite) TestCloudFloatingIPCreateCmdError(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating",
		httpmock.NewStringResponder(400, `{"message": "invalid region"}`))

	out, _ := cmd.Execute("cloud", "ip", "floating", "create", "--cloud-project", "fakeProjectID", "--region", "GRA11")
	assert.Cmp(out, td.Contains("failed to create floating IP"))
}

func (ms *MockSuite) TestCloudFloatingIPEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating/1.2.3.4",
		httpmock.NewStringResponder(200, `{
			"id": "1.2.3.4",
			"checksum": "abc123",
			"targetSpec": {"description": "old description"}
		}`).Once())

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating/1.2.3.4",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc123",
			"targetSpec": {"description": "new description"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "1.2.3.4"}`).Once())

	out, err := cmd.Execute("cloud", "ip", "floating", "edit", "1.2.3.4", "--cloud-project", "fakeProjectID", "--description", "new description")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudFloatingIPDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/floating/1.2.3.4",
		httpmock.NewStringResponder(204, ``))

	out, err := cmd.Execute("cloud", "ip", "floating", "delete", "1.2.3.4", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.String(out, `✅ Floating IP 1.2.3.4 deleted successfully`)
}

// ---------------------------------------------------------------------------
// Additional IP – list / get (API v2, read-only)
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudAdditionalIPListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/additional",
		httpmock.NewStringResponder(200, `[
			{"id": "1.2.3.4", "resourceStatus": "READY", "currentState": {"ip": "1.2.3.4"}}
		]`))

	out, err := cmd.Execute("cloud", "ip", "additional", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"id": "1.2.3.4", "resourceStatus": "READY", "currentState": {"ip": "1.2.3.4"}}
	]`))
}

func (ms *MockSuite) TestCloudAdditionalIPGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/additional/1.2.3.4",
		httpmock.NewStringResponder(200, `{"id": "1.2.3.4", "resourceStatus": "READY", "currentState": {"ip": "1.2.3.4"}}`))

	out, err := cmd.Execute("cloud", "ip", "additional", "get", "1.2.3.4", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id": "1.2.3.4", "resourceStatus": "READY", "currentState": {"ip": "1.2.3.4"}}`))
}

// ---------------------------------------------------------------------------
// Ext-Net IP – list / get / delete (API v2)
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudExtNetIPListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/extNet",
		httpmock.NewStringResponder(200, `[
			{"id": "5.6.7.8", "resourceStatus": "READY", "currentState": {"ip": "5.6.7.8", "location": {"region": "GRA11"}}}
		]`))

	out, err := cmd.Execute("cloud", "ip", "extnet", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"id": "5.6.7.8", "resourceStatus": "READY", "currentState": {"ip": "5.6.7.8", "location": {"region": "GRA11"}}}
	]`))
}

func (ms *MockSuite) TestCloudExtNetIPGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/extNet/5.6.7.8",
		httpmock.NewStringResponder(200, `{"id": "5.6.7.8", "resourceStatus": "READY", "currentState": {"ip": "5.6.7.8"}}`))

	out, err := cmd.Execute("cloud", "ip", "extnet", "get", "5.6.7.8", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id": "5.6.7.8", "resourceStatus": "READY", "currentState": {"ip": "5.6.7.8"}}`))
}

func (ms *MockSuite) TestCloudExtNetIPDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/publicIp/extNet/5.6.7.8",
		httpmock.NewStringResponder(204, ``))

	out, err := cmd.Execute("cloud", "ip", "extnet", "delete", "5.6.7.8", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.String(out, `✅ Ext-Net IP 5.6.7.8 deleted successfully`)
}

func (ms *MockSuite) TestCloudAdditionalIPAttachCmd(assert, require *td.T) {
	// Attach is not available on the v2 publicIp API yet, so it still uses v1.
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/ip/failover/additional-001/attach",
		httpmock.NewStringResponder(200, `{}`))

	out, err := cmd.Execute("cloud", "ip", "additional", "attach", "additional-001", "instance-123", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Additional IP additional-001 attached to instance instance-123 successfully"
	}`))
}
