package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/emailpro"
)

func init() {
	emailproCmd := &cobra.Command{
		Use:   "email-pro",
		Short: "Retrieve information and manage your EmailPro services",
	}

	// Command to list EmailPro services
	emailproListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your EmailPro services",
		Run:   emailpro.ListEmailPro,
	}
	emailproCmd.AddCommand(withFilterFlag(emailproListCmd))

	// Command to get a single EmailPro
	emailproCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific EmailPro",
		Args:  cobra.ExactArgs(1),
		Run:   emailpro.GetEmailPro,
	})

	rootCmd.AddCommand(emailproCmd)
}
