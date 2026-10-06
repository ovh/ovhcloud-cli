// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
)

const managedKubernetesV2Endpoint = "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/managedKubernetes"

//
// LIST TESTS
//

// TestManagedKubernetesListCmd tests that listing kubes returns the expected output.
func (ms *MockSuite) TestManagedKubernetesListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint,
		httpmock.NewStringResponder(200, `[{
			"id": "kube-12345",
			"resourceStatus": "READY",
			"createdAt": "2026-09-10T09:00:00Z",
			"currentState": {
				"name": "test-kube",
				"region": "GRA11",
				"plan": "FREE",
				"version": "1.34"
			}
		}]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬───────────┬────────┬──────┬─────────┬────────┬──────────────────────┐
│     ID     │   name    │ region │ plan │ version │ status │      created at      │
├────────────┼───────────┼────────┼──────┼─────────┼────────┼──────────────────────┤
│ kube-12345 │ test-kube │ GRA11  │ FREE │ 1.34    │ READY  │ 2026-09-10T09:00:00Z │
└────────────┴───────────┴────────┴──────┴─────────┴────────┴──────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

//
// GET TESTS
//

// TestManagedKubernetesGetCmd tests that a cluster is rendered with its template.
func (ms *MockSuite) TestManagedKubernetesGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"createdAt": "2021-10-12T14:23:45Z",
			"updatedAt": "2021-10-13T14:23:45Z",
			"currentState": {
				"name": "test-kube",
				"region": "GRA11",
				"plan": "FREE",
				"version": "1.34",
				"deployedVersion": "1.34.1-2",
				"cni": {"type": "CANAL"},
				"network": {"type": "PUBLIC"},
				"apiServer": {
					"access": {
						"publicEndpoint": {"url": "https://kube-12345.gra11.k8s.ovh.net"},
						"authorizedIps": {"rules": [{"cidr": "203.0.113.0/24"}]}
					}
				}
			},
			"targetSpec": {
				"name": "test-kube",
				"version": "1.34"
			}
		}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "get", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Cmp(out, td.All(
		td.Contains("test-kube"),
		td.Contains("**Deployed version**:   1.34.1-2"),
		td.Contains("203.0.113.0/24"),
	))
}

//
// CREATION TESTS
//

// TestManagedKubernetesCreateCmd tests that a cluster is created from the mandatory flags only.
// TestManagedKubernetesCreateCmd tests that the CNI is derived from the plan unless
// --cni-type is given, so an unforeseen combination stays reachable.
func (ms *MockSuite) TestManagedKubernetesCreateCmd(assert, require *td.T) {
	for _, tc := range []struct {
		plan, wantCni string
		extraArgs     []string
	}{
		{plan: "FREE", wantCni: "CANAL"},
		{plan: "STANDARD", wantCni: "CILIUM"},
		{plan: "FREE", wantCni: "CILIUM", extraArgs: []string{"--cni-type", "CILIUM"}},
	} {
		httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
			tdhttpmock.JSONBody(td.JSON(`{
				"targetSpec": {
					"cni": {"type": $1},
					"name": "test-kube",
					"network": {"type": "PUBLIC"},
					"plan": $2,
					"region": "GRA11",
					"version": "1.34"
				}
			}`, tc.wantCni, tc.plan)),
			httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

		out, err := cmd.Execute(append([]string{
			"cloud", "managed-kubernetes", "create",
			"--cloud-project", "fakeProjectID",
			"--name", "test-kube",
			"--region", "GRA11",
			"--plan", tc.plan,
			"--version", "1.34",
		}, tc.extraArgs...)...)

		require.CmpNoError(err, "plan %s, args %v", tc.plan, tc.extraArgs)
		assert.String(out, `✅ Cluster created successfully (id: kube-12345)`)
	}
}

// TestManagedKubernetesCreateCmdUnknownPlanKeepsCniUnset tests that a plan the CLI does
// not know about is sent as-is, letting the API report the real problem.
func (ms *MockSuite) TestManagedKubernetesCreateCmdUnknownPlanKeepsCniUnset(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"name": "test-kube",
				"network": {"type": "PUBLIC"},
				"plan": "PREMIUM",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(400, `{"message": "invalid plan"}`).Once())

	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "PREMIUM",
	)

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("invalid plan"))
}

// TestManagedKubernetesCreateCmdMissingMandatoryFields tests that the mandatory parts of the
// target specification are checked before calling the API.
func (ms *MockSuite) TestManagedKubernetesCreateCmdMissingMandatoryFields(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
	)

	require.CmpError(err)
	assert.Contains(err.Error(), `required flag(s) "region", "plan" not set`)
}

// TestManagedKubernetesCreateVrackCluster tests the creation of a cluster attached to a vRack.
func (ms *MockSuite) TestManagedKubernetesCreateVrackCluster(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {"type": "CANAL"},
				"name": "test-kube",
				"network": {
					"type": "VRACK",
					"vrack": {
						"privateNetworkId": "pn-1",
						"subnetId": "sn-1",
						"loadBalancerSubnetId": "sn-2",
						"defaultEgress": {
							"type": "USER_GATEWAY",
							"userGateway": {"ip": "10.0.0.1"}
						}
					}
				},
				"plan": "FREE",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--network-vrack-private-network-id", "pn-1",
		"--network-vrack-subnet-id", "sn-1",
		"--network-vrack-load-balancer-subnet-id", "sn-2",
		"--network-vrack-default-egress-type", "USER_GATEWAY",
		"--network-vrack-default-egress-user-gateway-ip", "10.0.0.1",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateFromFileOverridingPlan tests that a --plan overriding the file
// also overrides the CNI the file pairs with its own plan.
func (ms *MockSuite) TestManagedKubernetesCreateFromFileOverridingPlan(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {"type": "CILIUM"},
				"name": "test-kube",
				"network": {"type": "PUBLIC"},
				"plan": "STANDARD",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

	file := filepath.Join(assert.TempDir(), "params.json")
	require.CmpNoError(os.WriteFile(file, []byte(`{
		"targetSpec": {
			"cni": {"type": "CANAL"},
			"name": "test-kube",
			"network": {"type": "PUBLIC"},
			"plan": "FREE",
			"region": "GRA11"
		}
	}`), 0o600))

	out, err := cmd.Execute("cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--from-file", file,
		"--plan", "STANDARD")

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateVrackWithoutPrivateNetwork tests that any vRack flag selects the
// VRACK topology, so the API reports the missing private network.
func (ms *MockSuite) TestManagedKubernetesCreateVrackWithoutPrivateNetwork(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSONPointer("/targetSpec/network", map[string]any{
			"type":  "VRACK",
			"vrack": map[string]any{"loadBalancerSubnetId": "sn-2"},
		})),
		httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--network-vrack-load-balancer-subnet-id", "sn-2")

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateWithAPIServerOptions tests that the IP allowlist and the admission
// plugins given as flat flags are sent in their nested API shape.
func (ms *MockSuite) TestManagedKubernetesCreateWithAPIServerOptions(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"apiServer": {
					"access": {
						"authorizedIps": {
							"rules": [{"cidr": "203.0.113.0/24"}, {"cidr": "198.51.100.1/32"}]
						}
					},
					"admissionPlugins": {
						"AlwaysPullImages": {"enabled": true},
						"NodeRestriction": {"enabled": false}
					}
				},
				"cni": {"type": "CANAL"},
				"name": "test-kube",
				"network": {"type": "PUBLIC"},
				"plan": "FREE",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--cni-type", "CANAL",
		"--network-type", "PUBLIC",
		"--api-server-authorized-ips", "203.0.113.0/24,198.51.100.1/32",
		"--api-server-admission-plugins-enabled", "AlwaysPullImages",
		"--api-server-admission-plugins-disabled", "NodeRestriction",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateWithConflictingAdmissionPlugins tests that a plugin cannot be both
// enabled and disabled.
func (ms *MockSuite) TestManagedKubernetesCreateWithConflictingAdmissionPlugins(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--api-server-admission-plugins-enabled", "AlwaysPullImages",
		"--api-server-admission-plugins-disabled", "AlwaysPullImages",
	)

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`admission plugin "AlwaysPullImages" is both enabled and disabled`))
}

// TestManagedKubernetesCreateWithOIDC tests that giving an OIDC flag selects the OIDC
// authentication variant.
func (ms *MockSuite) TestManagedKubernetesCreateWithOIDC(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"authentication": {
					"type": "OIDC",
					"oidc": {
						"clientId": "my-client",
						"issuerUrl": "https://issuer.example.com",
						"signingAlgs": ["RS256"]
					}
				},
				"cni": {"type": "CANAL"},
				"name": "test-kube",
				"network": {"type": "PUBLIC"},
				"plan": "FREE",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-12345"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--cni-type", "CANAL",
		"--network-type", "PUBLIC",
		"--oidc-client-id", "my-client",
		"--oidc-issuer-url", "https://issuer.example.com",
		"--oidc-signing-algs", "RS256",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateCiliumHubbleDisabled tests that creating a kube with Cilium Hubble disabled results in a successful creation.
func (ms *MockSuite) TestManagedKubernetesCreateCiliumHubbleDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {
					"type": "CILIUM",
					"cilium": {
						"hubble": {"enabled": false}
					}
				},
				"name": "test-hubble-kube",
				"network": {"type": "PUBLIC"},
				"plan": "STANDARD",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-99999"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-hubble-kube",
		"--region", "GRA11",
		"--plan", "STANDARD",
		"--cni-type", "CILIUM",
		"--network-type", "PUBLIC",
		"--cilium-hubble-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateCiliumHubbleEnabled tests that creating a kube with both Cilium Hubble and Cilium Hubble UI enabled and all required UI resource flags results in a successful creation.
func (ms *MockSuite) TestManagedKubernetesCreateCiliumHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {
					"type": "CILIUM",
					"cilium": {
						"hubble": {
							"enabled": true,
							"relay": {"enabled": true},
							"ui": {
								"enabled": true,
								"frontendResources": {
									"limits": {"cpu": "10", "memory": "100m"},
									"requests": {"cpu": "10", "memory": "200m"}
								},
								"backendResources": {
									"limits": {"cpu": "10", "memory": "200m"},
									"requests": {"cpu": "10", "memory": "200m"}
								}
							}
						}
					}
				},
				"name": "test-hubble-kube",
				"network": {"type": "PUBLIC"},
				"plan": "STANDARD",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-99999"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-hubble-kube",
		"--region", "GRA11",
		"--plan", "STANDARD",
		"--cni-type", "CILIUM",
		"--network-type", "PUBLIC",
		"--cilium-hubble-enabled=true",
		"--cilium-hubble-relay-enabled=true",
		"--cilium-hubble-ui-enabled=true",
		"--cilium-hubble-ui-frontend-limits-cpu", "10",
		"--cilium-hubble-ui-frontend-limits-memory", "100m",
		"--cilium-hubble-ui-frontend-requests-cpu", "10",
		"--cilium-hubble-ui-frontend-requests-memory", "200m",
		"--cilium-hubble-ui-backend-limits-cpu", "10",
		"--cilium-hubble-ui-backend-limits-memory", "200m",
		"--cilium-hubble-ui-backend-requests-cpu", "10",
		"--cilium-hubble-ui-backend-requests-memory", "200m",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateCiliumClusterMeshWithAllOptions tests that creating a kube with Cilium ClusterMesh enabled and all its options results in a successful creation.
func (ms *MockSuite) TestManagedKubernetesCreateCiliumClusterMeshWithAllOptions(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {
					"type": "CILIUM",
					"cilium": {
						"clusterId": 42,
						"clusterMesh": {
							"enabled": true,
							"apiServer": {
								"serviceType": "NodePort",
								"nodePort": 32000
							}
						}
					}
				},
				"name": "test-mesh-kube",
				"network": {"type": "PUBLIC"},
				"plan": "STANDARD",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-88888"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-mesh-kube",
		"--region", "GRA11",
		"--plan", "STANDARD",
		"--cni-type", "CILIUM",
		"--network-type", "PUBLIC",
		"--cilium-cluster-id", "42",
		"--cilium-cluster-mesh-enabled=true",
		"--cilium-cluster-mesh-apiserver-service-type", "NodePort",
		"--cilium-cluster-mesh-apiserver-node-port", "32000",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestManagedKubernetesCreateWithIPAllocationPolicyCIDRs tests that the pods and services CIDRs
// are sent in the ipAllocationPolicy object.
func (ms *MockSuite) TestManagedKubernetesCreateWithIPAllocationPolicyCIDRs(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint,
		tdhttpmock.JSONBody(td.JSON(`{
			"targetSpec": {
				"cni": {"type": "CANAL"},
				"ipAllocationPolicy": {
					"pods": "10.100.0.0/16",
					"services": "10.200.0.0/16"
				},
				"name": "test-cidr-kube",
				"network": {"type": "PUBLIC"},
				"plan": "FREE",
				"region": "GRA11"
			}
		}`)),
		httpmock.NewStringResponder(200, `{"id": "kube-77777"}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--name", "test-cidr-kube",
		"--region", "GRA11",
		"--plan", "FREE",
		"--cni-type", "CANAL",
		"--network-type", "PUBLIC",
		"--ip-allocation-policy-pods", "10.100.0.0/16",
		"--ip-allocation-policy-services", "10.200.0.0/16",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

//
// EDITION TESTS
//

// TestManagedKubernetesEditCmd tests that editing a cluster preserves its checksum (optimistic
// locking) and drops the fields the update endpoint rejects.
func (ms *MockSuite) TestManagedKubernetesEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"checksum": "abc123",
			"resourceStatus": "READY",
			"currentState": {
				"name": "old-name",
				"version": "1.33"
			},
			"targetSpec": {
				"name": "old-name",
				"version": "1.33.1-1",
				"region": "GRA11",
				"plan": "FREE",
				"cni": {"type": "CANAL"},
				"network": {"type": "PUBLIC"},
				"ipAllocationPolicy": {"pods": "10.100.0.0/16"}
			}
		}`))

	httpmock.RegisterMatcherResponder(http.MethodPut, managedKubernetesV2Endpoint+"/kube-12345",
		tdhttpmock.JSONBody(td.JSON(`{
			"checksum": "abc123",
			"targetSpec": {
				"cni": {},
				"name": "new-name",
				"network": {},
				"version": "1.34.1-2"
			}
		}`)),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("cloud", "managed-kubernetes", "edit", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--name", "new-name",
		"--version", "1.34.1-2")

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}

// TestManagedKubernetesEditCmdReplacesLists tests that a list flag replaces the current list
// instead of being appended to it.
func (ms *MockSuite) TestManagedKubernetesEditCmdReplacesLists(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"checksum": "abc123",
			"targetSpec": {
				"name": "test-kube",
				"authentication": {
					"type": "OIDC",
					"oidc": {"clientId": "my-client", "issuerUrl": "https://issuer.example.com", "signingAlgs": ["RS256"]}
				}
			}
		}`).Once())

	httpmock.RegisterMatcherResponder(http.MethodPut, managedKubernetesV2Endpoint+"/kube-12345",
		tdhttpmock.JSONBody(td.JSONPointer("/targetSpec/authentication/oidc/signingAlgs", []any{"ES256"})),
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "edit", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--oidc-signing-algs", "ES256")

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}

//
// DELETION TESTS
//

func (ms *MockSuite) TestManagedKubernetesDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodDelete, managedKubernetesV2Endpoint+"/kube-12345",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "delete", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Cluster deleted successfully`)
}

//
// KUBECONFIG TESTS
//

func (ms *MockSuite) TestManagedKubernetesConfigGenerateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/kubeconfig",
		httpmock.NewStringResponder(200, `{"content": "apiVersion: v1"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "kubeconfig", "generate", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `apiVersion: v1`)
}

func (ms *MockSuite) TestManagedKubernetesConfigResetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodPost, managedKubernetesV2Endpoint+"/kube-12345/kubeconfig/reset",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "kubeconfig", "reset", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `✅ Kube config reset successfully`)
}

//
// RESTART TESTS
//

func (ms *MockSuite) TestManagedKubernetesRestartCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost, managedKubernetesV2Endpoint+"/kube-12345/restart",
		tdhttpmock.JSONBody(td.JSON(`{"force": true}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "restart", "kube-12345", "--cloud-project", "fakeProjectID", "--force")

	require.CmpNoError(err)
	assert.String(out, `⚡️ Kubernetes cluster restarting…`)
}

//
// TASK AND EVENT TESTS
//

func (ms *MockSuite) TestManagedKubernetesTaskListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/task",
		httpmock.NewStringResponder(200, `[
			{"id": "task-1", "type": "ClusterCreate", "status": "DONE", "createdAt": "2026-09-10T09:00:00Z", "finishedAt": "2026-09-10T09:05:00Z"},
			{"id": "task-2", "type": "ClusterUpdate", "status": "RUNNING", "createdAt": "2026-09-11T09:00:00Z"}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "task", "list", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌────────┬───────────────┬─────────┬──────────────────────┬──────────────────────┐
│   ID   │     type      │ status  │      created at      │     finished at      │
├────────┼───────────────┼─────────┼──────────────────────┼──────────────────────┤
│ task-1 │ ClusterCreate │ DONE    │ 2026-09-10T09:00:00Z │ 2026-09-10T09:05:00Z │
│ task-2 │ ClusterUpdate │ RUNNING │ 2026-09-11T09:00:00Z │                      │
└────────┴───────────────┴─────────┴──────────────────────┴──────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesTaskGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/task/task-1",
		httpmock.NewStringResponder(200, `{"id": "task-1", "type": "ClusterCreate", "status": "DONE"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "task", "get", "kube-12345", "task-1", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id": "task-1", "type": "ClusterCreate", "status": "DONE"}`))
}

func (ms *MockSuite) TestManagedKubernetesEventListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/event?taskId=task-1",
		httpmock.NewStringResponder(200, `[
			{"createdAt": "2026-09-10T09:00:00Z", "type": "TASK_START", "kind": "ClusterCreate", "link": "/v2/publicCloud/project/fakeProjectID/managedKubernetes/kube-12345/task/task-1", "message": "Task started"}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "event", "list", "kube-12345", "--task-id", "task-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌──────────────────────┬────────────┬───────────────┬─────────┬──────────────┐
│      created at      │    type    │     kind      │ task ID │   message    │
├──────────────────────┼────────────┼───────────────┼─────────┼──────────────┤
│ 2026-09-10T09:00:00Z │ TASK_START │ ClusterCreate │ task-1  │ Task started │
└──────────────────────┴────────────┴───────────────┴─────────┴──────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesNodepoolTaskListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/nodepool/np-1/task",
		httpmock.NewStringResponder(200, `[
			{"id": "task-3", "type": "NodePoolCreate", "status": "DONE", "createdAt": "2026-09-10T09:00:00Z", "finishedAt": "2026-09-10T09:05:00Z"}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "task", "list", "kube-12345", "np-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌────────┬────────────────┬────────┬──────────────────────┬──────────────────────┐
│   ID   │      type      │ status │      created at      │     finished at      │
├────────┼────────────────┼────────┼──────────────────────┼──────────────────────┤
│ task-3 │ NodePoolCreate │ DONE   │ 2026-09-10T09:00:00Z │ 2026-09-10T09:05:00Z │
└────────┴────────────────┴────────┴──────────────────────┴──────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesNodepoolTaskGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/nodepool/np-1/task/task-3",
		httpmock.NewStringResponder(200, `{"id": "task-3", "status": "DONE"}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "task", "get", "kube-12345", "np-1", "task-3", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"id": "task-3", "status": "DONE"}`))
}

func (ms *MockSuite) TestManagedKubernetesNodepoolEventListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/nodepool/np-1/event",
		httpmock.NewStringResponder(200, `[
			{"createdAt": "2026-09-10T09:00:00Z", "type": "TASK_SUCCESS", "kind": "NodePoolCreate", "message": "Task succeeded"}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "nodepool", "event", "list", "kube-12345", "np-1", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌──────────────────────┬──────────────┬────────────────┬─────────┬────────────────┐
│      created at      │     type     │      kind      │ task ID │    message     │
├──────────────────────┼──────────────┼────────────────┼─────────┼────────────────┤
│ 2026-09-10T09:00:00Z │ TASK_SUCCESS │ NodePoolCreate │         │ Task succeeded │
└──────────────────────┴──────────────┴────────────────┴─────────┴────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

//
// USAGE AND CAPABILITIES TESTS
//

func (ms *MockSuite) TestManagedKubernetesUsageEtcdCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/usage/etcd",
		httpmock.NewStringResponder(200, `{"used": 1024, "limit": 4096}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "usage", "etcd", "kube-12345", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{"used": 1024, "limit": 4096}`))
}

func (ms *MockSuite) TestManagedKubernetesUsageEtcdTextCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/usage/etcd",
		httpmock.NewStringResponder(200, `{"used": 1306624, "limit": 8589934592}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "usage", "etcd", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "1.25 MiB / 8 GiB")
}

func (ms *MockSuite) TestManagedKubernetesVersionCapabilitiesCmd(assert, require *td.T) {
	body := `{"currentVersion": "1.32.8-1", "availableUpdates": [{"version": "1.32.8-2", "status": "SUPPORTED"}], "availableUpgrades": [{"version": "1.33.1-2", "status": "SUPPORTED"}]}`
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/capabilities/version",
		httpmock.NewStringResponder(200, body).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "version", "capabilities", "kube-12345", "--cloud-project", "fakeProjectID", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(body))
}

func (ms *MockSuite) TestManagedKubernetesVersionCapabilitiesTextCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, managedKubernetesV2Endpoint+"/kube-12345/capabilities/version",
		httpmock.NewStringResponder(200, `{"currentVersion": "1.32.8-1", "availableUpdates": [], "availableUpgrades": [{"version": "1.33.1-2", "status": "SUPPORTED"}]}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "version", "capabilities", "kube-12345", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.Contains(out, "1.32.8-1")
	assert.Contains(out, "None")
	assert.Contains(out, "1.33.1-2 (SUPPORTED)")
}

func (ms *MockSuite) TestManagedKubernetesFlavorListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/managedKubernetes/flavor?region=EU-WEST-PAR",
		httpmock.NewStringResponder(200, `[
			{"name": "b3-8", "category": "GENERAL_PURPOSE", "status": "AVAILABLE", "resources": {"vcpus": 2, "memoryGiB": 8, "gpus": 0}}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "flavor", "list", "--region", "EU-WEST-PAR", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌──────┬─────────────────┬───────┬──────────────┬──────┬───────────┐
│ name │    category     │ vCPUs │ memory (GiB) │ GPUs │  status   │
├──────┼─────────────────┼───────┼──────────────┼──────┼───────────┤
│ b3-8 │ GENERAL_PURPOSE │ 2     │ 8            │ 0    │ AVAILABLE │
└──────┴─────────────────┴───────┴──────────────┴──────┴───────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

//
// REFERENCE TESTS
//

func (ms *MockSuite) TestManagedKubernetesPlanListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/managedKubernetes/plan?region=EU-WEST-PAR",
		httpmock.NewStringResponder(200, `[
			{"name": "FREE", "status": "AVAILABLE", "limits": {"maxNodes": 100}, "controlPlane": {"etcd": {"dedicated": false}}},
			{"name": "STANDARD", "status": "AVAILABLE", "limits": {"maxNodes": 500}, "controlPlane": {"etcd": {"dedicated": true}}}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "plan", "list", "--region", "EU-WEST-PAR", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌──────────┬───────────┬───────────┬────────────────┐
│   name   │  status   │ max nodes │ dedicated etcd │
├──────────┼───────────┼───────────┼────────────────┤
│ FREE     │ AVAILABLE │ 100       │ false          │
│ STANDARD │ AVAILABLE │ 500       │ true           │
└──────────┴───────────┴───────────┴────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesRegionListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/managedKubernetes/region",
		httpmock.NewStringResponder(200, `[
			{"name": "EU-WEST-PAR", "availablePlans": ["FREE", "STANDARD"], "availabilityZones": ["eu-west-par-a", "eu-west-par-b"]},
			{"name": "GRA11", "availablePlans": ["FREE"], "availabilityZones": []}
		]`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "region", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌─────────────┬─────────────────┬──────────────────────────────┐
│    name     │ available plans │      availability zones      │
├─────────────┼─────────────────┼──────────────────────────────┤
│ EU-WEST-PAR │ FREE, STANDARD  │ eu-west-par-a, eu-west-par-b │
│ GRA11       │ FREE            │                              │
└─────────────┴─────────────────┴──────────────────────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestManagedKubernetesVersionListCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/managedKubernetes/version",
		httpmock.NewStringResponder(200, `{
			"defaultVersion": "1.33",
			"availableVersions": [{"version": "1.34", "status": "SUPPORTED"}, {"version": "1.33", "status": "SUPPORTED"}, {"version": "1.31", "status": "END_OF_SUPPORT"}]
		}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "version", "list", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(cleanWhitespacesHelper(out), `
┌─────────┬────────────────┬─────────┐
│ version │     status     │ default │
├─────────┼────────────────┼─────────┤
│ 1.34    │ SUPPORTED      │ false   │
│ 1.33    │ SUPPORTED      │ true    │
│ 1.31    │ END_OF_SUPPORT │ false   │
└─────────┴────────────────┴─────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

// TestManagedKubernetesCreationExamples tests that the cluster and nodepool examples of the API schema are offered.
func (ms *MockSuite) TestManagedKubernetesCreationExamples(assert, require *td.T) {
	examples, err := openapi.GetOperationRequestExamples(assets.CloudV2OpenapiSchema, cloud.ManagedKubernetesV2SchemaPath, "post", "", nil)
	require.CmpNoError(err)
	assert.Cmp(examples, td.SuperMapOf(map[string]string{}, td.MapEntries{
		"Create a FREE cluster on the public network":    td.Ignore(),
		"Create a FREE cluster on a private network":     td.Ignore(),
		"Create a STANDARD cluster on a private network": td.Ignore(),
	}))

	examples, err = openapi.GetOperationRequestExamples(assets.CloudV2OpenapiSchema, cloud.ManagedKubernetesV2NodepoolsSchemaPath, "post", "", nil)
	require.CmpNoError(err)
	assert.Cmp(examples, td.SuperMapOf(map[string]string{}, td.MapEntries{
		"Create a fixed-size nodepool":                        td.Ignore(),
		"Create an autoscaled nodepool":                       td.Ignore(),
		"Create a nodepool in an availability zone":           td.Ignore(),
		"Create a monthly-billed nodepool with anti-affinity": td.Ignore(),
		"Create a dedicated nodepool with labels and taints":  td.Ignore(),
	}))
}
