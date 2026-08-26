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

func (ms *MockSuite) TestCloudPrivateNetworkCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "TestFromTheCLI",
					"location": {
						"region": "BHS5"
					}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "80c1de3e-9b09-11f0-993b-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "create", "BHS5", "--cloud-project", "fakeProjectID",
		"--name", "TestFromTheCLI", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Network creation started successfully (id: 80c1de3e-9b09-11f0-993b-0050568ce122)",
		"details": {"id": "80c1de3e-9b09-11f0-993b-0050568ce122"}
	}`))
}

func (ms *MockSuite) TestCloudPrivateNetworkCreateCmdWithWait(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "TestFromTheCLI",
					"vlanId": 1234,
					"location": {
						"region": "BHS5"
					}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "80c1de3e-9b09-11f0-993b-0050568ce122"}`),
	)

	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/80c1de3e-9b09-11f0-993b-0050568ce122",
		httpmock.NewStringResponder(200, `{
			"id": "80c1de3e-9b09-11f0-993b-0050568ce122",
			"resourceStatus": "READY"
		}`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "create", "BHS5", "--cloud-project", "fakeProjectID",
		"--name", "TestFromTheCLI", "--vlan-id", "1234", "--wait", "-o", "yaml")
	require.CmpNoError(err)
	assert.String(out, `details:
  id: 80c1de3e-9b09-11f0-993b-0050568ce122
message: ✅ Network 80c1de3e-9b09-11f0-993b-0050568ce122 created successfully in region
  BHS5
`)
}

func (ms *MockSuite) TestCloudPrivateNetworkListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network",
		httpmock.NewStringResponder(200, `[
			{
				"id": "net-1",
				"resourceStatus": "READY",
				"currentState": {
					"name": "network-one",
					"vlanId": 42,
					"location": {"region": "BHS5"}
				}
			}
		]`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "net-1",
			"resourceStatus": "READY",
			"currentState": {
				"name": "network-one",
				"vlanId": 42,
				"location": {"region": "BHS5"}
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudPrivateNetworkDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/net-123",
		httpmock.NewStringResponder(204, ``),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "delete", "net-123", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.String(out, `✅ Private network net-123 is being deleted…`)
}

func (ms *MockSuite) TestCloudPrivateNetworkEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/net-123",
		httpmock.NewStringResponder(200, `{
			"id": "net-123",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"targetSpec": {
				"name": "old-name",
				"vlanId": 42,
				"location": {"region": "BHS5"}
			}
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/net-123",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "abc123",
				"targetSpec": {
					"name": "new-name"
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "edit", "net-123", "--cloud-project", "fakeProjectID", "--name", "new-name")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetCreateCmd(assert, require *td.T) {
	// The subnet inherits the region of its parent network.
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456",
		httpmock.NewStringResponder(200, `{
			"id": "pn-123456",
			"targetSpec": {
				"name": "test-network",
				"location": {"region": "BHS5"}
			}
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456/subnet",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-subnet",
					"cidr": "192.168.1.0/24",
					"dhcpEnabled": true,
					"location": {"region": "BHS5"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "5e625f90-9ec3-11f0-9f75-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "subnet", "create", "pn-123456", "--cloud-project", "fakeProjectID",
		"--name", "my-subnet", "--cidr", "192.168.1.0/24", "--dhcp-enabled", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Subnet creation started successfully (id: 5e625f90-9ec3-11f0-9f75-0050568ce122)",
		"details": {"id": "5e625f90-9ec3-11f0-9f75-0050568ce122"}
	}`))
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetCreateCmdWithRegionFlag(assert, require *td.T) {
	// When --region is provided, no parent network lookup should be performed.
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456/subnet",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "my-subnet",
					"cidr": "192.168.1.0/24",
					"dhcpEnabled": false,
					"allocationPools": [
						{"start": "192.168.1.2", "end": "192.168.1.254"}
					],
					"location": {"region": "GRA11"}
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "subnet-with-region"}`),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "subnet", "create", "pn-123456", "--cloud-project", "fakeProjectID",
		"--name", "my-subnet", "--cidr", "192.168.1.0/24", "--region", "GRA11",
		"--allocation-pools", "192.168.1.2:192.168.1.254", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Subnet creation started successfully (id: subnet-with-region)",
		"details": {"id": "subnet-with-region"}
	}`))
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456/subnet/subnet-1",
		httpmock.NewStringResponder(200, `{
			"id": "subnet-1",
			"checksum": "chk-1",
			"targetSpec": {
				"name": "old-subnet",
				"cidr": "192.168.1.0/24",
				"location": {"region": "BHS5"}
			}
		}`),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456/subnet/subnet-1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"checksum": "chk-1",
				"targetSpec": {
					"name": "new-subnet"
				}
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "subnet", "edit", "pn-123456", "subnet-1",
		"--cloud-project", "fakeProjectID", "--name", "new-subnet")
	require.CmpNoError(err)
	assert.String(out, `✅ Resource updated successfully`)
}

func (ms *MockSuite) TestCloudPrivateNetworkSubnetDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE",
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/network/pn-123456/subnet/subnet-1",
		httpmock.NewStringResponder(204, ``),
	)

	out, err := cmd.Execute("cloud", "network", "private", "vrack", "subnet", "delete", "pn-123456", "subnet-1", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.String(out, `✅ Subnet subnet-1 is being deleted from network pn-123456…`)
}

// TestCloudPrivateNetworkCreateMissingNameCmd checks that a create without the
// mandatory --name fails client-side and does NOT hit the API (no POST
// responder is registered, so any call would error differently).
func (ms *MockSuite) TestCloudPrivateNetworkCreateMissingNameCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "network", "private", "vrack", "create", "GRA11",
		"--cloud-project", "fakeProjectID")

	require.CmpError(err)
	// The error keeps the full path to the missing value (review feedback).
	assert.Cmp(err.Error(), td.Contains(`mandatory field "targetSpec.name"`))
}
