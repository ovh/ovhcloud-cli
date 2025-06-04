package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudRancherCommand(cloudCmd *cobra.Command) {
	rancherCmd := &cobra.Command{
		Use:   "rancher",
		Short: "Manage Rancher services in the given cloud project",
	}
	rancherCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	rancherListCmd := &cobra.Command{
		Use:   "list",
		Short: "List Rancher services",
		Run:   cloud.ListCloudRanchers,
	}
	rancherCmd.AddCommand(withFilterFlag(rancherListCmd))

	rancherCmd.AddCommand(&cobra.Command{
		Use:   "get <rancher_id>",
		Short: "Get a specific Rancher service",
		Run:   cloud.GetRancher,
		Args:  cobra.ExactArgs(1),
	})

	rancherCmd.AddCommand(&cobra.Command{
		Use:   "edit <rancher_id>",
		Short: "Edit the given Rancher service",
		Run:   cloud.EditRancher,
	})

	cloudCmd.AddCommand(rancherCmd)
}
