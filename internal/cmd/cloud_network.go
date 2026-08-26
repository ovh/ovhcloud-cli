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

	// Security group commands (Cloud API v2)
	initCloudSecurityGroupCommand(networkCmd)

	// Private network commands
	privateNetworkCmd := &cobra.Command{
		Use:   "private",
		Short: "Manage private networks in the given cloud project",
	}
	networkCmd.AddCommand(privateNetworkCmd)

	// vRack-based private network commands (private > vrack), managed through
	// the Public Cloud API v2.
	vrackPrivateNetworkCmd := &cobra.Command{
		Use:   "vrack",
		Short: "Manage vRack-based private networks in the given cloud project",
	}
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

	privateNetworkEditCmd := &cobra.Command{
		Use:   "edit <network_id>",
		Short: "Edit a specific private network",
		Run:   cloud.EditPrivateNetwork,
		Args:  cobra.ExactArgs(1),
	}
	privateNetworkEditCmd.Flags().StringVar(&cloud.CloudNetworkEditSpec.TargetSpec.Name, "name", "", "Name of the private network")
	addInteractiveEditorFlag(privateNetworkEditCmd)
	vrackPrivateNetworkCmd.AddCommand(privateNetworkEditCmd)

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

	privateNetworkSubnetCmd.AddCommand(getSubnetEditCmd())

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
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Name, "name", "", "Name of the gateway")
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Description, "description", "", "Description of the gateway")
	gatewayEditCmd.Flags().BoolVar(&cloud.CloudGatewaySpec.TargetSpec.ExternalGateway.Enabled, "external-gateway-enabled", false, "Whether the external gateway is enabled")
	gatewayEditCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.ExternalGateway.Model, "external-gateway-model", "", "External gateway sizing model (S, M, L, XL, 2XL, 3XL)")
	gatewayEditCmd.Flags().StringSliceVar(&cloud.CloudGatewaySpec.TargetSpec.CliSubnets, "subnet", nil, "ID of a subnet to attach to the gateway (repeatable)")
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

	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.TargetSpec.Name, "name", "", "Name of the private network")
	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.TargetSpec.Description, "description", "", "Description of the private network")
	privateNetworkCreateCmd.Flags().IntVar(&cloud.CloudNetworkSpec.TargetSpec.VlanId, "vlan-id", 0, "VLAN ID for the private network (not supported on localzone regions)")
	privateNetworkCreateCmd.Flags().StringVar(&cloud.CloudNetworkSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")

	// Common flags for other means to define parameters
	addParameterFileFlags(privateNetworkCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/network", "post", cloud.PrivateNetworkCreationExample, nil)
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

	ovhcloud cloud network private vrack subnet create <network_id> --name MySubnet --cidr 192.168.1.0/24

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

	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.Name, "name", "", "Name of the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.Cidr, "cidr", "", "CIDR of the subnet (eg: 192.168.1.0/24)")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.Description, "description", "", "Description of the subnet")
	privateNetworkSubnetCreateCmd.Flags().BoolVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.DhcpEnabled, "dhcp-enabled", false, "Enable DHCP for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.GatewayIp, "gateway-ip", "", "Gateway IP address for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.DnsNameservers, "dns-nameservers", nil, "DNS nameservers for the subnet")
	privateNetworkSubnetCreateCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.CliAllocationPools, "allocation-pools", nil, "Allocation pools for the subnet in format start:end")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.Location.Region, "region", "", "Region of the subnet (defaults to the parent network region)")
	privateNetworkSubnetCreateCmd.Flags().StringVar(&cloud.CloudNetworkSubnetSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")

	// Common flags for other means to define parameters
	addParameterFileFlags(privateNetworkSubnetCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/network/{networkId}/subnet", "post", cloud.PrivateNetworkSubnetCreationExample, nil)
	addInteractiveEditorFlag(privateNetworkSubnetCreateCmd)
	privateNetworkSubnetCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for subnet creation to be done before exiting")
	markFlagsMutuallyExclusive(privateNetworkSubnetCreateCmd, "from-file", "editor")

	return privateNetworkSubnetCreateCmd
}

func getSubnetEditCmd() *cobra.Command {
	privateNetworkSubnetEditCmd := &cobra.Command{
		Use:   "edit <network_id> <subnet_id>",
		Short: "Edit a specific subnet in a private network",
		Run:   cloud.EditPrivateNetworkSubnet,
		Args:  cobra.ExactArgs(2),
	}

	privateNetworkSubnetEditCmd.Flags().StringVar(&cloud.CloudNetworkSubnetEditSpec.TargetSpec.Name, "name", "", "Name of the subnet")
	privateNetworkSubnetEditCmd.Flags().StringVar(&cloud.CloudNetworkSubnetEditSpec.TargetSpec.Description, "description", "", "Description of the subnet")
	privateNetworkSubnetEditCmd.Flags().StringVar(&cloud.CloudNetworkSubnetEditSpec.TargetSpec.GatewayIp, "gateway-ip", "", "Gateway IP address for the subnet")
	privateNetworkSubnetEditCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetEditSpec.TargetSpec.DnsNameservers, "dns-nameservers", nil, "DNS nameservers for the subnet")
	privateNetworkSubnetEditCmd.Flags().StringSliceVar(&cloud.CloudNetworkSubnetEditSpec.TargetSpec.CliAllocationPools, "allocation-pools", nil, "Allocation pools for the subnet in format start:end")

	var dhcpEnabled bool
	privateNetworkSubnetEditCmd.Flags().BoolVar(&dhcpEnabled, "dhcp-enabled", false, "Enable DHCP for the subnet")
	privateNetworkSubnetEditCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("dhcp-enabled") {
			cloud.CloudNetworkSubnetEditSpec.TargetSpec.DhcpEnabled = &dhcpEnabled
		} else {
			cloud.CloudNetworkSubnetEditSpec.TargetSpec.DhcpEnabled = nil
		}
		return nil
	}

	// Common flags for other means to define parameters
	addParameterFileFlags(privateNetworkSubnetEditCmd, true, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/network/{networkId}/subnet/{subnetId}", "put", "", nil)
	addInteractiveEditorFlag(privateNetworkSubnetEditCmd)
	markFlagsMutuallyExclusive(privateNetworkSubnetEditCmd, "from-file", "editor")

	return privateNetworkSubnetEditCmd
}

func getGatewayCreationCmd() *cobra.Command {
	gatewayCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a gateway in the given cloud project",
		Long: `Use this command to create a new gateway in the given public cloud project.

Subnets are nested objects: to attach them, use the repeatable --subnet flag, a
configuration file or your text editor. There are three ways to define the
creation parameters:

1. Using only CLI flags:

	ovhcloud cloud network gateway create --name MyGateway --region GRA11 --external-gateway-enabled --external-gateway-model S

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network gateway create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network gateway create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network gateway create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network gateway create --from-file ./params.json --name MyGateway

3. Using your default text editor:

	ovhcloud cloud network gateway create --editor
`,
		Run: cloud.CreateGateway,
	}

	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Name, "name", "", "Name of the gateway")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Description, "description", "", "Description of the gateway")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Location.Region, "region", "", "Region where the gateway will be created")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")
	gatewayCreateCmd.Flags().BoolVar(&cloud.CloudGatewaySpec.TargetSpec.ExternalGateway.Enabled, "external-gateway-enabled", false, "Whether the external gateway is enabled")
	gatewayCreateCmd.Flags().StringVar(&cloud.CloudGatewaySpec.TargetSpec.ExternalGateway.Model, "external-gateway-model", "", "External gateway sizing model (S, M, L, XL, 2XL, 3XL)")
	gatewayCreateCmd.Flags().StringSliceVar(&cloud.CloudGatewaySpec.TargetSpec.CliSubnets, "subnet", nil, "ID of a subnet to attach to the gateway (repeatable)")

	// Common flags for other means to define parameters
	addParameterFileFlags(gatewayCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/gateway", "post", cloud.GatewayCreationExample, nil)
	addInteractiveEditorFlag(gatewayCreateCmd)
	gatewayCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for gateway creation to be done before exiting")
	markFlagsMutuallyExclusive(gatewayCreateCmd, "from-file", "editor")

	return gatewayCreateCmd
}
