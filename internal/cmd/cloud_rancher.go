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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Rancher services",
		Run:     cloud.ListCloudRanchers,
	}
	rancherCmd.AddCommand(withFilterFlag(rancherListCmd))

	rancherCmd.AddCommand(&cobra.Command{
		Use:   "get <rancher_id>",
		Short: "Get a specific Rancher service",
		Run:   cloud.GetRancher,
		Args:  cobra.ExactArgs(1),
	})

	editRancherCmd := &cobra.Command{
		Use:   "edit <rancher_id>",
		Short: "Edit the given Rancher service",
		Run:   cloud.EditRancher,
		Args:  cobra.ExactArgs(1),
	}
	editRancherCmd.Flags().StringVar(&cloud.RancherTargetSpec.Name, "name", "", "Name of the managed Rancher service")
	editRancherCmd.Flags().StringVar(&cloud.RancherTargetSpec.Plan, "plan", "", "Plan of the managed Rancher service (OVHCLOUD_EDITION, STANDARD)")
	editRancherCmd.Flags().StringVar(&cloud.RancherTargetSpec.Version, "version", "", "Version of the managed Rancher service")
	editRancherCmd.Flags().StringArrayVar(&cloud.RancherTargetSpec.CLIIPRestrictions, "ip-restrictions", nil, "List of IP restrictions (expected format: '<cidrBlock>,<description>')")
	addInteractiveEditorFlag(editRancherCmd)
	rancherCmd.AddCommand(editRancherCmd)

	cloudCmd.AddCommand(rancherCmd)
}
