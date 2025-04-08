
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	veeamcloudconnectColumnsToDisplay = []string{ "serviceName","productOffer","location","vmCount" }
)

func listVeeamCloudConnect(_ *cobra.Command, _ []string) {
	manageListRequest("/veeamCloudConnect", veeamcloudconnectColumnsToDisplay)
}

func getVeeamCloudConnect(_ *cobra.Command, args []string) {
	manageObjectRequest("/veeamCloudConnect", args[0], veeamcloudconnectColumnsToDisplay[0])
}

func init() {
	veeamcloudconnectCmd := &cobra.Command{
		Use:   "veeamcloudconnect",
		Short: "Retrieve information and manage your VeeamCloudConnect services",
	}

	// Command to list VeeamCloudConnect services
	veeamcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VeeamCloudConnect services",
		Run:   listVeeamCloudConnect,
	})

	// Command to get a single VeeamCloudConnect
	veeamcloudconnectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VeeamCloudConnect",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVeeamCloudConnect,
	})

	rootCmd.AddCommand(veeamcloudconnectCmd)
}
