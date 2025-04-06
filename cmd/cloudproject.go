
package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal"
)

var (
	cloudprojectColumnsToDisplay = []string{ "project_id","projectName","status","description" }
)

func listCloudProject(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/cloud/project", nil, true)
	if err != nil {
		fmt.Printf("error crafting request: %s\n", err)
		return
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error fetching /cloud/project: %s\n", err)
		return
	}

	var unmarshalled []map[string]any
	if err := client.UnmarshalResponse(resp, &unmarshalled); err != nil {
		fmt.Printf("error unmarshalling response: %s\n", err)
		return
	}

	internal.RenderTable(unmarshalled, cloudprojectColumnsToDisplay)
}

func getCloudProject(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/cloud/project/%s", url.PathEscape(args[0])), &object); err != nil {
		fmt.Printf("error fetching CloudProject: %s\n", err)
		return
	}

	internal.RenderObject(object, cloudprojectColumnsToDisplay[0])
}

func init() {
	cloudprojectCmd := &cobra.Command{
		Use:   "cloudproject",
		Short: "Retrieve information and manage your CloudProject services",
	}

	// Command to list CloudProject services
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your CloudProject services",
		Run:   listCloudProject,
	})

	// Command to get a single CloudProject
	cloudprojectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific CloudProject",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getCloudProject,
	})

	rootCmd.AddCommand(cloudprojectCmd)
}
