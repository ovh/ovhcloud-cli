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
		Use:   "email-domain",
		Short: "Retrieve information and manage your Email Domain services",
	}

	// Command to list EmailDomain services
	emaildomainListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Email Domain services",
		Run:   listEmailDomain,
	}
	emaildomainCmd.AddCommand(withFilterFlag(emaildomainListCmd))

	// Command to get a single EmailDomain
	emaildomainCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Email Domain",
		Args:  cobra.ExactArgs(1),
		Run:   getEmailDomain,
	})

	rootCmd.AddCommand(emaildomainCmd)
}
