package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	veeamcloudconnectColumnsToDisplay = []string{ "serviceName","productOffer","location","vmCount" }

	//go:embed templates/veeamcloudconnect.tmpl
	veeamcloudconnectTemplate string
)

func listVeeamCloudConnect(_ *cobra.Command, _ []string) {
	manageListRequest("/veeamCloudConnect", veeamcloudconnectColumnsToDisplay, genericFilters)
}

func getVeeamCloudConnect(_ *cobra.Command, args []string) {
	manageObjectRequest("/veeamCloudConnect", args[0], veeamcloudconnectTemplate)
}

func init() {
	veeamcloudconnectCmd := &cobra.Command{
		Use:   "veeamcloudconnect",
		Short: "Retrieve information and manage your VeeamCloudConnect services",
	}

	// Command to list VeeamCloudConnect services
	veeamcloudconnectListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VeeamCloudConnect services",
		Run:   listVeeamCloudConnect,
	}
	veeamcloudconnectListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	veeamcloudconnectCmd.AddCommand(veeamcloudconnectListCmd)

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
