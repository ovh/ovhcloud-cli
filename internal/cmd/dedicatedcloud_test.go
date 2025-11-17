// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestDedicatedCloudListCmd(assert, require *td.T) {
	// Mock the list of dedicated clouds
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud",
		httpmock.NewStringResponder(200, `["pcc-12345","pcc-67890"]`).Once())

	// Mock expanded dedicated cloud details
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345",
		httpmock.NewStringResponder(200, `{
			"serviceName": "pcc-12345",
			"location": "pcc-12345",
			"state": "delivered",
			"description": "Test PCC",
			"version": {
				"major": "8",
				"minor": "0",
				"build": "U3e.24674346"
			}
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-67890",
		httpmock.NewStringResponder(200, `{
			"serviceName": "pcc-67890",
			"location": "pcc-67890",
			"state": "delivered",
			"description": "Test PCC 2",
			"version": {
				"major": "8",
				"minor": "0",
				"build": "U3g.24853646"
			}
		}`).Once())

	// Mock location list
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/location",
		httpmock.NewStringResponder(200, `["pcc-12345","pcc-67890"]`).Once())

	// Mock location details
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/location/pcc-12345",
		httpmock.NewStringResponder(200, `{
			"regionLocation": "Europe (France - Gravelines)"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/location/pcc-67890",
		httpmock.NewStringResponder(200, `{
			"regionLocation": "Europe (United Kingdom - Erith)"
		}`).Once())

	out, err := cmd.Execute("dedicated-cloud", "list")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("pcc-12345"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("pcc-67890"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Europe (France - Gravelines)"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("8.0.U3e.24674346"))
}

func (ms *MockSuite) TestDedicatedCloudGetCmd(assert, require *td.T) {
	// Mock dedicated cloud details
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345",
		httpmock.NewStringResponder(200, `{
			"serviceName": "pcc-12345",
			"location": "pcc-12345",
			"state": "delivered",
			"generation": "8",
			"version": {
				"major": "8",
				"minor": "0",
				"build": "U3e.24674346"
			},
			"commercialRange": "nsx-t",
			"billingType": "hourly",
			"webInterfaceUrl": "https://pcc-12345.ovh.com",
			"vScopeUrl": "https://vscope.ovh.com/pcc-12345",
			"advancedSecurity": true,
			"bandwidth": "1Gbps",
			"spla": false,
			"sslV3": false,
			"userAccessPolicy": "readOnly",
			"iam": {
				"displayName": "PCC-12345",
				"urn": "urn:v1:eu:resource:pccVMware:pcc-12345"
			}
		}`).Once())

	// Mock datacenters list
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter",
		httpmock.NewStringResponder(200, `[1, 2]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1",
		httpmock.NewStringResponder(200, `{
			"datacenterId": 1,
			"name": "datacenter-1",
			"version": "8.0",
			"commercialName": "Enterprise"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/2",
		httpmock.NewStringResponder(200, `{
			"datacenterId": 2,
			"name": "datacenter-2",
			"version": "8.0",
			"commercialName": "Standard"
		}`).Once())

	// Mock location list and details for regionLocation
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/location",
		httpmock.NewStringResponder(200, `["pcc-12345"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/location/pcc-12345",
		httpmock.NewStringResponder(200, `{
			"regionLocation": "Europe (France - Gravelines)"
		}`).Once())

	out, err := cmd.Execute("dedicated-cloud", "get", "pcc-12345")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("pcc-12345"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Europe (France - Gravelines)"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("8.0.U3e.24674346"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("datacenter-1"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("datacenter-2"))
}

func (ms *MockSuite) TestDedicatedCloudDatacenterListCmd(assert, require *td.T) {
	// Mock datacenters list
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter",
		httpmock.NewStringResponder(200, `[1]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1",
		httpmock.NewStringResponder(200, `{
			"datacenterId": 1,
			"name": "datacenter-1",
			"version": "8.0",
			"commercialName": "Enterprise"
		}`).Once())

	// Mock hosts for totals calculation
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/host",
		httpmock.NewStringResponder(200, `[1481, 1511]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/host/1481",
		httpmock.NewStringResponder(200, `{
			"hostId": 1481,
			"cpuNum": 32,
			"ram": {
				"value": 768,
				"unit": "GB"
			},
			"vmTotal": 36
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/host/1511",
		httpmock.NewStringResponder(200, `{
			"hostId": 1511,
			"cpuNum": 32,
			"ram": {
				"value": 768,
				"unit": "GB"
			},
			"vmTotal": 16
		}`).Once())

	// Mock filers for totals calculation
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/filer",
		httpmock.NewStringResponder(200, `[1557]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/filer/1557",
		httpmock.NewStringResponder(200, `{
			"filerId": 1557,
			"size": {
				"value": 3000,
				"unit": "GB"
			}
		}`).Once())

	// Mock global filers
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/filer",
		httpmock.NewStringResponder(200, `[]`).Once())

	out, err := cmd.Execute("dedicated-cloud", "datacenter", "list", "pcc-12345")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("datacenter-1"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("64"))      // totalCores: 32 + 32
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("1536 GB")) // totalRAM: 768 + 768
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("52"))      // totalVMs: 36 + 16
}

func (ms *MockSuite) TestDedicatedCloudDatacenterGetCmd(assert, require *td.T) {
	// Mock datacenter details
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1",
		httpmock.NewStringResponder(200, `{
			"datacenterId": 1,
			"name": "datacenter-1",
			"version": "8.0",
			"commercialName": "Enterprise",
			"commercialRange": "nsx-t",
			"removable": false
		}`).Once())

	// Mock clusters
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/cluster",
		httpmock.NewStringResponder(200, `[25035]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/cluster/25035",
		httpmock.NewStringResponder(200, `{
			"clusterId": 25035,
			"name": "Cluster1",
			"drsStatus": "enabled",
			"drsMode": "fullyAutomated",
			"haStatus": "enabled",
			"evcMode": "disabled"
		}`).Once())

	// Mock hosts
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/host",
		httpmock.NewStringResponder(200, `[1481]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/host/1481",
		httpmock.NewStringResponder(200, `{
			"hostId": 1481,
			"name": "172.17.136.56",
			"clusterName": "Cluster1",
			"clusterId": 25035,
			"connectionState": "connected",
			"inMaintenance": false,
			"cpuNum": 32,
			"cpu": {
				"value": 93,
				"unit": "GHz"
			},
			"ram": {
				"value": 768,
				"unit": "GB"
			},
			"profile": "PRE 768 NSX",
			"vmTotal": 36,
			"billingType": "hourly"
		}`).Once())

	// Mock local filers
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/filer",
		httpmock.NewStringResponder(200, `[1557]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/datacenter/1/filer/1557",
		httpmock.NewStringResponder(200, `{
			"filerId": 1557,
			"name": "ssd-001557",
			"connectionState": "online",
			"state": "delivered",
			"size": {
				"value": 3000,
				"unit": "GB"
			},
			"spaceFree": {
				"value": 2972,
				"unit": "GB"
			},
			"master": "cluster5068.example.com",
			"vmTotal": 0,
			"billingType": "hourly",
			"activeNode": "master"
		}`).Once())

	// Mock global filers
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345/filer",
		httpmock.NewStringResponder(200, `[]`).Once())

	// Mock dedicated cloud for IAM
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicatedCloud/pcc-12345",
		httpmock.NewStringResponder(200, `{
			"iam": {
				"displayName": "PCC-12345",
				"urn": "urn:v1:eu:resource:pccVMware:pcc-12345"
			}
		}`).Once())

	out, err := cmd.Execute("dedicated-cloud", "datacenter", "get", "pcc-12345", "1")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("datacenter-1"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Total CPU (Cores)"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Total RAM"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Total VMs"))
	assert.Cmp(cleanWhitespacesHelper(out), td.Contains("Total Disk Space"))
}
