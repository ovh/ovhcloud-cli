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
		Use:   "vmwareclouddirectorbackup",
		Short: "Retrieve information and manage your VmwareCloudDirectorBackup services",
	}

	// Command to list VmwareCloudDirectorBackup services
	vmwareclouddirectorbackupListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VmwareCloudDirectorBackup services",
		Run:   listVmwareCloudDirectorBackup,
	}
	vmwareclouddirectorbackupListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	vmwareclouddirectorbackupCmd.AddCommand(vmwareclouddirectorbackupListCmd)

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
