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
	vrackservicesColumnsToDisplay = []string{"id", "currentState.region", "currentState.productStatus", "resourceStatus"}
)

func listVrackServices(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/v2/vrackServices/resource", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /v2/vrackServices/resource: %s", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, vrackservicesColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVrackServices(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/v2/vrackServices/resource/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching VrackServices: %s\n", err)
	}

	internal.OutputObject(object, vrackservicesColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vrackservicesCmd := &cobra.Command{
		Use:   "vrackservices",
		Short: "Retrieve information and manage your VrackServices services",
	}

	// Command to list VrackServices services
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VrackServices services",
		Run:   listVrackServices,
	})

	// Command to get a single VrackServices
	vrackservicesCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VrackServices",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVrackServices,
	})

	rootCmd.AddCommand(vrackservicesCmd)
}
