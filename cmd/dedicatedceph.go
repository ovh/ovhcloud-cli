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
		Use:   "dedicated-ceph",
		Short: "Retrieve information and manage your Dedicated Ceph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Dedicated Ceph services",
		Run:   listDedicatedCeph,
	}
	dedicatedcephCmd.AddCommand(withFilterFlag(dedicatedcephListCmd))

	// Command to get a single DedicatedCeph
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Dedicated Ceph",
		Args:  cobra.ExactArgs(1),
		Run:   getDedicatedCeph,
	})

	rootCmd.AddCommand(dedicatedcephCmd)
}
