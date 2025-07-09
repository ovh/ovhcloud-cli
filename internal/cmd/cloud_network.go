package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/cloud"
)

func initCloudNetworkCommand(cloudCmd *cobra.Command) {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks in the given cloud project",
	}
	networkCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	cloudCmd.AddCommand(networkCmd)

	privateNetworkCmd := &cobra.Command{
		Use:   "private",
		Short: "Manage private networks in the given cloud project",
	}
	networkCmd.AddCommand(privateNetworkCmd)

	privateNetworkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your private networks",
		Run:   cloud.ListCloudPrivateNetworks,
	}
	privateNetworkCmd.AddCommand(withFilterFlag(privateNetworkListCmd))

	privateNetworkCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific private network",
		Run:   cloud.GetCloudPrivateNetwork,
		Args:  cobra.ExactArgs(1),
	})

	privateNetworkEditCmd := &cobra.Command{
		Use:   "edit <network_id>",
		Short: "Edit the given private network",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditCloudPrivateNetwork,
	}
	privateNetworkEditCmd.Flags().StringVar(&cloud.CloudNetworkName, "name", "", "Name of the private network")
	privateNetworkEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	privateNetworkCmd.AddCommand(privateNetworkEditCmd)

	publicNetworkCmd := &cobra.Command{
		Use:   "public",
		Short: "Manage public networks in the given cloud project",
	}
	networkCmd.AddCommand(publicNetworkCmd)

	publicNetworkListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your public networks",
		Run:   cloud.ListCloudPublicNetworks,
	}
	publicNetworkCmd.AddCommand(withFilterFlag(publicNetworkListCmd))

	publicNetworkCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific public network",
		Run:   cloud.GetCloudPublicNetwork,
		Args:  cobra.ExactArgs(1),
	})

	gatewayCmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage gateways in the given cloud project",
	}
	networkCmd.AddCommand(gatewayCmd)

	gatewayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your gateways",
		Run:   cloud.ListCloudGateways,
	}
	gatewayCmd.AddCommand(withFilterFlag(gatewayListCmd))

	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "get <gateway_id>",
		Short: "Get a specific gateway",
		Run:   cloud.GetCloudGateway,
		Args:  cobra.ExactArgs(1),
	})

	gatewayEditCmd := &cobra.Command{
		Use:   "edit <gateway_id>",
		Short: "Edit the given gateway",
		Run:   cloud.EditCloudGateway,
		Args:  cobra.ExactArgs(1),
	}
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Name, "name", "", "Name of the gateway")
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Model, "model", "", "Model of the gateway (s, m, l, xl, 2xl, 3xl)")
	gatewayEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define paraxmeters")
	gatewayCmd.AddCommand(gatewayEditCmd)
}
