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

func initCloudPrivateEndpointCommand(cloudCmd *cobra.Command) {
	privateEndpointCmd := &cobra.Command{
		Use:     "private-endpoint",
		Aliases: []string{"pe"},
		Short:   "Manage private endpoints in the given cloud project",
	}
	privateEndpointCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")
	cloudCmd.AddCommand(privateEndpointCmd)

	privateEndpointListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List private endpoints",
		Run:     cloud.ListCloudPrivateEndpoint,
	}
	privateEndpointCmd.AddCommand(withFilterFlag(privateEndpointListCmd))

	privateEndpointCmd.AddCommand(&cobra.Command{
		Use:   "get <private_endpoint_id>",
		Short: "Get a specific private endpoint",
		Run:   cloud.GetCloudPrivateEndpoint,
		Args:  cobra.ExactArgs(1),
	})

	privateEndpointCmd.AddCommand(getCloudPrivateEndpointCreateCmd())
	privateEndpointCmd.AddCommand(getCloudPrivateEndpointEditCmd())

	privateEndpointCmd.AddCommand(&cobra.Command{
		Use:   "delete <private_endpoint_id>",
		Short: "Delete a specific private endpoint",
		Run:   cloud.DeleteCloudPrivateEndpoint,
		Args:  cobra.ExactArgs(1),
	})

	privateEndpointCmd.AddCommand(getCloudPrivateEndpointServiceAvailabilityCmd())
}

func getCloudPrivateEndpointCreateCmd() *cobra.Command {
	var publicDnsEnabled bool

	spec := &cloud.PrivateEndpointSpec.TargetSpec

	privateEndpointCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new private endpoint",
		Long: `Use this command to create a new private endpoint in the given public cloud project.

Use the "service-availability list" command to discover the services and tiers
that can be exposed. There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud private-endpoint create --name MyEndpoint --region GRA11 --service-name MyService --tier S --subnet-id <subnet_id>

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud private-endpoint create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud private-endpoint create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud private-endpoint create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud private-endpoint create --from-file ./params.json --name MyEndpoint

3. Using your default text editor:

	ovhcloud cloud private-endpoint create --editor
`,
		Run: cloud.CreateCloudPrivateEndpoint,
	}

	privateEndpointCreateCmd.Flags().StringVar(&spec.Name, "name", "", "Name of the private endpoint")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Description, "description", "", "Description of the private endpoint")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Location.Region, "region", "", "Region where the private endpoint will be created")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Service.Name, "service-name", "", "Name of the service to expose")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Service.Tier, "tier", "", "Tier of the service (available tiers can be listed using 'ovhcloud cloud private-endpoint service-availability list' command)")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Network.Id, "network-id", "", "Network to allocate the endpoint port in")
	privateEndpointCreateCmd.Flags().StringVar(&spec.Subnet.Id, "subnet-id", "", "Subnet to allocate the endpoint port in")
	privateEndpointCreateCmd.Flags().StringVar(&spec.VipAddress, "vip-address", "", "Private IP address to assign to the endpoint inside the subnet")
	privateEndpointCreateCmd.Flags().BoolVar(&publicDnsEnabled, "public-dns-enabled", false, "Register the endpoint hostname in the public DNS zone")

	// Handle optional public-dns-enabled boolean
	privateEndpointCreateCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("public-dns-enabled") {
			spec.PublicDnsEnabled = &publicDnsEnabled
		} else {
			spec.PublicDnsEnabled = nil
		}

		return nil
	}

	// Common flags for other means to define parameters
	addParameterFileFlags(privateEndpointCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/privateEndpoint", "post", cloud.PrivateEndpointCreateExample, nil)
	addInteractiveEditorFlag(privateEndpointCreateCmd)
	privateEndpointCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for private endpoint creation to be done before exiting")
	markFlagsMutuallyExclusive(privateEndpointCreateCmd, "from-file", "editor")

	return privateEndpointCreateCmd
}

func getCloudPrivateEndpointEditCmd() *cobra.Command {
	privateEndpointEditCmd := &cobra.Command{
		Use:   "edit <private_endpoint_id>",
		Short: "Edit the given private endpoint (only name and description are mutable)",
		Run:   cloud.EditCloudPrivateEndpoint,
		Args:  cobra.ExactArgs(1),
	}

	var description string

	privateEndpointEditCmd.Flags().StringVar(&cloud.PrivateEndpointEditSpec.TargetSpec.Name, "name", "", "Name of the private endpoint")
	privateEndpointEditCmd.Flags().StringVar(&description, "description", "", "Description of the private endpoint (an empty value clears it)")
	addInteractiveEditorFlag(privateEndpointEditCmd)

	// Handle optional description: only sent when the flag is given, so that
	// an explicit empty value clears it
	privateEndpointEditCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("description") {
			cloud.PrivateEndpointEditSpec.TargetSpec.Description = &description
		} else {
			cloud.PrivateEndpointEditSpec.TargetSpec.Description = nil
		}

		return nil
	}

	return privateEndpointEditCmd
}

func getCloudPrivateEndpointServiceAvailabilityCmd() *cobra.Command {
	serviceAvailabilityCmd := &cobra.Command{
		Use:     "service-availability",
		Aliases: []string{"sa"},
		Short:   "Consult the services that can be exposed through a private endpoint",
	}

	serviceAvailabilityListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available private endpoint services",
		Run:     cloud.ListCloudPrivateEndpointServiceAvailabilities,
	}
	serviceAvailabilityCmd.AddCommand(withFilterFlag(serviceAvailabilityListCmd))

	serviceAvailabilityCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Get a specific private endpoint service availability",
		Run:   cloud.GetCloudPrivateEndpointServiceAvailability,
		Args:  cobra.ExactArgs(1),
	})

	return serviceAvailabilityCmd
}
