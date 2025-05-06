package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	telephonyColumnsToDisplay = []string{"billingAccount", "description", "status"}

	//go:embed templates/telephony.tmpl
	telephonyTemplate string
)

func listTelephony(_ *cobra.Command, _ []string) {
	manageListRequest("/telephony", "", telephonyColumnsToDisplay, genericFilters)
}

func getTelephony(_ *cobra.Command, args []string) {
	manageObjectRequest("/telephony", args[0], telephonyTemplate)
}

func init() {
	telephonyCmd := &cobra.Command{
		Use:   "telephony",
		Short: "Retrieve information and manage your Telephony services",
	}

	// Command to list Telephony services
	telephonyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Telephony services",
		Run:   listTelephony,
	}
	telephonyCmd.AddCommand(withFilterFlag(telephonyListCmd))

	// Command to get a single Telephony
	telephonyCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Telephony",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getTelephony,
	})

	rootCmd.AddCommand(telephonyCmd)
}
