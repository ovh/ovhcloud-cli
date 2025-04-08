
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	dedicatedclusterColumnsToDisplay = []string{ "id","region","model","status" }
)

func listDedicatedCluster(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/cluster", dedicatedclusterColumnsToDisplay)
}

func getDedicatedCluster(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/cluster", args[0], dedicatedclusterColumnsToDisplay[0])
}

func init() {
	dedicatedclusterCmd := &cobra.Command{
		Use:   "dedicatedcluster",
		Short: "Retrieve information and manage your DedicatedCluster services",
	}

	// Command to list DedicatedCluster services
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCluster services",
		Run:   listDedicatedCluster,
	})

	// Command to get a single DedicatedCluster
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCluster",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCluster,
	})

	rootCmd.AddCommand(dedicatedclusterCmd)
}
