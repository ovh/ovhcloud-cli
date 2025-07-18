package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/vmwareclouddirectororganization"
)

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirector-organization",
		Short: "Retrieve information and manage your VmwareCloudDirector Organizations",
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
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VmwareCloudDirector Organization",
		Args:  cobra.ExactArgs(1),
		Run:   vmwareclouddirectororganization.GetVmwareCloudDirectorOrganization,
	})

	// Command to update a single VmwareCloudDirector Organization
	vmwareclouddirectororganizationEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VmwareCloudDirector Organization",
		Args:  cobra.ExactArgs(1),
		Run:   vmwareclouddirectororganization.EditVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationEditCmd.Flags().StringVar(&vmwareclouddirectororganization.VmwareCloudDirectorOrganizationSpec.TargetSpec.Description, "description", "", "Description of the organization")
	vmwareclouddirectororganizationEditCmd.Flags().StringVar(&vmwareclouddirectororganization.VmwareCloudDirectorOrganizationSpec.TargetSpec.FullName, "full-name", "", "Full name of the organization")
	vmwareclouddirectororganizationEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	vmwareclouddirectororganizationCmd.AddCommand(vmwareclouddirectororganizationEditCmd)

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
