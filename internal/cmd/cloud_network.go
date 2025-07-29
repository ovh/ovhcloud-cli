package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudNetworkCommand(cloudCmd *cobra.Command) {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks in the given cloud project",
	}
	networkCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	cloudCmd.AddCommand(networkCmd)

	// Private network commands
	privateNetworkCmd := &cobra.Command{
		Use:   "private",
		Short: "Manage private networks in the given cloud project",
	}
	networkCmd.AddCommand(privateNetworkCmd)

	privateNetworkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your private networks",
		Run:     cloud.ListPrivateNetworks,
	}
	privateNetworkCmd.AddCommand(withFilterFlag(privateNetworkListCmd))

	privateNetworkCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific private network",
		Run:   cloud.GetPrivateNetwork,
		Args:  cobra.ExactArgs(1),
	})

	privateNetworkEditCmd := &cobra.Command{
		Use:   "edit <network_id>",
		Short: "Edit the given private network",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditPrivateNetwork,
	}
	privateNetworkEditCmd.Flags().StringVar(&cloud.CloudNetworkName, "name", "", "Name of the private network")
	addInteractiveEditorFlag(privateNetworkEditCmd)
	privateNetworkCmd.AddCommand(privateNetworkEditCmd)

	privateNetworkCmd.AddCommand(getPrivateNetworkCreationCmd())

	privateNetworkCmd.AddCommand(&cobra.Command{
		Use:   "delete <network_id>",
		Short: "Delete a specific private network",
		Run:   cloud.DeletePrivateNetwork,
		Args:  cobra.ExactArgs(1),
	})

	// Private network region commands
	privateNetworkRegionCmd := &cobra.Command{
		Use:   "region",
		Short: "Manage regions in a specific private network",
	}
	privateNetworkCmd.AddCommand(privateNetworkRegionCmd)

	privateNetworkRegionCmd.AddCommand(&cobra.Command{
		Use:   "delete <network_id> <region>",
		Short: "Delete the given region from a private network",
		Run:   cloud.DeletePrivateNetworkRegion,
		Args:  cobra.ExactArgs(2),
	})

	privateNetworkRegionCmd.AddCommand(&cobra.Command{
		Use:   "add <network_id> <region>",
		Short: "Add a region to a private network",
		Run:   cloud.AddPrivateNetworkRegion,
		Args:  cobra.ExactArgs(2),
	})

	// Private network subnet commands
	privateNetworkSubnetCmd := &cobra.Command{
		Use:   "subnet",
		Short: "Manage subnets in a specific private network",
	}
	privateNetworkCmd.AddCommand(privateNetworkSubnetCmd)

	privateNetworkSubnetCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <network_id>",
		Aliases: []string{"ls"},
		Short:   "List subnets in a private network",
		Run:     cloud.ListPrivateNetworkSubnets,
		Args:    cobra.ExactArgs(1),
	}))

	privateNetworkSubnetCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id> <subnet_id>",
		Short: "Get a specific subnet in a private network",
		Run:   cloud.GetPrivateNetworkSubnet,
		Args:  cobra.ExactArgs(2),
	})

	privateNetworkSubnetEditCmd := &cobra.Command{
		Use:   "edit <network_id> <subnet_id>",
		Short: "Edit a specific subnet in a private network",
		Run:   cloud.EditPrivateNetworkSubnet,
		Args:  cobra.ExactArgs(2),
	}
	privateNetworkSubnetEditCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.Dhcp, "enable-dhcp", false, "Enable DHCP (set to true if you don't want to set a default gateway IP)")
	privateNetworkSubnetEditCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.DisableGateway, "disable-gateway", false, "Set to true if you want to disable the default gateway")
	privateNetworkSubnetEditCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.GatewayIp, "gateway-ip", "", "Gateway IP address")
	addInteractiveEditorFlag(privateNetworkSubnetEditCmd)
	privateNetworkSubnetCmd.AddCommand(privateNetworkSubnetEditCmd)

	privateNetworkSubnetCmd.AddCommand(&cobra.Command{
		Use:   "delete <network_id> <subnet_id>",
		Short: "Delete a specific subnet in a private network",
		Run:   cloud.DeletePrivateNetworkSubnet,
		Args:  cobra.ExactArgs(2),
	})

	// Public network commands
	publicNetworkCmd := &cobra.Command{
		Use:   "public",
		Short: "Manage public networks in the given cloud project",
	}
	networkCmd.AddCommand(publicNetworkCmd)

	publicNetworkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your public networks",
		Run:     cloud.ListPublicNetworks,
	}
	publicNetworkCmd.AddCommand(withFilterFlag(publicNetworkListCmd))

	publicNetworkCmd.AddCommand(&cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific public network",
		Run:   cloud.GetPublicNetwork,
		Args:  cobra.ExactArgs(1),
	})

	gatewayCmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage gateways in the given cloud project",
	}
	networkCmd.AddCommand(gatewayCmd)

	gatewayListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your gateways",
		Run:     cloud.ListGateways,
	}
	gatewayCmd.AddCommand(withFilterFlag(gatewayListCmd))

	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "get <gateway_id>",
		Short: "Get a specific gateway",
		Run:   cloud.GetGateway,
		Args:  cobra.ExactArgs(1),
	})

	gatewayEditCmd := &cobra.Command{
		Use:   "edit <gateway_id>",
		Short: "Edit the given gateway",
		Run:   cloud.EditGateway,
		Args:  cobra.ExactArgs(1),
	}
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Name, "name", "", "Name of the gateway")
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Model, "model", "", "Model of the gateway (s, m, l, xl, 2xl, 3xl)")
	addInteractiveEditorFlag(gatewayEditCmd)
	gatewayCmd.AddCommand(gatewayEditCmd)
}

func getPrivateNetworkCreationCmd() *cobra.Command {
	privateNetworkCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a private network in the given cloud project",
		Run:   cloud.CreatePrivateNetwork,
		Args:  cobra.ExactArgs(1),
	}

	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Name, "name", "", "Name of the private network")
	privateNetworkCreateCmd.Flags().IntVar(&cloud.CloudNetworkSpec.VlanId, "vlan-id", 0, "VLAN ID for the private network")

	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Gateway.Model, "gateway-model", "", "Gateway model (s, m, l, xl, 2xl, 3xl)")
	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Gateway.Name, "gateway-name", "", "Name of the gateway")

	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Subnet.Name, "subnet-name", "", "Name of the subnet")
	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Subnet.Cidr, "subnet-cidr", "", "CIDR of the subnet")
	privateNetworkCreateCmd.Flags().IntVar(&cloud.CloudNetworkSpec.Subnet.IPVersion, "subnet-ip-version", 0, "IP version (4 or 6)")
	privateNetworkCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSpec.Subnet.EnableDhcp, "subnet-enable-dhcp", false, "Enable DHCP for the subnet")
	privateNetworkCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSpec.Subnet.EnableGatewayIp, "subnet-enable-gateway-ip", false, "Set a gateway ip for the subnet")
	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Subnet.GatewayIp, "subnet-gateway-ip", "", "Gateway IP address for the subnet")
	privateNetworkCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSpec.Subnet.UseDefaultPublicDNSResolver, "subnet-use-default-public-dns-resolver", false, "Use default DNS resolver for the subnet")

	privateNetworkCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSpec.Subnet.DnsNameServers, "subnet-dns-name-servers", nil, "DNS name servers for the subnet")
	privateNetworkCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSpec.Subnet.CliAllocationPools, "subnet-allocation-pools", nil, "Allocation pools for the subnet in format start:end")
	privateNetworkCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSpec.Subnet.CliHostRoutes, "subnet-host-routes", nil, "Host routes for the subnet in format destination:nextHop")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(privateNetworkCreateCmd, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/network", "post", cloud.PrivateNetworkCreationExample, nil)
	addInteractiveEditorFlag(privateNetworkCreateCmd)
	addFromFileFlag(privateNetworkCreateCmd)
	privateNetworkCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for network creation to be done before exiting")
	privateNetworkCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return privateNetworkCreateCmd
}
