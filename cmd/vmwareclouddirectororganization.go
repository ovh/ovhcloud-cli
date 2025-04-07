
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
	vmwareclouddirectororganizationColumnsToDisplay = []string{ "id","currentState.fullName","currentState.region","resourceStatus" }
)

func listVmwareCloudDirectorOrganization(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/v2/vmwareCloudDirector/organization", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /v2/vmwareCloudDirector/organization: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, vmwareclouddirectororganizationColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVmwareCloudDirectorOrganization(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/v2/vmwareCloudDirector/organization/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching VmwareCloudDirectorOrganization: %s\n", err)
	}

	internal.OutputObject(object, vmwareclouddirectororganizationColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirectororganization",
		Short: "Retrieve information and manage your VmwareCloudDirectorOrganization services",
	}

	// Command to list VmwareCloudDirectorOrganization services
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorOrganization services",
		Run:   listVmwareCloudDirectorOrganization,
	})

	// Command to get a single VmwareCloudDirectorOrganization
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VmwareCloudDirectorOrganization",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVmwareCloudDirectorOrganization,
	})

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
