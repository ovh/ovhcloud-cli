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

func (ms *MockSuite) TestCloudPrivateEndpointListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint",
		httpmock.NewStringResponder(200, `[
			{
				"id": "pe-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-endpoint",
					"location": {"region": "GRA11"},
					"service": {"name": "my-service"},
					"vipAddress": "10.0.0.5"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "private-endpoint", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "pe-1",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-endpoint",
				"location": {"region": "GRA11"},
				"service": {"name": "my-service"},
				"vipAddress": "10.0.0.5"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudPrivateEndpointGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		httpmock.NewStringResponder(200, `{
			"id": "pe-1",
			"resourceStatus": "READY",
			"checksum": "abc123",
			"currentState": {
				"name": "my-endpoint",
				"location": {"region": "GRA11"},
				"service": {"name": "my-service", "tier": "S"},
				"subnet": {"id": "subnet-1"},
				"vipAddress": "10.0.0.5",
				"hostname": "my-endpoint.example.net"
			}
		}`))

	out, err := cmd.Execute("cloud", "private-endpoint", "get", "pe-1", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("pe-1"))
	assert.Cmp(out, td.Contains("my-endpoint"))
	assert.Cmp(out, td.Contains("my-service"))
}

func (ms *MockSuite) TestCloudPrivateEndpointGetPartialStateCmd(assert, require *td.T) {
	// A freshly created endpoint has no location yet: the template must not
	// dereference it.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-2",
		httpmock.NewStringResponder(200, `{
			"id": "pe-2",
			"resourceStatus": "CREATING",
			"currentState": {"name": "my-endpoint"},
			"targetSpec": {"name": "my-endpoint"}
		}`))

	out, err := cmd.Execute("cloud", "private-endpoint", "get", "pe-2", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-endpoint"))
	assert.Cmp(out, td.Contains("CREATING"))
}

func (ms *MockSuite) TestCloudPrivateEndpointCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-endpoint",
					"location": {"region": "GRA11"},
					"service": {"name": "my-service", "tier": "S"},
					"subnet": {"id": "subnet-1"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "pe-123"}`),
	)

	out, err := cmd.Execute("cloud", "private-endpoint", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-endpoint", "--region", "GRA11", "--service-name", "my-service",
		"--tier", "S", "--subnet-id", "subnet-1")

	require.CmpNoError(err)
	assert.String(out, `✅ Private endpoint creation started successfully (id: pe-123)`)
}

func (ms *MockSuite) TestCloudPrivateEndpointCreateFromFileCmd(assert, require *td.T) {
	// publicDnsEnabled is immutable after creation: a value set in the
	// parameters file must not be overridden by the flag default.
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-endpoint",
					"location": {"region": "GRA11"},
					"service": {"name": "my-service", "tier": "S"},
					"subnet": {"id": "subnet-1"},
					"publicDnsEnabled": true
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "pe-123"}`),
	)

	// … but an explicit flag still wins over the file.
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-endpoint",
					"location": {"region": "GRA11"},
					"service": {"name": "my-service", "tier": "S"},
					"subnet": {"id": "subnet-1"},
					"publicDnsEnabled": false
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "pe-456"}`),
	)

	tmpFile, err := os.CreateTemp("", "private-endpoint-create-*.json")
	require.CmpNoError(err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(`{
		"targetSpec": {
			"name": "my-endpoint",
			"location": {"region": "GRA11"},
			"service": {"name": "my-service", "tier": "S"},
			"subnet": {"id": "subnet-1"},
			"publicDnsEnabled": true
		}
	}`)
	require.CmpNoError(err)
	tmpFile.Close()

	out, err := cmd.Execute("cloud", "private-endpoint", "create", "--cloud-project", "fakeProjectID",
		"--from-file", tmpFile.Name())
	require.CmpNoError(err)
	assert.String(out, `✅ Private endpoint creation started successfully (id: pe-123)`)

	out, err = cmd.Execute("cloud", "private-endpoint", "create", "--cloud-project", "fakeProjectID",
		"--from-file", tmpFile.Name(), "--public-dns-enabled=false")
	require.CmpNoError(err)
	assert.String(out, `✅ Private endpoint creation started successfully (id: pe-456)`)
}

func (ms *MockSuite) TestCloudPrivateEndpointCreateErrorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint",
		httpmock.NewStringResponder(400, `{"message": "invalid service"}`),
	)

	out, err := cmd.Execute("cloud", "private-endpoint", "create", "--cloud-project", "fakeProjectID",
		"--name", "my-endpoint", "--region", "GRA11", "--service-name", "my-service",
		"--tier", "S", "--subnet-id", "subnet-1")

	assert.CmpError(err)
	assert.Cmp(out, td.Contains("failed to create private endpoint"))
}

func (ms *MockSuite) TestCloudPrivateEndpointDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "private-endpoint", "delete", "pe-1", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.String(out, `✅ Private endpoint pe-1 is being deleted…`)
}

func (ms *MockSuite) TestCloudPrivateEndpointEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		httpmock.NewStringResponder(200, `{
			"id": "pe-1",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "old-name",
				"location": {"region": "GRA11"},
				"service": {"name": "my-service", "tier": "S"}
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"name": "new-name"
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "private-endpoint", "edit", "pe-1", "--cloud-project", "fakeProjectID", "--name", "new-name")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudPrivateEndpointServiceAvailabilityListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/privateEndpoint/serviceAvailability",
		httpmock.NewStringResponder(200, `[
			{
				"name": "my-service",
				"description": "My exposable service",
				"tiers": [
					{"name": "S", "bandwidth": 1000, "port": 443, "transport": "TCP", "location": {"region": "GRA1.PREPROD"}}
				]
			}
		]`))

	out, err := cmd.Execute("cloud", "private-endpoint", "service-availability", "list", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	// Table view is flattened to one row per (service × tier), not a
	// raw tiers map.
	assert.Cmp(out, td.Not(td.Contains("map[")))
	assert.Cmp(out, td.Contains("my-service"))
	assert.Cmp(out, td.Contains("1000"))
	assert.Cmp(out, td.Contains("GRA1.PREPROD"))
	assert.Cmp(out, td.Contains("TCP"))
	assert.Cmp(out, td.Contains("443"))
}

func (ms *MockSuite) TestCloudPrivateEndpointServiceAvailabilityListJSONCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/privateEndpoint/serviceAvailability",
		httpmock.NewStringResponder(200, `[
			{
				"name": "my-service",
				"description": "My exposable service",
				"tiers": [
					{"name": "S", "bandwidth": 1000, "port": 443, "transport": "TCP", "location": {"region": "GRA1.PREPROD"}},
					{"name": "M", "bandwidth": 5000, "port": 53, "transport": "UDP", "location": {"region": "SBG5"}}
				]
			}
		]`))

	// JSON output must be the raw API response, not the flattened table rows.
	out, err := cmd.Execute("cloud", "private-endpoint", "service-availability", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"name": "my-service",
			"description": "My exposable service",
			"tiers": [
				{"name": "S", "bandwidth": 1000, "port": 443, "transport": "TCP", "location": {"region": "GRA1.PREPROD"}},
				{"name": "M", "bandwidth": 5000, "port": 53, "transport": "UDP", "location": {"region": "SBG5"}}
			]
		}
	]`))
}

func (ms *MockSuite) TestCloudPrivateEndpointServiceAvailabilityGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/privateEndpoint/serviceAvailability/my-service",
		httpmock.NewStringResponder(200, `{
			"name": "my-service",
			"description": "My exposable service",
			"tiers": [
				{"name": "S", "bandwidth": 1000, "port": 443, "transport": "TCP", "location": {"region": "GRA1.PREPROD"}}
			]
		}`))

	out, err := cmd.Execute("cloud", "private-endpoint", "service-availability", "get", "my-service", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("my-service"))
	assert.Cmp(out, td.Contains("My exposable service"))
	// Tier line rendered from the template, not raw JSON.
	assert.Cmp(out, td.Contains("1000 Mbit/s"))
	assert.Cmp(out, td.Contains("TCP/443"))
	assert.Cmp(out, td.Contains("GRA1.PREPROD"))
}

func (ms *MockSuite) TestCloudPrivateEndpointGetNullCurrentStateCmd(assert, require *td.T) {
	// currentState is nullable: the template must only render the target state.
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-3",
		httpmock.NewStringResponder(200, `{
			"id": "pe-3",
			"resourceStatus": "CREATING",
			"currentState": null,
			"targetSpec": {"name": "my-endpoint", "location": {"region": "GRA11"}}
		}`))

	out, err := cmd.Execute("cloud", "private-endpoint", "get", "pe-3", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("pe-3"))
	assert.Cmp(out, td.Contains("my-endpoint"))
	assert.Cmp(out, td.Contains("CREATING"))
}

func (ms *MockSuite) TestCloudPrivateEndpointEditClearDescriptionCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		httpmock.NewStringResponder(200, `{
			"id": "pe-1",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "my-endpoint",
				"description": "old description",
				"location": {"region": "GRA11"},
				"service": {"name": "my-service", "tier": "S"}
			}
		}`))

	// An explicit empty --description must be sent, not dropped.
	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/privateEndpoint/pe-1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"name": "my-endpoint",
					"description": ""
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "private-endpoint", "edit", "pe-1", "--cloud-project", "fakeProjectID", "--description", "")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudPrivateEndpointServiceAvailabilityListFilterCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/privateEndpoint/serviceAvailability",
		httpmock.NewStringResponder(200, `[
			{
				"name": "svc-a",
				"tiers": [{"name": "S", "port": 443, "transport": "TCP", "location": {"region": "GRA11"}}]
			},
			{
				"name": "svc-b",
				"tiers": [{"name": "M", "port": 53, "transport": "UDP", "location": {"region": "SBG5"}}]
			}
		]`))

	// --filter applies to the displayed (flattened) columns in table mode.
	out, err := cmd.Execute("cloud", "private-endpoint", "service-availability", "list", "--cloud-project", "fakeProjectID", "--filter", `region=="GRA11"`)
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("svc-a"))
	assert.Cmp(out, td.Not(td.Contains("svc-b")))
}
