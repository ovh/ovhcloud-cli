package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initInstanceCommand(cloudCmd *cobra.Command) {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage instances in the given cloud project",
	}
	instanceCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	instanceListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your instances",
		Run:   cloud.ListInstances,
	}
	instanceCmd.AddCommand(withFilterFlag(instanceListCmd))

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "get <instance_id>",
		Short: "Get a specific instance",
		Run:   cloud.GetInstance,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(instanceCmd)
}
