// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// List nodes of a nodepool.
func (ms *MockSuite) TestManagedKubernetesNodeListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedKubernetes/MyMksID-12345/nodepool/MyNodePool/node",
		httpmock.NewStringResponder(200, `[
		{
			"id": "node1",
			"createdAt": "2026-09-10T09:00:12Z",
			"updatedAt": "2026-09-10T09:04:41Z",
			"resourceStatus": "READY",
			"currentState": {
				"name": "nodepool1-node-abcdef",
				"flavor": "b3-8",
				"nodePoolId": "MyNodePool",
				"version": "1.33.1",
				"upToDate": true,
				"deployedAt": "2026-09-10T09:04:41Z"
			}
		},
		{
			"id": "node2",
			"createdAt": "2026-09-10T09:00:12Z",
			"updatedAt": "2026-09-10T09:04:41Z",
			"resourceStatus": "UPDATING",
			"currentState": {
				"name": "nodepool1-node-123456",
				"flavor": "b3-8",
				"nodePoolId": "MyNodePool",
				"version": "1.32.4",
				"upToDate": false
			}
		}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "node", "list", "MyMksID-12345", "MyNodePool", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌───────┬───────────────────────┬────────┬─────────┬────────────┬──────────┬──────────────────────┐
│  ID   │         name          │ flavor │ version │ up-to-date │  status  │     deployed at      │
├───────┼───────────────────────┼────────┼─────────┼────────────┼──────────┼──────────────────────┤
│ node1 │ nodepool1-node-abcdef │ b3-8   │ 1.33.1  │ true       │ READY    │ 2026-09-10T09:04:41Z │
│ node2 │ nodepool1-node-123456 │ b3-8   │ 1.32.4  │ false      │ UPDATING │                      │
└───────┴───────────────────────┴────────┴─────────┴────────────┴──────────┴──────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// List nodes of a nodepool, filtering on the lifecycle status.
func (ms *MockSuite) TestManagedKubernetesNodeListCmdWithFilters(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedKubernetes/MyMksID-12345/nodepool/MyNodePool/node?includeDeleted=true&resourceStatus=DELETED",
		httpmock.NewStringResponder(200, `[
		{
			"id": "node3",
			"createdAt": "2026-09-10T09:00:12Z",
			"updatedAt": "2026-09-10T09:04:41Z",
			"resourceStatus": "DELETED",
			"currentState": {
				"name": "nodepool1-node-deleted",
				"flavor": "b3-8",
				"nodePoolId": "MyNodePool",
				"version": "1.33.1",
				"upToDate": true
			}
		}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "node", "list", "MyMksID-12345", "MyNodePool",
		"--include-deleted", "--resource-status", "DELETED", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌───────┬────────────────────────┬────────┬─────────┬────────────┬─────────┬─────────────┐
│  ID   │          name          │ flavor │ version │ up-to-date │ status  │ deployed at │
├───────┼────────────────────────┼────────┼─────────┼────────────┼─────────┼─────────────┤
│ node3 │ nodepool1-node-deleted │ b3-8   │ 1.33.1  │ true       │ DELETED │             │
└───────┴────────────────────────┴────────┴─────────┴────────────┴─────────┴─────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// Get a node.
func (ms *MockSuite) TestManagedKubernetesNodeGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedKubernetes/MyMksID-12345/nodepool/MyNodePool/node/MyNodeID",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodeID",
			"createdAt": "2025-12-04T10:00:00Z",
			"updatedAt": "2025-12-05T09:00:00Z",
			"resourceStatus": "READY",
			"currentState": {
				"name": "nodepool1-node-abcdef",
				"flavor": "b3-8",
				"instanceId": "MyInstanceID",
				"nodePoolId": "MyNodePool",
				"deployedAt": "2025-12-04T10:05:00Z",
				"upToDate": true,
				"version": "1.33.1"
			}
		}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "node", "get", "MyMksID-12345", "MyNodePool", "MyNodeID", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Managed Kubernetes Node MyNodeID

  *nodepool1-node-abcdef*

  ## General information

  **Flavor**:          b3-8
  **Instance ID**:     MyInstanceID
  **Nodepool ID**:     MyNodePool
  **Deployment date**: 2025-12-04T10:05:00Z
  **Creation date**:   2025-12-04T10:00:00Z
  **Update date**:     2025-12-05T09:00:00Z

  ## Node state

  **Version**:    1.33.1
  **Status**:     READY
  **Up to date**: true

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}
