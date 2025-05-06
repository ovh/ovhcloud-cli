package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	dedicatedcephColumnsToDisplay = []string{"serviceName", "region", "state", "status"}

	//go:embed templates/dedicatedceph.tmpl
	dedicatedcephTemplate string
)

func listDedicatedCeph(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/ceph", "", dedicatedcephColumnsToDisplay, genericFilters)
}

func getDedicatedCeph(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/ceph", args[0], dedicatedcephTemplate)
}

func init() {
	dedicatedcephCmd := &cobra.Command{
		Use:   "dedicatedceph",
		Short: "Retrieve information and manage your DedicatedCeph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DedicatedCeph services",
		Run:   listDedicatedCeph,
	}
	dedicatedcephCmd.AddCommand(withFilterFlag(dedicatedcephListCmd))

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
