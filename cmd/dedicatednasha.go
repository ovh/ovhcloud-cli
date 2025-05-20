package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	dedicatednashaColumnsToDisplay = []string{"serviceName", "customName", "datacenter"}

	//go:embed templates/dedicatednasha.tmpl
	dedicatednashaTemplate string
)

func listDedicatedNasHA(_ *cobra.Command, _ []string) {
	manageListRequest("/dedicated/nasha", "", dedicatednashaColumnsToDisplay, genericFilters)
}

func getDedicatedNasHA(_ *cobra.Command, args []string) {
	manageObjectRequest("/dedicated/nasha", args[0], dedicatednashaTemplate)
}

func init() {
	dedicatednashaCmd := &cobra.Command{
		Use:   "dedicated-nasha",
		Short: "Retrieve information and manage your Dedicated NasHA services",
	}

	// Command to list DedicatedNasHA services
	dedicatednashaListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DedicatedNasHA services",
		Run:   listDedicatedNasHA,
	}
	dedicatednashaCmd.AddCommand(withFilterFlag(dedicatednashaListCmd))

	// Command to get a single DedicatedNasHA
	dedicatednashaCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific DedicatedNasHA",
		Args:  cobra.ExactArgs(1),
		Run:   getDedicatedNasHA,
	})

	rootCmd.AddCommand(dedicatednashaCmd)
}
