// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudLoadbalancerCommand(cloudCmd *cobra.Command) {
	loadbalancerCmd := &cobra.Command{
		Use:   "loadbalancer",
		Short: "Manage loadbalancers in the given cloud project",
	}
	loadbalancerCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	loadbalancerListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your loadbalancers",
		Run:     cloud.ListCloudLoadbalancers,
	}
	loadbalancerCmd.AddCommand(withFilterFlag(loadbalancerListCmd))

	loadbalancerCmd.AddCommand(&cobra.Command{
		Use:   "get <loadbalancer_id>",
		Short: "Get a specific loadbalancer",
		Run:   cloud.GetCloudLoadbalancer,
		Args:  cobra.ExactArgs(1),
	})

	editLoadbalancerCmd := &cobra.Command{
		Use:   "edit <loadbalancer_id>",
		Short: "Edit the given loadbalancer",
		Run:   cloud.EditCloudLoadbalancer,
		Args:  cobra.ExactArgs(1),
	}
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.Name, "name", "", "Name of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.Description, "description", "", "Description of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateSpec.FlavorId, "flavor", "", "Flavor ID of the loadbalancer (can be retrieved with 'cloud reference loadbalancer list-flavors <region>')")
	addInteractiveEditorFlag(editLoadbalancerCmd)
	loadbalancerCmd.AddCommand(editLoadbalancerCmd)

	cloudCmd.AddCommand(loadbalancerCmd)
}
