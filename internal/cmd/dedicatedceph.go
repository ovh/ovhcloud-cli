package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/dedicatedceph"
	"github.com/spf13/cobra"
)

func init() {
	dedicatedcephCmd := &cobra.Command{
		Use:   "dedicated-ceph",
		Short: "Retrieve information and manage your Dedicated Ceph services",
	}

	// Command to list DedicatedCeph services
	dedicatedcephListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Dedicated Ceph services",
		Run:     dedicatedceph.ListDedicatedCeph,
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
	editCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given Dedicated Ceph",
		Args:  cobra.ExactArgs(1),
		Run:   dedicatedceph.EditDedicatedCeph,
	}
	editCmd.Flags().StringVar(&dedicatedceph.DedicatedCephSpec.CrushTunables, "crush-tunables", "", "Tunables of cluster (ARGONAUT, BOBTAIL, DEFAULT, FIREFLY, HAMMER, JEWEL, LEGACY, OPTIMAL)")
	editCmd.Flags().StringVar(&dedicatedceph.DedicatedCephSpec.Label, "label", "", "Name of the cluster")
	addInteractiveEditorFlag(editCmd)
	dedicatedcephCmd.AddCommand(editCmd)

	rootCmd.AddCommand(dedicatedcephCmd)
}
