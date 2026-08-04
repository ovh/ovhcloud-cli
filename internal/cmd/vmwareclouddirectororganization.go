// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/vmwareclouddirectororganization"
	"github.com/spf13/cobra"
)

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirector-organization",
		Short: "Manage VMware Cloud Director organizations",
	}

	// Command to list VmwareCloudDirector Organizations
	vmwareclouddirectororganizationListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your VmwareCloudDirector Organizations",
		Run:     vmwareclouddirectororganization.ListVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationCmd.AddCommand(withFilterFlag(vmwareclouddirectororganizationListCmd))

	// Command to get a single VmwareCloudDirector Organization
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific VmwareCloudDirector Organization",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v2/vmwareCloudDirector/organization"),
		Run:               vmwareclouddirectororganization.GetVmwareCloudDirectorOrganization,
	})

	// Command to update a single VmwareCloudDirector Organization
	vmwareclouddirectororganizationEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit the given VmwareCloudDirector Organization",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v2/vmwareCloudDirector/organization"),
		Run:               vmwareclouddirectororganization.EditVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationEditCmd.Flags().StringVar(&vmwareclouddirectororganization.VmwareCloudDirectorOrganizationSpec.TargetSpec.Description, "description", "", "Description of the organization")
	vmwareclouddirectororganizationEditCmd.Flags().StringVar(&vmwareclouddirectororganization.VmwareCloudDirectorOrganizationSpec.TargetSpec.FullName, "full-name", "", "Full name of the organization")
	addInteractiveEditorFlag(vmwareclouddirectororganizationEditCmd)
	vmwareclouddirectororganizationCmd.AddCommand(vmwareclouddirectororganizationEditCmd)

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
