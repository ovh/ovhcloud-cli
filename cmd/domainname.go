package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	domainnameColumnsToDisplay = []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"}

	//go:embed templates/domainname.tmpl
	domainnameTemplate string
)

func listDomainName(_ *cobra.Command, _ []string) {
	manageListRequest("/domain", "", domainnameColumnsToDisplay, genericFilters)
}

func getDomainName(_ *cobra.Command, args []string) {
	manageObjectRequest("/domain", args[0], domainnameTemplate)
}

func init() {
	domainnameCmd := &cobra.Command{
		Use:   "domain-name",
		Short: "Retrieve information and manage your domain names",
	}

	// Command to list DomainName services
	domainnameListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your domain names",
		Run:   listDomainName,
	}
	domainnameCmd.AddCommand(withFilterFlag(domainnameListCmd))

	// Command to get a single DomainName
	domainnameCmd.AddCommand(&cobra.Command{
		Use:   "get <domain_name>",
		Short: "Retrieve information of a specific domain name",
		Args:  cobra.ExactArgs(1),
		Run:   getDomainName,
	})

	rootCmd.AddCommand(domainnameCmd)
}
