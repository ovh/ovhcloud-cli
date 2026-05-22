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

//
// LIST TESTS
//

// TestCloudKubeListCmd tests that listing kubes returns the expected output.
func (ms *MockSuite) TestCloudKubeListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		httpmock.NewStringResponder(200, `["kube-12345"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"name": "test-kube",
			"region": "GRA11",
			"plan": "free",
			"version": "1.21.5",
			"status": "INSTALLING",
			"createdAt": "2021-10-12T14:23:45+00:00"
		}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬───────────┬────────┬──────┬─────────┬────────────┐
│     id     │   name    │ region │ plan │ version │   status   │
├────────────┼───────────┼────────┼──────┼─────────┼────────────┤
│ kube-12345 │ test-kube │ GRA11  │ free │ 1.21.5  │ INSTALLING │
└────────────┴───────────┴────────┴──────┴─────────┴────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

//
// CREATION CLUSTER WITH CILIUM HUBBLE CUSTOMIZATION TESTS
//

// TestCloudKubeCreateCiliumHubbleDisabled tests that creating a kube with Cilium Hubble disabled results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": false
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-hubble-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestCloudKubeCreateCiliumHubbleEnabled tests that creating a kube with both Cilium Hubble and Cilium Hubble UI enabled and all required UI resource flags results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": true,
						"relay": {
     						"enabled": true,
						},
						"ui": {
							"enabled": true,
							"frontendResources": {
								"limits": {
									"cpu": "10",
									"memory": "100m"
								},
								"requests": {
									"cpu": "10",
									"memory": "200m"
								}
							},
							"backendResources": {
								"limits": {
									"cpu": "10",
									"memory": "200m"
								},
								"requests": {
									"cpu": "10",
									"memory": "200m"
								}
							}
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-hubble-ui-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-hubble-enabled",
		"--cilium-hubble-ui-enabled",
		"--cilium-hubble-ui-frontend-limits-cpu=10",
		"--cilium-hubble-ui-frontend-limits-memory=100m",
		"--cilium-hubble-ui-frontend-requests-cpu=10",
		"--cilium-hubble-ui-frontend-requests-memory=200m",
		"--cilium-hubble-ui-backend-limits-cpu=10",
		"--cilium-hubble-ui-backend-limits-memory=200m",
		"--cilium-hubble-ui-backend-requests-cpu=10",
		"--cilium-hubble-ui-backend-requests-memory=200m",
		"--cilium-hubble-relay-enabled",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

//
// CREATION CLUSTER WITH CILIUM CLUSTERMESH CUSTOMIZATION TESTS
//

// TestCloudKubeCreateCiliumClusterMeshWithAllOptions tests that creating a kube with all ClusterMesh options set results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterMeshWithAllOptions(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterId": 123,
					"clusterMesh": {
						"enabled": true,
						"apiServer": {
							"serviceType": "NodePort",
							"nodePort": 30000
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-clustermesh-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-id=123",
		"--cilium-cluster-mesh-enabled",
		"--cilium-cluster-mesh-apiserver-service-type=NodePort",
		"--cilium-cluster-mesh-apiserver-node-port=30000",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// TestCloudKubeCreateCiliumClusterMeshDisabled tests that creating a kube with --cilium-cluster-mesh-enabled=false results in a successful creation with clusterMesh disabled.
func (ms *MockSuite) TestCloudKubeCreateCiliumClusterMeshDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterMesh": {
						"enabled": false
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-clustermesh-disabled-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--cilium-cluster-mesh-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

// CREATION CLUSTER WITH IP ALLOCATION POLICY TESTS

// TestCloudKubeCreateWithIPAllocationPolicyCIDRs tests that creating a kube with both CIDR flags set results in a successful creation.
func (ms *MockSuite) TestCloudKubeCreateWithIPAllocationPolicyCIDRs(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"ipAllocationPolicy": {
				"podsIpv4Cidr": "10.0.0.0/16",
				"servicesIpv4Cidr": "10.1.0.0/16"
			}
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "kube-99999",
			"name": "test-cidr-kube"
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "create",
		"--cloud-project", "fakeProjectID",
		"--region", "GRA999",
		"--ip-allocation-policy-pods-ipv4-cidr=10.0.0.0/16",
		"--ip-allocation-policy-services-ipv4-cidr=10.1.0.0/16",
	)

	require.CmpNoError(err)
	assert.Contains(out, "created successfully")
}

//
// RESET CLUSTER TESTS
//

// TestCloudKubeResetCmd tests that resetting a kube with basic flags results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"version": "1.32",
			"workerNodesPolicy": "reinstall"
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--version", "1.32",
		"--worker-nodes-policy", "reinstall",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetCmdMissingClusterID tests that resetting a kube without a cluster_id argument results in an error.
func (ms *MockSuite) TestCloudKubeResetCmdMissingClusterID(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpError(err)
}

// TestCloudKubeResetWithIPAllocationPolicyCIDRs tests that resetting a kube with both CIDR flags set results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetWithIPAllocationPolicyCIDRs(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"ipAllocationPolicy": {
				"podsIpv4Cidr": "10.0.0.0/16",
				"servicesIpv4Cidr": "10.1.0.0/16"
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--ip-allocation-policy-pods-ipv4-cidr=10.0.0.0/16",
		"--ip-allocation-policy-services-ipv4-cidr=10.1.0.0/16",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetCiliumHubbleDisabled tests that resetting a kube with Cilium Hubble disable results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetCiliumHubbleDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": false
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-hubble-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetCiliumHubbleEnabled tests that resetting a kube with Hubble, Relay, UI and all resource flags results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetCiliumHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"hubble": {
						"enabled": true,
						"relay": {
							"enabled": true
						},
						"ui": {
							"enabled": true,
							"frontendResources": {
								"limits": {
									"cpu": "500m",
									"memory": "256Mi"
								},
								"requests": {
									"cpu": "100m",
									"memory": "128Mi"
								}
							},
							"backendResources": {
								"limits": {
									"cpu": "500m",
									"memory": "256Mi"
								},
								"requests": {
									"cpu": "100m",
									"memory": "128Mi"
								}
							}
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-hubble-enabled",
		"--cilium-hubble-relay-enabled",
		"--cilium-hubble-ui-enabled",
		"--cilium-hubble-ui-frontend-limits-cpu=500m",
		"--cilium-hubble-ui-frontend-limits-memory=256Mi",
		"--cilium-hubble-ui-frontend-requests-cpu=100m",
		"--cilium-hubble-ui-frontend-requests-memory=128Mi",
		"--cilium-hubble-ui-backend-limits-cpu=500m",
		"--cilium-hubble-ui-backend-limits-memory=256Mi",
		"--cilium-hubble-ui-backend-requests-cpu=100m",
		"--cilium-hubble-ui-backend-requests-memory=128Mi",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetCiliumClusterMeshDisabled tests that resetting a kube with --cilium-cluster-mesh-enabled=false results in a successful reset with clusterMesh disabled.
func (ms *MockSuite) TestCloudKubeResetCiliumClusterMeshDisabled(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterMesh": {
						"enabled": false
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-cluster-mesh-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetCiliumClusterMeshWithAllOptions tests that resetting a kube with all ClusterMesh options set results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetCiliumClusterMeshWithAllOptions(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"customization": {
				"cilium": {
					"clusterId": 42,
					"clusterMesh": {
						"enabled": true,
						"apiServer": {
							"serviceType": "LoadBalancer",
							"nodePort": 31000
						}
					}
				}
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-cluster-id=42",
		"--cilium-cluster-mesh-enabled",
		"--cilium-cluster-mesh-apiserver-service-type=LoadBalancer",
		"--cilium-cluster-mesh-apiserver-node-port=31000",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

// TestCloudKubeResetWithPrivateNetworkConfig tests that resetting a kube with private network configuration results in a successful reset.
func (ms *MockSuite) TestCloudKubeResetWithPrivateNetworkConfig(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(
		http.MethodPost,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/reset",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"privateNetworkConfiguration": {
				"defaultVrackGateway": "10.0.0.1",
				"privateNetworkRoutingAsDefault": true
			}
		}`)),
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "reset", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--private-network.default-vrack-gateway", "10.0.0.1",
		"--private-network.routing-as-default",
	)

	require.CmpNoError(err)
	assert.Contains(out, "Kubernetes cluster is being reset")
}

//
// CUSTOMIZATION SUBCOMMAND TESTS
//

// TestCloudKubeCustomizationGetCmdMissingClusterID tests that getting a customization without a cluster_id argument results in an error.
func (ms *MockSuite) TestCloudKubeCustomizationGetCmdMissingClusterID(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "customization", "get",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpError(err)
}

// TestCloudKubeCustomizationEditCmdMissingClusterID tests that editing a customization without a cluster_id argument results in an error.
func (ms *MockSuite) TestCloudKubeCustomizationEditCmdMissingClusterID(assert, require *td.T) {
	_, err := cmd.Execute(
		"cloud", "managed-kubernetes", "customization", "edit",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpError(err)
}

// TestCloudKubeCustomizationGetCmd tests that getting a cluster customization returns the expected output.
func (ms *MockSuite) TestCloudKubeCustomizationGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/customization",
		httpmock.NewStringResponder(200, `{
			"apiServer": {
				"admissionPlugins": {
					"enabled": ["NodeRestriction"],
					"disabled": ["AlwaysPullImages"]
				}
			},
			"kubeProxy": {
				"iptables": {
					"minSyncPeriod": "PT30S",
					"syncPeriod": "PT60S"
				},
				"ipvs": {
					"minSyncPeriod": "PT30S",
					"syncPeriod": "PT60S",
					"scheduler": "rr",
					"tcpFinTimeout": "PT60S",
					"tcpTimeout": "PT120S",
					"udpTimeout": "PT60S"
				}
			}
		}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "customization", "get", "kube-12345",
		"--cloud-project", "fakeProjectID",
	)

	require.CmpNoError(err)
	assert.Contains(out, "NodeRestriction")
	assert.Contains(out, "AlwaysPullImages")
}

// TestCloudKubeCustomizationEditCiliumHubbleEnabled tests that editing customization with Cilium Hubble enabled succeeds.
func (ms *MockSuite) TestCloudKubeCustomizationEditCiliumHubbleEnabled(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/customization",
		httpmock.NewStringResponder(200, `{
			"apiServer": {"admissionPlugins": {"enabled": [], "disabled": []}},
			"kubeProxy": {"iptables": {}, "ipvs": {}}
		}`).Once())

	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/customization",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "customization", "edit", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-hubble-enabled",
	)

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}

// TestCloudKubeCustomizationEditCiliumClusterMeshDisabled tests that editing customization with ClusterMesh explicitly disabled succeeds.
func (ms *MockSuite) TestCloudKubeCustomizationEditCiliumClusterMeshDisabled(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/customization",
		httpmock.NewStringResponder(200, `{
			"apiServer": {"admissionPlugins": {"enabled": [], "disabled": []}},
			"kubeProxy": {"iptables": {}, "ipvs": {}}
		}`).Once())

	httpmock.RegisterResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/customization",
		httpmock.NewStringResponder(200, `{}`).Once())

	out, err := cmd.Execute(
		"cloud", "managed-kubernetes", "customization", "edit", "kube-12345",
		"--cloud-project", "fakeProjectID",
		"--cilium-cluster-mesh-enabled=false",
	)

	require.CmpNoError(err)
	assert.Contains(out, "updated successfully")
}
