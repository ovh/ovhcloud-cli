// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/vmwareclouddirectorbackup"
	"github.com/spf13/cobra"
)

func init() {
	vmwareclouddirectorbackupCmd := &cobra.Command{
		Use:   "vmwareclouddirector-backup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your VmwareCloudDirector Backup services",
		Run:     vmwareclouddirectorbackup.ListVmwareCloudDirectorBackup,
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
	addInteractiveEditorFlag(vmwareclouddirectorbackupEditCmd)
	vmwareclouddirectorbackupCmd.AddCommand(vmwareclouddirectorbackupEditCmd)

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
