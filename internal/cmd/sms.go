package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/sms"
)

func init() {
	smsCmd := &cobra.Command{
		Use:   "sms",
		Short: "Retrieve information and manage your SMS services",
	}

	// Command to list Sms services
	smsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your SMS services",
		Run:   sms.ListSms,
	}
	smsCmd.AddCommand(withFilterFlag(smsListCmd))

	// Command to get a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SMS account",
		Args:  cobra.ExactArgs(1),
		Run:   sms.GetSms,
	})

	// Command to update a single Sms
	smsCmd.AddCommand(&cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given SMS account",
		Run:   sms.EditSms,
	})

	rootCmd.AddCommand(smsCmd)
}
