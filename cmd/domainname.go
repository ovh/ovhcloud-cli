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
		Use:   "domainname",
		Short: "Retrieve information and manage your DomainName services",
	}

	// Command to list DomainName services
	domainnameListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your DomainName services",
		Run:   listDomainName,
	}
	domainnameCmd.AddCommand(withFilterFlag(domainnameListCmd))

	// Command to get a single DomainName
	domainnameCmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific DomainName",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        getDomainName,
	})

	rootCmd.AddCommand(domainnameCmd)
}
