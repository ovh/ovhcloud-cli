// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
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
	cloudprojectPrivateNetworkColumnsToDisplay = []string{"id", "name", "region", "visibility", "vlanId"}
	cloudprojectPublicNetworkColumnsToDisplay  = []string{"id", "name", "region"}
	cloudprojectGatewayColumnsToDisplay        = []string{"id", "currentState.name name", "currentState.location.region region", "resourceStatus"}

	//go:embed templates/cloud_network_private.tmpl
	cloudNetworkPrivateTemplate string

	//go:embed templates/cloud_network_public.tmpl
	cloudNetworkPublicTemplate string

	//go:embed templates/cloud_network_gateway.tmpl
	cloudGatewayTemplate string

	//go:embed templates/cloud_network_gateway_v2.tmpl
	cloudGatewayV2Template string

	//go:embed templates/cloud_network_private_subnet.tmpl
	cloudNetworkPrivateSubnetTemplate string

	// CloudNetworkRegionFilter is used to filter networks by region
	CloudNetworkRegionFilter string

	//go:embed parameter-samples/private-network-create.json
	PrivateNetworkCreationExample string

	//go:embed parameter-samples/private-network-subnet-create.json
	PrivateNetworkSubnetCreationExample string

	//go:embed parameter-samples/gateway-create.json
	GatewayCreationExample string

	// CloudGatewaySpec contains the parameters for creating or updating a v2 cloud gateway
	CloudGatewaySpec struct {
		TargetSpec CloudGatewayTargetSpec `json:"targetSpec"`
	}

	CloudNetworkSpec struct {
		Name   string `json:"name,omitempty"`
		VlanId int    `json:"vlanId,omitempty"`
	}

	CloudNetworkSubnetSpec struct {
		Name                        string                         `json:"name,omitempty"`
		Cidr                        string                         `json:"cidr,omitempty"`
		IPVersion                   int                            `json:"ipVersion,omitempty"`
		EnableDhcp                  bool                           `json:"enableDhcp"`
		EnableGatewayIp             bool                           `json:"enableGatewayIp"`
		GatewayIp                   string                         `json:"gatewayIp,omitempty"`
		DnsNameServers              []string                       `json:"dnsNameServers,omitempty"`
		UseDefaultPublicDNSResolver bool                           `json:"useDefaultPublicDNSResolver,omitempty"`
		AllocationPools             []PrivateNetworkAllocationPool `json:"allocationPools,omitempty"`
		HostRoutes                  []PrivateNetworkHostRoute      `json:"hostRoutes,omitempty"`

		CliAllocationPools []string `json:"-"`
		CliHostRoutes      []string `json:"-"`
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

	// CloudGatewayTargetSpec is the desired specification of a v2 gateway.
	CloudGatewayTargetSpec struct {
		Name            string                 `json:"name,omitempty"`
		Description     string                 `json:"description,omitempty"`
		Location        gatewayLocation        `json:"location,omitzero"`
		ExternalGateway gatewayExternalGateway `json:"externalGateway,omitzero"`
		Subnets         []gatewaySubnetRef     `json:"subnets,omitempty"`

		// CliSubnets holds subnet IDs passed on the command line; converted into Subnets.
		CliSubnets []string `json:"-"`
	}

	gatewayLocation struct {
		Region           string `json:"region,omitempty"`
		AvailabilityZone string `json:"availabilityZone,omitempty"`
	}

	gatewayExternalGateway struct {
		Enabled bool   `json:"enabled"`
		Model   string `json:"model,omitempty"`
	}

	gatewaySubnetRef struct {
		Id string `json:"id,omitempty"`
	}
)

func ListPrivateNetworks(_ *cobra.Command, _ []string) {
	listNetworksByVisibility("private", cloudprojectPrivateNetworkColumnsToDisplay)
}

func listNetworksByVisibility(visibility string, columns []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var regions []any

	// If a region filter is set, use only that region
	if CloudNetworkRegionFilter != "" {
		regions = []any{CloudNetworkRegionFilter}
	} else {
		// Fetch regions with network feature available
		regions, err = getCloudRegionsWithFeatureAvailable(projectID, "network")
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to fetch regions with network feature available: %s", err)
			return
		}
	}

	// Fetch networks in targeted regions
	baseURL := fmt.Sprintf("/v1/cloud/project/%s/region", projectID)
	networks, err := httpLib.FetchObjectsParallel[[]map[string]any](baseURL+"/%s/network", regions, true)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch networks: %s", err)
		return
	}

	// Flatten networks in a single array and filter by visibility
	var allNetworks []map[string]any
	for i, regionNetworks := range networks {
		for _, network := range regionNetworks {
			if v, ok := network["visibility"]; ok && v == visibility {
				// Ensure the region field is set for table display
				if _, ok := network["region"]; !ok {
					network["region"] = fmt.Sprint(regions[i])
				}
				allNetworks = append(allNetworks, network)
			}
		}
	}

	// Filter results
	allNetworks, err = filtersLib.FilterLines(allNetworks, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(allNetworks, columns, &flags.OutputFormatConfig)
}

func findNetwork(networkID string) (string, map[string]any, error) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		return "", nil, err
	}

	// If a region is provided, go directly to that region
	if CloudNetworkRegionFilter != "" {
		var (
			network  map[string]any
			endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s",
				projectID, url.PathEscape(CloudNetworkRegionFilter), url.PathEscape(networkID))
		)
		if err := httpLib.Client.Get(endpoint, &network); err != nil {
			return "", nil, fmt.Errorf("network %s not found in region %s: %w", networkID, CloudNetworkRegionFilter, err)
		}
		return endpoint, network, nil
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions with network feature available: %w", err)
	}

	// Search for the given network in all regions in parallel
	type networkResult struct {
		endpoint string
		network  map[string]any
	}

	var (
		wg     sync.WaitGroup
		result = make(chan networkResult, 1)
	)

	for _, region := range regions {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			var (
				network  map[string]any
				endpoint = fmt.Sprintf("/v1/cloud/project/%s/region/%s/network/%s",
					projectID, url.PathEscape(r), url.PathEscape(networkID))
			)
			if err := httpLib.Client.Get(endpoint, &network); err == nil {
				select {
				case result <- networkResult{endpoint: endpoint, network: network}:
				default:
				}
			}
		}(region.(string))
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait()
		close(result)
	}()

	if found, ok := <-result; ok {
		return found.endpoint, found.network, nil
	}

	return "", nil, errors.New("no network found with given ID")
}

func GetPrivateNetwork(_ *cobra.Command, args []string) {
	networkID := args[0]
	foundURL, object, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch subnets of the network
	var subnets []map[string]any
	if err := httpLib.Client.Get(foundURL+"/subnet", &subnets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching subnets: %s", err)
		return
	}

	object["subnets"] = subnets

	display.OutputObject(object, networkID, cloudNetworkPrivateTemplate, &flags.OutputFormatConfig)
}

func CreatePrivateNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
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
		[]string{"name"})
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

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Network %s created successfully in region %s", networkID, region)
}

func DeletePrivateNetwork(_ *cobra.Command, args []string) {
	networkID := args[0]
	foundURL, _, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if err := httpLib.Client.Delete(foundURL, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private network: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Private network %s deleted successfully", networkID)
}

func ListPrivateNetworkSubnets(_ *cobra.Command, args []string) {
	networkID := args[0]
	foundURL, _, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(foundURL+"/subnet", []string{"id", "name", "cidr", "gatewayIp", "dhcpEnabled", "ipVersion"}, flags.GenericFilters)
}

func GetPrivateNetworkSubnet(_ *cobra.Command, args []string) {
	networkID := args[0]
	subnetID := args[1]
	foundURL, _, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var object map[string]any
	if err := httpLib.Client.Get(foundURL+"/subnet/"+url.PathEscape(subnetID), &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching subnet %s: %s", subnetID, err)
		return
	}

	display.OutputObject(object, subnetID, cloudNetworkPrivateSubnetTemplate, &flags.OutputFormatConfig)
}

func CreatePrivateNetworkSubnet(cmd *cobra.Command, args []string) {
	networkID := args[0]
	foundURL, _, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Transform CLI flags into the CloudNetworkSubnetSpec structure
	for _, allocationPool := range CloudNetworkSubnetSpec.CliAllocationPools {
		parts := strings.Split(allocationPool, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid allocation pool format, expected start:end, got %s", allocationPool)
			return
		}

		CloudNetworkSubnetSpec.AllocationPools = append(CloudNetworkSubnetSpec.AllocationPools, PrivateNetworkAllocationPool{
			Start: parts[0],
			End:   parts[1],
		})
	}
	for _, hostRoute := range CloudNetworkSubnetSpec.CliHostRoutes {
		parts := strings.Split(hostRoute, ":")
		if len(parts) != 2 {
			display.OutputError(&flags.OutputFormatConfig, "invalid host route format, expected destination:nextHop, got %s", hostRoute)
			return
		}

		CloudNetworkSubnetSpec.HostRoutes = append(CloudNetworkSubnetSpec.HostRoutes, PrivateNetworkHostRoute{
			Destination: parts[0],
			NextHop:     parts[1],
		})
	}

	if CloudNetworkSubnetSpec.IPVersion == 0 && CloudNetworkSubnetSpec.Cidr != "" {
		CloudNetworkSubnetSpec.IPVersion = ipVersionFromCIDR(CloudNetworkSubnetSpec.Cidr)
	}

	subnet, err := common.CreateResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/network/{networkId}/subnet",
		foundURL+"/subnet",
		PrivateNetworkSubnetCreationExample,
		CloudNetworkSubnetSpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "cidr", "ipVersion"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create subnet: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, subnet, "✅ Subnet %s created successfully", subnet["id"])
}

func DeletePrivateNetworkSubnet(_ *cobra.Command, args []string) {
	networkID := args[0]
	subnetID := args[1]
	foundURL, _, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := foundURL + "/subnet/" + url.PathEscape(subnetID)
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete private network subnet: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Subnet %s deleted successfully from network %s", subnetID, networkID)
}

func ListPublicNetworks(_ *cobra.Command, _ []string) {
	listNetworksByVisibility("public", cloudprojectPublicNetworkColumnsToDisplay)
}

func GetPublicNetwork(_ *cobra.Command, args []string) {
	networkID := args[0]
	foundURL, object, err := findNetwork(networkID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Fetch subnets of the network
	var subnets []map[string]any
	if err := httpLib.Client.Get(foundURL+"/subnet", &subnets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "error fetching subnets: %s", err)
		return
	}

	object["subnets"] = subnets

	display.OutputObject(object, networkID, cloudNetworkPublicTemplate, &flags.OutputFormatConfig)
}

func ListGateways(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageListRequestNoExpand(
		fmt.Sprintf("/v2/publicCloud/project/%s/gateway", projectID),
		cloudprojectGatewayColumnsToDisplay,
		flags.GenericFilters,
	)
}

// findGateway searches for a v1 gateway across all regions. It is kept to serve
// the commands that have no Cloud API v2 equivalent yet (expose, interfaces).
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
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.ManageObjectRequest(
		fmt.Sprintf("/v2/publicCloud/project/%s/gateway", projectID),
		args[0],
		cloudGatewayV2Template,
	)
}

func EditGateway(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Convert subnet IDs given on the command line into the expected structure
	for _, subnetID := range CloudGatewaySpec.TargetSpec.CliSubnets {
		CloudGatewaySpec.TargetSpec.Subnets = append(CloudGatewaySpec.TargetSpec.Subnets, gatewaySubnetRef{Id: subnetID})
	}

	if err := common.EditResource(
		cmd,
		"/publicCloud/project/{projectId}/gateway/{gatewayId}",
		fmt.Sprintf("/v2/publicCloud/project/%s/gateway/%s", projectID, url.PathEscape(args[0])),
		CloudGatewaySpec,
		assets.CloudV2OpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func CreateGateway(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Convert subnet IDs given on the command line into the expected structure
	for _, subnetID := range CloudGatewaySpec.TargetSpec.CliSubnets {
		CloudGatewaySpec.TargetSpec.Subnets = append(CloudGatewaySpec.TargetSpec.Subnets, gatewaySubnetRef{Id: subnetID})
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/gateway", projectID)
	gateway, err := common.CreateResource(
		cmd,
		"/publicCloud/project/{projectId}/gateway",
		endpoint,
		GatewayCreationExample,
		CloudGatewaySpec,
		assets.CloudV2OpenapiSchema,
		[]string{"targetSpec"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create gateway: %s", err)
		return
	}

	gatewayID, _ := gateway["id"].(string)

	// Wait for the resource to be ready if --wait flag is set
	if !flags.WaitForTask {
		display.OutputInfo(&flags.OutputFormatConfig, gateway, "⚡️ Gateway creation started successfully (id: %s)", gatewayID)
		return
	}

	if err := waitForCloudResourceReady(fmt.Sprintf("%s/%s", endpoint, url.PathEscape(gatewayID)), 30*time.Minute); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to wait for gateway creation: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, gateway, "✅ Gateway %s created successfully", gatewayID)
}

func DeleteGateway(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v2/publicCloud/project/%s/gateway/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to delete gateway: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Gateway %s is being deleted…", args[0])
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

// ipVersionFromCIDR returns 4 or 6 based on the CIDR string.
// Returns 0 if the CIDR cannot be parsed.
func ipVersionFromCIDR(cidr string) int {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0
	}
	if ip.To4() != nil {
		return 4
	}
	return 6
}
