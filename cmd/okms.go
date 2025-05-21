package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/okms"
)

func init() {
	okmsCmd := &cobra.Command{
		Use:   "okms",
		Short: "Retrieve information and manage your Okms services",
	}

	// Command to list Okms services
	okmsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your Okms services",
		Run:   okms.ListOkms,
	}
	okmsCmd.AddCommand(withFilterFlag(okmsListCmd))

	// Command to get a single Okms
	okmsCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Okms",
		Args:  cobra.ExactArgs(1),
		Run:   okms.GetOkms,
	})

	rootCmd.AddCommand(okmsCmd)
}
