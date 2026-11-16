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

const managedKubernetesNodepoolV2Endpoint = managedKubernetesV2Endpoint + "/kube-12345/nodepool"

// managedKubernetesNodepoolV2Object is a node pool as returned by the API v2.
const managedKubernetesNodepoolV2Object = `{
	"id": "np-12345",
	"checksum": "abc123",
	"resourceStatus": "READY",
	"createdAt": "2025-12-04T15:29:32Z",
	"updatedAt": "2025-12-05T15:51:08Z",
	"currentState": {
		"name": "nodepool1",
		"availabilityZone": "eu-west-par-a",
		"billingType": "HOURLY",
		"desiredNodes": 3,
		"nodeConfig": {"flavor": "b3-8", "attachFloatingIp": {"enabled": false}},
		"antiAffinity": {"enabled": false},
		"autoscaler": {"enabled": true, "minNodes": 1, "maxNodes": 5, "scaleDown": {"utilizationThreshold": 0.5}},
		"status": {"currentNodes": 3, "readyNodes": 2, "upToDateNodes": 3},
		"template": {
			"metadata": {"labels": {"env": "prod"}},
			"spec": {"taints": [{"key": "k", "value": "v", "effect": "NoSchedule"}], "unschedulable": false}
		}
	},
	"targetSpec": {
		"name": "nodepool1",
		"availabilityZone": "eu-west-par-a",
		"billingType": "HOURLY",
		"desiredNodes": 3,
		"nodeConfig": {"flavor": "b3-8"},
		"autoscaler": {"enabled": true, "minNodes": 1, "maxNodes": 5},
		"template": {"metadata": {"labels": {"env": "prod"}}}
	}
}`

func (ms *MockSuite) TestManagedKubernetesNodepoolListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesNodepoolV2Endpoint,
		httpmock.NewStringResponder(200, `[`+managedKubernetesNodepoolV2Object+`]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "list", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌──────────┬───────────┬────────┬───────────────┬───────────────┬──────────────────┬────────┬──────────────────────┐
│    ID    │   name    │ flavor │ desired nodes │ current nodes │ up-to-date nodes │ status │      created at      │
├──────────┼───────────┼────────┼───────────────┼───────────────┼──────────────────┼────────┼──────────────────────┤
│ np-12345 │ nodepool1 │ b3-8   │ 3             │ 3             │ 3                │ READY  │ 2025-12-04T15:29:32Z │
└──────────┴───────────┴────────┴───────────────┴───────────────┴──────────────────┴────────┴──────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesNodepoolGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		httpmock.NewStringResponder(200, managedKubernetesNodepoolV2Object).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "get", "kube-12345", "np-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(cleanWhitespacesHelper(out), td.All(
		td.Contains("nodepool1"),
		td.Contains("**Flavor**:            b3-8"),
		td.Contains("**Availability zone**: eu-west-par-a"),
		td.Contains("**Ready nodes**:      2"),
		td.Contains("**Max nodes**: 5"),
		td.Contains("env: prod"),
		td.Contains("k=v:NoSchedule"),
	))
}

func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesNodepoolV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "mynodepool",
				"desiredNodes": 3,
				"availabilityZone": "eu-west-par-a",
				"billingType": "MONTHLY",
				"nodeConfig": {"flavor": "b3-8", "attachFloatingIp": {"enabled": true}},
				"antiAffinity": {"enabled": true},
				"autoscaler": {"enabled": true, "minNodes": 0, "maxNodes": 5, "scaleDown": {"utilizationThreshold": 0.4}},
				"template": {
					"spec": {"taints": [{"key": "k", "value": "v", "effect": "NoSchedule"}], "unschedulable": false}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "np-12345"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--name", "mynodepool",
		"--flavor", "b3-8",
		"--desired-nodes", "3",
		"--availability-zone", "eu-west-par-a",
		"--billing-type", "MONTHLY",
		"--attach-floating-ip",
		"--anti-affinity-enabled",
		"--autoscaler-enabled",
		"--autoscaler-min-nodes", "0",
		"--autoscaler-max-nodes", "5",
		"--autoscaler-scale-down-utilization-threshold", "0.4",
		"--template-taints", "k=v:NoSchedule",
		"--template-unschedulable=false")

	require.CmpNoError(err)
	assert.String(out, `✅ Node pool created successfully (id: np-12345)`)
}

// TestManagedKubernetesNodepoolCreateCmdEmpty tests that zero desired nodes is sent rather than omitted.
func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdEmpty(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesNodepoolV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{"targetSpec": {"desiredNodes": 0, "nodeConfig": {"flavor": "b3-8"}}}`)),
		httpmock.NewStringResponder(200, `{"id": "np-12345"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID", "--flavor", "b3-8", "--desired-nodes", "0")

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdMissingClusterID(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "--cloud-project", "fakeProjectID")

	require.CmpError(err)
}

func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdMissingFlavor(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID", "--desired-nodes", "3")
	require.CmpError(err)
	assert.Contains(err.Error(), `required flag(s) "flavor" not set`)
}

func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdWithoutDesiredNodes(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesNodepoolV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{"targetSpec": {"nodeConfig": {"flavor": "b3-8"}}}`)),
		httpmock.NewStringResponder(200, managedKubernetesNodepoolV2Object))

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID", "--flavor", "b3-8")
	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdInvalidTaint(assert, require *td.T) {
	_, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID", "--flavor", "b3-8", "--desired-nodes", "1", "--template-taints", "novalue")

	require.CmpError(err)
	assert.Contains(err.Error(), "invalid taint format")
}

// TestManagedKubernetesNodepoolCreateCmdTaintWithoutValue tests that a taint value is optional, as with kubectl.
func (ms *MockSuite) TestManagedKubernetesNodepoolCreateCmdTaintWithoutValue(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesNodepoolV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{"targetSpec": {"desiredNodes": 1, "nodeConfig": {"flavor": "b3-8"},
			"template": {"spec": {"taints": [{"key": "dedicated", "value": "", "effect": "NoSchedule"}]}}}}`)),
		httpmock.NewStringResponder(200, `{"id": "np-12345"}`).Once())

	_, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "create", "kube-12345",
		"--cloud-project", "fakeProjectID", "--flavor", "b3-8", "--desired-nodes", "1", "--template-taints", "dedicated:NoSchedule")

	require.CmpNoError(err)
}

// TestManagedKubernetesNodepoolEditCmd tests that editing a node pool preserves its checksum
// (optimistic locking) and only sends the mutable fields.
func (ms *MockSuite) TestManagedKubernetesNodepoolEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		httpmock.NewStringResponder(200, managedKubernetesNodepoolV2Object).Once())

	httpmock.RegisterMatcherResponder(http.MethodPut, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc123",
			"targetSpec": {
				"desiredNodes": 4,
				"autoscaler": {"enabled": false, "minNodes": 1, "maxNodes": 5},
				"template": {
					"metadata": {"labels": {"env": "prod", "team": "a"}},
					"spec": {"unschedulable": true}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "edit", "kube-12345", "np-12345",
		"--cloud-project", "fakeProjectID",
		"--desired-nodes", "4",
		"--autoscaler-enabled=false",
		"--template-labels", "team=a",
		"--template-unschedulable")

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}

// TestManagedKubernetesNodepoolEditCmdReplacesTaints tests that the given taints replace the
// current ones, so a taint effect can be changed without duplicating its key.
func (ms *MockSuite) TestManagedKubernetesNodepoolEditCmdReplacesTaints(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		httpmock.NewStringResponder(200, `{
			"id": "np-12345",
			"checksum": "abc123",
			"targetSpec": {
				"desiredNodes": 3,
				"template": {"spec": {"taints": [{"key": "dedicated", "value": "gpu", "effect": "NoSchedule"}]}}
			}
		}`).Once())

	httpmock.RegisterMatcherResponder(http.MethodPut, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		tdhttpmock.JSONBody(td.JSONPointer("/targetSpec/template/spec/taints", []any{
			map[string]any{"key": "dedicated", "value": "gpu", "effect": "NoExecute"},
		})),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "edit", "kube-12345", "np-12345",
		"--cloud-project", "fakeProjectID",
		"--template-taints", "dedicated=gpu:NoExecute")

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}

func (ms *MockSuite) TestManagedKubernetesNodepoolDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, managedKubernetesNodepoolV2Endpoint+"/np-12345",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "delete", "kube-12345", "np-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Node pool deleted successfully`)
}
