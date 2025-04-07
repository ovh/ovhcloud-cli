
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
	vmwareclouddirectorbackupColumnsToDisplay = []string{ "id","iam.displayName","currentState.azName","resourceStatus" }
)

func listVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	req, err := client.NewRequest(http.MethodGet, "/v2/vmwareCloudDirector/backup", nil, true)
	if err != nil {
		log.Fatalf("error crafting request: %s\n", err)
	}

	req.Header.Set("X-Pagination-Mode", "CachedObjectList-Pages")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error fetching /v2/vmwareCloudDirector/backup: %s\n", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %s", err)
	}

	internal.OutputTable(bodyBytes, vmwareclouddirectorbackupColumnsToDisplay, jsonOutput, yamlOutput)
}

func getVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	var object map[string]any
	if err := client.Get(fmt.Sprintf("/v2/vmwareCloudDirector/backup/%s", url.PathEscape(args[0])), &object); err != nil {
		log.Fatalf("error fetching VmwareCloudDirectorBackup: %s\n", err)
	}

	internal.OutputObject(object, vmwareclouddirectorbackupColumnsToDisplay[0], jsonOutput, yamlOutput)
}

func init() {
	vmwareclouddirectorbackupCmd := &cobra.Command{
		Use:   "vmwareclouddirectorbackup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorBackup services",
		Run:   listVmwareCloudDirectorBackup,
	})

	// Command to get a single VmwareCloudDirectorBackup
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VmwareCloudDirectorBackup",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVmwareCloudDirectorBackup,
	})

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
