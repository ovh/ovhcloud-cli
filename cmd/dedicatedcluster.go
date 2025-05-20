package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	dedicatedclusterColumnsToDisplay = []string{"id", "region", "model", "status"}

	//go:embed templates/dedicatedcluster.tmpl
	dedicatedclusterTemplate string
)

func listDedicatedCluster(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/cluster", "", dedicatedclusterColumnsToDisplay, genericFilters)
}

func getDedicatedCluster(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/cluster", args[0], dedicatedclusterTemplate)
}

func init() {
	dedicatedclusterCmd := &cobra.Command{
		Use:   "dedicated-cluster",
		Short: "Retrieve information and manage your DedicatedCluster services",
	}

	// Command to list DedicatedCluster services
	dedicatedclusterListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCluster services",
		Run:   listDedicatedCluster,
	}
	dedicatedclusterCmd.AddCommand(withFilterFlag(dedicatedclusterListCmd))

	// Command to get a single DedicatedCluster
	dedicatedclusterCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific DedicatedCluster",
		Args:  cobra.ExactArgs(1),
		Run:   getDedicatedCluster,
	})

	rootCmd.AddCommand(dedicatedclusterCmd)
}
