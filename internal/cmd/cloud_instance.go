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

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "start <instance_id>",
		Short: "Start the given instance",
		Run:   cloud.StartInstance,
		Args:  cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "stop <instance_id>",
		Short: "Stop the given instance",
		Run:   cloud.StopInstance,
		Args:  cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "shelve <instance_id>",
		Short: "Shelve the given instance",
		Long: `The resources dedicated to the Public Cloud instance are released.
The data of the local storage will be stored, the duration of the operation depends on the size of the local disk.
The instance can be unshelved at any time. Meanwhile hourly instances will not be billed.
The Snapshot Storage used to store the instance's data will be billed.`,
		Run:  cloud.ShelveInstance,
		Args: cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "unshelve <instance_id>",
		Short: "Unshelve the given instance",
		Long: `The resources dedicated to the Public Cloud instance are restored.
The duration of the operation depends on the size of the local disk.
Instance billing will get back to normal and the snapshot used to store the instance's data will be deleted.`,
		Run:  cloud.UnshelveInstance,
		Args: cobra.ExactArgs(1),
	})

	instanceCmd.AddCommand(&cobra.Command{
		Use:   "resume <instance_id>",
		Short: "Resume the given suspended instance",
		Run:   cloud.ResumeInstance,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(instanceCmd)
}
