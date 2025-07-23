package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/emaildomain"
	"github.com/spf13/cobra"
)

func init() {
	emaildomainCmd := &cobra.Command{
		Use:   "email-domain",
		Short: "Retrieve information and manage your Email Domain services",
	}

	// Command to list EmailDomain services
	emaildomainListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your Email Domain services",
		Run:     emaildomain.ListEmailDomain,
	}
	emaildomainCmd.AddCommand(withFilterFlag(emaildomainListCmd))

	// Command to get a single EmailDomain
	emaildomainCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Email Domain",
		Args:  cobra.ExactArgs(1),
		Run:   emaildomain.GetEmailDomain,
	})

	rootCmd.AddCommand(emaildomainCmd)
}
