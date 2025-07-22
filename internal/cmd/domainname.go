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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your domain names",
		Run:     domainname.ListDomainName,
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
	editDomainNameCmd := &cobra.Command{
		Use:   "edit <domain_name>",
		Short: "Edit the given domain name service",
		Args:  cobra.ExactArgs(1),
		Run:   domainname.EditDomainName,
	}
	editDomainNameCmd.Flags().StringVar(&domainname.DomainSpec.NameServerType, "name-server-type", "", "Type of name server (anycast, dedicated, empty, external, hold, hosted, hosting, mixed, parking)")
	editDomainNameCmd.Flags().StringVar(&domainname.DomainSpec.TranferLockStatus, "transfer-lock-status", "", "Transfer lock status (locked, locking, unavailable, unlocked, unlocking)")
	addInteractiveEditorFlag(editDomainNameCmd)
	domainnameCmd.AddCommand(editDomainNameCmd)

	rootCmd.AddCommand(domainnameCmd)
}
