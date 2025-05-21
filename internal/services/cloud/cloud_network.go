package cloud

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	httpLib "stash.ovh.net/api/ovh-cli/internal/http"
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
)

func ListCloudPrivateNetworks(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	body, err := httpLib.FetchExpandedArray(fmt.Sprintf("/cloud/project/%s/network/private", projectID), "id")
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
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
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudPrivateNetwork(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	path := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(path, &object); err != nil {
		display.ExitError("error fetching %s: %s", path, err)
	}

	for _, region := range object["regions"].([]any) {
		region := region.(map[string]any)

		// Fetch subnets of region network
		path = fmt.Sprintf("/cloud/project/%s/region/%s/network/%s/subnet",
			projectID,
			url.PathEscape(region["region"].(string)),
			url.PathEscape(region["openstackId"].(string)),
		)
		var subnets []map[string]any
		if err := httpLib.Client.Get(path, &subnets); err != nil {
			display.ExitError("error fetching %s: %s", path, err)
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPrivateTemplate, &flags.OutputFormatConfig)
}

func ListCloudPublicNetworks(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var body []map[string]any
	err := httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &body)
	if err != nil {
		display.ExitError("failed to fetch results: %s", err)
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
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudPublicNetwork(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var allNetworks []map[string]any
	err := httpLib.Client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &allNetworks)
	if err != nil {
		display.ExitError("failed to fetch public networks: %s", err)
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
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPublicTemplate, &flags.OutputFormatConfig)
}

func ListCloudGateways(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		display.ExitError("failed to fetch regions with network feature available: %s", err)
	}

	// Fetch gateways in all regions
	url := fmt.Sprintf("/cloud/project/%s/region", projectID)
	gateways, err := httpLib.FetchObjectsParallel[[]map[string]any](url+"/%s/gateway", regions, true)
	if err != nil {
		display.ExitError("failed to fetch gateways: %s", err)
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
	}

	display.RenderTable(allGateways, cloudprojectGatewayColumnsToDisplay, &flags.OutputFormatConfig)
}

func GetCloudGateway(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	// Fetch regions with network feature available
	regions, err := getCloudRegionsWithFeatureAvailable(projectID, "network")
	if err != nil {
		display.ExitError("failed to fetch regions with network feature available: %s", err)
	}

	// Search for the given gateway in all regions
	// TODO: speed up with parallel search or by adding a required region argument
	var foundGateway map[string]any
	for _, region := range regions {
		url := fmt.Sprintf("/cloud/project/%s/region/%s/gateway/%s",
			projectID, url.PathEscape(region.(string)), url.PathEscape(args[0]))
		if err := httpLib.Client.Get(url, &foundGateway); err == nil {
			break
		}
		foundGateway = nil
	}

	if foundGateway == nil {
		display.ExitError("no gateway found with given ID")
	}

	display.OutputObject(foundGateway, args[0], cloudGatewayTemplate, &flags.OutputFormatConfig)
}
