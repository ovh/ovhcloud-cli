package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"
)

var (
	sslgatewayColumnsToDisplay = []string{"serviceName", "displayName", "state", "zones"}

	//go:embed templates/sslgateway.tmpl
	sslgatewayTemplate string
)

func listSslGateway(_ *cobra.Command, _ []string) {
	manageListRequest("/sslGateway", "", sslgatewayColumnsToDisplay, genericFilters)
}

func getSslGateway(_ *cobra.Command, args []string) {
	manageObjectRequest("/sslGateway", args[0], sslgatewayTemplate)
}

func init() {
	sslgatewayCmd := &cobra.Command{
		Use:   "ssl-gateway",
		Short: "Retrieve information and manage your SSL Gateway services",
	}

	// Command to list SslGateway services
	sslgatewayListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your SSL Gateway services",
		Run:   listSslGateway,
	}
	sslgatewayCmd.AddCommand(withFilterFlag(sslgatewayListCmd))

	// Command to get a single SslGateway
	sslgatewayCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SSL Gateway",
		Args:  cobra.ExactArgs(1),
		Run:   getSslGateway,
	})

	rootCmd.AddCommand(sslgatewayCmd)
}
