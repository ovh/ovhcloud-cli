package cloud

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/editor"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
	"stash.ovh.net/api/ovh-cli/internal/services/common"
)

var (
	cloudprojectKubeColumnsToDisplay = []string{"id", "name", "region", "version", "status"}

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
		Name                        string `json:"name"`
		NodesSubnetId               string `json:"nodesSubnetId,omitempty"`
		PrivateNetworkConfiguration struct {
			DefaultVrackGateway            string `json:"defaultVrackGateway,omitempty"`
			PrivateNetworkRoutingAsDefault bool   `json:"privateNetworkRoutingAsDefault,omitempty"`
		} `json:"privateNetworkConfiguration,omitzero"`
		PrivateNetworkId string `json:"privateNetworkId,omitempty"`
		Region           string `json:"region,omitempty"`
		UpdatePolicy     string `json:"updatePolicy,omitempty"`
		Version          string `json:"version,omitempty"`
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
}

type kubeNodepoolSpecTaintType struct {
	Effect string `json:"effect,omitempty"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
}

func ListKubes(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), "", cloudprojectKubeColumnsToDisplay, flags.GenericFilters)
}

func GetKube(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(fmt.Sprintf("/cloud/project/%s/kube", projectID), args[0], cloudKubeTemplate)
}

func CreateKube(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube", projectID)
	cluster, err := common.CreateResource("/cloud/project/{serviceName}/kube",
		endpoint,
		CloudKubeCreationExample,
		KubeSpec,
		CloudOpenapiSchema,
		[]string{"region"})
	if err != nil {
		display.ExitError("failed to create Kubernetes cluster: %s", err)
		return
	}

	fmt.Printf("\n✅ Cluster %s created successfully (id: %s)\n", cluster["name"], cluster["id"])
}

func EditKube(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/kube/{kubeId}", endpoint, CloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}

func DeleteKube(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete MKS cluster: %s", err)
		return
	}

	fmt.Println("\n✅ MKS cluster deleted successfully")
}

func GetKubeCustomization(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/customization", projectID, url.PathEscape(args[0]))
	var customization map[string]any
	if err := httpLib.Client.Get(endpoint, &customization); err != nil {
		display.ExitError("failed to fetch MKS cluster customization: %s", err)
		return
	}

	display.OutputObject(customization, args[0], cloudKubeCustomizationTemplate, &flags.OutputFormatConfig)
}

func EditKubeCustomization(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/customization", projectID, url.PathEscape(args[0]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/kube/{kubeId}/customization", endpoint, CloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func ListKubeIPRestrictions(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions", projectID, url.PathEscape(args[0]))

	body, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.ExitError("failed to fetch IP restrictions: %s", err)
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
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions", projectID, url.PathEscape(args[0]))

	// Fetch current IP restrictions
	currentRestrictions, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.ExitError("failed to fetch IP restrictions: %s", err)
		return
	}

	// Prepare editable body
	editableBody := map[string]any{
		"ips": currentRestrictions,
	}

	// Format editable body
	editableOutput, err := json.MarshalIndent(editableBody, "", "  ")
	if err != nil {
		display.ExitError("failed to marshal writable body: %s", err)
		return
	}

	// Edit value
	updatedBody, err := editor.EditValueWithEditor(editableOutput)
	if err != nil {
		display.ExitError("failed to edit properties: %s", err)
		return
	}

	if slices.Equal(updatedBody, editableOutput) {
		display.ExitWarning("\n🟠 Resource not updated, exiting")
		return
	}

	// Update API call
	if err := httpLib.Client.Put(endpoint, json.RawMessage(updatedBody), nil); err != nil {
		display.ExitError("failed to update resource: %s", err)
		return
	}

	fmt.Println("\n✅ IP restrictions updated succesfully ...")
}

func GenerateKubeConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/kubeconfig", projectID, url.PathEscape(args[0]))

	var kubeConfig map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &kubeConfig); err != nil {
		display.ExitError("failed to generate kube config: %s", err)
		return
	}

	fmt.Println(kubeConfig["content"])
}

func ResetKubeConfig(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/kubeconfig/reset", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.ExitError("failed to reset kube config: %s", err)
		return
	}

	fmt.Println("\n✅ Kube config reset successfully")
}

func ListKubeNodes(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/node", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "flavor", "version", "status"}, flags.GenericFilters)
}

func GetKubeNode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/node", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudKubeNodeTemplate)
}

func DeleteKubeNode(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/node/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete MKS node: %s", err)
		return
	}

	fmt.Println("\n✅ MKS node deleted successfully")
}

func ListKubeNodepools(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "name", "flavor", "currentNodes", "status"}, flags.GenericFilters)
}

func GetKubeNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))

	common.ManageObjectRequest(endpoint, args[1], cloudKubeNodepoolTemplate)
}

func EditKubeNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/kube/{kubeId}/nodepool/{nodepoolId}", endpoint, CloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func DeleteKubeNodepool(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete MKS node pool: %s", err)
		return
	}

	fmt.Println("\n✅ MKS node pool deleted successfully")
}

func CreateKubeNodepool(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		display.ExitError("create command requires the MKS cluster ID as the first argument.\nUsage:\n%s", cmd.UsageString())
		return
	}

	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Transform spec taints into a format suitable for the API
	for _, taint := range KubeNodepoolSpec.Template.Spec.CommandLineTaints {
		parts := strings.Split(taint, ":")
		if len(parts) != 2 {
			display.ExitError("invalid taint format: %q, expected format is key=value:effect", taint)
			return
		}

		kvParts := strings.Split(parts[0], "=")
		if len(kvParts) != 2 {
			display.ExitError("invalid taint format: %q, expected format is key=value:effect", taint)
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
		display.ExitError("failed to get flavor from interactive selector: %s", err)
		return
	}
	if flavor != nil {
		KubeNodepoolSpec.FlavorName = flavor["flavorName"].(string)
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, url.PathEscape(args[0]))
	nodepool, err := common.CreateResource("/cloud/project/{serviceName}/kube/{kubeId}/nodepool",
		endpoint,
		CloudKubeNodePoolCreationExample,
		KubeNodepoolSpec,
		CloudOpenapiSchema,
		[]string{"flavorName"})
	if err != nil {
		display.ExitError("failed to create Kubernetes node pool: %s", err)
		return
	}

	fmt.Printf("\n✅ Node pool %s created successfully\n", nodepool["id"])
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
	if err := httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/kube/%s", projectID, url.PathEscape(clusterID)), &clusterDetails); err != nil {
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
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))

	var oidcConfig map[string]any
	if err := httpLib.Client.Get(endpoint, &oidcConfig); err != nil {
		display.ExitError("failed to fetch OIDC configuration: %s", err)
		return
	}

	display.OutputObject(oidcConfig, args[0], cloudKubeOIDCTemplate, &flags.OutputFormatConfig)
}

func EditKubeOIDCIntegration(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))

	if err := editor.EditResource(httpLib.Client, "/cloud/project/{serviceName}/kube/{kubeId}/openIdConnect", endpoint, CloudOpenapiSchema); err != nil {
		display.ExitError(err.Error())
	}
}

func CreateKubeOIDCIntegration(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))
	if _, err := common.CreateResource("/cloud/project/{serviceName}/kube/{kubeId}/openIdConnect",
		endpoint,
		CloudKubeOIDCCreationExample,
		KubeOIDCConfig,
		CloudOpenapiSchema,
		[]string{"clientId", "issuerUrl"}); err != nil {
		display.ExitError("failed to create OIDC integration: %s", err)
		return
	}

	fmt.Println("\n✅ OIDC integration created successfully")
}

func DeleteKubeOIDCIntegration(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect", projectID, url.PathEscape(args[0]))

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete OIDC integration: %s", err)
		return
	}

	fmt.Println("\n✅ OIDC integration deleted successfully")
}
