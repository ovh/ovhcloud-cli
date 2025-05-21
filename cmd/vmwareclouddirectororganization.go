package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/vmwareclouddirectororganization"
)

func init() {
	vmwareclouddirectororganizationCmd := &cobra.Command{
		Use:   "vmwareclouddirector-organization",
		Short: "Retrieve information and manage your VmwareCloudDirector Organizations",
	}

	// Command to list VmwareCloudDirectorOrganization services
	vmwareclouddirectororganizationListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirector Organizations",
		Run:   vmwareclouddirectororganization.ListVmwareCloudDirectorOrganization,
	}
	vmwareclouddirectororganizationCmd.AddCommand(withFilterFlag(vmwareclouddirectororganizationListCmd))

	// Command to get a single VmwareCloudDirectorOrganization
	vmwareclouddirectororganizationCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VmwareCloudDirector Organization",
		Args:  cobra.ExactArgs(1),
		Run:   vmwareclouddirectororganization.GetVmwareCloudDirectorOrganization,
	})

	rootCmd.AddCommand(vmwareclouddirectororganizationCmd)
}
