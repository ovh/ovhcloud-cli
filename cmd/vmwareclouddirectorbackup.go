package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/vmwareclouddirectorbackup"
)

func init() {
	vmwareclouddirectorbackupCmd := &cobra.Command{
		Use:   "vmwareclouddirector-backup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirector Backup services",
		Run:   vmwareclouddirectorbackup.ListVmwareCloudDirectorBackup,
	}
	vmwareclouddirectorbackupCmd.AddCommand(withFilterFlag(vmwareclouddirectorbackupListCmd))

	// Command to get a single VmwareCloudDirectorBackup
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VmwareCloudDirector Backup",
		Args:  cobra.ExactArgs(1),
		Run:   vmwareclouddirectorbackup.GetVmwareCloudDirectorBackup,
	})

	// Command to update a single VmwareCloudDirectorBackup
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VmwareCloudDirector Backup",
		Run:   vmwareclouddirectorbackup.EditVmwareCloudDirectorBackup,
	})

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
