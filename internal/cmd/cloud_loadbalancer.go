package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
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
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateFields.Name, "name", "", "Name of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateFields.Description, "description", "", "Description of the loadbalancer")
	editLoadbalancerCmd.Flags().StringVar(&cloud.CloudLoadbalancerUpdateFields.Size, "size", "", "Size of the load balancer (S, M, L)")
	editLoadbalancerCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define private network configuration parameters")
	loadbalancerCmd.AddCommand(editLoadbalancerCmd)

	cloudCmd.AddCommand(loadbalancerCmd)
}
