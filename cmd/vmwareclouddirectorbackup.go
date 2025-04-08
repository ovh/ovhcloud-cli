
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vmwareclouddirectorbackupColumnsToDisplay = []string{ "id","iam.displayName","currentState.azName","resourceStatus" }
)

func listVmwareCloudDirectorBackup(_ *cobra.Command, _ []string) {
	manageListRequest("/v2/vmwareCloudDirector/backup", vmwareclouddirectorbackupColumnsToDisplay)
}

func getVmwareCloudDirectorBackup(_ *cobra.Command, args []string) {
	manageObjectRequest("/v2/vmwareCloudDirector/backup", args[0], vmwareclouddirectorbackupColumnsToDisplay[0])
}

func init() {
	vmwareclouddirectorbackupCmd := &cobra.Command{
		Use:   "vmwareclouddirectorbackup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorBackup services",
		Run:   listVmwareCloudDirectorBackup,
	})

	// Command to get a single VmwareCloudDirectorBackup
	vmwareclouddirectorbackupCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific VmwareCloudDirectorBackup",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getVmwareCloudDirectorBackup,
	})

	rootCmd.AddCommand(vmwareclouddirectorbackupCmd)
}
