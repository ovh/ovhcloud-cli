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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your SSL Gateway services",
		Run:     sslgateway.ListSslGateway,
	}
	sslgatewayCmd.AddCommand(withFilterFlag(sslgatewayListCmd))

	// Command to get a single SslGateway
	sslgatewayCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name>",
		Short: "Retrieve information of a specific SSL Gateway",
		Args:  cobra.ExactArgs(1),
		Run:   sslgateway.GetSslGateway,
	})

	// Command to update a single SslGateway
	sslgatewayEditCmd := &cobra.Command{
		Use:   "edit <service_name>",
		Short: "Edit the given SSL Gateway",
		Args:  cobra.ExactArgs(1),
		Run:   sslgateway.EditSslGateway,
	}
	sslgatewayEditCmd.Flags().StringSliceVar(&sslgateway.SSLGatewaySpec.AllowedSource, "allowed-source", nil, "Restrict SSL Gateway access to these ip block")
	sslgatewayEditCmd.Flags().StringVar(&sslgateway.SSLGatewaySpec.DisplayName, "display-name", "", "Display name of the SSL Gateway")
	sslgatewayEditCmd.Flags().BoolVar(&sslgateway.SSLGatewaySpec.Hsts, "hsts", false, "Enable HSTS")
	sslgatewayEditCmd.Flags().BoolVar(&sslgateway.SSLGatewaySpec.HttpsRedirect, "https-redirect", false, "Enable HTTPS redirect")
	sslgatewayEditCmd.Flags().StringVar(&sslgateway.SSLGatewaySpec.Reverse, "reverse", "", "Custom reverse for your SSL Gateway")
	sslgatewayEditCmd.Flags().BoolVar(&sslgateway.SSLGatewaySpec.ServerHttps, "server-https", false, "Contact backend servers over HTTPS")
	sslgatewayEditCmd.Flags().StringVar(&sslgateway.SSLGatewaySpec.SslConfiguration, "ssl-configuration", "", "SSL configuration (intermediate, internal, modern)")
	addInteractiveEditorFlag(sslgatewayEditCmd)
	sslgatewayCmd.AddCommand(sslgatewayEditCmd)

	rootCmd.AddCommand(sslgatewayCmd)
}
