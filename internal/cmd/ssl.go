package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/ssl"
)

func init() {
	sslCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Retrieve information and manage your SSL services",
	}

	// Command to list Ssl services
	sslListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your SSL services",
		Run:     ssl.ListSsl,
	}
	sslCmd.AddCommand(withFilterFlag(sslListCmd))

	// Command to get a single Ssl
	sslCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific Ssl",
		Args:  cobra.ExactArgs(1),
		Run:   ssl.GetSsl,
	})

	rootCmd.AddCommand(sslCmd)
}
