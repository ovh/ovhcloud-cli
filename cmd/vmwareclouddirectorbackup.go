package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	vmwareclouddirectorbackupColumnsToDisplay = []string{"id", "iam.displayName", "currentState.azName", "resourceStatus"}

	//go:embed templates/vmwareclouddirectorbackup.tmpl
	vmwareclouddirectorbackupTemplate string
)

func listVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vmwareCloudDirector/backup", "id", vmwareclouddirectorbackupColumnsToDisplay, genericFilters)
}

func getVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vmwareCloudDirector/backup", args[0], vmwareclouddirectorbackupTemplate)
}

func init() {
	vmwareclouddirectorbackupCmd := &cobra.Command{
		Use:   "vmwareclouddirector-backup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirector Backup services",
		Run:   listVmwareCloudDirectorBackup,
	}
	vmwareclouddirectorbackupCmd.AddCommand(withFilterFlag(vmwareclouddirectorbackupListCmd))

	// Command to get a single VmwareCloudDirectorBackup
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific VmwareCloudDirector Backup",
		Args:  cobra.ExactArgs(1),
		Run:   getVmwareCloudDirectorBackup,
	})

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
