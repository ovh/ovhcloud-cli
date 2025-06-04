package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/dedicatedceph"
)

func init() {
	dedicatedcephCmd := &cobra.Command{
		Use:   "dedicated-ceph",
		Short: "Retrieve information and manage your Dedicated Ceph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Dedicated Ceph services",
		Run:   dedicatedceph.ListDedicatedCeph,
	}
	dedicatedcephCmd.AddCommand(withFilterFlag(dedicatedcephListCmd))

	// Command to get a single DedicatedCeph
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Dedicated Ceph",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatedceph.GetDedicatedCeph,
	})

	// Command to update a single DedicatedCeph
	dedicatedcephCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Dedicated Ceph",
		Run:   dedicatedceph.EditDedicatedCeph,
	})

	rootCmd.AddCommand(dedicatedcephCmd)
}
