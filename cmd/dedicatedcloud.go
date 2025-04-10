package cmd

import (
	"github.com/spf13/cobra"
)

var (
	dedicatedcloudColumnsToDisplay = []string{ "serviceName","location","state","description" }
)

func listDedicatedCloud(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicatedCloud", dedicatedcloudColumnsToDisplay, genericFilters)
}

func getDedicatedCloud(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicatedCloud", args[0], dedicatedcloudColumnsToDisplay[0])
}

func init() {
	dedicatedcloudCmd := &cobra.Command{
		Use:   "dedicatedcloud",
		Short: "Retrieve information and manage your DedicatedCloud services",
	}

	// Command to list DedicatedCloud services
	dedicatedcloudListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCloud services",
		Run:   listDedicatedCloud,
	}
	dedicatedcloudListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	dedicatedcloudCmd.AddCommand(dedicatedcloudListCmd)

	// Command to get a single DedicatedCloud
	dedicatedcloudCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCloud",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCloud,
	})

	rootCmd.AddCommand(dedicatedcloudCmd)
}
