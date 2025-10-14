// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
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
	editRancherCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Name, "name", "", "Name of the managed Rancher service")
	editRancherCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Plan, "plan", "", "Plan of the managed Rancher service (OVHCLOUD_EDITION, STANDARD)")
	editRancherCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Version, "version", "", "Version of the managed Rancher service")
	editRancherCmd.Flags().StringArrayVar(&cloud.RancherSpec.TargetSpec.CLIIPRestrictions, "ip-restrictions", nil, "List of IP restrictions (expected format: '<cidrBlock>,<description>')")
	addInteractiveEditorFlag(editRancherCmd)
	rancherCmd.AddCommand(editRancherCmd)

	rancherCmd.AddCommand(getRancherCreateCmd())

	rancherCmd.AddCommand(&cobra.Command{
		Use:   "reset-admin-credentials <rancher_id>",
		Short: "Reset admin user credentials",
		Run:   cloud.ResetRancherAdminCredentials,
		Args:  cobra.ExactArgs(1),
	})

	rancherCmd.AddCommand(&cobra.Command{
		Use:   "delete <rancher_id>",
		Short: "Delete a specific Rancher service",
		Run:   cloud.DeleteRancher,
		Args:  cobra.ExactArgs(1),
	})

	cloudCmd.AddCommand(rancherCmd)
}

func getRancherCreateCmd() *cobra.Command {
	rancherCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Rancher service",
		Long: `Use this command to create a managed Rancher service in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud rancher create --name MyNewRancher --plan OVHCLOUD_EDITION --version 2.11.3

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud rancher create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud rancher create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud rancher create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud rancher create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud rancher create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud rancher create --editor --region BHS5
`,
		Run: cloud.CreateRancher,
	}

	// Specific flags for Rancher creation
	rancherCreateCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Name, "name", "", "Name of the managed Rancher service")
	rancherCreateCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Plan, "plan", "", "Plan of the managed Rancher service (available plans can be listed using 'cloud reference rancher list-plans' command)")
	rancherCreateCmd.Flags().StringVar(&cloud.RancherSpec.TargetSpec.Version, "version", "", "Version of the managed Rancher service (available versions can be listed using 'cloud reference rancher list-versions' command)")
	rancherCreateCmd.Flags().BoolVar(&cloud.RancherSpec.TargetSpec.IAMAuthEnabled, "iam-auth-enabled", false, "Allow Rancher to use identities managed by OVHcloud IAM (Identity and Access Management) to control access")

	// Common flags for other means to define parameters
	addInitParameterFileFlag(rancherCreateCmd, assets.CloudV2OpenapiSchema, "/cloud/project/{serviceName}/rancher", "post", cloud.CloudRancherCreationExample, nil)
	addInteractiveEditorFlag(rancherCreateCmd)
	addFromFileFlag(rancherCreateCmd)
	rancherCreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	return rancherCreateCmd
}
