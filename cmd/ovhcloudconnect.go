
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	ovhcloudconnectColumnsToDisplay = []string{ "uuid","provider","status","description" }
)

func listOvhCloudConnect(_ *cobra.Command, _ []string) {
	manageListRequest("/ovhCloudConnect", ovhcloudconnectColumnsToDisplay)
}

func getOvhCloudConnect(_ *cobra.Command, args []string) {
	manageObjectRequest("/ovhCloudConnect", args[0], ovhcloudconnectColumnsToDisplay[0])
}

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your OvhCloudConnect services",
		Run:   listOvhCloudConnect,
	})

	// Command to get a single OvhCloudConnect
	ovhcloudconnectCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific OvhCloudConnect",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getOvhCloudConnect,
	})

	rootCmd.AddCommand(ovhcloudconnectCmd)
}
