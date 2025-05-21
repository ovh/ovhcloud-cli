package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/sslgateway"
)

func init() {
	sslgatewayCmd := &cobra.Command{
		Use:   "ssl-gateway",
		Short: "Retrieve information and manage your SSL Gateway services",
	}

	// Command to list SslGateway services
	sslgatewayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your SSL Gateway services",
		Run:   sslgateway.ListSslGateway,
	}
	sslgatewayCmd.AddCommand(withFilterFlag(sslgatewayListCmd))

	// Command to get a single SslGateway
	sslgatewayCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SSL Gateway",
		Args:  cobra.ExactArgs(1),
		Run:   sslgateway.GetSslGateway,
	})

	rootCmd.AddCommand(sslgatewayCmd)
}
