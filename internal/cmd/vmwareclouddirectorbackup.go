package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
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
	vmwareclouddirectorbackupEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given VmwareCloudDirector Backup",
		Args:  cobra.ExactArgs(1),
		Run:   vmwareclouddirectorbackup.EditVmwareCloudDirectorBackup,
	}
	vmwareclouddirectorbackupEditCmd.Flags().StringSliceVar(
		&vmwareclouddirectorbackup.VmwareCloudDirectorBackupSpec.TargetSpec.CliOffers,
		"offers", nil, "List of your VMware Cloud Director backup offers formatted as '<name>:<quotaInTB>' (available names: BRONZE, GOLD, SILVER)",
	)
	vmwareclouddirectorbackupEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define parameters")
	vmwareclouddirectorbackupCmd.AddCommand(vmwareclouddirectorbackupEditCmd)

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
