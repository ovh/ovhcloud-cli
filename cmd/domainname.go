package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/domainname"
)

func init() {
	domainnameCmd := &cobra.Command{
		Use:   "domain-name",
		Short: "Retrieve information and manage your domain names",
	}

	// Command to list DomainName services
	domainnameListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your domain names",
		Run:   domainname.ListDomainName,
	}
	domainnameCmd.AddCommand(withFilterFlag(domainnameListCmd))

	// Command to get a single DomainName
	domainnameCmd.AddCommand(&cobra.Command{
		Use:   "get <domain_name>",
		Short: "Retrieve information of a specific domain name",
		Args:  cobra.ExactArgs(1),
		Run:   domainname.GetDomainName,
	})

	// Command to update a single DomainName
	domainnameCmd.AddCommand(&cobra.Command{
		Use:   "edit <domain_name>",
		Short: "Edit the given domain name service",
		Run:   domainname.EditDomainName,
	})

	rootCmd.AddCommand(domainnameCmd)
}
