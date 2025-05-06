package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	emaildomainColumnsToDisplay = []string{"domain", "status", "offer"}

	//go:embed templates/emaildomain.tmpl
	emaildomainTemplate string
)

func listEmailDomain(_ *cobra.Command, _ []string) {
	manageListRequest("/email/domain", "", emaildomainColumnsToDisplay, genericFilters)
}

func getEmailDomain(_ *cobra.Command, args []string) {
	manageObjectRequest("/email/domain", args[0], emaildomainTemplate)
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
	emaildomainCmd.AddCommand(withFilterFlag(emaildomainListCmd))

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
