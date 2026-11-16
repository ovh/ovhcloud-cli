// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"dario.cat/mergo"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/ovh/ovhcloud-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Paths of the API v2 cluster operations in the OpenAPI schema.
const (
	ManagedKubernetesV2SchemaPath          = "/publicCloud/project/{projectId}/managedKubernetes"
	ManagedKubernetesV2ClusterSchemaPath   = ManagedKubernetesV2SchemaPath + "/{managedKubernetesId}"
	ManagedKubernetesV2NodepoolsSchemaPath = ManagedKubernetesV2ClusterSchemaPath + "/nodepool"
	ManagedKubernetesV2NodepoolSchemaPath  = ManagedKubernetesV2NodepoolsSchemaPath + "/{nodePoolId}"
)

var (
	managedKubernetesColumnsToDisplay = []string{
		"id ID",
		"currentState.name name",
		"currentState.region region",
		"currentState.plan plan",
		"currentState.version version",
		"resourceStatus status",
		"createdAt created at",
	}

	managedKubernetesNodepoolColumnsToDisplay = []string{
		"id ID",
		"currentState.name name",
		"currentState.nodeConfig.flavor flavor",
		"currentState.desiredNodes desired nodes",
		"currentState.status.currentNodes current nodes",
		"currentState.status.upToDateNodes up-to-date nodes",
		"resourceStatus status",
		"createdAt created at",
	}

	managedKubernetesNodeColumnsToDisplay = []string{
		"id ID",
		"currentState.name name",
		"currentState.flavor flavor",
		"currentState.version version",
		"currentState.upToDate up-to-date",
		"resourceStatus status",
		"currentState.deployedAt deployed at",
	}

	managedKubernetesTaskColumnsToDisplay = []string{
		"id ID",
		"type",
		"status",
		"createdAt created at",
		"finishedAt finished at",
	}

	managedKubernetesEventColumnsToDisplay = []string{"createdAt created at", "type", "kind", "taskId task ID", "message"}

	managedKubernetesPlanColumnsToDisplay = []string{
		"name",
		"status",
		"limits.maxNodes max nodes",
		"controlPlane.etcd.dedicated dedicated etcd",
	}

	managedKubernetesRegionColumnsToDisplay = []string{
		"name",
		"availablePlans available plans",
		"availabilityZones availability zones",
	}

	managedKubernetesFlavorColumnsToDisplay = []string{
		"name",
		"category",
		"resources.vcpus vCPUs",
		"resources.memoryGiB memory (GiB)",
		"resources.gpus GPUs",
		"status",
	}

	managedKubernetesVersionColumnsToDisplay = []string{"version", "status", "default"}

	// managedKubernetesPlanCNI maps a plan to the CNI it imposes: the API requires CANAL on
	// the FREE plan and CILIUM on the STANDARD one, and rejects any other combination.
	managedKubernetesPlanCNI = map[string]string{
		"FREE":     "CANAL",
		"STANDARD": "CILIUM",
	}

	//go:embed templates/cloud_managed_kubernetes.tmpl
	managedKubernetesTemplate string

	//go:embed templates/cloud_managed_kubernetes_node.tmpl
	managedKubernetesNodeTemplate string

	//go:embed templates/cloud_managed_kubernetes_nodepool.tmpl
	managedKubernetesNodepoolTemplate string

	//go:embed templates/cloud_managed_kubernetes_task.tmpl
	managedKubernetesTaskTemplate string

	//go:embed templates/cloud_managed_kubernetes_etcd_usage.tmpl
	managedKubernetesEtcdUsageTemplate string

	//go:embed templates/cloud_managed_kubernetes_version_capabilities.tmpl
	managedKubernetesVersionCapabilitiesTemplate string

	// ManagedKubernetesSpec is the cluster payload sent to the API v2 create and update endpoints.
	ManagedKubernetesSpec ManagedKubernetesClusterSpec

	// ManagedKubernetesNodepoolSpec is the node pool payload sent to the API v2 create and update endpoints.
	ManagedKubernetesNodepoolSpec struct {
		TargetSpec ManagedKubernetesNodepoolTargetSpec `json:"targetSpec,omitzero"`
	}

	// ManagedKubernetesForceAction indicates whether to force an action
	// It is set by a command line flag
	ManagedKubernetesForceAction bool

	// Filters of the node listing, set by command line flags.
	ManagedKubernetesNodeIncludeDeleted bool
	ManagedKubernetesNodeResourceStatus string

	// ManagedKubernetesEventTaskID restricts the event listing to the given task.
	ManagedKubernetesEventTaskID string

	// ManagedKubernetesReferenceRegion is the region to list the plans and flavors for.
	ManagedKubernetesReferenceRegion string
)

type (
	// ManagedKubernetesClusterSpec is the cluster payload of the API v2 create and update endpoints.
	ManagedKubernetesClusterSpec struct {
		TargetSpec ManagedKubernetesTargetSpec `json:"targetSpec,omitzero"`
	}

	// ManagedKubernetesFlavor is a node pool flavor of the MKS reference catalog.
	ManagedKubernetesFlavor struct {
		Name      string `json:"name"`
		Status    string `json:"status"`
		Resources struct {
			Vcpus     int `json:"vcpus"`
			MemoryGiB int `json:"memoryGiB"`
			Gpus      int `json:"gpus"`
		} `json:"resources"`
	}

	// ManagedKubernetesVersionCatalog is the version catalog of the MKS reference.
	// Versions stay maps so they can be flagged, filtered and rendered as is.
	ManagedKubernetesVersionCatalog struct {
		AvailableVersions []map[string]any `json:"availableVersions"`
		DefaultVersion    string           `json:"defaultVersion"`
	}

	// ManagedKubernetesTargetSpec is the desired state of a cluster. Fields not accepted on update
	// (region, plan, network and CNI variants, ipAllocationPolicy) are pruned from the
	// body by the OpenAPI filter, so the same structure serves create and edit.
	ManagedKubernetesTargetSpec struct {
		ApiServer          ManagedKubernetesAPIServer          `json:"apiServer,omitzero"`
		Authentication     ManagedKubernetesAuthentication     `json:"authentication,omitzero"`
		AutoUpgrade        ManagedKubernetesAutoUpgrade        `json:"autoUpgrade,omitzero"`
		Cni                ManagedKubernetesCni                `json:"cni,omitzero"`
		IPAllocationPolicy ManagedKubernetesIPAllocationPolicy `json:"ipAllocationPolicy,omitzero"`
		Name               string                              `json:"name,omitempty"`
		Network            ManagedKubernetesNetwork            `json:"network,omitzero"`
		Plan               string                              `json:"plan,omitempty"`
		Region             string                              `json:"region,omitempty"`
		Version            string                              `json:"version,omitempty"`
	}

	// ManagedKubernetesAPIServer configures the cluster API server.
	ManagedKubernetesAPIServer struct {
		Access           ManagedKubernetesAPIServerAccess            `json:"access,omitzero"`
		AdmissionPlugins map[string]ManagedKubernetesAdmissionPlugin `json:"admissionPlugins,omitempty"`
	}

	// ManagedKubernetesAPIServerAccess restricts the network access to the API server.
	ManagedKubernetesAPIServerAccess struct {
		AuthorizedIps ManagedKubernetesAuthorizedIps `json:"authorizedIps,omitzero"`
	}

	// ManagedKubernetesAuthorizedIps lists the rules allowed to reach the API server.
	ManagedKubernetesAuthorizedIps struct {
		Rules []ManagedKubernetesAuthorizedIpRule `json:"rules,omitempty"`
	}

	// ManagedKubernetesAuthorizedIpRule allows a CIDR to reach the API server.
	ManagedKubernetesAuthorizedIpRule struct {
		Cidr string `json:"cidr"`
	}

	// ManagedKubernetesAdmissionPlugin toggles an API server admission plugin.
	ManagedKubernetesAdmissionPlugin struct {
		Enabled bool `json:"enabled"`
	}

	// ManagedKubernetesAuthentication configures how users authenticate against the cluster.
	ManagedKubernetesAuthentication struct {
		Oidc ManagedKubernetesOidc `json:"oidc,omitzero"`
		Type string                `json:"type,omitempty"`
	}

	// ManagedKubernetesOidc configures the OIDC provider used to authenticate users.
	ManagedKubernetesOidc struct {
		CaContent      string   `json:"caContent,omitempty"`
		ClientId       string   `json:"clientId,omitempty"`
		GroupsClaim    []string `json:"groupsClaim,omitempty"`
		GroupsPrefix   string   `json:"groupsPrefix,omitempty"`
		IssuerUrl      string   `json:"issuerUrl,omitempty"`
		RequiredClaim  []string `json:"requiredClaim,omitempty"`
		SigningAlgs    []string `json:"signingAlgs,omitempty"`
		UsernameClaim  string   `json:"usernameClaim,omitempty"`
		UsernamePrefix string   `json:"usernamePrefix,omitempty"`
	}

	// ManagedKubernetesAutoUpgrade configures the automatic upgrades of the cluster.
	ManagedKubernetesAutoUpgrade struct {
		Patch ManagedKubernetesAutoUpgradeTrack `json:"patch,omitzero"`
	}

	// ManagedKubernetesAutoUpgradeTrack toggles the automatic upgrades of a version track.
	ManagedKubernetesAutoUpgradeTrack struct {
		Enabled *bool `json:"enabled,omitempty"`
	}

	// ManagedKubernetesCni configures the cluster CNI.
	ManagedKubernetesCni struct {
		Canal  ManagedKubernetesCniCanal `json:"canal,omitzero"`
		Cilium Cilium                    `json:"cilium,omitzero"`
		Type   string                    `json:"type,omitempty"`
	}

	// ManagedKubernetesCniCanal configures the Canal CNI.
	ManagedKubernetesCniCanal struct {
		KubeProxy KubeProxy `json:"kubeProxy,omitzero"`
	}

	// KubeProxy configures kube-proxy and its proxy mode.
	KubeProxy struct {
		IPTables KubeProxyIPTables `json:"iptables,omitzero"`
		IPVS     KubeProxyIPVS     `json:"ipvs,omitzero"`
		Mode     string            `json:"mode,omitempty"`
	}

	// KubeProxyIPTables configures kube-proxy in iptables mode.
	KubeProxyIPTables struct {
		MinSyncPeriod string `json:"minSyncPeriod,omitempty"`
		SyncPeriod    string `json:"syncPeriod,omitempty"`
	}

	// KubeProxyIPVS configures kube-proxy in IPVS mode.
	KubeProxyIPVS struct {
		MinSyncPeriod string `json:"minSyncPeriod,omitempty"`
		Scheduler     string `json:"scheduler,omitempty"`
		SyncPeriod    string `json:"syncPeriod,omitempty"`
		TCPFinTimeout string `json:"tcpFinTimeout,omitempty"`
		TCPTimeout    string `json:"tcpTimeout,omitempty"`
		UDPTimeout    string `json:"udpTimeout,omitempty"`
	}

	// ManagedKubernetesIPAllocationPolicy sets the CIDRs of the pods and services.
	ManagedKubernetesIPAllocationPolicy struct {
		Pods     string `json:"pods,omitempty"`
		Services string `json:"services,omitempty"`
	}

	// ManagedKubernetesNetwork configures the network the cluster is attached to.
	ManagedKubernetesNetwork struct {
		Type  string                        `json:"type,omitempty"`
		Vrack ManagedKubernetesNetworkVrack `json:"vrack,omitzero"`
	}

	// ManagedKubernetesNetworkVrack attaches the cluster to a vRack private network.
	ManagedKubernetesNetworkVrack struct {
		DefaultEgress        ManagedKubernetesDefaultEgress `json:"defaultEgress,omitzero"`
		LoadBalancerSubnetId string                         `json:"loadBalancerSubnetId,omitempty"`
		PrivateNetworkId     string                         `json:"privateNetworkId,omitempty"`
		SubnetId             string                         `json:"subnetId,omitempty"`
	}

	// ManagedKubernetesDefaultEgress configures the default egress of the vRack network.
	ManagedKubernetesDefaultEgress struct {
		Type        string                       `json:"type,omitempty"`
		UserGateway ManagedKubernetesUserGateway `json:"userGateway,omitzero"`
	}

	// ManagedKubernetesUserGateway is the user-managed gateway used as default egress.
	ManagedKubernetesUserGateway struct {
		IP string `json:"ip,omitempty"`
	}

	// Cilium configures the Cilium CNI.
	Cilium struct {
		ClusterID   *uint8      `json:"clusterId,omitempty"`
		ClusterMesh ClusterMesh `json:"clusterMesh,omitzero"`
		Hubble      Hubble      `json:"hubble,omitzero"`
	}

	// ClusterMesh configures the Cilium Cluster Mesh.
	ClusterMesh struct {
		APIServer ClusterMeshAPIServer `json:"apiServer,omitzero"`
		Enabled   *bool                `json:"enabled,omitempty"`
	}

	// ClusterMeshAPIServer configures how the Cluster Mesh API server is exposed.
	ClusterMeshAPIServer struct {
		NodePort    uint16 `json:"nodePort,omitempty"`
		ServiceType string `json:"serviceType,omitempty"`
	}

	// Hubble configures the Hubble observability layer of Cilium.
	Hubble struct {
		Enabled *bool       `json:"enabled,omitempty"`
		Relay   HubbleRelay `json:"relay,omitzero"`
		UI      HubbleUI    `json:"ui,omitzero"`
	}

	// HubbleRelay toggles the Hubble relay.
	HubbleRelay struct {
		Enabled *bool `json:"enabled,omitempty"`
	}

	// HubbleUI configures the Hubble UI.
	HubbleUI struct {
		BackendResources  ResourceRequirements `json:"backendResources,omitzero"`
		Enabled           *bool                `json:"enabled,omitempty"`
		FrontendResources ResourceRequirements `json:"frontendResources,omitzero"`
	}

	// ResourceRequirements sets the resource limits and requests of a component.
	ResourceRequirements struct {
		Limits   ResourceList `json:"limits,omitzero"`
		Requests ResourceList `json:"requests,omitzero"`
	}

	// ResourceList is a CPU and memory resource quantity.
	ResourceList struct {
		Cpu    string `json:"cpu,omitempty"`
		Memory string `json:"memory,omitempty"`
	}

	// ManagedKubernetesNodepoolTargetSpec is the desired state of a node pool. Fields not accepted
	// on update are pruned from the body by the OpenAPI filter, so it serves create and edit.
	ManagedKubernetesNodepoolTargetSpec struct {
		AntiAffinity     ManagedKubernetesAntiAffinity `json:"antiAffinity,omitzero"`
		Autoscaler       ManagedKubernetesAutoscaler   `json:"autoscaler,omitzero"`
		AvailabilityZone string                        `json:"availabilityZone,omitempty"`
		BillingType      string                        `json:"billingType,omitempty"`
		DesiredNodes     *int                          `json:"desiredNodes,omitempty"`
		Name             string                        `json:"name,omitempty"`
		NodeConfig       ManagedKubernetesNodeConfig   `json:"nodeConfig,omitzero"`
		Template         ManagedKubernetesNodeTemplate `json:"template,omitzero"`
	}

	// ManagedKubernetesAntiAffinity spreads the node pool nodes across hypervisors.
	ManagedKubernetesAntiAffinity struct {
		Enabled *bool `json:"enabled,omitempty"`
	}

	// ManagedKubernetesAutoscaler configures the node pool autoscaling.
	ManagedKubernetesAutoscaler struct {
		Enabled   *bool                                `json:"enabled,omitempty"`
		MaxNodes  *int                                 `json:"maxNodes,omitempty"`
		MinNodes  *int                                 `json:"minNodes,omitempty"`
		ScaleDown ManagedKubernetesAutoscalerScaleDown `json:"scaleDown,omitzero"`
	}

	// ManagedKubernetesAutoscalerScaleDown tunes when the autoscaler removes nodes.
	ManagedKubernetesAutoscalerScaleDown struct {
		UnneededTimeSeconds  int     `json:"unneededTimeSeconds,omitempty"`
		UnreadyTimeSeconds   int     `json:"unreadyTimeSeconds,omitempty"`
		UtilizationThreshold float64 `json:"utilizationThreshold,omitempty"`
	}

	// ManagedKubernetesNodeConfig configures the instances backing the nodes.
	ManagedKubernetesNodeConfig struct {
		AttachFloatingIp ManagedKubernetesAttachFloatingIp `json:"attachFloatingIp,omitzero"`
		Flavor           string                            `json:"flavor,omitempty"`
	}

	// ManagedKubernetesAttachFloatingIp toggles a floating IP on each node.
	ManagedKubernetesAttachFloatingIp struct {
		Enabled *bool `json:"enabled,omitempty"`
	}

	// ManagedKubernetesNodeTemplate is the template applied to the node pool nodes.
	ManagedKubernetesNodeTemplate struct {
		Metadata ManagedKubernetesNodeTemplateMetadata `json:"metadata,omitzero"`
		Spec     ManagedKubernetesNodeTemplateSpec     `json:"spec,omitzero"`
	}

	// ManagedKubernetesNodeTemplateMetadata is the metadata applied to the node pool nodes.
	ManagedKubernetesNodeTemplateMetadata struct {
		Annotations map[string]string `json:"annotations,omitempty"`
		Finalizers  []string          `json:"finalizers,omitempty"`
		Labels      map[string]string `json:"labels,omitempty"`
	}

	// ManagedKubernetesNodeTemplateSpec is the spec applied to the node pool nodes.
	ManagedKubernetesNodeTemplateSpec struct {
		Taints        []ManagedKubernetesTaint `json:"taints,omitempty"`
		Unschedulable *bool                    `json:"unschedulable,omitempty"`
	}

	// ManagedKubernetesTaint is a taint applied to the node pool nodes.
	ManagedKubernetesTaint struct {
		Effect string `json:"effect"`
		Key    string `json:"key"`
		Value  string `json:"value"`
	}
)

// API v2 endpoints, formatted with the project ID then the IDs of the resources along the path.
const (
	ManagedKubernetesV2EndpointFormat            = "/v2/publicCloud/project/%s/managedKubernetes"
	managedKubernetesClusterEndpointFormat       = ManagedKubernetesV2EndpointFormat + "/%s"
	ManagedKubernetesTasksEndpointFormat         = managedKubernetesClusterEndpointFormat + "/task"
	ManagedKubernetesNodepoolsEndpointFormat     = managedKubernetesClusterEndpointFormat + "/nodepool"
	managedKubernetesNodepoolEndpointFormat      = ManagedKubernetesNodepoolsEndpointFormat + "/%s"
	ManagedKubernetesNodepoolTasksEndpointFormat = managedKubernetesNodepoolEndpointFormat + "/task"
	ManagedKubernetesNodesEndpointFormat         = managedKubernetesNodepoolEndpointFormat + "/node"
)

// managedKubernetesEndpoint fills the given endpoint format with the path-escaped IDs.
func managedKubernetesEndpoint(format string, ids ...string) string {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = url.PathEscape(id)
	}
	return fmt.Sprintf(format, args...)
}

// managedKubernetesReferenceEndpoint returns the API v2 managed Kubernetes reference endpoint for the given project.
func managedKubernetesReferenceEndpoint(projectID string) string {
	return fmt.Sprintf("/v2/publicCloud/project/%s/reference/managedKubernetes", projectID)
}

// managedKubernetesFlavorEndpoint returns the API v2 endpoint of the flavors available in the given region.
func managedKubernetesFlavorEndpoint(projectID, region string) string {
	return managedKubernetesReferenceEndpoint(projectID) + "/flavor?region=" + url.QueryEscape(region)
}

// ListManagedKubernetes lists the clusters of the project.
func ListManagedKubernetes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesV2EndpointFormat, projectID)

	common.ManageListRequestNoExpand(endpoint, managedKubernetesColumnsToDisplay, flags.GenericFilters)
}

// GetManagedKubernetes displays the given cluster.
func GetManagedKubernetes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesV2EndpointFormat, projectID)

	common.ManageObjectRequest(endpoint, args[0], managedKubernetesTemplate)
}

// managedKubernetesCreationSpec completes the CLI values with the CNI and network type implied
// by the plan and vRack flags, so they also override stale values of a file or an editor example.
func managedKubernetesCreationSpec() ManagedKubernetesClusterSpec {
	// Works on a copy so derived values, backed by no flag, don't leak into the next run.
	spec := ManagedKubernetesSpec

	// An unknown plan is left alone, so that a plan added later is not given a wrong CNI.
	if spec.TargetSpec.Cni.Type == "" {
		spec.TargetSpec.Cni.Type = managedKubernetesPlanCNI[spec.TargetSpec.Plan]
	}

	if spec.TargetSpec.Network.Type == "" {
		switch {
		// Any vRack flag means the cluster goes in a vRack.
		case spec.TargetSpec.Network.Vrack != (ManagedKubernetesNetworkVrack{}):
			spec.TargetSpec.Network.Type = "VRACK"
		// Defaults to PUBLIC only without file, editor or pipe, so the type they set is kept.
		case !flags.ParametersViaEditor && flags.ParametersFile == "" && !utils.IsInputFromPipe():
			spec.TargetSpec.Network.Type = "PUBLIC"
		}
	}

	return spec
}

// CreateManagedKubernetes creates a cluster from the flags, a file or the editor.
func CreateManagedKubernetes(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesV2EndpointFormat, projectID)

	cluster, err := common.CreateResource(
		cmd,
		ManagedKubernetesV2SchemaPath,
		endpoint,
		"", // Examples come from the OpenAPI spec
		managedKubernetesCreationSpec(),
		assets.CloudV2OpenapiSchema,
		[]string{
			"targetSpec.name",
			"targetSpec.region",
			"targetSpec.plan",
		})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, cluster, "✅ Cluster created successfully (id: %s)", cluster["id"])
}

// EditManagedKubernetes updates the given cluster.
func EditManagedKubernetes(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0])

	// The checksum read from the fetched cluster is kept by the PUT filter (optimistic locking).
	if err := common.EditResourceWithMergeOpts(
		cmd,
		ManagedKubernetesV2ClusterSchemaPath,
		endpoint,
		ManagedKubernetesSpec,
		assets.CloudV2OpenapiSchema,
		[]func(*mergo.Config){utils.ReplaceSlices},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

// DeleteManagedKubernetes deletes the given cluster.
func DeleteManagedKubernetes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0])
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete MKS cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Cluster deleted successfully")
}

// GenerateKubeConfig prints the kubeconfig of the given cluster.
func GenerateKubeConfig(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0]) + "/kubeconfig"

	var kubeConfig map[string]any
	if err := httpLib.Client.Get(endpoint, &kubeConfig); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate kube config: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "%s", kubeConfig["content"])
}

// ResetKubeConfig resets the kubeconfig of the given cluster, invalidating the previous one.
func ResetKubeConfig(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0]) + "/kubeconfig/reset"

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset kube config: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Kube config reset successfully")
}

// RestartManagedKubernetes restarts the control plane of the given cluster.
func RestartManagedKubernetes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0]) + "/restart"
	if err := httpLib.Client.Post(endpoint, map[string]any{
		"force": ManagedKubernetesForceAction,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restart Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Kubernetes cluster restarting…")
}

// ListManagedKubernetesNodes lists the nodes of the given node pool.
func ListManagedKubernetesNodes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	query := url.Values{}
	if ManagedKubernetesNodeIncludeDeleted {
		query.Set("includeDeleted", "true")
	}
	if ManagedKubernetesNodeResourceStatus != "" {
		query.Set("resourceStatus", ManagedKubernetesNodeResourceStatus)
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodesEndpointFormat, projectID, args[0], args[1])
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	common.ManageListRequestNoExpand(endpoint, managedKubernetesNodeColumnsToDisplay, flags.GenericFilters)
}

// GetManagedKubernetesNode displays the given node.
func GetManagedKubernetesNode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodesEndpointFormat, projectID, args[0], args[1])

	common.ManageObjectRequest(
		endpoint,
		args[2],
		managedKubernetesNodeTemplate)
}

// ListManagedKubernetesNodepools lists the node pools of the given cluster.
func ListManagedKubernetesNodepools(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodepoolsEndpointFormat, projectID, args[0])

	common.ManageListRequestNoExpand(endpoint, managedKubernetesNodepoolColumnsToDisplay, flags.GenericFilters)
}

// GetManagedKubernetesNodepool displays the given node pool.
func GetManagedKubernetesNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodepoolsEndpointFormat, projectID, args[0])

	common.ManageObjectRequest(endpoint, args[1], managedKubernetesNodepoolTemplate)
}

// EditManagedKubernetesNodepool updates the given node pool.
func EditManagedKubernetesNodepool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesNodepoolEndpointFormat, projectID, args[0], args[1])

	// The checksum read from the fetched node pool is kept by the PUT filter (optimistic locking).
	if err := common.EditResourceWithMergeOpts(
		cmd,
		ManagedKubernetesV2NodepoolSchemaPath,
		endpoint,
		ManagedKubernetesNodepoolSpec,
		assets.CloudV2OpenapiSchema,
		[]func(*mergo.Config){utils.ReplaceSlices},
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

// DeleteManagedKubernetesNodepool deletes the given node pool.
func DeleteManagedKubernetesNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesNodepoolEndpointFormat, projectID, args[0], args[1])
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete MKS node pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Node pool deleted successfully")
}

// CreateManagedKubernetesNodepool creates a node pool in the given cluster.
func CreateManagedKubernetesNodepool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Works on a copy so the selected flavor, backed by no flag, doesn't leak into the next run.
	spec := ManagedKubernetesNodepoolSpec
	if InstanceFlavorViaInteractiveSelector {
		flavor, err := selectManagedKubernetesFlavor(args[0])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to get flavor from interactive selector: %s", err)
			return
		}
		spec.TargetSpec.NodeConfig.Flavor = flavor
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodepoolsEndpointFormat, projectID, args[0])

	nodepool, err := common.CreateResource(
		cmd,
		ManagedKubernetesV2NodepoolsSchemaPath,
		endpoint,
		"", // Examples come from the OpenAPI spec
		spec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec.nodeConfig.flavor"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Kubernetes node pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nodepool, "✅ Node pool created successfully (id: %s)", nodepool["id"])
}

// GetManagedKubernetesFlavorInteractiveSelector returns the flavor picked with the interactive
// selector, shaped as the node pool creation payload, or nil when the selector is not requested.
func GetManagedKubernetesFlavorInteractiveSelector(_ *cobra.Command, args []string) (map[string]any, error) {
	if !InstanceFlavorViaInteractiveSelector {
		return nil, nil
	}

	flavor, err := selectManagedKubernetesFlavor(args[0])
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"targetSpec": map[string]any{
			"nodeConfig": map[string]any{"flavor": flavor},
		},
	}, nil
}

// selectManagedKubernetesFlavor runs the flavor selector on the region of the given cluster.
func selectManagedKubernetesFlavor(clusterID string) (string, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", err
	}

	region, err := managedKubernetesRegion(projectID, clusterID)
	if err != nil {
		return "", err
	}

	// Picks from the MKS catalog: the instance one lists flavors node pools don't support.
	var flavors []ManagedKubernetesFlavor
	endpoint := managedKubernetesFlavorEndpoint(projectID, region)
	if err := httpLib.Client.Get(endpoint, &flavors); err != nil {
		return "", fmt.Errorf("failed to fetch available flavors: %w", err)
	}

	choices := make(map[string]string, len(flavors))
	for _, flavor := range flavors {
		if flavor.Status != "AVAILABLE" {
			continue
		}
		label := fmt.Sprintf("%s (%d vCPUs, %d GiB", flavor.Name, flavor.Resources.Vcpus, flavor.Resources.MemoryGiB)
		if flavor.Resources.Gpus > 0 {
			label += fmt.Sprintf(", %d GPUs", flavor.Resources.Gpus)
		}
		choices[label+")"] = flavor.Name
	}
	if len(choices) == 0 {
		return "", fmt.Errorf("no flavor available in region %s", region)
	}

	_, selectedFlavor, err := display.RunGenericChoicePicker("Please select a flavor", choices, 30)
	if err != nil {
		return "", fmt.Errorf("failed to select a flavor: %w", err)
	}

	if selectedFlavor == "" {
		return "", errors.New("no flavor selected, exiting")
	}

	return selectedFlavor, nil
}

// managedKubernetesRegion returns the region the given cluster is deployed in.
func managedKubernetesRegion(projectID, clusterID string) (string, error) {
	// The region is immutable, so the target spec holds the deployed one.
	var cluster ManagedKubernetesClusterSpec
	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, clusterID)
	if err := httpLib.Client.Get(endpoint, &cluster); err != nil {
		return "", fmt.Errorf("failed to fetch MKS cluster details: %w", err)
	}
	if cluster.TargetSpec.Region == "" {
		return "", errors.New("failed to read MKS cluster region")
	}

	return cluster.TargetSpec.Region, nil
}

// listManagedKubernetesEvents lists the events of the given cluster or node pool endpoint.
func listManagedKubernetesEvents(endpoint string) {
	endpoint += "/event"
	if ManagedKubernetesEventTaskID != "" {
		endpoint += "?taskId=" + url.QueryEscape(ManagedKubernetesEventTaskID)
	}

	body, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	// Exposes the task ID carried by the event link, so it can be passed to "task get".
	events := make([]map[string]any, 0, len(body))
	for _, object := range body {
		event := object.(map[string]any)
		if link, ok := event["link"].(string); ok {
			if _, taskID, found := strings.Cut(link, "/task/"); found {
				event["taskId"] = taskID
			}
		}
		events = append(events, event)
	}

	events, err = filtersLib.FilterLines(events, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(events, managedKubernetesEventColumnsToDisplay, &flags.OutputFormatConfig)
}

// ListManagedKubernetesTasks lists the tasks of the given cluster.
func ListManagedKubernetesTasks(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesTasksEndpointFormat, projectID, args[0])

	common.ManageListRequestNoExpand(endpoint, managedKubernetesTaskColumnsToDisplay, flags.GenericFilters)
}

// GetManagedKubernetesTask displays the given cluster task.
func GetManagedKubernetesTask(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesTasksEndpointFormat, projectID, args[0])

	common.ManageObjectRequest(endpoint, args[1], managedKubernetesTaskTemplate)
}

// ListManagedKubernetesEvents lists the events of the given cluster.
func ListManagedKubernetesEvents(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0])

	listManagedKubernetesEvents(endpoint)
}

// ListManagedKubernetesNodepoolTasks lists the tasks of the given node pool.
func ListManagedKubernetesNodepoolTasks(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodepoolTasksEndpointFormat, projectID, args[0], args[1])

	common.ManageListRequestNoExpand(endpoint, managedKubernetesTaskColumnsToDisplay, flags.GenericFilters)
}

// GetManagedKubernetesNodepoolTask displays the given node pool task.
func GetManagedKubernetesNodepoolTask(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(ManagedKubernetesNodepoolTasksEndpointFormat, projectID, args[0], args[1])

	common.ManageObjectRequest(endpoint, args[2], managedKubernetesTaskTemplate)
}

// ListManagedKubernetesNodepoolEvents lists the events of the given node pool.
func ListManagedKubernetesNodepoolEvents(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesNodepoolEndpointFormat, projectID, args[0], args[1])

	listManagedKubernetesEvents(endpoint)
}

// GetManagedKubernetesEtcdUsage displays the etcd storage usage of the given cluster.
func GetManagedKubernetesEtcdUsage(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0]) + "/usage"

	common.ManageObjectRequest(endpoint, "etcd", managedKubernetesEtcdUsageTemplate)
}

// GetManagedKubernetesVersionCapabilities displays the versions the given cluster can be updated (patch) or upgraded (minor) to.
func GetManagedKubernetesVersionCapabilities(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesEndpoint(managedKubernetesClusterEndpointFormat, projectID, args[0]) + "/capabilities"

	common.ManageObjectRequest(endpoint, "version", managedKubernetesVersionCapabilitiesTemplate)
}

// ListManagedKubernetesPlans lists the plans available in the reference region.
func ListManagedKubernetesPlans(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesReferenceEndpoint(projectID) + "/plan?region=" + url.QueryEscape(ManagedKubernetesReferenceRegion)

	common.ManageListRequestNoExpand(endpoint, managedKubernetesPlanColumnsToDisplay, flags.GenericFilters)
}

// ListManagedKubernetesRegions lists the regions where clusters can be created.
func ListManagedKubernetesRegions(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesReferenceEndpoint(projectID) + "/region"

	common.ManageListRequestNoExpand(endpoint, managedKubernetesRegionColumnsToDisplay, flags.GenericFilters)
}

// ListManagedKubernetesVersions lists the versions available for cluster creation, flagging the default one.
func ListManagedKubernetesVersions(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var catalog ManagedKubernetesVersionCatalog
	endpoint := managedKubernetesReferenceEndpoint(projectID) + "/version"
	if err := httpLib.Client.Get(endpoint, &catalog); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch Kubernetes versions: %s", err)
		return
	}

	for _, version := range catalog.AvailableVersions {
		version["default"] = version["version"] == catalog.DefaultVersion
	}

	versions, err := filtersLib.FilterLines(catalog.AvailableVersions, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(versions, managedKubernetesVersionColumnsToDisplay, &flags.OutputFormatConfig)
}

// ListManagedKubernetesFlavors lists the node pool flavors available in the reference region.
func ListManagedKubernetesFlavors(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := managedKubernetesFlavorEndpoint(projectID, ManagedKubernetesReferenceRegion)

	common.ManageListRequestNoExpand(endpoint, managedKubernetesFlavorColumnsToDisplay, flags.GenericFilters)
}
