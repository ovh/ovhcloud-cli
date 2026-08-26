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

func initCloudIPCommand(cloudCmd *cobra.Command) {
	ipCmd := &cobra.Command{
		Use:   "ip",
		Short: "Manage public IPs (floating, additional and ext-net) in the given cloud project",
	}
	ipCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// list (all public IPs, API v2)
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all public IPs (floating, additional and ext-net) of the project",
		Run:     cloud.ListAllPublicIPs,
	}
	ipCmd.AddCommand(withFilterFlag(listCmd))

	// Floating, additional and ext-net public IPs (API v2)
	initCloudIPFloatingCommand(ipCmd)
	initCloudIPAdditionalCommand(ipCmd)
	initCloudIPExtNetCommand(ipCmd)

	cloudCmd.AddCommand(ipCmd)
}

// initCloudIPFloatingCommand registers the `cloud ip floating` subcommands (API v2).
func initCloudIPFloatingCommand(ipCmd *cobra.Command) {
	floatingCmd := &cobra.Command{
		Use:   "floating",
		Short: "Manage floating public IPs in the given cloud project",
	}

	floatingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List floating IPs",
		Run:     cloud.ListPublicIPFloating,
	}
	floatingCmd.AddCommand(withFilterFlag(floatingListCmd))

	floatingCmd.AddCommand(&cobra.Command{
		Use:   "get <ip>",
		Short: "Get a specific floating IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetPublicIPFloating,
	})

	floatingCmd.AddCommand(getCloudIPFloatingCreateCmd())
	floatingCmd.AddCommand(getCloudIPFloatingEditCmd())

	floatingCmd.AddCommand(&cobra.Command{
		Use:   "delete <ip>",
		Short: "Delete a specific floating IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.DeletePublicIPFloating,
	})

	ipCmd.AddCommand(floatingCmd)
}

func getCloudIPFloatingCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new floating IP",
		Long: `Use this command to create a floating IP in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud ip floating create --region GRA11 --description "My floating IP"

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud ip floating create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud ip floating create --from-file ./params.json

3. Using your default text editor:

	ovhcloud cloud ip floating create --editor
`,
		Run: cloud.CreatePublicIPFloating,
	}

	createCmd.Flags().StringVar(&cloud.PublicIPFloatingCreationSpec.TargetSpec.Location.Region, "region", "", "Region where the floating IP will be created")
	createCmd.Flags().StringVar(&cloud.PublicIPFloatingCreationSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")
	createCmd.Flags().StringVar(&cloud.PublicIPFloatingCreationSpec.TargetSpec.Description, "description", "", "Description of the floating IP")
	createCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for floating IP creation to be done before exiting")

	addParameterFileFlags(createCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/publicIp/floating", "post", cloud.PublicIPFloatingCreationExample, nil)
	addInteractiveEditorFlag(createCmd)
	markFlagsMutuallyExclusive(createCmd, "from-file", "editor")

	return createCmd
}

func getCloudIPFloatingEditCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit <ip>",
		Short: "Edit the given floating IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditPublicIPFloating,
	}

	editCmd.Flags().StringVar(&cloud.PublicIPFloatingUpdateSpec.TargetSpec.Description, "description", "", "Description of the floating IP")

	addParameterFileFlags(editCmd, true, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/publicIp/floating/{id}", "put", "", nil)
	addInteractiveEditorFlag(editCmd)
	markFlagsMutuallyExclusive(editCmd, "from-file", "editor")

	return editCmd
}

// initCloudIPAdditionalCommand registers the `cloud ip additional` subcommands.
// List and get use API v2; attach still uses API v1 (not available in v2 yet).
func initCloudIPAdditionalCommand(ipCmd *cobra.Command) {
	additionalCmd := &cobra.Command{
		Use:   "additional",
		Short: "Manage additional public IPs in the given cloud project",
	}

	additionalListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List additional IPs",
		Run:     cloud.ListPublicIPAdditional,
	}
	additionalCmd.AddCommand(withFilterFlag(additionalListCmd))

	additionalCmd.AddCommand(&cobra.Command{
		Use:   "get <ip>",
		Short: "Get a specific additional IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetPublicIPAdditional,
	})

	additionalCmd.AddCommand(&cobra.Command{
		Use:   "attach <ip_id> <instance_id>",
		Short: "Attach an additional IP to an instance",
		Args:  cobra.ExactArgs(2),
		Run:   cloud.AttachPublicIPAdditional,
	})

	ipCmd.AddCommand(additionalCmd)
}

// initCloudIPExtNetCommand registers the `cloud ip extNet` subcommands (API v2).
func initCloudIPExtNetCommand(ipCmd *cobra.Command) {
	extNetCmd := &cobra.Command{
		Use:   "extnet",
		Short: "Manage ext-net public IPs in the given cloud project",
	}

	extNetListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List ext-net IPs",
		Run:     cloud.ListPublicIPExtNet,
	}
	extNetCmd.AddCommand(withFilterFlag(extNetListCmd))

	extNetCmd.AddCommand(&cobra.Command{
		Use:   "get <ip>",
		Short: "Get a specific ext-net IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetPublicIPExtNet,
	})

	extNetCmd.AddCommand(&cobra.Command{
		Use:   "delete <ip>",
		Short: "Delete a specific ext-net IP",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.DeletePublicIPExtNet,
	})

	ipCmd.AddCommand(extNetCmd)
}
