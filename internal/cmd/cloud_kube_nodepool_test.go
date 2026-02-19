// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// List Nodepool.
func (ms *MockSuite) TestCloudKubeNodepoolListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12345/nodepool",
		httpmock.NewStringResponder(200, `[
		{
			"id": "rototo",
			"name": "nodepool-2025-12-04",
			"flavor": "b3-8",
			"currentNodes": 2,
			"status": "READY"
		},
		{
			"id": "rototo2",
			"name": "nodepool-2025-12-05",
			"flavor": "b3-8",
			"currentNodes": 3,
			"status": "UPSCALING"
		}
		]`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "list", "MyMksID-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌─────────┬─────────────────────┬────────┬──────────────┬───────────┐
│   id    │        name         │ flavor │ currentNodes │  status   │
├─────────┼─────────────────────┼────────┼──────────────┼───────────┤
│ rototo  │ nodepool-2025-12-04 │ b3-8   │ 2            │ READY     │
│ rototo2 │ nodepool-2025-12-05 │ b3-8   │ 3            │ UPSCALING │
└─────────┴─────────────────────┴────────┴──────────────┴───────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// Get a Nodepool.
func (ms *MockSuite) TestCloudKubeNodepoolGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12345/nodepool/MyNodePool",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodePoolId",
			"projectId": "fakeProjectID",
			"name": "nodepool1",
			"flavor": "b3-32",
			"status": "READY",
			"sizeStatus": "CAPACITY_OK",
			"autoscale": false,
			"monthlyBilled": false,
			"antiAffinity": false,
			"desiredNodes": 0,
			"minNodes": 0,
			"maxNodes": 100,
			"currentNodes": 0,
			"availableNodes": 0,
			"upToDateNodes": 0,
			"createdAt": "2025-12-04T15:29:32.487775Z",
			"updatedAt": "2025-12-05T15:51:08Z",
			"autoscaling": {
				"scaleDownUtilizationThreshold": 0.5,
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200
			},
			"template": {
				"metadata": {
					"labels": {},
					"annotations": {},
					"finalizers": []
			},
				"spec": {
					"unschedulable": false,
					"taints": []
			 		}
			},
			"attachFloatingIps": {
				"enabled": false
			},
			"availabilityZones": [
				"myzone"
			]
		}`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "get", "MyMksID-12345", "MyNodePool", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `
  # 🚀 Managed Kubernetes Node Pool MyNodePool

  *nodepool1*

  ## General information

  **Status**:             READY
  **Project ID**:         fakeProjectID
  **Availability Zones**: myzone
  **Monthly Billed**:     false
  **Flavor**:             b3-32
  **Creation date**:      2025-12-04T15:29:32.487775Z
  **Update date**:        2025-12-05T15:51:08Z

  **Anti Affinity**:      false
  **Autoscale**:          false

  **AttachFloatingIP**:   false

  **Autoscaling**:
  - Scale Down Unneeded Time (s): 600
  - Scale Down Unready Time (s):  1200
  - Scale Down Utilization Threshold*: 0.5

  * Sum of CPU or memory of all pods running on the node divided by node's
  corresponding allocatable resource.

  ## Node pool state

  **Ready nodes**:        0
  **Current Nodes**:      0
  **Desired Nodes**:      0
  **Max Nodes**:          100
  **Min Nodes**:          0
  **Size Status**:        CAPACITY_OK
  **Up To Date Nodes**:   0

  ## Template

  ### Metadata

  **Annotations**:
  **Finalizers**:
  **Labels**:

  ### Spec

  **Taints**:
  **Unschedulable**: false

  💡 Use option -o json or -o yaml to get the raw output with all information

`)
}

// Create a Nodepool with the attachFloatingIps flag.
// The nodepool spec must be set to true.
func (ms *MockSuite) TestCloudKubeNodepoolCreateCmdWithAttachFloatingIps(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12345/nodepool",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"attachFloatingIps": {
					"enabled": true
				},
				"availabilityZones": [
					"myzone"
				],
				"desiredNodes": 11,
				"flavorName": "b3-8",
				"name": "mynodepoolname"
			}`)),
		httpmock.NewStringResponder(200, `{
			"id": "mynodepoolid",
			"name": "mynodepoolname"
		}`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "create", "MyMksID-12345", "--flavor-name", "b3-8", "--name", "mynodepoolname", "--availability-zones", "myzone", "--desired-nodes", "11", "--cloud-project", "fakeProjectID", "--attach-floating-ips")

	require.CmpNoError(err)
	assert.String(out, `✅ Node pool mynodepoolid created successfully`)
}

// Create a Nodepool without the attachFloatingIps flag.
// The nodepool spec must be set to false
func (ms *MockSuite) TestCloudKubeNodepoolCreateCmdWithoutAttachFloatingIps(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"attachFloatingIps": {
					"enabled": false
				},
				"availabilityZones": [
					"myzone2"
				],
				"desiredNodes": 12,
				"flavorName": "b3-16",
				"name": "mynodepoolname2"
			}`)),
		httpmock.NewStringResponder(200, `{
			"id": "mynodepool2id",
			"name": "mynodepoolname2"
		}`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "create", "MyMksID-12346", "--flavor-name", "b3-16", "--name", "mynodepoolname2", "--availability-zones", "myzone2", "--desired-nodes", "12", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Node pool mynodepool2id created successfully`)
}

// Create a Nodepool with the attachFloatingIps flag set to false.
// The nodepool spec must be set to false.
func (ms *MockSuite) TestCloudKubeNodepoolCreateCmdWithAttachFloatingIpsSetFalse(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"attachFloatingIps": {
					"enabled": false
				},
				"availabilityZones": [
					"myzone3"
				],
				"desiredNodes": 12,
				"flavorName": "b3-16",
				"name": "mynodepoolname3"
			}`)),
		httpmock.NewStringResponder(200, `{
			"id": "mynodepool3id",
			"name": "mynodepoolname3"
		}`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "create", "MyMksID-12346", "--flavor-name", "b3-16", "--name", "mynodepoolname3", "--availability-zones", "myzone3", "--desired-nodes", "12", "--cloud-project", "fakeProjectID", "--attach-floating-ips=false")

	require.CmpNoError(err)
	assert.String(out, `✅ Node pool mynodepool3id created successfully`)
}

// Update a Nodepool with attachFloatingIps disabled with the flag attachFloatingIps
// The spec must be updated from false to true.
func (ms *MockSuite) TestCloudKubeNodepoolEditCmdWithAttachFloatingIpsTrue(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodePoolId",
			"projectId": "fakeProjectID",
			"name": "nodepool1",
			"flavor": "b3-32",
			"status": "READY",
			"sizeStatus": "CAPACITY_OK",
			"autoscale": false,
			"monthlyBilled": false,
			"antiAffinity": false,
			"desiredNodes": 0,
			"minNodes": 0,
			"maxNodes": 100,
			"currentNodes": 0,
			"availableNodes": 0,
			"upToDateNodes": 0,
			"createdAt": "2025-12-04T15:29:32.487775Z",
			"updatedAt": "2025-12-05T15:51:08Z",
			"autoscaling": {
				"scaleDownUtilizationThreshold": 0.5,
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200
			},
			"template": {
				"metadata": {
					"labels": {},
					"annotations": {},
					"finalizers": []
			},
				"spec": {
					"unschedulable": false,
					"taints": []
			 		}
			},
			"attachFloatingIps": {
				"enabled": false
			},
			"availabilityZones": [
				"myzone"
			]
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		tdhttpmock.JSONBody(td.JSON(`
		{
			"attachFloatingIps": {
				"enabled": true
			},
			"autoscale": false,
			"autoscaling": {
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200,
				"scaleDownUtilizationThreshold": 0.5
			},
			"desiredNodes": 0,
			"maxNodes": 100,
			"minNodes": 0,
			"template": {
				"metadata": {
					"annotations": {},
					"finalizers": [],
					"labels": {}
				},
				"spec": {
					"taints": [],
					"unschedulable": false
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `✅ Resource updated successfully`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "edit", "MyMksID-12346", "MyNodePoolId", "--cloud-project", "fakeProjectID", "--attach-floating-ips")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `✅ Resource updated successfully`)
}

// Update a Nodepool with attachFloatingIps enabled specifying the flag attachFloatingIps=false
// The spec must be updated from true to false.
func (ms *MockSuite) TestCloudKubeNodepoolEditCmdWithAttachFloatingIpsFalse(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodePoolId",
			"projectId": "fakeProjectID",
			"name": "nodepool1",
			"flavor": "b3-32",
			"status": "READY",
			"sizeStatus": "CAPACITY_OK",
			"autoscale": false,
			"monthlyBilled": false,
			"antiAffinity": false,
			"desiredNodes": 0,
			"minNodes": 0,
			"maxNodes": 100,
			"currentNodes": 0,
			"availableNodes": 0,
			"upToDateNodes": 0,
			"createdAt": "2025-12-04T15:29:32.487775Z",
			"updatedAt": "2025-12-05T15:51:08Z",
			"autoscaling": {
				"scaleDownUtilizationThreshold": 0.5,
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200
			},
			"template": {
				"metadata": {
					"labels": {},
					"annotations": {},
					"finalizers": []
			},
				"spec": {
					"unschedulable": false,
					"taints": []
			 		}
			},
			"attachFloatingIps": {
				"enabled": true
			},
			"availabilityZones": [
				"myzone"
			]
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		tdhttpmock.JSONBody(td.JSON(`
		{
			"attachFloatingIps": {
				"enabled": false
			},
			"autoscale": false,
			"autoscaling": {
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200,
				"scaleDownUtilizationThreshold": 0.5
			},
			"desiredNodes": 0,
			"maxNodes": 100,
			"minNodes": 0,
			"template": {
				"metadata": {
					"annotations": {},
					"finalizers": [],
					"labels": {}
				},
				"spec": {
					"taints": [],
					"unschedulable": false
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `✅ Resource updated successfully`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "edit", "MyMksID-12346", "MyNodePoolId", "--cloud-project", "fakeProjectID", "--attach-floating-ips=false")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `✅ Resource updated successfully`)
}

// Update a Nodepool with attachFloatingIps enabled without specify the flag
// The spec must not be updated and must be true.
func (ms *MockSuite) TestCloudKubeNodepoolEditCmdWithoutAttachFloatingIpsTrue(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodePoolId",
			"projectId": "fakeProjectID",
			"name": "nodepool1",
			"flavor": "b3-32",
			"status": "READY",
			"sizeStatus": "CAPACITY_OK",
			"autoscale": false,
			"monthlyBilled": false,
			"antiAffinity": false,
			"desiredNodes": 0,
			"minNodes": 0,
			"maxNodes": 100,
			"currentNodes": 0,
			"availableNodes": 0,
			"upToDateNodes": 0,
			"createdAt": "2025-12-04T15:29:32.487775Z",
			"updatedAt": "2025-12-05T15:51:08Z",
			"autoscaling": {
				"scaleDownUtilizationThreshold": 0.5,
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200
			},
			"template": {
				"metadata": {
					"labels": {},
					"annotations": {},
					"finalizers": []
			},
				"spec": {
					"unschedulable": false,
					"taints": []
			 		}
			},
			"attachFloatingIps": {
				"enabled": true
			},
			"availabilityZones": [
				"myzone"
			]
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		tdhttpmock.JSONBody(td.JSON(`
		{
			"attachFloatingIps": {
				"enabled": true
			},
			"autoscale": false,
			"autoscaling": {
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200,
				"scaleDownUtilizationThreshold": 0.5
			},
			"desiredNodes": 0,
			"maxNodes": 100,
			"minNodes": 0,
			"template": {
				"metadata": {
					"annotations": {},
					"finalizers": [],
					"labels": {}
				},
				"spec": {
					"taints": [],
					"unschedulable": false
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `✅ Resource updated successfully`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "edit", "MyMksID-12346", "MyNodePoolId", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `✅ Resource updated successfully`)
}

// Update Nodepool with attachFloatingIps disabled without specify the flag
// The spec must not be updated and must be false.
func (ms *MockSuite) TestCloudKubeNodepoolEditCmdWithoutAttachFloatingIpsFalse(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		httpmock.NewStringResponder(200, `
		{
			"id": "MyNodePoolId",
			"projectId": "fakeProjectID",
			"name": "nodepool1",
			"flavor": "b3-32",
			"status": "READY",
			"sizeStatus": "CAPACITY_OK",
			"autoscale": false,
			"monthlyBilled": false,
			"antiAffinity": false,
			"desiredNodes": 0,
			"minNodes": 0,
			"maxNodes": 100,
			"currentNodes": 0,
			"availableNodes": 0,
			"upToDateNodes": 0,
			"createdAt": "2025-12-04T15:29:32.487775Z",
			"updatedAt": "2025-12-05T15:51:08Z",
			"autoscaling": {
				"scaleDownUtilizationThreshold": 0.5,
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200
			},
			"template": {
				"metadata": {
					"labels": {},
					"annotations": {},
					"finalizers": []
			},
				"spec": {
					"unschedulable": false,
					"taints": []
			 		}
			},
			"attachFloatingIps": {
				"enabled": false
			},
			"availabilityZones": [
				"myzone"
			]
		}`).Once())

	httpmock.RegisterMatcherResponder(
		http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/MyMksID-12346/nodepool/MyNodePoolId",
		tdhttpmock.JSONBody(td.JSON(`
		{
			"attachFloatingIps": {
				"enabled": false
			},
			"autoscale": false,
			"autoscaling": {
				"scaleDownUnneededTimeSeconds": 600,
				"scaleDownUnreadyTimeSeconds": 1200,
				"scaleDownUtilizationThreshold": 0.5
			},
			"desiredNodes": 0,
			"maxNodes": 100,
			"minNodes": 0,
			"template": {
				"metadata": {
					"annotations": {},
					"finalizers": [],
					"labels": {}
				},
				"spec": {
					"taints": [],
					"unschedulable": false
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `✅ Resource updated successfully`).Once())

	out, err := cmd.Execute("cloud", "kube", "nodepool", "edit", "MyMksID-12346", "MyNodePoolId", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), `✅ Resource updated successfully`)
}
