// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudprojectNetworkColumnsToDisplay = []string{"id", "name", "openstackId", "region", "status"}
	cloudprojectGatewayColumnsToDisplay = []string{"id", "name", "region", "model", "status"}

	//go:embed templates/cloud_network_private.tmpl
	cloudNetworkPrivateTemplate string

	//go:embed templates/cloud_network_public.tmpl
	cloudNetworkPublicTemplate string

	//go:embed templates/cloud_network_gateway.tmpl
	cloudGatewayTemplate string

	// CloudNetworkName is used to store the name of the cloud network
	CloudNetworkName string

	//go:embed parameter-samples/private-network-create.json
	PrivateNetworkCreationExample string

	//go:embed parameter-samples/private-network-subnet-create.json
	PrivateNetworkSubnetCreationExample string

	//go:embed parameter-samples/gateway-create.json
	GatewayCreationExample string

	// CloudGatewaySpec contains the parameters for updating a cloud gateway
	CloudGatewaySpec struct {
		Model             string `json:"model,omitempty"`
		Name              string `json:"name,omitempty"`
		ExistingNetworkID string `json:"-"`
		ExistingSubnetID  string `json:"-"`
		Network           struct {
			Name   string `json:"name,omitempty"`
			VlanId int    `json:"vlanId,omitempty"`
			Subnet struct {
				Name                        string                         `json:"name,omitempty"`
				Cidr                        string                         `json:"cidr,omitempty"`
				EnableDhcp                  bool                           `json:"enableDhcp,omitempty"`
				GatewayIp                   string                         `json:"gatewayIp,omitempty"`
				DnsNameServers              []string                       `json:"dnsNameServers,omitempty"`
				UseDefaultPublicDNSResolver bool                           `json:"useDefaultPublicDNSResolver,omitempty"`
				IPVersion                   int                            `json:"ipVersion,omitempty"`
				AllocationPools             []PrivateNetworkAllocationPool `json:"allocationPools,omitempty"`
				HostRoutes                  []PrivateNetworkHostRoute      `json:"hostRoutes,omitempty"`

				CliAllocationPools []string `json:"-"`
				CliHostRoutes      []string `json:"-"`
			} `json:"subnet,omitzero"`
		} `json:"network,omitzero"`
	}

	CloudNetworkSpec struct {
		Name    string `json:"name,omitempty"`
		VlanId  int    `json:"vlanId,omitempty"`
		Gateway struct {
			Model string `json:"model,omitempty"`
			Name  string `json:"name,omitempty"`
		} `json:"gateway,omitzero"`
		Subnet struct {
			Name                        string                         `json:"name,omitempty"`
			Cidr                        string                         `json:"cidr,omitempty"`
			EnableDhcp                  bool                           `json:"enableDhcp"`
			EnableGatewayIp             bool                           `json:"enableGatewayIp"`
			GatewayIp                   string                         `json:"gatewayIp,omitempty"`
			DnsNameServers              []string                       `json:"dnsNameServers,omitempty"`
			UseDefaultPublicDNSResolver bool                           `json:"useDefaultPublicDNSResolver,omitempty"`
			IPVersion                   int                            `json:"ipVersion,omitempty"`
			AllocationPools             []PrivateNetworkAllocationPool `json:"allocationPools,omitempty"`
			HostRoutes                  []PrivateNetworkHostRoute      `json:"hostRoutes,omitempty"`

			CliAllocationPools []string `json:"-"`
			CliHostRoutes      []string `json:"-"`
		} `json:"subnet,omitzero"`
	}

	CloudNetworkSubnetEditSpec struct {
		Dhcp           bool   `json:"dhcp"`
		DisableGateway bool   `json:"disableGateway"`
		GatewayIp      string `json:"gatewayIp,omitempty"`
	}

	CloudNetworkSubnetSpec struct {
		DHCP      bool   `json:"dhcp"`
		NoGateway bool   `json:"noGateway"`
		Network   string `json:"network"`
		Start     string `json:"start"`
		End       string `json:"end"`
		Region    string `json:"region"`
	}

	GatewayInterfaceSpec struct {
		SubnetID string `json:"subnetId,omitempty"`
	}
)

type (
	PrivateNetworkAllocationPool struct {
		Start string `json:"start,omitempty"`
		End   string `json:"end,omitempty"`
	}

	PrivateNetworkHostRoute struct {
		Destination string `json:"destination,omitempty"`
		NextHop     string `json:"nextHop,omitempty"`
	}

	NetworkRegionDetails struct {
		OpenstackID string `json:"openstackId"`
		Region      string `json:"region"`
	}

	PrivateNetwork struct {
		ID      string                 `json:"id"`
		Regions []NetworkRegionDetails `json:"regions"`
	}
)

func ListPrivateNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	body, err := httpLib.FetchExpandedArray(fmt.Sprintf("/v1/cloud/project/%s/network/private", projectID), "id")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	flattenedBody := []map[string]any{}

	for _, line := range body {
		regions := line["regions"].([]any)

		for _, region := range regions {
			region := region.(map[string]any)

			flattenedBody = append(flattenedBody, map[string]any{
				"id":          line["id"],
				"name":        line["name"],
				"vlanId":      line["vlanId"],
				"openstackId": region["openstackId"],
				"region":      region["region"],
				"status":      region["status"],
				"type":        line["type"],
			})
		}
	}

	body, err = filtersLib.FilterLines(flattenedBody, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetPrivateNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	path := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
		return
	}

	for _, region := range object["regions"].([]any) {
		region := region.(map[string]any)

		// Skip regions without openstackId
		if openstackId, ok := region["openstackId"]; !ok || openstackId == nil {
			continue
		}

		// Fetch subnets of region network
		path = fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
			projectID,
			url.PathEscape(region["region"].(string)),
			url.PathEscape(region["openstackId"].(string)),
		)
		var subnets []map[string]any
		if err := httpLib.Client.Get(path, &subnets); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
			return
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPrivateTemplate, &flags.OutputFormatConfig)
}

func EditPrivateNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/network/private/{networkId}",
		fmt.Sprintf("/v1/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0])),
		map[string]any{
			"name": CloudNetworkName,
		},
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreatePrivateNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Transform CLI flags into the CloudNetworkSpec structure
	for _, allocationPool := range CloudNetworkSpec.Subnet.CliAllocationPools {
		parts := strings.Split(allocationPool, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid allocation pool format, expected start:end, got %s", allocationPool)
			return
		}

		CloudNetworkSpec.Subnet.AllocationPools = append(CloudNetworkSpec.Subnet.AllocationPools, PrivateNetworkAllocationPool{
			Start: parts[0],
			End:   parts[1],
		})
	}
	for _, hostRoute := range CloudNetworkSpec.Subnet.CliHostRoutes {
		parts := strings.Split(hostRoute, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid host route format, expected destination:nextHop, got %s", hostRoute)
			return
		}

		CloudNetworkSpec.Subnet.HostRoutes = append(CloudNetworkSpec.Subnet.HostRoutes, PrivateNetworkHostRoute{
			Destination: parts[0],
			NextHop:     parts[1],
		})
	}

	// Create resource
	region := args[0]
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network", projectID, url.PathEscape(region))
	task, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/network",
		endpoint,
		PrivateNetworkCreationExample,
		CloudNetworkSpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "subnet"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create private network: %s", err)
		return
	}

	// Wait for task to complete if --wait flag is set
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, task, `✅ Network creation started successfully (operation ID: %s)
You can check the status of the operation with: 'ovhcloud cloud operation get %[1]s'`, task["id"])
		return
	}

	networkID, err := waitForCloudOperation(projectID, task["id"].(string), "network#create", 10*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for network creation: %s", err)
		return
	}

	// Fetch all private networks
	var networks []PrivateNetwork
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/network/private", projectID), &networks); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch private networks: %s", err)
		return
	}

	// Find the created network
	var (
		foundNetwork       *PrivateNetwork
		foundRegionNetwork *NetworkRegionDetails
	)

eachNetwork:
	for _, network := range networks {
		for _, regionDetails := range network.Regions {
			if regionDetails.OpenstackID == networkID && regionDetails.Region == region {
				foundNetwork = &network
				foundRegionNetwork = &regionDetails
				break eachNetwork
			}
		}
	}

	if foundNetwork == nil {
		display.OutputError(&flags.OutputFormatConfig, "created network not found, this is unexpected")
		return
	}

	// Fetch subnets of created network
	endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet", projectID, url.PathEscape(region), url.PathEscape(foundRegionNetwork.OpenstackID))
	var subnets []map[string]any
	if err := httpLib.Client.Get(endpoint, &subnets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch subnets of created network: %s", err)
		return
	}

	// Fetch gateway of created subnets and prepare output
	var outputSubnets []map[string]any
	for _, subnet := range subnets {
		endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway?subnetId=%s",
			projectID,
			url.PathEscape(region),
			url.PathEscape(subnet["id"].(string)),
		)

		var gateways []map[string]any
		if err := httpLib.Client.Get(endpoint, &gateways); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch gateways of created network: %s", err)
			return
		}

		var outputGateways []map[string]any
		for _, gateway := range gateways {
			outputGateway := map[string]any{
				"id":   gateway["id"],
				"name": gateway["name"],
			}
			outputGateways = append(outputGateways, outputGateway)
		}

		outputSubnets = append(outputSubnets, map[string]any{
			"id":       subnet["id"],
			"name":     subnet["name"],
			"gateways": outputGateways,
		})
	}

	networkObject := map[string]any{
		"id":          foundNetwork.ID,
		"openstackId": foundRegionNetwork.OpenstackID,
		"region":      foundRegionNetwork.Region,
		"subnets":     outputSubnets,
	}

	display.OutputInfo(&flags.OutputFormatConfig, networkObject, "✅ Network %s created successfully (Openstack ID: %s)", foundNetwork.ID, foundRegionNetwork.OpenstackID)
}

func DeletePrivateNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private network: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private network %s deleted successfully", args[0])
}

func DeletePrivateNetworkRegion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/region/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private network region: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private network %s region %s deleted successfully", args[0], args[1])
}

func AddPrivateNetworkRegion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/region", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, map[string]string{"region": args[1]}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to add private network region: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private network %s region %s added successfully", args[0], args[1])
}

func ListPrivateNetworkSubnets(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "cidr", "gatewayIp", "dhcpEnabled"}, flags.GenericFilters)
}

func GetPrivateNetworkSubnet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet", projectID, url.PathEscape(args[0]))
	common.ManageObjectRequest(endpoint, args[1], "")
}

func CreatePrivateNetworkSubnet(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	subnet, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/network/private/{networkId}/subnet",
		fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet", projectID, url.PathEscape(args[0])),
		PrivateNetworkSubnetCreationExample,
		CloudNetworkSubnetSpec,
		assets.CloudOpenapiSchema,
		[]string{"dhcp", "network", "start", "end", "region", "noGateway"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create subnet: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, subnet, "✅ Subnet %s created successfully", subnet["id"])
}

func EditPrivateNetworkSubnet(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/network/private/{networkId}/subnet/{subnetId}",
		endpoint,
		CloudNetworkSubnetEditSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func DeletePrivateNetworkSubnet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/cloud/project/%s/network/private/%s/subnet/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private network subnet: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private network %s subnet %s deleted successfully", args[0], args[1])
}

func ListPublicNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var body []map[string]any
	err = httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/network/public", projectID), &body)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch results: %s", err)
		return
	}

	flattenedBody := []map[string]any{}

	for _, line := range body {
		regions := line["regions"].([]any)

		for _, region := range regions {
			region := region.(map[string]any)

			flattenedBody = append(flattenedBody, map[string]any{
				"id":          line["id"],
				"name":        line["name"],
				"vlanId":      line["vlanId"],
				"openstackId": region["openstackId"],
				"region":      region["region"],
				"status":      region["status"],
				"type":        line["type"],
			})
		}
	}

	body, err = filtersLib.FilterLines(flattenedBody, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetPublicNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var allNetworks []map[string]any
	err = httpLib.Client.Get(fmt.Sprintf("/v1/cloud/project/%s/network/public", projectID), &allNetworks)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch public networks: %s", err)
		return
	}

	var object map[string]any
	for _, network := range allNetworks {
		networkID := network["id"].(string)
		if networkID == args[0] {
			object = network
			break
		}
	}

	if object == nil {
		display.OutputError(&flags.OutputFormatConfig, "no public network found with ID %q", args[0])
		return
	}

	for _, region := range object["regions"].([]any) {
		region := region.(map[string]any)

		// Fetch subnets of region network
		path := fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s/subnet",
			projectID,
			url.PathEscape(region["region"].(string)),
			url.PathEscape(region["openstackId"].(string)),
		)
		var subnets []map[string]any
		if err := httpLib.Client.Get(path, &subnets); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", path, err)
			return
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPublicTemplate, &flags.OutputFormatConfig)
}

func ListGateways(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with network feature available: %s", err)
		return
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	gateways, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/gateway", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch gateways: %s", err)
		return
	}

	// Flatten gateways in a single array
	var allGateways []map[string]any
	for _, regionGateways := range gateways {
		allGateways = append(allGateways, regionGateways...)
	}

	// Filter results
	allGateways, err = filtersLib.FilterLines(allGateways, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allGateways, cloudprojectGatewayColumnsToDisplay, &flags.OutputFormatConfig)
}

func findGateway(gatewayId string) (string, map[string]any, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", nil, err
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with network feature available: %w", err)
	}

	// Search for the given gateway in all regions
	// TODO: speed up with parallel search or by adding a required region argument
	for _, region := range regions {
		var (
			gateway  map[string]any
			endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway/%s",
				projectID, url.PathEscape(region.(string)), url.PathEscape(gatewayId))
		)
		if err := httpLib.Client.Get(endpoint, &gateway); err == nil {
			return endpoint, gateway, nil
		}
	}

	return "", nil, errors.New("no gateway found with given ID")
}

func GetGateway(_ *cobra.Command, args []string) {
	_, foundGateway, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(foundGateway, args[0], cloudGatewayTemplate, &flags.OutputFormatConfig)
}

func EditGateway(cmd *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/gateway/{id}",
		foundURL,
		CloudGatewaySpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateGateway(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Transform CLI flags into the CloudGatewaySpec structure
	for _, allocationPool := range CloudGatewaySpec.Network.Subnet.CliAllocationPools {
		parts := strings.Split(allocationPool, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid allocation pool format, expected start:end, got %s", allocationPool)
			return
		}
		CloudGatewaySpec.Network.Subnet.AllocationPools = append(CloudGatewaySpec.Network.Subnet.AllocationPools, PrivateNetworkAllocationPool{
			Start: parts[0],
			End:   parts[1],
		})
	}
	for _, hostRoute := range CloudGatewaySpec.Network.Subnet.CliHostRoutes {
		parts := strings.Split(hostRoute, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid host route format, expected destination:nextHop, got %s", hostRoute)
			return
		}
		CloudGatewaySpec.Network.Subnet.HostRoutes = append(CloudGatewaySpec.Network.Subnet.HostRoutes, PrivateNetworkHostRoute{
			Destination: parts[0],
			NextHop:     parts[1],
		})
	}

	var (
		endpoint, path string
		region         = args[0]
	)
	if CloudGatewaySpec.ExistingNetworkID != "" {
		path = "/cloud/project/{serviceName}/region/{regionName}/network/{networkId}/subnet/{subnetId}/gateway"
		endpoint = fmt.Sprintf(
			"/v1/cloud/project/%s/region/%s/network/%s/subnet/%s/gateway",
			projectID, url.PathEscape(region), url.PathEscape(CloudGatewaySpec.ExistingNetworkID),
			url.PathEscape(CloudGatewaySpec.ExistingSubnetID))
	} else {
		path = "/cloud/project/{serviceName}/region/{regionName}/gateway"
		endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/gateway", projectID, url.PathEscape(region))
	}

	// Create resource
	task, err := common.CreateResource(
		cmd,
		path,
		endpoint,
		GatewayCreationExample,
		CloudGatewaySpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "model"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create gateway: %s", err)
		return
	}

	// Wait for task to complete if --wait flag is set
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, task, `⚡️ Gateway creation started successfully (operation ID: %s)
You can check the status of the operation with: 'ovhcloud cloud operation get %s'`, task["id"], task["id"])
		return
	}

	gatewayID, err := waitForCloudOperation(projectID, task["id"].(string), "gateway#create", 30*time.Minute)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for gateway creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Gateway %s created successfully", gatewayID)
}

func DeleteGateway(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(foundURL, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete gateway: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Gateway %s deleted successfully", args[0])
}

func ExposeGateway(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(foundURL+"/expose", nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to expose gateway: %s", err)
		return
	}

	log.Printf("✅ Gateway %s exposed successfully", args[0])

	// Display updated gateway information
	var object map[string]any
	if err := httpLib.Client.Get(foundURL, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching %s: %s", foundURL, err)
		return
	}
	display.OutputObject(object, args[0], cloudGatewayTemplate, &flags.OutputFormatConfig)
}

func ListGatewayInterfaces(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(foundURL+"/interface", []string{"id", "ip", "networkId", "subnetId"}, flags.GenericFilters)
}

func GetGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(foundURL+"/interface", args[1], "")
}

func CreateGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Post(
		foundURL+"/interface",
		GatewayInterfaceSpec,
		nil,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create gateway interface: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Gateway %s interface created successfully", args[0])
}

func DeleteGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(foundURL+"/interface/"+url.PathEscape(args[1]), nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete gateway interface: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Gateway %s interface %s deleted successfully", args[0], args[1])
}
