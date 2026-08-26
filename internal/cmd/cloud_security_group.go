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

func initCloudSecurityGroupCommand(networkCmd *cobra.Command) {
	securityGroupCmd := &cobra.Command{
		Use:     "security-group",
		Aliases: []string{"sg"},
		Short:   "Manage security groups in the given cloud project",
	}
	networkCmd.AddCommand(securityGroupCmd)

	securityGroupListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List security groups",
		Run:     cloud.ListSecurityGroups,
	}
	securityGroupCmd.AddCommand(withFilterFlag(securityGroupListCmd))

	securityGroupCmd.AddCommand(&cobra.Command{
		Use:   "get <security_group_id>",
		Short: "Get a specific security group",
		Run:   cloud.GetSecurityGroup,
		Args:  cobra.ExactArgs(1),
	})

	securityGroupCmd.AddCommand(getSecurityGroupCreateCmd())
	securityGroupCmd.AddCommand(getSecurityGroupEditCmd())

	securityGroupCmd.AddCommand(&cobra.Command{
		Use:   "delete <security_group_id>",
		Short: "Delete a specific security group",
		Run:   cloud.DeleteSecurityGroup,
		Args:  cobra.ExactArgs(1),
	})
}

func getSecurityGroupCreateCmd() *cobra.Command {
	securityGroupCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new security group",
		Long: `Use this command to create a security group in the given public cloud project.

Rules are nested objects: to define them, use a configuration file or your text
editor. There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud network security-group create --name my-sg --region GRA11

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network security-group create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network security-group create --from-file ./params.json

  You can also pipe the content of the parameters file:

	cat ./params.json | ovhcloud cloud network security-group create

  In both cases, you can override values using command line flags, for example:

	ovhcloud cloud network security-group create --from-file ./params.json --name my-sg

3. Using your default text editor:

	ovhcloud cloud network security-group create --editor
`,
		Run: cloud.CreateSecurityGroup,
	}

	securityGroupCreateCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Name, "name", "", "Name of the security group")
	securityGroupCreateCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Description, "description", "", "Description of the security group")
	securityGroupCreateCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Location.Region, "region", "", "Region where the security group will be created")
	securityGroupCreateCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Location.AvailabilityZone, "availability-zone", "", "Availability zone within the region")

	addParameterFileFlags(securityGroupCreateCmd, false, assets.CloudV2OpenapiSchema, "/publicCloud/project/{projectId}/securityGroup", "post", cloud.SecurityGroupCreationExample, nil)
	addInteractiveEditorFlag(securityGroupCreateCmd)
	securityGroupCreateCmd.Flags().BoolVar(&flags.WaitForTask, "wait", false, "Wait for security group creation to be done before exiting")
	markFlagsMutuallyExclusive(securityGroupCreateCmd, "from-file", "editor")

	return securityGroupCreateCmd
}

func getSecurityGroupEditCmd() *cobra.Command {
	securityGroupEditCmd := &cobra.Command{
		Use:   "edit <security_group_id>",
		Short: "Edit the given security group",
		Run:   cloud.EditSecurityGroup,
		Args:  cobra.ExactArgs(1),
	}

	securityGroupEditCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Name, "name", "", "Name of the security group")
	securityGroupEditCmd.Flags().StringVar(&cloud.SecurityGroupSpec.TargetSpec.Description, "description", "", "Description of the security group")
	addInteractiveEditorFlag(securityGroupEditCmd)

	return securityGroupEditCmd
}
