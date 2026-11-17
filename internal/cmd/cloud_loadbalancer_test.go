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
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/fakeLB",
		httpmock.NewStringResponder(200, `{
			"id": "334fc97e-a8db-11f0-944d-0050568ce122",
			"createdAt": "2024-07-30T08:26:51Z",
			"updatedAt": "2025-10-14T08:48:33Z",
			"resourceStatus": "READY",
			"currentState": {
				"name": "loadbalancer-sbg5-2024-07-30",
				"description": "my loadbalancer",
				"location": {"region": "SBG5"},
				"flavor": {"name": "MEDIUM"},
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE",
				"network": {
					"id": "3f29f530-a8db-11f0-9ab2-0050568ce122",
					"subnetId": "44a869c4-a8db-11f0-899f-0050568ce122",
					"addresses": [
						{"ip": "1.2.3.4", "type": "FIXED"},
						{"ip": "5.6.7.8", "type": "FLOATING"}
					]
				}
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Load balancer fakeLB

  *loadbalancer-sbg5-2024-07-30*

  ## General information

  **ID**:                  334fc97e-a8db-11f0-944d-0050568ce122
  **Status**:              READY
  **Description**:         my loadbalancer
  **Region**:              SBG5
  **Operating status**:    ONLINE
  **Provisioning status**: ACTIVE
  **Flavor**:              MEDIUM
  **Creation date**:       2024-07-30T08:26:51Z
  **Update date**:         2025-10-14T08:48:33Z

  ## Network

  **VIP network ID**:     3f29f530-a8db-11f0-9ab2-0050568ce122
  **VIP subnet ID**:      44a869c4-a8db-11f0-899f-0050568ce122

   Address                             | Type
  -------------------------------------|------------------------------------
   1.2.3.4                             | FIXED
   5.6.7.8                             | FLOATING

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

// currentState is null while the load balancer is being created: the command
// must not fail, and falls back on targetSpec.
func (ms *MockSuite) TestCloudLoadbalancerGetCreatingCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/fakeLB",
		httpmock.NewStringResponder(200, `{
			"id": "334fc97e-a8db-11f0-944d-0050568ce122",
			"createdAt": "2024-07-30T08:26:51Z",
			"updatedAt": "2024-07-30T08:26:51Z",
			"resourceStatus": "CREATING",
			"currentState": null,
			"targetSpec": {
				"name": "creating-lb",
				"description": "being created",
				"location": {"region": "SBG5"},
				"flavor": {"name": "SMALL"},
				"network": {
					"id": "3f29f530-a8db-11f0-9ab2-0050568ce122",
					"subnetId": "44a869c4-a8db-11f0-899f-0050568ce122"
				}
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.All(
		td.Contains("*creating-lb*"),
		td.Contains("**Status**:              CREATING"),
		td.Contains("**Description**:         being created"),
		td.Contains("**Region**:              SBG5"),
		td.Contains("**Operating status**:    *not yet available*"),
		td.Contains("**Provisioning status**: *not yet available*"),
		td.Contains("**Flavor**:              SMALL"),
		td.Contains("**VIP network ID**:     3f29f530-a8db-11f0-9ab2-0050568ce122"),
		td.Contains("**VIP subnet ID**:      44a869c4-a8db-11f0-899f-0050568ce122"),
		td.Contains("*Addresses not yet available*"),
	))
}

// Neither currentState nor its nested objects are guaranteed to be present.
func (ms *MockSuite) TestCloudLoadbalancerGetPartialStateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/fakeLB",
		httpmock.NewStringResponder(200, `{
			"id": "334fc97e-a8db-11f0-944d-0050568ce122",
			"resourceStatus": "CREATING",
			"currentState": {"name": "partial-lb"}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "get", "fakeLB", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.All(
		td.Contains("*partial-lb*"),
		td.Contains("*Addresses not yet available*"),
	))
}

// ---------------------------------------------------------------------------
// Loadbalancer – list
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer",
		httpmock.NewStringResponder(200, `[
			{
				"id": "lb-gra-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-lb-gra",
					"location": {"region": "GRA11"},
					"provisioningStatus": "ACTIVE",
					"operatingStatus": "ONLINE"
				}
			},
			{
				"id": "lb-sbg-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-lb-sbg",
					"location": {"region": "SBG5"},
					"provisioningStatus": "ACTIVE",
					"operatingStatus": "ONLINE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "ls", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "lb-gra-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-lb-gra",
				"location": {"region": "GRA11"},
				"provisioningStatus": "ACTIVE",
				"operatingStatus": "ONLINE"
			}
		},
		{
			"id": "lb-sbg-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-lb-sbg",
				"location": {"region": "SBG5"},
				"provisioningStatus": "ACTIVE",
				"operatingStatus": "ONLINE"
			}
		}
	]`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – delete
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/lb-sbg-001",
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
// Loadbalancer – create
// ---------------------------------------------------------------------------

func (ms *MockSuite) TestCloudLoadbalancerCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer",
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-lb",
				"flavor": {"name": "SMALL"},
				"location": {"region": "GRA11"},
				"network": {
					"id": "net-001",
					"subnetId": "sub-001"
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "lb-new-001"}`),
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
		"details": {"id": "lb-new-001"}
	}`))
}

func (ms *MockSuite) TestCloudLoadbalancerCreateWithNetworkIPCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer",
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-lb",
				"description": "with a floating ip",
				"flavor": {"name": "MEDIUM"},
				"location": {"region": "GRA11"},
				"network": {
					"subnetId": "sub-001",
					"ip": "203.0.113.10"
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "lb-new-002"}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "create", "GRA11",
		"--cloud-project", "fakeProjectID",
		"--name", "my-lb",
		"--description", "with a floating ip",
		"--size", "MEDIUM",
		"--subnet-id", "sub-001",
		"--network-ip", "203.0.113.10",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Loadbalancer created successfully (ID: lb-new-002)",
		"details": {"id": "lb-new-002"}
	}`))
}

func (ms *MockSuite) TestCloudLoadbalancerCreateMissingSubnetCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "loadbalancer", "create", "GRA11",
		"--cloud-project", "fakeProjectID",
		"--name", "my-lb",
		"--size", "SMALL")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`targetSpec.network.subnetId`))
}

// ---------------------------------------------------------------------------
// Loadbalancer – edit
// ---------------------------------------------------------------------------
// Listener
// ---------------------------------------------------------------------------

const listenersURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/lb-001/listener"

func (ms *MockSuite) TestCloudLoadbalancerListenerListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, listenersURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "lis-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-listener",
					"protocol": "HTTP",
					"protocolPort": 80,
					"operatingStatus": "ONLINE",
					"provisioningStatus": "ACTIVE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "ls", "lb-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "lis-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-listener",
				"protocol": "HTTP",
				"protocolPort": 80,
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudLoadbalancerListenerGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, listenersURL+"/lis-001",
		httpmock.NewStringResponder(200, `{
			"id": "lis-001",
			"resourceStatus": "READY",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:05:00Z",
			"currentState": {
				"name": "my-listener",
				"description": "My HTTP listener",
				"protocol": "HTTP",
				"protocolPort": 80,
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE",
				"defaultPool": {"id": "pool-001"}
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "get", "lb-001", "lis-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🔊 Listener lis-001

  *my-listener*

  ## General information

  **ID**:                  lis-001
  **Status**:              READY
  **Protocol**:            HTTP
  **Port**:                80
  **Operating status**:    ONLINE
  **Provisioning status**: ACTIVE
  **Default pool ID**:     pool-001
  **Description**:         My HTTP listener
  **Creation date**:       2026-11-10T10:00:00Z
  **Update date**:         2026-11-10T10:05:00Z

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudLoadbalancerListenerGetCreatingCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, listenersURL+"/lis-001",
		httpmock.NewStringResponder(200, `{
			"id": "lis-001",
			"resourceStatus": "CREATING",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:00:00Z",
			"currentState": null,
			"targetSpec": {"name": "my-listener", "protocol": "TCP", "protocolPort": 443}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "get", "lb-001", "lis-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("*my-listener*"))
	assert.Cmp(out, td.Re(`\*\*Protocol\*\*:\s+TCP`))
	assert.Cmp(out, td.Re(`\*\*Operating status\*\*:\s+\*not yet available\*`))
}

func (ms *MockSuite) TestCloudLoadbalancerListenerCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, listenersURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-listener",
				"protocol": "HTTP",
				"protocolPort": 80,
				"defaultPool": {"id": "pool-001"}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "lis-001"}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--name", "my-listener",
		"--protocol", "http",
		"--port", "80",
		"--default-pool-id", "pool-001",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Listener created successfully (ID: lis-001)", "details": {"id": "lis-001"}}`))
}

func (ms *MockSuite) TestCloudLoadbalancerListenerCreateMissingPortCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "loadbalancer", "listener", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--name", "my-listener",
		"--protocol", "HTTP")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`targetSpec.protocolPort`))
}

func (ms *MockSuite) TestCloudLoadbalancerListenerEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, listenersURL+"/lis-001",
		httpmock.NewStringResponder(200, `{
			"id": "lis-001",
			"checksum": "abc",
			"targetSpec": {"name": "my-listener", "description": "old", "protocol": "HTTP", "protocolPort": 80}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, listenersURL+"/lis-001",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc",
			"targetSpec": {
				"name": "my-listener",
				"description": "new",
				"defaultPool": {"id": "pool-002"}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "lis-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "listener", "edit", "lb-001", "lis-001",
		"--cloud-project", "fakeProjectID",
		"--description", "new",
		"--default-pool-id", "pool-002")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerListenerDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, listenersURL+"/lis-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "listener", "delete", "lb-001", "lis-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Listener lis-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Pool
// ---------------------------------------------------------------------------

const poolsURL = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/loadbalancer/lb-001/pool"

func (ms *MockSuite) TestCloudLoadbalancerPoolListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, poolsURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "pool-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "my-pool",
					"algorithm": "ROUND_ROBIN",
					"protocol": "HTTP",
					"operatingStatus": "ONLINE",
					"provisioningStatus": "ACTIVE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "ls", "lb-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "pool-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "my-pool",
				"algorithm": "ROUND_ROBIN",
				"protocol": "HTTP",
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, poolsURL+"/pool-001",
		httpmock.NewStringResponder(200, `{
			"id": "pool-001",
			"resourceStatus": "READY",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:05:00Z",
			"currentState": {
				"name": "my-pool",
				"algorithm": "ROUND_ROBIN",
				"protocol": "HTTP",
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE",
				"healthMonitor": {
					"type": "HTTP",
					"delay": 5,
					"timeout": 3,
					"maxRetries": 2,
					"maxRetriesDown": 4,
					"httpMethod": "GET",
					"urlPath": "/health",
					"expectedCodes": "200"
				}
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "get", "lb-001", "pool-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🏊 Pool pool-001

  *my-pool*

  ## General information

  **ID**:                  pool-001
  **Status**:              READY
  **Algorithm**:           ROUND_ROBIN
  **Protocol**:            HTTP
  **Operating status**:    ONLINE
  **Provisioning status**: ACTIVE
  **Creation date**:       2026-11-10T10:00:00Z
  **Update date**:         2026-11-10T10:05:00Z

  ## Health monitor

  **Type**:             HTTP
  **Delay**:            5s
  **Timeout**:          3s
  **Max retries**:      2
  **Max retries down**: 4
  **HTTP method**:      GET
  **URL path**:         /health
  **Expected codes**:   200

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolGetWithoutHealthMonitorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, poolsURL+"/pool-001",
		httpmock.NewStringResponder(200, `{
			"id": "pool-001",
			"resourceStatus": "CREATING",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:00:00Z",
			"currentState": null,
			"targetSpec": {"name": "my-pool", "algorithm": "SOURCE_IP", "protocol": "TCP"}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "get", "lb-001", "pool-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Re(`\*\*Algorithm\*\*:\s+SOURCE_IP`))
	assert.Cmp(out, td.Contains("No health monitor"))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, poolsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-pool",
				"algorithm": "LEAST_CONNECTIONS",
				"protocol": "HTTP",
				"healthMonitor": {
					"type": "HTTP",
					"delay": 5,
					"timeout": 3,
					"maxRetries": 2,
					"httpMethod": "GET",
					"urlPath": "/health"
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pool-001"}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--name", "my-pool",
		"--algorithm", "least_connections",
		"--protocol", "http",
		"--health-monitor-type", "http",
		"--health-monitor-delay", "5",
		"--health-monitor-timeout", "3",
		"--health-monitor-max-retries", "2",
		"--health-monitor-http-method", "get",
		"--health-monitor-url-path", "/health",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Pool created successfully (ID: pool-001)", "details": {"id": "pool-001"}}`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolCreateWithoutHealthMonitorCmd(assert, require *td.T) {
	// The healthMonitor key must be absent from the body when no health monitor flag is given
	httpmock.RegisterMatcherResponder(http.MethodPost, poolsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"algorithm": "ROUND_ROBIN", "protocol": "TCP"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pool-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--algorithm", "ROUND_ROBIN",
		"--protocol", "TCP")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, poolsURL+"/pool-001",
		httpmock.NewStringResponder(200, `{
			"id": "pool-001",
			"checksum": "abc",
			"targetSpec": {
				"name": "my-pool",
				"algorithm": "ROUND_ROBIN",
				"protocol": "HTTP",
				"healthMonitor": {"type": "HTTP", "delay": 5, "timeout": 3, "maxRetries": 2}
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, poolsURL+"/pool-001",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc",
			"targetSpec": {
				"name": "my-pool",
				"algorithm": "SOURCE_IP",
				"healthMonitor": {"type": "HTTP", "delay": 10, "timeout": 3, "maxRetries": 2}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pool-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "edit", "lb-001", "pool-001",
		"--cloud-project", "fakeProjectID",
		"--algorithm", "source_ip",
		"--health-monitor-delay", "10")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolCreateIncompleteHealthMonitorCmd(assert, require *td.T) {
	// No POST responder registered: the command must fail before calling the API
	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--algorithm", "ROUND_ROBIN",
		"--protocol", "HTTP",
		"--health-monitor-type", "HTTP",
		"--health-monitor-delay", "5")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("the following parameters are required: --health-monitor-timeout, --health-monitor-max-retries"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolEditIncompleteHealthMonitorCmd(assert, require *td.T) {
	// The pool has no health monitor: a single --health-monitor-* flag cannot create one
	httpmock.RegisterResponder(http.MethodGet, poolsURL+"/pool-001",
		httpmock.NewStringResponder(200, `{
			"id": "pool-001",
			"checksum": "abc",
			"targetSpec": {"name": "my-pool", "algorithm": "ROUND_ROBIN", "protocol": "HTTP"}
		}`))

	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "edit", "lb-001", "pool-001",
		"--cloud-project", "fakeProjectID",
		"--health-monitor-delay", "10")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("this pool has no health monitor yet"))
	assert.Cmp(err.Error(), td.Contains("--health-monitor-type, --health-monitor-timeout, --health-monitor-max-retries"))
	// Only the GET happened, no PUT
	assert.Cmp(httpmock.GetCallCountInfo()["PUT "+poolsURL+"/pool-001"], 0)
}

// writeTempJSON writes content to a temporary JSON file and returns its path.
func writeTempJSON(require *td.T, content string) string {
	f, err := os.CreateTemp("", "lb-params-*.json")
	require.CmpNoError(err)
	_, err = f.WriteString(content)
	require.CmpNoError(err)
	require.CmpNoError(f.Close())
	return f.Name()
}

func (ms *MockSuite) TestCloudLoadbalancerPoolCreateFileWithHealthMonitorAndFlagCmd(assert, require *td.T) {
	// The file defines a complete health monitor, a single flag overrides one field:
	// the flags alone are incomplete but the merged body is valid.
	file := writeTempJSON(require, `{"targetSpec": {"name": "my-pool", "algorithm": "ROUND_ROBIN", "protocol": "HTTP",
		"healthMonitor": {"type": "HTTP", "delay": 5, "timeout": 3, "maxRetries": 2}}}`)
	defer os.Remove(file)

	httpmock.RegisterMatcherResponder(http.MethodPost, poolsURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "my-pool",
				"algorithm": "ROUND_ROBIN",
				"protocol": "HTTP",
				"healthMonitor": {"type": "HTTP", "delay": 10, "timeout": 3, "maxRetries": 2}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pool-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "create", "lb-001",
		"--cloud-project", "fakeProjectID",
		"--from-file", file,
		"--health-monitor-delay", "10")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, poolsURL+"/pool-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "delete", "lb-001", "pool-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Pool pool-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// Pool Member
// ---------------------------------------------------------------------------

const membersURL = poolsURL + "/pool-001/member"

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, membersURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "mem-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "web-1",
					"address": "10.0.0.10",
					"protocolPort": 80,
					"weight": 1,
					"operatingStatus": "ONLINE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "ls", "lb-001", "pool-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "mem-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "web-1",
				"address": "10.0.0.10",
				"protocolPort": 80,
				"weight": 1,
				"operatingStatus": "ONLINE"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, membersURL+"/mem-001",
		httpmock.NewStringResponder(200, `{
			"id": "mem-001",
			"resourceStatus": "READY",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:05:00Z",
			"currentState": {
				"name": "web-1",
				"address": "10.0.0.10",
				"protocolPort": 80,
				"weight": 2,
				"backup": true,
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE",
				"subnet": {"id": "sub-001"}
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "get", "lb-001", "pool-001", "mem-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 👤 Pool member mem-001

  *web-1*

  ## General information

  **ID**:                  mem-001
  **Status**:              READY
  **Address**:             10.0.0.10
  **Protocol port**:       80
  **Weight**:              2
  **Backup**:              true
  **Operating status**:    ONLINE
  **Provisioning status**: ACTIVE
  **Subnet ID**:           sub-001
  **Creation date**:       2026-11-10T10:00:00Z
  **Update date**:         2026-11-10T10:05:00Z

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberGetCreatingCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, membersURL+"/mem-001",
		httpmock.NewStringResponder(200, `{
			"id": "mem-001",
			"resourceStatus": "CREATING",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:00:00Z",
			"currentState": null,
			"targetSpec": {"name": "web-1", "address": "10.0.0.10", "protocolPort": 8080}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "get", "lb-001", "pool-001", "mem-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("*web-1*"))
	assert.Cmp(out, td.Re(`\*\*Address\*\*:\s+10\.0\.0\.10`))
	assert.Cmp(out, td.Re(`\*\*Operating status\*\*:\s+\*not yet available\*`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, membersURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"address": "10.0.0.10",
				"protocolPort": 80,
				"name": "web-1",
				"weight": 2,
				"backup": true,
				"subnet": {"id": "sub-001"}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "mem-001"}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "create", "lb-001", "pool-001",
		"--cloud-project", "fakeProjectID",
		"--address", "10.0.0.10",
		"--protocol-port", "80",
		"--name", "web-1",
		"--weight", "2",
		"--backup",
		"--subnet-id", "sub-001",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ Pool member created successfully (ID: mem-001)", "details": {"id": "mem-001"}}`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberCreateMissingAddressCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "create", "lb-001", "pool-001",
		"--cloud-project", "fakeProjectID",
		"--protocol-port", "80")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`targetSpec.address`))
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, membersURL+"/mem-001",
		httpmock.NewStringResponder(200, `{
			"id": "mem-001",
			"checksum": "abc",
			"targetSpec": {"name": "web-1", "address": "10.0.0.10", "protocolPort": 80, "weight": 1}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, membersURL+"/mem-001",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc",
			"targetSpec": {"name": "web-1", "weight": 5}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "mem-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "edit", "lb-001", "pool-001", "mem-001",
		"--cloud-project", "fakeProjectID",
		"--weight", "5")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerPoolMemberDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, membersURL+"/mem-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "pool", "member", "delete", "lb-001", "pool-001", "mem-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ Pool member mem-001 deleted successfully"
	}`))
}

// ---------------------------------------------------------------------------
// L7 Policy
// ---------------------------------------------------------------------------

const l7PoliciesURL = listenersURL + "/lis-001/l7policy"

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, l7PoliciesURL,
		httpmock.NewStringResponder(200, `[
			{
				"id": "pol-001",
				"resourceStatus": "READY",
				"currentState": {
					"name": "to-api",
					"action": "REDIRECT_TO_POOL",
					"position": 1,
					"operatingStatus": "ONLINE",
					"provisioningStatus": "ACTIVE"
				}
			}
		]`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "ls", "lb-001", "lis-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"id": "pol-001",
			"resourceStatus": "READY",
			"currentState": {
				"name": "to-api",
				"action": "REDIRECT_TO_POOL",
				"position": 1,
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE"
			}
		}
	]`))
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, l7PoliciesURL+"/pol-001",
		httpmock.NewStringResponder(200, `{
			"id": "pol-001",
			"resourceStatus": "READY",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:05:00Z",
			"currentState": {
				"name": "to-api",
				"description": "Send /api to the API pool",
				"action": "REDIRECT_TO_POOL",
				"position": 1,
				"operatingStatus": "ONLINE",
				"provisioningStatus": "ACTIVE",
				"redirectPool": {"id": "pool-001"},
				"rules": [
					{"id": "rule-001", "type": "PATH", "compareType": "STARTS_WITH", "value": "/api", "key": "", "invert": false}
				]
			}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "get", "lb-001", "lis-001", "pol-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("*to-api*"))
	assert.Cmp(out, td.Re(`\*\*Action\*\*:\s+REDIRECT_TO_POOL`))
	assert.Cmp(out, td.Re(`\*\*Redirect pool ID\*\*:\s+pool-001`))
	assert.Cmp(out, td.Contains("/api"))
	assert.Cmp(out, td.Contains("STARTS_WITH"))
	assert.Cmp(out, td.Not(td.Contains("No rule")))
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyGetCreatingWithoutRuleCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, l7PoliciesURL+"/pol-001",
		httpmock.NewStringResponder(200, `{
			"id": "pol-001",
			"resourceStatus": "CREATING",
			"createdAt": "2026-11-10T10:00:00Z",
			"updatedAt": "2026-11-10T10:00:00Z",
			"currentState": null,
			"targetSpec": {"name": "deny", "action": "REJECT", "position": 2}
		}`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "get", "lb-001", "lis-001", "pol-001", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("*deny*"))
	assert.Cmp(out, td.Re(`\*\*Action\*\*:\s+REJECT`))
	assert.Cmp(out, td.Re(`\*\*Operating status\*\*:\s+\*not yet available\*`))
	assert.Cmp(out, td.Contains("No rule"))
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyCreateWithRuleCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, l7PoliciesURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "to-api",
				"action": "REDIRECT_TO_POOL",
				"position": 1,
				"redirectPool": {"id": "pool-001"},
				"rules": [
					{"type": "PATH", "compareType": "STARTS_WITH", "value": "/api"}
				]
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pol-001"}`),
	)

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "create", "lb-001", "lis-001",
		"--cloud-project", "fakeProjectID",
		"--name", "to-api",
		"--action", "redirect_to_pool",
		"--position", "1",
		"--redirect-pool-id", "pool-001",
		"--rule-type", "path",
		"--rule-compare-type", "starts_with",
		"--rule-value", "/api",
		"-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"message": "✅ L7 policy created successfully (ID: pol-001)", "details": {"id": "pol-001"}}`))
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyCreateWithoutRuleCmd(assert, require *td.T) {
	// The rules key must be absent from the body when no --rule-* flag is given
	httpmock.RegisterMatcherResponder(http.MethodPost, l7PoliciesURL,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {"action": "REJECT", "name": "deny"}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pol-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "create", "lb-001", "lis-001",
		"--cloud-project", "fakeProjectID",
		"--name", "deny",
		"--action", "REJECT")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyCreateMissingActionCmd(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "create", "lb-001", "lis-001",
		"--cloud-project", "fakeProjectID",
		"--name", "deny")
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`targetSpec.action`))
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyEditKeepsRulesCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, l7PoliciesURL+"/pol-001",
		httpmock.NewStringResponder(200, `{
			"id": "pol-001",
			"checksum": "abc",
			"targetSpec": {
				"name": "to-api",
				"action": "REDIRECT_TO_POOL",
				"position": 1,
				"redirectPool": {"id": "pool-001"},
				"rules": [{"type": "PATH", "compareType": "STARTS_WITH", "value": "/api"}]
			}
		}`))

	// The rules of the policy are sent back unchanged: the PUT replaces the whole list
	httpmock.RegisterMatcherResponder(http.MethodPut, l7PoliciesURL+"/pol-001",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc",
			"targetSpec": {
				"name": "to-api",
				"action": "REDIRECT_TO_POOL",
				"position": 3,
				"redirectPool": {"id": "pool-001"},
				"rules": [{"type": "PATH", "compareType": "STARTS_WITH", "value": "/api"}]
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "pol-001"}`),
	)

	_, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "edit", "lb-001", "lis-001", "pol-001",
		"--cloud-project", "fakeProjectID",
		"--position", "3")
	require.CmpNoError(err)
}

func (ms *MockSuite) TestCloudLoadbalancerL7PolicyDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, l7PoliciesURL+"/pol-001",
		httpmock.NewStringResponder(200, `null`))

	out, err := cmd.Execute("cloud", "loadbalancer", "l7policy", "delete", "lb-001", "lis-001", "pol-001", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"message": "✅ L7 policy pol-001 deleted successfully"
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
