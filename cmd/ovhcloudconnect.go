package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	ovhcloudconnectColumnsToDisplay = []string{"uuid", "provider", "status", "description"}

	//go:embed templates/ovhcloudconnect.tmpl
	ovhcloudconnectTemplate string
)

func listOvhCloudConnect(_ *cobra.Command, _ []string) {
	manageListRequest("/ovhCloudConnect", "", ovhcloudconnectColumnsToDisplay, genericFilters)
}

func getOvhCloudConnect(_ *cobra.Command, args []string) {
	manageObjectRequest("/ovhCloudConnect", args[0], ovhcloudconnectTemplate)
}

func init() {
	ovhcloudconnectCmd := &cobra.Command{
		Use:   "ovhcloudconnect",
		Short: "Retrieve information and manage your OvhCloudConnect services",
	}

	// Command to list OvhCloudConnect services
	ovhcloudconnectListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your OvhCloudConnect services",
		Run:   listOvhCloudConnect,
	}
	ovhcloudconnectListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	ovhcloudconnectCmd.AddCommand(ovhcloudconnectListCmd)

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
