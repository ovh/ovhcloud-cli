// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/editor"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectKubeColumnsToDisplay = []string{"id", "name", "region", "plan", "version", "status"}

	//go:embed templates/cloud_kube.tmpl
	cloudKubeTemplate string

	//go:embed templates/cloud_kube_customization.tmpl
	cloudKubeCustomizationTemplate string

	//go:embed templates/cloud_kube_node.tmpl
	cloudKubeNodeTemplate string

	//go:embed templates/cloud_kube_nodepool.tmpl
	cloudKubeNodepoolTemplate string

	//go:embed templates/cloud_kube_oidc.tmpl
	cloudKubeOIDCTemplate string

	//go:embed parameter-samples/kube-create.json
	CloudKubeCreationExample string

	//go:embed parameter-samples/kube-reset.json
	CloudKubeResetExample string

	//go:embed parameter-samples/kube-nodepool-create.json
	CloudKubeNodePoolCreationExample string

	//go:embed parameter-samples/kube-oidc-create.json
	CloudKubeOIDCCreationExample string

	// KubeSpec defines the structure for a Kubernetes cluster specification
	KubeSpec struct {
		Customization struct {
			APIServer struct {
				AdmissionPlugins struct {
					Disabled []string `json:"disabled,omitempty"`
					Enabled  []string `json:"enabled,omitempty"`
				} `json:"admissionPlugins,omitzero"`
			} `json:"apiServer,omitzero"`
			KubeProxy struct {
				IPTables struct {
					MinSyncPeriod string `json:"minSyncPeriod,omitempty"`
					SyncPeriod    string `json:"syncPeriod,omitempty"`
				} `json:"iptables,omitzero"`
				IPVS struct {
					MinSyncPeriod string `json:"minSyncPeriod,omitempty"`
					Scheduler     string `json:"scheduler,omitempty"`
					SyncPeriod    string `json:"syncPeriod,omitempty"`
					TCPFinTimeout string `json:"tcpFinTimeout,omitempty"`
					TCPTimeout    string `json:"tcpTimeout,omitempty"`
					UDPTimeout    string `json:"udpTimeout,omitempty"`
				} `json:"ipvs,omitzero"`
			} `json:"kubeProxy,omitzero"`
		} `json:"customization,omitzero"`
		KubeProxyMode               string `json:"kubeProxyMode,omitempty"`
		LoadBalancersSubnetId       string `json:"loadBalancersSubnetId,omitempty"`
		Name                        string `json:"name,omitempty"`
		NodesSubnetId               string `json:"nodesSubnetId,omitempty"`
		PrivateNetworkConfiguration struct {
			DefaultVrackGateway            string `json:"defaultVrackGateway,omitempty"`
			PrivateNetworkRoutingAsDefault bool   `json:"privateNetworkRoutingAsDefault,omitempty"`
		} `json:"privateNetworkConfiguration,omitzero"`
		PrivateNetworkId  string `json:"privateNetworkId,omitempty"`
		Region            string `json:"region,omitempty"`
		UpdatePolicy      string `json:"updatePolicy,omitempty"`
		Version           string `json:"version,omitempty"`
		WorkerNodesPolicy string `json:"workerNodesPolicy,omitempty"`
		Plan              string `json:"plan,omitempty"`
	}

	// KubeNodepoolSpec defines the structure for a Kubernetes node pool specification
	KubeNodepoolSpec kubeNodepoolSpec

	// KubeOIDCConfig defines the structure for OpenID Connect configuration in Kubernetes
	KubeOIDCConfig struct {
		CaContent         string   `json:"caContent,omitempty"`
		ClientId          string   `json:"clientId,omitempty"`
		GroupsClaim       []string `json:"groupsClaim,omitempty"`
		GroupsPrefix      string   `json:"groupsPrefix,omitempty"`
		IssuerUrl         string   `json:"issuerUrl,omitempty"`
		RequiredClaim     []string `json:"requiredClaim,omitempty"`
		SigningAlgorithms []string `json:"signingAlgorithms,omitempty"`
		UsernameClaim     string   `json:"usernameClaim,omitempty"`
		UsernamePrefix    string   `json:"usernamePrefix,omitempty"`
	}

	// KubeForceAction indicates whether to force an action
	// It is set by a command line flag
	KubeForceAction bool

	// KubeUpdateStrategy defines the strategy for updating Kubernetes clusters
	// It is set by a command line flag
	KubeUpdateStrategy string

	// KubeIPRestrictions defines the IP restrictions for Kubernetes clusters
	// It is set by a command line flag
	KubeIPRestrictions []string
)

type kubeNodepoolSpec struct {
	AntiAffinity bool `json:"antiAffinity,omitempty"`
	Autoscale    bool `json:"autoscale,omitempty"`
	Autoscaling  struct {
		ScaleDownUnneededTimeSeconds  int     `json:"scaleDownUnneededTimeSeconds,omitempty"`
		ScaleDownUnreadyTimeSeconds   int     `json:"scaleDownUnreadyTimeSeconds,omitempty"`
		ScaleDownUtilizationThreshold float64 `json:"scaleDownUtilizationThreshold,omitempty"`
	} `json:"autoscaling,omitzero"`
	AvailabilityZones []string `json:"availabilityZones,omitempty"`
	DesiredNodes      int      `json:"desiredNodes,omitempty"`
	FlavorName        string   `json:"flavorName,omitempty"`
	MaxNodes          int      `json:"maxNodes,omitempty"`
	MinNodes          int      `json:"minNodes,omitempty"`
	MonthlyBilled     bool     `json:"monthlyBilled,omitempty"`
	Name              string   `json:"name,omitempty"`
	Template          struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations,omitempty"`
			Finalizers  []string          `json:"finalizers,omitempty"`
			Labels      map[string]string `json:"labels,omitempty"`
		} `json:"metadata,omitzero"`
		Spec struct {
			Taints            []kubeNodepoolSpecTaintType `json:"taints,omitempty"`
			CommandLineTaints []string                    `json:"-"`
			Unschedulable     bool                        `json:"unschedulable,omitempty"`
		} `json:"spec,omitzero"`
	} `json:"template,omitzero"`

	// Only used when updating a node pool
	NodesToRemove []string `json:"nodesToRemove,omitempty"`
}

type kubeNodepoolSpecTaintType struct {
	Effect string `json:"effect,omitempty"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

func ListKubes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequest(fmt.Sprintf("/v1/cloud/project/%s/kube", projectID), "", cloudprojectKubeColumnsToDisplay, flags.GenericFilters)
}

func GetKube(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", endpoint, err)
		return
	}

	// Fetch etcd usage, ignore potential errors
	var etcdUsage map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("%s/metrics/etcdUsage", endpoint), &etcdUsage); err != nil {
		log.Printf("failed to fetch etcd usage: %s", err)
	} else {
		etcdUsage["usage"], err = etcdUsage["usage"].(json.Number).Float64()
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse etcd usage 'usage' value: %s", err)
			return
		}
		etcdUsage["quota"], err = etcdUsage["quota"].(json.Number).Float64()
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to parse etcd usage 'quota' value: %s", err)
			return
		}
		object["etcdUsage"] = etcdUsage
	}

	display.OutputObject(object, args[0], cloudKubeTemplate, &flags.OutputFormatConfig)
}

func CreateKube(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube", projectID)
	cluster, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/kube",
		endpoint,
		CloudKubeCreationExample,
		KubeSpec,
		assets.CloudOpenapiSchema,
		[]string{"region"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, cluster, "✅ Cluster %s created successfully (id: %s)", cluster["name"], cluster["id"])
}

func EditKube(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}",
		fmt.Sprintf("/v1/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0])),
		KubeSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteKube(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete MKS cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ MKS cluster is being deleted…")
}

func GetKubeCustomization(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/customization", projectID, url.PathEscape(args[0]))
	var customization map[string]any
	if err := httpLib.Client.Get(endpoint, &customization); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch MKS cluster customization: %s", err)
		return
	}

	display.OutputObject(customization, args[0], cloudKubeCustomizationTemplate, &flags.OutputFormatConfig)
}

func EditKubeCustomization(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/customization",
		fmt.Sprintf("/v1/cloud/project/%s/kube/%s/customization", projectID, url.PathEscape(args[0])),
		KubeSpec.Customization,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func ListKubeIPRestrictions(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/ipRestrictions", projectID, url.PathEscape(args[0]))

	body, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch IP restrictions: %s", err)
		return
	}

	objects := make([]map[string]any, 0, len(body))
	for _, ip := range body {
		objects = append(objects, map[string]any{
			"ip": ip,
		})
	}

	display.RenderTable(objects, []string{"ip"}, &flags.OutputFormatConfig)
}

func EditKubeIPRestrictions(cmd *cobra.Command, args []string) {
	if cmd.Flags().NFlag() == 0 {
		display.OutputWarning(&flags.OutputFormatConfig, "No parameters given, nothing to edit")
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/ipRestrictions", projectID, url.PathEscape(args[0]))

	if flags.ParametersViaEditor {
		// Fetch resource
		var ips []string
		if err := httpLib.Client.Get(endpoint, &ips); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching resource %s", err)
			return
		}

		// Format editable body
		editableOutput, err := json.MarshalIndent(map[string]any{
			"ips": ips,
		}, "", "  ")
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to marshal writable body: %s", err)
			return
		}

		// Edit value
		updatedBody, err := editor.EditValueWithEditor(editableOutput)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to edit properties: %s", err)
			return
		}

		// Update API call
		if err := httpLib.Client.Put(endpoint, json.RawMessage(updatedBody), nil); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to update resource: %s", err)
			return
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP restrictions updated successfully…")
		return
	}

	// Update API call with IPS set via flags
	if err := httpLib.Client.Put(endpoint, map[string]any{
		"ips": KubeIPRestrictions,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update IP restrictions: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ IP restrictions updated successfully…")
}

func GenerateKubeConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/kubeconfig", projectID, url.PathEscape(args[0]))

	var kubeConfig map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &kubeConfig); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to generate kube config: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "%s", kubeConfig["content"])
}

func ResetKubeConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/kubeconfig/reset", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset kube config: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Kube config reset successfully")
}

func ListKubeNodes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/node", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "flavor", "version", "status"}, flags.GenericFilters)
}

func GetKubeNode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/node", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudKubeNodeTemplate)
}

func DeleteKubeNode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/node/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete MKS node: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ MKS node deleted successfully")
}

func ListKubeNodepools(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "flavor", "currentNodes", "status"}, flags.GenericFilters)
}

func GetKubeNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudKubeNodepoolTemplate)
}

func EditKubeNodepool(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/nodepool/{nodepoolId}",
		fmt.Sprintf("/v1/cloud/project/%s/kube/%s/nodepool/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1])),
		KubeNodepoolSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeleteKubeNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/nodepool/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete MKS node pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ MKS node pool deleted successfully")
}

func CreateKubeNodepool(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		display.OutputError(&flags.OutputFormatConfig, "create command requires the MKS cluster ID as the first argument.\nUsage:\n%s", cmd.UsageString())
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Transform spec taints into a format suitable for the API
	for _, taint := range KubeNodepoolSpec.Template.Spec.CommandLineTaints {
		parts := strings.Split(taint, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid taint format: %q, expected format is key=value:effect", taint)
			return
		}

		kvParts := strings.Split(parts[0], "=")
		if len(kvParts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid taint format: %q, expected format is key=value:effect", taint)
			return
		}

		KubeNodepoolSpec.Template.Spec.Taints = append(KubeNodepoolSpec.Template.Spec.Taints,
			kubeNodepoolSpecTaintType{
				Key:    kvParts[0],
				Value:  kvParts[1],
				Effect: parts[1],
			},
		)
	}

	// Run interactive flavor selector if the flag is set
	flavor, err := GetKubeFlavorInteractiveSelector(cmd, args)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get flavor from interactive selector: %s", err)
		return
	}
	if flavor != nil {
		KubeNodepoolSpec.FlavorName = flavor["flavorName"].(string)
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))
	nodepool, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/nodepool",
		endpoint,
		CloudKubeNodePoolCreationExample,
		KubeNodepoolSpec,
		assets.CloudOpenapiSchema,
		[]string{"flavorName"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create Kubernetes node pool: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nodepool, "✅ Node pool %s created successfully", nodepool["id"])
}

func GetKubeFlavorInteractiveSelector(cmd *cobra.Command, args []string) (map[string]any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("create command requires the MKS cluster as the first argument.\nUsage:\n%s", cmd.UsageString())
	}

	if !InstanceFlavorViaInteractiveSelector {
		return nil, nil
	}

	clusterID := args[0]

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return nil, err
	}

	// Fetch MKS cluster to extract its region
	var clusterDetails map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/kube/%s", projectID, url.PathEscape(clusterID)), &clusterDetails); err != nil {
		return nil, fmt.Errorf("failed to fetch MKS cluster details: %w", err)
	}
	region := clusterDetails["region"].(string)

	selectedFlavor, _, err := runFlavorSelector(projectID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to select a flavor: %w", err)
	}

	if selectedFlavor == "" {
		return nil, errors.New("no flavor selected, exiting")
	}

	return map[string]any{
		"flavorName": selectedFlavor,
	}, nil
}

func GetKubeOIDCIntegration(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))

	var oidcConfig map[string]any
	if err := httpLib.Client.Get(endpoint, &oidcConfig); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch OIDC configuration: %s", err)
		return
	}

	display.OutputObject(oidcConfig, args[0], cloudKubeOIDCTemplate, &flags.OutputFormatConfig)
}

func EditKubeOIDCIntegration(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/openIdConnect",
		fmt.Sprintf("/v1/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0])),
		KubeOIDCConfig,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateKubeOIDCIntegration(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))
	if _, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/openIdConnect",
		endpoint,
		CloudKubeOIDCCreationExample,
		KubeOIDCConfig,
		assets.CloudOpenapiSchema,
		[]string{"clientId", "issuerUrl"}); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create OIDC integration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ OIDC integration created successfully")
}

func DeleteKubeOIDCIntegration(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete OIDC integration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ OIDC integration deleted successfully")
}

func GetKubePrivateNetworkConfiguration(_cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/privateNetworkConfiguration", projectID, url.PathEscape(args[0]))

	var privateNetworkConfig map[string]any
	if err := httpLib.Client.Get(endpoint, &privateNetworkConfig); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch private network configuration: %s", err)
		return
	}

	display.RenderTable([]map[string]any{privateNetworkConfig}, []string{"defaultVrackGateway", "privateNetworkRoutingAsDefault"}, &flags.OutputFormatConfig)
}

func EditKubePrivateNetworkConfiguration(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/privateNetworkConfiguration",
		fmt.Sprintf("/v1/cloud/project/%s/kube/%s/privateNetworkConfiguration", projectID, url.PathEscape(args[0])),
		KubeSpec.PrivateNetworkConfiguration,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func ResetKubeCluster(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/reset", projectID, url.PathEscape(args[0]))
	_, err = common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/kube/{kubeId}/reset",
		endpoint,
		CloudKubeResetExample,
		KubeSpec,
		assets.CloudOpenapiSchema,
		nil)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reset Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Kubernetes cluster is being reset…")
}

func RestartKubeCluster(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(fmt.Sprintf("/v1/cloud/project/%s/kube/%s/restart", projectID, url.PathEscape(args[0])), map[string]any{
		"force": KubeForceAction,
	}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to restart Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Kubernetes cluster restarting…")
}

func UpdateKubeCluster(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/update", projectID, url.PathEscape(args[0]))

	body := map[string]any{
		"force": KubeForceAction,
	}
	if KubeUpdateStrategy != "" {
		body["strategy"] = KubeUpdateStrategy
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update Kubernetes cluster: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "⚡️ Kubernetes cluster update in progress…")
}

func UpdateKubeLoadBalancersSubnet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/kube/%s/updateLoadBalancersSubnetId", projectID, url.PathEscape(args[0]))
	body := map[string]any{
		"loadBalancersSubnetId": args[1],
	}

	if err := httpLib.Client.Put(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update load balancers subnet: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Load balancers subnet updated successfully")
}
