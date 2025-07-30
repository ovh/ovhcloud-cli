package cloud

import (
	_ "embed"
	"errors"
	"fmt"
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
			EnableDhcp                  bool                           `json:"enableDhcp,omitempty"`
			EnableGatewayIp             bool                           `json:"enableGatewayIp,omitempty"`
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

	CloudNetworkSubnetSpec struct {
		Dhcp           bool   `json:"dhcp,omitempty"`
		DisableGateway bool   `json:"disableGateway,omitempty"`
		GatewayIp      string `json:"gatewayIp,omitempty"`
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
)

func ListPrivateNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	body, err := httpLib.FetchExpandedArray(fmt.Sprintf("/cloud/project/%s/network/private", projectID), "id")
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
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
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetPrivateNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}
	path := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s", path, err)
		return
	}

	for _, region := range object["regions"].([]any) {
		region := region.(map[string]any)

		// Skip regions without openstackId
		if openstackId, ok := region["openstackId"]; !ok || openstackId == nil {
			continue
		}

		// Fetch subnets of region network
		path = fmt.Sprintf("/cloud/project/%s/region/%s/network/%s/subnet",
			projectID,
			url.PathEscape(region["region"].(string)),
			url.PathEscape(region["openstackId"].(string)),
		)
		var subnets []map[string]any
		if err := httpLib.Client.Get(path, &subnets); err != nil {
			display.ExitError("error fetching %s: %s", path, err)
			return
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPrivateTemplate, &flags.OutputFormatConfig)
}

func EditPrivateNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/network/private/{networkId}",
		fmt.Sprintf("/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0])),
		map[string]any{
			"name": CloudNetworkName,
		},
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func CreatePrivateNetwork(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Transform CLI flags into the CloudNetworkSpec structure
	for _, allocationPool := range CloudNetworkSpec.Subnet.CliAllocationPools {
		parts := strings.Split(allocationPool, ":")
		if len(parts) != 2 {
			display.ExitError("invalid allocation pool format, expected start:end, got %s", allocationPool)
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
			display.ExitError("invalid host route format, expected destination:nextHop, got %s", hostRoute)
			return
		}

		CloudNetworkSpec.Subnet.HostRoutes = append(CloudNetworkSpec.Subnet.HostRoutes, PrivateNetworkHostRoute{
			Destination: parts[0],
			NextHop:     parts[1],
		})
	}

	// Create resource
	region := args[0]
	endpoint := fmt.Sprintf("/cloud/project/%s/region/%s/network", projectID, url.PathEscape(region))
	task, err := common.CreateResource(
		"/cloud/project/{serviceName}/region/{regionName}/network",
		endpoint,
		PrivateNetworkCreationExample,
		CloudNetworkSpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "subnet"})
	if err != nil {
		display.ExitError("failed to create private network: %s", err)
		return
	}

	// Wait for task to complete if --wait flag is set
	if !flags.WaitForTask {
		fmt.Printf("\n⚡️ Network creation started successfully (operation ID: %s)\n", task["id"].(string))
		fmt.Printf("You can check the status of the operation with: `ovhcloud cloud operation get %s`\n", task["id"].(string))
		return
	}

	networkID, err := waitForCloudOperation(projectID, task["id"].(string), "network#create", 10*time.Minute)
	if err != nil {
		display.ExitError("failed to wait for network creation: %s", err)
		return
	}

	// Fetch all private networks
	var networks []struct {
		ID      string `json:"id"`
		Regions []struct {
			OpenstackID string `json:"openstackId"`
			Region      string `json:"region"`
		} `json:"regions"`
	}
	if err := httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/network/private", projectID), &networks); err != nil {
		display.ExitError("failed to fetch private networks: %s", err)
		return
	}

	// Find the created network
	for _, network := range networks {
		for _, regionDetails := range network.Regions {
			if regionDetails.OpenstackID == networkID && regionDetails.Region == region {
				fmt.Printf("\n✅ Network %s created successfully (Openstack ID: %s)\n", network.ID, regionDetails.OpenstackID)
				return
			}
		}
	}

	display.ExitError("created network not found, this is unexpected")
}

func DeletePrivateNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete private network: %s", err)
		return
	}

	fmt.Printf("✅ Private network %s deleted successfully\n", args[0])
}

func DeletePrivateNetworkRegion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/region/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete private network region: %s", err)
		return
	}

	fmt.Printf("✅ Private network %s region %s deleted successfully\n", args[0], args[1])
}

func AddPrivateNetworkRegion(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/region", projectID, url.PathEscape(args[0]))
	if err := httpLib.Client.Post(endpoint, map[string]string{"region": args[1]}, nil); err != nil {
		display.ExitError("failed to add private network region: %s", err)
		return
	}

	fmt.Printf("✅ Private network %s region %s added successfully\n", args[0], args[1])
}

func ListPrivateNetworkSubnets(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet", projectID, url.PathEscape(args[0]))

	common.ManageListRequestNoExpand(endpoint, []string{"id", "cidr", "gatewayIp", "dhcpEnabled"}, flags.GenericFilters)
}

func GetPrivateNetworkSubnet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet", projectID, url.PathEscape(args[0]))
	common.ManageObjectRequest(endpoint, args[1], "")
}

func EditPrivateNetworkSubnet(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/network/private/{networkId}/subnet/{subnetId}",
		endpoint,
		CloudNetworkSubnetSpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func DeletePrivateNetworkSubnet(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/network/private/%s/subnet/%s", projectID, url.PathEscape(args[0]), url.PathEscape(args[1]))
	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.ExitError("failed to delete private network subnet: %s", err)
		return
	}

	fmt.Printf("✅ Private network %s subnet %s deleted successfully\n", args[0], args[1])
}

func ListPublicNetworks(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	var body []map[string]any
	err = httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &body)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
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
		display.ExitError("failed to filter results: %s", err)
		return
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetPublicNetwork(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	var allNetworks []map[string]any
	err = httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &allNetworks)
	if err != nil {
		display.ExitError("failed to fetch public networks: %s", err)
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
		display.ExitError("no public network found with ID %q", args[0])
		return
	}

	for _, region := range object["regions"].([]any) {
		region := region.(map[string]any)

		// Fetch subnets of region network
		path := fmt.Sprintf("/cloud/project/%s/region/%s/network/%s/subnet",
			projectID,
			url.PathEscape(region["region"].(string)),
			url.PathEscape(region["openstackId"].(string)),
		)
		var subnets []map[string]any
		if err := httpLib.Client.Get(path, &subnets); err != nil {
			display.ExitError("error fetching %s: %s", path, err)
			return
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPublicTemplate, &flags.OutputFormatConfig)
}

func ListGateways(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		display.ExitError("failed to fetch regions with network feature available: %s", err)
		return
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	gateways, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/gateway", regions, true)
	if err != nil {
		display.ExitError("failed to fetch gateways: %s", err)
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
		display.ExitError("failed to filter results: %s", err)
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
			endpoint = fmt.Sprintf("/cloud/project/%s/region/%s/gateway/%s",
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
		display.ExitError(err.Error())
		return
	}

	display.OutputObject(foundGateway, args[0], cloudGatewayTemplate, &flags.OutputFormatConfig)
}

func EditGateway(cmd *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := common.EditResource(
		cmd,
		"/cloud/project/{serviceName}/region/{regionName}/gateway/{id}",
		foundURL,
		CloudGatewaySpec,
		assets.CloudOpenapiSchema,
	); err != nil {
		display.ExitError(err.Error())
		return
	}
}

func CreateGateway(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	// Transform CLI flags into the CloudGatewaySpec structure
	for _, allocationPool := range CloudGatewaySpec.Network.Subnet.CliAllocationPools {
		parts := strings.Split(allocationPool, ":")
		if len(parts) != 2 {
			display.ExitError("invalid allocation pool format, expected start:end, got %s", allocationPool)
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
			display.ExitError("invalid host route format, expected destination:nextHop, got %s", hostRoute)
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
			"/cloud/project/%s/region/%s/network/%s/subnet/%s/gateway",
			projectID, url.PathEscape(region), url.PathEscape(CloudGatewaySpec.ExistingNetworkID),
			url.PathEscape(CloudGatewaySpec.ExistingSubnetID))
	} else {
		path = "/cloud/project/{serviceName}/region/{regionName}/gateway"
		endpoint = fmt.Sprintf("/cloud/project/%s/region/%s/gateway", projectID, url.PathEscape(region))
	}

	// Create resource
	task, err := common.CreateResource(
		path,
		endpoint,
		GatewayCreationExample,
		CloudGatewaySpec,
		assets.CloudOpenapiSchema,
		[]string{"name", "model"})
	if err != nil {
		display.ExitError("failed to create gateway: %s", err)
		return
	}

	// Wait for task to complete if --wait flag is set
	if !flags.WaitForTask {
		fmt.Printf("\n⚡️ Gateway creation started successfully (operation ID: %s)\n", task["id"])
		fmt.Printf("You can check the status of the operation with: `ovhcloud cloud operation get %s`\n", task["id"])
		return
	}

	gatewayID, err := waitForCloudOperation(projectID, task["id"].(string), "gateway#create", 30*time.Minute)
	if err != nil {
		display.ExitError("failed to wait for gateway creation: %s", err)
		return
	}

	fmt.Printf("\n✅ Gateway %s created successfully\n", gatewayID)
}

func DeleteGateway(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Delete(foundURL, nil); err != nil {
		display.ExitError("failed to delete gateway: %s", err)
		return
	}

	fmt.Printf("✅ Gateway %s deleted successfully\n", args[0])
}

func ExposeGateway(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Post(foundURL+"/expose", nil, nil); err != nil {
		display.ExitError("failed to expose gateway: %s", err)
		return
	}

	fmt.Printf("✅ Gateway %s exposed successfully\n", args[0])

	// Display updated gateway information
	var object map[string]any
	if err := httpLib.Client.Get(foundURL, &object); err != nil {
		display.ExitError("error fetching %s: %s", foundURL, err)
		return
	}
	display.OutputObject(object, args[0], cloudGatewayTemplate, &flags.OutputFormatConfig)
}

func ListGatewayInterfaces(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageListRequestNoExpand(foundURL+"/interface", []string{"id", "ip", "networkId", "subnetId"}, flags.GenericFilters)
}

func GetGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	common.ManageObjectRequest(foundURL+"/interface", args[1], "")
}

func CreateGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Post(
		foundURL+"/interface",
		GatewayInterfaceSpec,
		nil,
	); err != nil {
		display.ExitError("failed to create gateway interface: %s", err)
		return
	}

	fmt.Printf("✅ Gateway %s interface created successfully\n", args[0])
}

func DeleteGatewayInterface(_ *cobra.Command, args []string) {
	foundURL, _, err := findGateway(args[0])
	if err != nil {
		display.ExitError(err.Error())
		return
	}

	if err := httpLib.Client.Delete(foundURL+"/interface/"+url.PathEscape(args[1]), nil); err != nil {
		display.ExitError("failed to delete gateway interface: %s", err)
		return
	}

	fmt.Printf("✅ Gateway %s interface %s deleted successfully\n", args[0], args[1])
}
