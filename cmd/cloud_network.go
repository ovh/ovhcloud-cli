package cmd

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/display"
	filtersLib "stash.ovh.net/api/ovh-cli/internal/filters"
)

var (
	cloudprojectNetworkColumnsToDisplay = []string{"id", "name", "openstackId", "region", "status"}

	//go:embed templates/cloud_network_private.tmpl
	cloudNetworkPrivateTemplate string

	//go:embed templates/cloud_network_public.tmpl
	cloudNetworkPublicTemplate string
)

func listCloudPrivateNetworks(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	body, err := fetchExpandedArray(fmt.Sprintf("/cloud/project/%s/network/private", projectID), "id")
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

	body, err = filtersLib.FilterLines(flattenedBody, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &outputFormatConfig)
}

func getCloudPrivateNetwork(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())
	path := fmt.Sprintf("/cloud/project/%s/network/private/%s", projectID, url.PathEscape(args[0]))

	var object map[string]any
	if err := client.Get(path, &object); err != nil {
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
		if err := client.Get(path, &subnets); err != nil {
			display.ExitError("error fetching %s: %s", path, err)
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPrivateTemplate, &outputFormatConfig)
}

func listCloudPublicNetworks(_ *cobra.Command, _ []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var body []map[string]any
	err := client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &body)
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

	body, err = filtersLib.FilterLines(flattenedBody, genericFilters)
	if err != nil {
		display.ExitError("failed to filter results: %s", err)
	}

	display.RenderTable(body, cloudprojectNetworkColumnsToDisplay, &outputFormatConfig)
}

func getCloudPublicNetwork(_ *cobra.Command, args []string) {
	projectID := url.PathEscape(getConfiguredCloudProject())

	var allNetworks []map[string]any
	err := client.Get(fmt.Sprintf("/cloud/project/%s/network/public", projectID), &allNetworks)
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
		if err := client.Get(path, &subnets); err != nil {
			display.ExitError("error fetching %s: %s", path, err)
		}

		region["subnets"] = subnets
	}

	display.OutputObject(object, args[0], cloudNetworkPublicTemplate, &outputFormatConfig)
}

func initCloudNetworkCommand(cloudCmd *cobra.Command) {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks in the given cloud project",
	}
	networkCmd.PersistentFlags().StringVar(&cloudProject, "cloud-project", "", "Cloud project ID")
	cloudCmd.AddCommand(networkCmd)

	privateNetworkCmd := &cobra.Command{
		Use:   "private",
		Short: "Manage private networks in the given cloud project",
	}
	networkCmd.AddCommand(privateNetworkCmd)

	privateNetworkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your private networks",
		Run:   listCloudPrivateNetworks,
	}
	privateNetworkCmd.AddCommand(withFilterFlag(privateNetworkListCmd))

	privateNetworkCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific private network",
		Run:        getCloudPrivateNetwork,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"network_id"},
	})

	publicNetworkCmd := &cobra.Command{
		Use:   "public",
		Short: "Manage public networks in the given cloud project",
	}
	networkCmd.AddCommand(publicNetworkCmd)

	publicNetworkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your public networks",
		Run:   listCloudPublicNetworks,
	}
	publicNetworkCmd.AddCommand(withFilterFlag(publicNetworkListCmd))

	publicNetworkCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Get a specific public network",
		Run:        getCloudPublicNetwork,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"network_id"},
	})
}
