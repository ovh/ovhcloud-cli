// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

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

	// vRack-based private network commands (private > vrack)
	vrackPrivateNetworkCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Manage vRack-based private networks in the given cloud project",
	}
	vrackPrivateNetworkCmd.PersistentFlags().StringVar(&cloud.CloudNetworkRegionFilter, "region", "", "Filter by region or specify the region of the network")
	privateNetworkCmd.AddCommand(vrackPrivateNetworkCmd)

	privateNetworkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your private networks",
		Run:     cloud.ListPrivateNetworks,
	}
	vrackPrivateNetworkCmd.AddCommand(withFilterFlag(privateNetworkListCmd))

	privateNetworkGetCmd := &cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific private network",
		Run:   cloud.GetPrivateNetwork,
		Args:  cobra.ExactArgs(1),
	}
	vrackPrivateNetworkCmd.AddCommand(privateNetworkGetCmd)

	vrackPrivateNetworkCmd.AddCommand(getPrivateNetworkCreationCmd())

	vrackPrivateNetworkCmd.AddCommand(&cobra.Command{
		Use:   "delete <network_id>",
		Short: "Delete a specific private network",
		Run:   cloud.DeletePrivateNetwork,
		Args:  cobra.ExactArgs(1),
	})

	// Private network subnet commands
	privateNetworkSubnetCmd := &cobra.Command{
		Use:   "subnet",
		Short: "Manage subnets in a specific private network",
	}
	vrackPrivateNetworkCmd.AddCommand(privateNetworkSubnetCmd)

	subnetListCmd := &cobra.Command{
		Use:     "list <network_id>",
		Aliases: []string{"ls"},
		Short:   "List subnets in a private network",
		Run:     cloud.ListPrivateNetworkSubnets,
		Args:    cobra.ExactArgs(1),
	}
	privateNetworkSubnetCmd.AddCommand(withFilterFlag(subnetListCmd))

	subnetGetCmd := &cobra.Command{
		Use:   "get <network_id> <subnet_id>",
		Short: "Get a specific subnet in a private network",
		Run:   cloud.GetPrivateNetworkSubnet,
		Args:  cobra.ExactArgs(2),
	}
	privateNetworkSubnetCmd.AddCommand(subnetGetCmd)

	privateNetworkSubnetCmd.AddCommand(getSubnetCreationCmd())

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
	publicNetworkCmd.PersistentFlags().StringVar(&cloud.CloudNetworkRegionFilter, "region", "", "Filter by region or specify the region of the network")
	networkCmd.AddCommand(publicNetworkCmd)

	publicNetworkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your public networks",
		Run:     cloud.ListPublicNetworks,
	}
	publicNetworkCmd.AddCommand(withFilterFlag(publicNetworkListCmd))

	publicNetworkGetCmd := &cobra.Command{
		Use:   "get <network_id>",
		Short: "Get a specific public network",
		Run:   cloud.GetPublicNetwork,
		Args:  cobra.ExactArgs(1),
	}
	publicNetworkCmd.AddCommand(publicNetworkGetCmd)

	// Gateway commands
	gatewayCmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage gateways in the given cloud project",
	}
	networkCmd.AddCommand(gatewayCmd)

	gatewayCmd.AddCommand(getGatewayCreationCmd())

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

	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "delete <gateway_id>",
		Short: "Delete a specific gateway",
		Run:   cloud.DeleteGateway,
		Args:  cobra.ExactArgs(1),
	})

	gatewayCmd.AddCommand(&cobra.Command{
		Use:   "expose <gateway_id>",
		Short: "Expose gateway to public network by adding a public port on it",
		Run:   cloud.ExposeGateway,
		Args:  cobra.ExactArgs(1),
	})

	// Gateway interfaces commands
	gatewayInterfaceCmd := &cobra.Command{
		Use:   "interface",
		Short: "Manage interfaces of a specific gateway",
	}
	gatewayCmd.AddCommand(gatewayInterfaceCmd)

	gatewayInterfaceCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list <gateway_id>",
		Aliases: []string{"ls"},
		Short:   "List interfaces of a specific gateway",
		Run:     cloud.ListGatewayInterfaces,
		Args:    cobra.ExactArgs(1),
	}))

	gatewayInterfaceCmd.AddCommand(&cobra.Command{
		Use:   "get <gateway_id> <interface_id>",
		Short: "Get a specific interface of a gateway",
		Run:   cloud.GetGatewayInterface,
		Args:  cobra.ExactArgs(2),
	})

	gatewayInterfaceCreateCmd := &cobra.Command{
		Use:   "create <gateway_id>",
		Short: "Create a new interface for the given gateway",
		Run:   cloud.CreateGatewayInterface,
		Args:  cobra.ExactArgs(1),
	}
	gatewayInterfaceCreateCmd.Flags().StringVar(&cloud.GatewayInterfaceSpec.SubnetID, "subnet-id", "", "ID of the subnet to attach the interface to")
	gatewayInterfaceCreateCmd.MarkFlagRequired("subnet-id")
	gatewayInterfaceCmd.AddCommand(gatewayInterfaceCreateCmd)

	gatewayInterfaceCmd.AddCommand(&cobra.Command{
		Use:   "delete <gateway_id> <interface_id>",
		Short: "Delete a specific interface of a gateway",
		Run:   cloud.DeleteGatewayInterface,
		Args:  cobra.ExactArgs(2),
	})
}

func getPrivateNetworkCreationCmd() *cobra.Command {
	privateNetworkCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a private network in the given cloud project",
		Long: `Use this command to create a private network.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network private vrack create <region> --name MyNetwork

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network private vrack create <region> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network private vrack create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network private vrack create <region>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network private vrack create <region> --from-file ./params.json --name MyNetwork

3. Using your default text editor:

	ovhcloud cloud network private vrack create <region> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network private vrack create <region> --editor --name MyNetwork
`,
		Run:  cloud.CreatePrivateNetwork,
		Args: cobra.ExactArgs(1),
	}

	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.Name, "name", "", "Name of the private network")
	privateNetworkCreateCmd.Flags().IntVar(&cloud.CloudNetworkSpec.VlanId, "vlan-id", 0, "VLAN ID for the private network")

	// Common flags for other means to define parameters
	addParameterFileFlags(privateNetworkCreateCmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/network", "post", cloud.PrivateNetworkCreationExample, nil)
	addInteractiveEditorFlag(privateNetworkCreateCmd)
	privateNetworkCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for network creation to be done before exiting")
	markFlagsMutuallyExclusive(privateNetworkCreateCmd, "from-file", "editor")

	return privateNetworkCreateCmd
}

func getSubnetCreationCmd() *cobra.Command {
	privateNetworkSubnetCreateCmd := &cobra.Command{
		Use:   "create <network_id>",
		Short: "Create a subnet in the given private network",
		Long: `Use this command to create a new subnet in a private network.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network private vrack subnet create <network_id> --region GRA9 --name MySubnet --cidr 192.168.1.0/24 --ip-version 4

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network private vrack subnet create <network_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network private vrack subnet create <network_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network private vrack subnet create <network_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network private vrack subnet create <network_id> --from-file ./params.json --name MySubnet

3. Using your default text editor:

	ovhcloud cloud network private vrack subnet create <network_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network private vrack subnet create <network_id> --editor --name MySubnet
`,
		Run:  cloud.CreatePrivateNetworkSubnet,
		Args: cobra.ExactArgs(1),
	}

	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.Name, "name", "", "Name of the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.Cidr, "cidr", "", "CIDR of the subnet (eg: 192.168.1.0/24)")
	privateNetworkSubnetCreateCmd.Flags().IntVar(&cloud.CloudNetworkSubnetSpec.IPVersion, "ip-version", 0, "IP version (4 or 6)")
	privateNetworkSubnetCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.EnableDhcp, "enable-dhcp", false, "Enable DHCP for the subnet")
	privateNetworkSubnetCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.EnableGatewayIp, "enable-gateway-ip", false, "Set a gateway IP for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.GatewayIp, "gateway-ip", "", "Gateway IP address for the subnet")
	privateNetworkSubnetCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.UseDefaultPublicDNSResolver, "use-default-public-dns-resolver", false, "Use default DNS resolver for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetSpec.DnsNameServers, "dns-name-servers", nil, "DNS name servers for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetSpec.CliAllocationPools, "allocation-pools", nil, "Allocation pools for the subnet in format start:end")
	privateNetworkSubnetCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetSpec.CliHostRoutes, "host-routes", nil, "Host routes for the subnet in format destination:nextHop")

	// Common flags for other means to define parameters
	addParameterFileFlags(privateNetworkSubnetCreateCmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/network/{networkId}/subnet", "post", cloud.PrivateNetworkSubnetCreationExample, nil)
	addInteractiveEditorFlag(privateNetworkSubnetCreateCmd)
	markFlagsMutuallyExclusive(privateNetworkSubnetCreateCmd, "from-file", "editor")

	return privateNetworkSubnetCreateCmd
}

func getGatewayCreationCmd() *cobra.Command {
	gatewayCreateCmd := &cobra.Command{
		Use:   "create <region>",
		Short: "Create a gateway in the given cloud project",
		Long: `Use this command to create a new gateway.

Two options are available to create a gateway:
	- Create a gateway in an existing private network
	- Create a gateway in a new private network

When creating a gateway in an existing private network, you must specify the network ID and subnet ID 
using the flags --network-id and --subnet-id.
In this case, only two parameters are supported and required: the gateway model and its name (respectively
--model and --name flags).

There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network gateway create <region> --name MyGateway --model xl

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network gateway create <region> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network gateway create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network gateway create <region>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network gateway create <region> --from-file ./params.json --name MyGateway

3. Using your default text editor:

	ovhcloud cloud network gateway create <region> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network gateway create <region> --editor --name MyGateway
`,
		Run:  cloud.CreateGateway,
		Args: cobra.ExactArgs(1),
	}

	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Model, "model", "", "Gateway model (s, m, l, xl, 2xl, 3xl)")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Name, "name", "", "Name of the gateway")

	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Network.Name, "network-name", "", "Name of the private network")
	gatewayCreateCmd.Flags().IntVar(&cloud.CloudGatewaySpec.Network.VlanId, "network-vlan-id", 0, "VLAN ID for the private network")

	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Network.Subnet.Name, "subnet-name", "", "Name of the subnet")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Network.Subnet.Cidr, "subnet-cidr", "", "CIDR of the subnet")
	gatewayCreateCmd.Flags().IntVar(&cloud.CloudGatewaySpec.Network.Subnet.IPVersion, "subnet-ip-version", 0, "IP version (4 or 6)")
	gatewayCreateCmd.Flags().BoolVar(&cloud.CloudGatewaySpec.Network.Subnet.EnableDhcp, "subnet-enable-dhcp", false, "Enable DHCP for the subnet")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.Network.Subnet.GatewayIp, "subnet-gateway-ip", "", "Gateway IP address for the subnet")
	gatewayCreateCmd.Flags().BoolVar(&cloud.CloudGatewaySpec.Network.Subnet.UseDefaultPublicDNSResolver, "subnet-use-default-public-dns-resolver", false, "Use default DNS resolver for the subnet")

	gatewayCreateCmd.Flags().StringSliceVar(&cloud.CloudGatewaySpec.Network.Subnet.DnsNameServers, "subnet-dns-name-servers", nil, "DNS name servers for the subnet")
	gatewayCreateCmd.Flags().StringSliceVar(&cloud.CloudGatewaySpec.Network.Subnet.CliAllocationPools, "subnet-allocation-pools", nil, "Allocation pools for the subnet in format start:end")
	gatewayCreateCmd.Flags().StringSliceVar(&cloud.CloudGatewaySpec.Network.Subnet.CliHostRoutes, "subnet-host-routes", nil, "Host routes for the subnet in format destination:nextHop")

	// Common flags for other means to define parameters
	addParameterFileFlags(gatewayCreateCmd, false, assets.CloudOpenapiSchema, "/cloud/project/{serviceName}/region/{regionName}/gateway", "post", cloud.GatewayCreationExample, nil)
	addInteractiveEditorFlag(gatewayCreateCmd)
	gatewayCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for gateway creation to be done before exiting")
	markFlagsMutuallyExclusive(gatewayCreateCmd, "from-file", "editor")

	// Add a flag to specify the network ID if creating in an existing private network
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.ExistingNetworkID, "network-id", "", "ID of the existing private network to create the gateway in")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.ExistingSubnetID, "subnet-id", "", "ID of the existing subnet to create the gateway in")
	markFlagsMutuallyExclusive(gatewayCreateCmd, "network-name", "network-id")
	markFlagsMutuallyExclusive(gatewayCreateCmd, "subnet-name", "subnet-id")

	return gatewayCreateCmd
}
