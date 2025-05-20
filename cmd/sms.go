package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	smsColumnsToDisplay = []string{"name", "status"}

	//go:embed templates/sms.tmpl
	smsTemplate string
)

func listSms(_ *cobra.Command, _ []string) {
	manageListRequest("/sms", "", smsColumnsToDisplay, genericFilters)
}

func getSms(_ *cobra.Command, args []string) {
	manageObjectRequest("/sms", args[0], smsTemplate)
}

func init() {
	smsCmd := &cobra.Command{
		Use:   "sms",
		Short: "Retrieve information and manage your SMS services",
	}

	// Command to list Sms services
	smsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your SMS services",
		Run:   listSms,
	}
	smsCmd.AddCommand(withFilterFlag(smsListCmd))

	// Command to get a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SMS",
		Args:  cobra.ExactArgs(1),
		Run:   getSms,
	})

	rootCmd.AddCommand(smsCmd)
}
