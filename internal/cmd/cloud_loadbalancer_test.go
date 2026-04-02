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
// Helpers: region mocks shared across loadbalancing tests
// ---------------------------------------------------------------------------

// registerLoadbalancingRegionMocks sets up the standard region discovery responses
// used by locateLoadbalancer, locateLoadbalancingResource, and listLoadbalancingResources.
// GRA11 and SBG5 have the octavialoadbalancer feature; BHS5 does not.
func registerLoadbalancingRegionMocks() {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA11", "SBG5", "BHS5"]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11",
		httpmock.NewStringResponder(200, `{
			"name": "GRA11",
			"type": "region",
			"status": "UP",
			"services": [{"name": "octavialoadbalancer", "status": "UP"}]
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5",
		httpmock.NewStringResponder(200, `{
			"name": "SBG5",
			"type": "region",
			"status": "UP",
			"services": [{"name": "octavialoadbalancer", "status": "UP"}]
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5",
		httpmock.NewStringResponder(200, `{
			"name": "BHS5",
			"type": "region",
			"status": "UP",
			"services": []
		}`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region",
		httpmock.NewStringResponder(200, `["GRA11", "SBG5", "BHS5"]`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11",
		httpmock.NewStringResponder(200, `{
			"name": "GRA11",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "octavialoadbalancer",
					"status": "UP"
				}
			],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "GRA11"
		}`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5",
		httpmock.NewStringResponder(200, `{
			"name": "SBG5",
			"type": "region",
			"status": "UP",
			"services": [
				{
					"name": "octavialoadbalancer",
					"status": "UP"
				}
			],
			"countryCode": "fr",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "SBG5"
		}`))

	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/BHS5",
		httpmock.NewStringResponder(200, `{
			"name": "BHS5",
			"type": "region",
			"status": "UP",
			"services": [],
			"countryCode": "ca",
			"ipCountries": [],
			"continentCode": "NA",
			"availabilityZones": [],
			"datacenterLocation": "BHS5"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer/fakeLB",
		httpmock.NewStringResponder(200, `{
			"createdAt": "2024-07-30T08:26:51Z",
			"flavorId": "f862fa22-6275-4f8f-885e-66a8faf5e44e",
			"floatingIp": null,
			"id": "334fc97e-a8db-11f0-944d-0050568ce122",
			"name": "loadbalancer-sbg5-2024-07-30",
			"operatingStatus": "online",
			"provisioningStatus": "active",
			"region": "SBG5",
			"updatedAt": "2025-10-14T08:48:33Z",
			"vipAddress": "1.2.3.4",
			"vipNetworkId": "3f29f530-a8db-11f0-9ab2-0050568ce122",
			"vipSubnetId": "44a869c4-a8db-11f0-899f-0050568ce122"
		}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/flavor/f862fa22-6275-4f8f-885e-66a8faf5e44e",
		httpmock.NewStringResponder(200, `{
			"id": "f862fa22-6275-4f8f-885e-66a8faf5e44e",
			"name": "medium",
			"description": "Medium Load Balancer Flavor"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Load balancer fakeLB

  *loadbalancer-sbg5-2024-07-30*

  ## General information

  **Region**:              SBG5
  **Operating status**:    online
  **Provisioning status**: active
  **Flavor**:              medium (ID: f862fa22-6275-4f8f-885e-66a8faf5e44e)
  **Creation date**:       2024-07-30T08:26:51Z

  ## Technical information

  **VIP address**:        1.2.3.4
  **VIP network ID**:     3f29f530-a8db-11f0-9ab2-0050568ce122
  **VIP subnet ID**:      44a869c4-a8db-11f0-899f-0050568ce122

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Loadbalancer – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer",
		httpmock.NewStringResponder(200, `[
			{
				"id": "lb-gra-001",
				"name": "my-lb-gra",
				"region": "GRA11",
				"provisioningStatus": "active",
				"operatingStatus": "online"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer",
		httpmock.NewStringResponder(200, `[
			{
				"id": "lb-sbg-001",
				"name": "my-lb-sbg",
				"region": "SBG5",
				"provisioningStatus": "active",
				"operatingStatus": "online"
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "lb-gra-001",
			"name": "my-lb-gra",
			"region": "GRA11",
			"provisioningStatus": "active",
			"operatingStatus": "online"
		},
		{
			"id": "lb-sbg-001",
			"name": "my-lb-sbg",
			"region": "SBG5",
			"provisioningStatus": "active",
			"operatingStatus": "online"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	// locateLoadbalancer: 404 on GRA11, found on SBG5
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-sbg-001",
		httpmock.NewStringResponder(404, `{"message":"not found"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer/lb-sbg-001",
		httpmock.NewStringResponder(200, `{"id":"lb-sbg-001","name":"my-lb-sbg","region":"SBG5"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/loadbalancer/lb-sbg-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "delete", "lb-sbg-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Loadbalancer lb-sbg-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – stats
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerStatsCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001",
		httpmock.NewStringResponder(200, `{"id":"lb-gra-001","name":"my-lb-gra","region":"GRA11"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001/stats",
		httpmock.NewStringResponder(200, `{
			"activeConnections": 42,
			"bytesIn": 1024000,
			"bytesOut": 2048000,
			"requestErrors": 3,
			"totalConnections": 500
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "stats", "lb-gra-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"activeConnections": 42,
		"bytesIn": 1024000,
		"bytesOut": 2048000,
		"requestErrors": 3,
		"totalConnections": 500
	}`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – create with --size name resolution
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerCreateWithSizeCmd(assert, require *td.T) {
	// Mock flavor list for size resolution
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/flavor",
		httpmock.NewStringResponder(200, `[
			{
				"id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				"name": "small",
				"region": "GRA11"
			},
			{
				"id": "ffffffff-bbbb-cccc-dddd-eeeeeeeeeeee",
				"name": "medium",
				"region": "GRA11"
			}
		]`))

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer",
		tdhttpmock.JSONBody(td.JSON(`{
			"name": "my-lb",
			"flavorId": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			"network": {
				"private": {
					"network": {
						"id": "net-001",
						"subnetId": "sub-001"
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "lb-new-001",
			"name": "my-lb",
			"region": "GRA11"
		}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "create", "GRA11",
		"--cloud-project", "fakeProjectID",
		"--name", "my-lb",
		"--size", "small",
		"--network-id", "net-001",
		"--subnet-id", "sub-001",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Loadbalancer created successfully (ID: lb-new-001)",
		"details": {
			"id": "lb-new-001",
			"name": "my-lb",
			"region": "GRA11"
		}
	}`))
}

// ---------------------------------------------------------------------------
// Listener – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerListenerListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/listener",
		httpmock.NewStringResponder(200, `[
			{
				"id": "lis-001",
				"name": "my-listener",
				"protocol": "http",
				"port": 80,
				"operatingStatus": "online",
				"provisioningStatus": "active"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/listener",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "lis-001",
			"name": "my-listener",
			"protocol": "http",
			"port": 80,
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}
	]`))
}

func (ms *MockSuite) TestCloudLoadbalancerListenerListWithLoadbalancerIDCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/listener?loadbalancerId=lb-gra-001",
		httpmock.NewStringResponder(200, `[
			{
				"id": "lis-001",
				"name": "my-listener",
				"protocol": "http",
				"port": 80,
				"operatingStatus": "online",
				"provisioningStatus": "active"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/listener?loadbalancerId=lb-gra-001",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "ls", "--cloud-project", "fakeProjectID", "--loadbalancer-id", "lb-gra-001", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "lis-001",
			"name": "my-listener",
			"protocol": "http",
			"port": 80,
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Listener – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerListenerGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/listener/lis-001",
		httpmock.NewStringResponder(404, `{"message":"not found"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/listener/lis-001",
		httpmock.NewStringResponder(200, `{
			"id": "lis-001",
			"name": "my-listener",
			"protocol": "http",
			"port": 80,
			"operatingStatus": "online",
			"provisioningStatus": "active",
			"defaultPoolId": "pool-001",
			"description": "My HTTP listener"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "get", "lis-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🔊 Listener lis-001

  *my-listener*

  ## General information

  **Protocol**:            http
  **Port**:                80
  **Operating status**:    online
  **Provisioning status**: active
  **Default pool ID**:     pool-001
  **Description**:         My HTTP listener

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Listener – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerListenerDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/listener/lis-001",
		httpmock.NewStringResponder(200, `{"id":"lis-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/listener/lis-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "delete", "lis-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Listener lis-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Pool – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool",
		httpmock.NewStringResponder(200, `[
			{
				"id": "pool-001",
				"name": "my-pool",
				"algorithm": "roundRobin",
				"protocol": "http",
				"operatingStatus": "online",
				"provisioningStatus": "active"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/pool",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "pool-001",
			"name": "my-pool",
			"algorithm": "roundRobin",
			"protocol": "http",
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Pool – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{
			"id": "pool-001",
			"name": "my-pool",
			"algorithm": "roundRobin",
			"protocol": "http",
			"operatingStatus": "online",
			"provisioningStatus": "active",
			"loadbalancerId": "lb-gra-001",
			"listenerId": "lis-001"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "get", "pool-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🏊 Pool pool-001

  *my-pool*

  ## General information

  **Algorithm**:           roundRobin
  **Protocol**:            http
  **Operating status**:    online
  **Provisioning status**: active
  **Loadbalancer ID**:     lb-gra-001
  **Listener ID**:         lis-001

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Pool – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{"id":"pool-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "delete", "pool-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Pool pool-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Pool Member – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	// Locate the pool first
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{"id":"pool-001"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001/member",
		httpmock.NewStringResponder(200, `[
			{
				"id": "mem-001",
				"name": "backend-1",
				"address": "10.0.0.10",
				"protocolPort": 8080,
				"weight": 1,
				"operatingStatus": "online"
			},
			{
				"id": "mem-002",
				"name": "backend-2",
				"address": "10.0.0.11",
				"protocolPort": 8080,
				"weight": 2,
				"operatingStatus": "online"
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "ls", "pool-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "mem-001",
			"name": "backend-1",
			"address": "10.0.0.10",
			"protocolPort": 8080,
			"weight": 1,
			"operatingStatus": "online"
		},
		{
			"id": "mem-002",
			"name": "backend-2",
			"address": "10.0.0.11",
			"protocolPort": 8080,
			"weight": 2,
			"operatingStatus": "online"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Pool Member – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{"id":"pool-001"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001/member/mem-001",
		httpmock.NewStringResponder(200, `{
			"id": "mem-001",
			"name": "backend-1",
			"address": "10.0.0.10",
			"protocolPort": 8080,
			"weight": 1,
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "get", "pool-001", "mem-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 👤 Pool member mem-001

  *backend-1*

  ## General information

  **Address**:             10.0.0.10
  **Protocol port**:       8080
  **Weight**:              1
  **Operating status**:    online
  **Provisioning status**: active

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Pool Member – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{"id":"pool-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001/member/mem-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "delete", "pool-001", "mem-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Pool member mem-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Pool Member – create with flags
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberCreateWithFlagsCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	// Locate pool
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001",
		httpmock.NewStringResponder(200, `{"id":"pool-001"}`))

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/pool/pool-001/member",
		tdhttpmock.JSONBody(td.JSON(`{
			"members": [
				{
					"address": "10.0.0.42",
					"name": "my-member",
					"protocolPort": 8080,
					"weight": 5
				}
			]
		}`)),
		httpmock.NewStringResponder(200, `{
			"members": [
				{
					"id": "mem-new-001",
					"address": "10.0.0.42",
					"name": "my-member",
					"protocolPort": 8080,
					"weight": 5,
					"operatingStatus": "online",
					"provisioningStatus": "active"
				}
			]
		}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "create", "pool-001",
		"--cloud-project", "fakeProjectID",
		"--address", "10.0.0.42",
		"--name", "my-member",
		"--protocol-port", "8080",
		"--weight", "5",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Pool member(s) created successfully",
		"details": {
			"members": [
				{
					"id": "mem-new-001",
					"address": "10.0.0.42",
					"name": "my-member",
					"protocolPort": 8080,
					"weight": 5,
					"operatingStatus": "online",
					"provisioningStatus": "active"
				}
			]
		}
	}`))
}

// ---------------------------------------------------------------------------
// Health Monitor – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerHealthMonitorListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/healthMonitor",
		httpmock.NewStringResponder(200, `[
			{
				"id": "hm-001",
				"name": "my-health-monitor",
				"monitorType": "http",
				"poolId": "pool-001",
				"operatingStatus": "online",
				"provisioningStatus": "active"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/healthMonitor",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "health-monitor", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "hm-001",
			"name": "my-health-monitor",
			"monitorType": "http",
			"poolId": "pool-001",
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Health Monitor – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerHealthMonitorGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/healthMonitor/hm-001",
		httpmock.NewStringResponder(200, `{
			"id": "hm-001",
			"name": "my-health-monitor",
			"monitorType": "http",
			"poolId": "pool-001",
			"delay": 5,
			"timeout": 10,
			"maxRetries": 3,
			"maxRetriesDown": 3,
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "health-monitor", "get", "hm-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 💓 Health monitor hm-001

  *my-health-monitor*

  ## General information

  **Monitor type**:        http
  **Pool ID**:             pool-001
  **Delay**:               5s
  **Timeout**:             10s
  **Max retries**:         3
  **Max retries down**:    3
  **Operating status**:    online
  **Provisioning status**: active

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Health Monitor – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerHealthMonitorDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/healthMonitor/hm-001",
		httpmock.NewStringResponder(200, `{"id":"hm-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/healthMonitor/hm-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "health-monitor", "delete", "hm-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Health monitor hm-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// L7 Policy – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy",
		httpmock.NewStringResponder(200, `[
			{
				"id": "l7p-001",
				"name": "my-policy",
				"action": "redirectToPool",
				"listenerId": "lis-001",
				"position": 1,
				"operatingStatus": "online"
			}
		]`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/SBG5/loadbalancing/l7Policy",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "l7p-001",
			"name": "my-policy",
			"action": "redirectToPool",
			"listenerId": "lis-001",
			"position": 1,
			"operatingStatus": "online"
		}
	]`))
}

// ---------------------------------------------------------------------------
// L7 Policy – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `{
			"id": "l7p-001",
			"name": "my-policy",
			"action": "redirectToPool",
			"listenerId": "lis-001",
			"position": 1,
			"operatingStatus": "online",
			"provisioningStatus": "active",
			"description": "Redirect to backend pool",
			"redirectPoolId": "pool-001",
			"redirectUrl": "",
			"redirectPrefix": ""
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "get", "l7p-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 📋 L7 policy l7p-001

  *my-policy*

  ## General information

  **Action**:              redirectToPool
  **Position**:            1
  **Listener ID**:         lis-001
  **Operating status**:    online
  **Provisioning status**: active
  **Description**:         Redirect to backend pool
  **Redirect pool ID**:    pool-001

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// L7 Policy – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `{"id":"l7p-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "delete", "l7p-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ L7 policy l7p-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// L7 Rule – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7RuleListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	// Locate the L7 policy first
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `{"id":"l7p-001"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001/l7Rule",
		httpmock.NewStringResponder(200, `[
			{
				"id": "l7r-001",
				"ruleType": "header",
				"compareType": "equalTo",
				"value": "application/json",
				"key": "Content-Type",
				"invert": false
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "l7rule", "ls", "l7p-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "l7r-001",
			"ruleType": "header",
			"compareType": "equalTo",
			"value": "application/json",
			"key": "Content-Type",
			"invert": false
		}
	]`))
}

// ---------------------------------------------------------------------------
// L7 Rule – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7RuleGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `{"id":"l7p-001"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001/l7Rule/l7r-001",
		httpmock.NewStringResponder(200, `{
			"id": "l7r-001",
			"ruleType": "header",
			"compareType": "equalTo",
			"value": "application/json",
			"key": "Content-Type",
			"invert": false,
			"operatingStatus": "online",
			"provisioningStatus": "active"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "l7rule", "get", "l7p-001", "l7r-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 📏 L7 rule l7r-001

  ## General information

  **Rule type**:           header
  **Compare type**:        equalTo
  **Value**:               application/json
  **Key**:                 Content-Type
  **Invert**:              false
  **Operating status**:    online
  **Provisioning status**: active

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// L7 Rule – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerL7RuleDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001",
		httpmock.NewStringResponder(200, `{"id":"l7p-001"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/l7Policy/l7p-001/l7Rule/l7r-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "l7rule", "delete", "l7p-001", "l7r-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ L7 rule l7r-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Log – list kinds
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogKindListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/log/kind",
		httpmock.NewStringResponder(200, `["haproxy"]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "list-kinds", "GRA11", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"kinds": [
			"haproxy"
		]
	}`))
}

// ---------------------------------------------------------------------------
// Log – get kind
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogKindGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/log/kind/haproxy",
		httpmock.NewStringResponder(200, `{
			"name": "haproxy",
			"additionalReturnedFields": ["remote_addr", "request_path"],
			"displayName": "HAProxy logs"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "get-kind", "GRA11", "haproxy", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"name": "haproxy",
		"additionalReturnedFields": ["remote_addr", "request_path"],
		"displayName": "HAProxy logs"
	}`))
}

// ---------------------------------------------------------------------------
// Log – generate URL
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogGenerateURLCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001",
		httpmock.NewStringResponder(200, `{"id":"lb-gra-001","name":"my-lb","region":"GRA11"}`))

	httpmock.RegisterResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001/log/url",
		httpmock.NewStringResponder(200, `{
			"url": "https://logs.example.com/temp/abc123"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "generate-url", "lb-gra-001", "--kind", "haproxy", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Temporary log URL generated successfully: https://logs.example.com/temp/abc123",
		"details": {
			"url": "https://logs.example.com/temp/abc123"
		}
	}`))
}

// ---------------------------------------------------------------------------
// Log Subscription – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogSubscriptionListCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001",
		httpmock.NewStringResponder(200, `{"id":"lb-gra-001","name":"my-lb","region":"GRA11"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001/log/subscription",
		httpmock.NewStringResponder(200, `[
			{
				"subscriptionId": "sub-001",
				"kind": "haproxy",
				"streamId": "stream-abc",
				"createdAt": "2024-01-15T10:30:00Z"
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "subscription", "ls", "lb-gra-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"subscriptionId": "sub-001",
			"kind": "haproxy",
			"streamId": "stream-abc",
			"createdAt": "2024-01-15T10:30:00Z"
		}
	]`))
}

// ---------------------------------------------------------------------------
// Log Subscription – get
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogSubscriptionGetCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001",
		httpmock.NewStringResponder(200, `{"id":"lb-gra-001","name":"my-lb","region":"GRA11"}`))

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001/log/subscription/sub-001",
		httpmock.NewStringResponder(200, `{
			"subscriptionId": "sub-001",
			"kind": "haproxy",
			"streamId": "stream-abc",
			"serviceName": "ldp-service",
			"createdAt": "2024-01-15T10:30:00Z",
			"updatedAt": "2024-01-15T10:30:00Z"
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "subscription", "get", "lb-gra-001", "sub-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 📝 Log subscription sub-001

  ## General information

  **Subscription ID**: sub-001
  **Kind**:            haproxy
  **Stream ID**:       stream-abc
  **Service name**:    ldp-service
  **Created at**:      2024-01-15T10:30:00Z
  **Updated at**:      2024-01-15T10:30:00Z

  💡 Use option --json or --yaml to get the raw output with all information

`)
}

// ---------------------------------------------------------------------------
// Log Subscription – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerLogSubscriptionDeleteCmd(assert, require *td.T) {
	registerLoadbalancingRegionMocks()

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001",
		httpmock.NewStringResponder(200, `{"id":"lb-gra-001","name":"my-lb","region":"GRA11"}`))

	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/region/GRA11/loadbalancing/loadbalancer/lb-gra-001/log/subscription/sub-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "log", "subscription", "delete", "lb-gra-001", "sub-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Log subscription sub-001 deleted successfully"
	}`))
}
