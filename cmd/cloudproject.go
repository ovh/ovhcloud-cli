
package cmd

import (
	"fmt"
	"io"
	"log"
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
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /cloud/project: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, cloudprojectColumnsToDisplay, jsonOutput, yamlOutput)
}

func getCloudProject(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/cloud/project/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching CloudProject: %s\n", err)
	}

	internal.OutputObject(object, cloudprojectColumnsToDisplay[0], jsonOutput, yamlOutput)
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
