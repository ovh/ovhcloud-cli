package cmd

import (
	"github.com/spf13/cobra"
)

var (
	emaildomainColumnsToDisplay = []string{ "domain","status","offer" }
)

func listEmailDomain(_ *cobra.Command, _ []string) {
	manageListRequest("/email/domain", emaildomainColumnsToDisplay, genericFilters)
}

func getEmailDomain(_ *cobra.Command, args []string) {
	manageObjectRequest("/email/domain", args[0], emaildomainColumnsToDisplay[0])
}

func init() {
	emaildomainCmd := &cobra.Command{
		Use:   "emaildomain",
		Short: "Retrieve information and manage your EmailDomain services",
	}

	// Command to list EmailDomain services
	emaildomainListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your EmailDomain services",
		Run:   listEmailDomain,
	}
	emaildomainListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	emaildomainCmd.AddCommand(emaildomainListCmd)

	// Command to get a single EmailDomain
	emaildomainCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific EmailDomain",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getEmailDomain,
	})

	rootCmd.AddCommand(emaildomainCmd)
}
