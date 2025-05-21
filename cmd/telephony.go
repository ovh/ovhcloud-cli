package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/telephony"
)

func init() {
	telephonyCmd := &cobra.Command{
		Use:   "telephony",
		Short: "Retrieve information and manage your Telephony services",
	}

	// Command to list Telephony services
	telephonyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Telephony services",
		Run:   telephony.ListTelephony,
	}
	telephonyCmd.AddCommand(withFilterFlag(telephonyListCmd))

	// Command to get a single Telephony
	telephonyCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Telephony service",
		Args:  cobra.ExactArgs(1),
		Run:   telephony.GetTelephony,
	})

	rootCmd.AddCommand(telephonyCmd)
}
