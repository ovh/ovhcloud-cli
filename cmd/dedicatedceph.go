
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	dedicatedcephColumnsToDisplay = []string{ "serviceName","region","state","status" }
)

func listDedicatedCeph(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/ceph", dedicatedcephColumnsToDisplay)
}

func getDedicatedCeph(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/ceph", args[0], dedicatedcephColumnsToDisplay[0])
}

func init() {
	dedicatedcephCmd := &cobra.Command{
		Use:   "dedicatedceph",
		Short: "Retrieve information and manage your DedicatedCeph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCeph services",
		Run:   listDedicatedCeph,
	})

	// Command to get a single DedicatedCeph
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DedicatedCeph",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDedicatedCeph,
	})

	rootCmd.AddCommand(dedicatedcephCmd)
}
