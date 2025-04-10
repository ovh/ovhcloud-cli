package cmd

import (
	"github.com/spf13/cobra"
)

var (
	smsColumnsToDisplay = []string{ "name","status" }
)

func listSms(_ *cobra.Command, _ []string) {
	manageListRequest("/sms", smsColumnsToDisplay, genericFilters)
}

func getSms(_ *cobra.Command, args []string) {
	manageObjectRequest("/sms", args[0], smsColumnsToDisplay[0])
}

func init() {
	smsCmd := &cobra.Command{
		Use:   "sms",
		Short: "Retrieve information and manage your Sms services",
	}

	// Command to list Sms services
	smsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Sms services",
		Run:   listSms,
	}
	smsListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	smsCmd.AddCommand(smsListCmd)

	// Command to get a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific Sms",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getSms,
	})

	rootCmd.AddCommand(smsCmd)
}
