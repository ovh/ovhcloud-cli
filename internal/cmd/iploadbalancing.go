// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	_ "embed"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/iploadbalancing"
	"github.com/spf13/cobra"
)

func init() {
	iploadbalancingCmd := &cobra.Command{
		Use:   "iploadbalancing",
		Short: "Retrieve information and manage your IP LoadBalancing services",
	}

	// Command to list IpLoadbalancing services
	iploadbalancingListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your IpLoadbalancing services",
		Run:     iploadbalancing.ListIpLoadbalancing,
	}
	iploadbalancingCmd.AddCommand(withFilterFlag(iploadbalancingListCmd))

	// Command to get a single IpLoadbalancing
	iploadbalancingCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name>",
		Short:             "Retrieve information of a specific IpLoadbalancing",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.GetIpLoadbalancing,
	})

	// Command to update a single IpLoadbalancing
	iploadbalancingEditCmd := &cobra.Command{
		Use:               "edit <service_name>",
		Short:             "Edit the given IpLoadbalancing",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.EditIpLoadbalancing,
	}
	iploadbalancingEditCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSpec.DisplayName, "display-name", "", "Display name of the load balancer")
	iploadbalancingEditCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSpec.SSLConfiguration, "ssl-configuration", "", "SSL configuration of the load balancer (intermediate, modern)")
	addInteractiveEditorFlag(iploadbalancingEditCmd)
	iploadbalancingCmd.AddCommand(iploadbalancingEditCmd)

	// SSL commands
	iploadbalancingSSLCmd := &cobra.Command{
		Use:   "ssl",
		Short: "Manage SSL certificates of your IpLoadbalancing services",
	}
	iploadbalancingCmd.AddCommand(iploadbalancingSSLCmd)

	// Command to list SSL certificates of an IpLoadbalancing
	iploadbalancingSSLCmd.AddCommand(&cobra.Command{
		Use:               "list <service_name>",
		Aliases:           []string{"ls"},
		Short:             "List SSL certificate IDs of the given IpLoadbalancing",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.ListIpLoadbalancingSSL,
	})

	// Command to get a single SSL certificate of an IpLoadbalancing
	iploadbalancingSSLCmd.AddCommand(&cobra.Command{
		Use:               "get <service_name> <ssl_id>",
		Short:             "Retrieve information of a specific SSL certificate of the given IpLoadbalancing",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.GetIpLoadbalancingSSL,
	})

	// Command to create a SSL certificate on an IpLoadbalancing
	iploadbalancingSSLCmd.AddCommand(getIpLoadbalancingSSLCreateCmd())

	// Command to delete a SSL certificate of an IpLoadbalancing
	iploadbalancingSSLCmd.AddCommand(&cobra.Command{
		Use:               "delete <service_name> <ssl_id>",
		Short:             "Delete a custom SSL certificate of the given IpLoadbalancing",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.DeleteIpLoadbalancingSSL,
	})

	rootCmd.AddCommand(iploadbalancingCmd)
}

// getIpLoadbalancingSSLCreateCmd returns the command to add a new custom SSL certificate on an IpLoadbalancing
func getIpLoadbalancingSSLCreateCmd() *cobra.Command {
	sslCreateCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Add a new custom SSL certificate on the given IpLoadbalancing",
		Long: `Use this command to add a new custom SSL certificate on the given IpLoadbalancing.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud iploadbalancing ssl create <service_name> --certificate "$(cat cert.pem)" --key "$(cat key.pem)"

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iploadbalancing ssl create <service_name> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud iploadbalancing ssl create <service_name> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iploadbalancing ssl create <service_name>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iploadbalancing ssl create <service_name> --from-file ./params.json --display-name nameoverriden

3. Using your default text editor:

	ovhcloud iploadbalancing ssl create <service_name> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iploadbalancing ssl create <service_name> --editor --display-name nameoverriden
`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/ipLoadbalancing"),
		Run:               iploadbalancing.CreateIpLoadbalancingSSL,
	}

	sslCreateCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSSLSpec.Certificate, "certificate", "", "SSL certificate (PEM)")
	sslCreateCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSSLSpec.Chain, "chain", "", "SSL certificate chain (PEM)")
	sslCreateCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSSLSpec.DisplayName, "display-name", "", "Human readable name for your SSL certificate")
	sslCreateCmd.Flags().StringVar(&iploadbalancing.IPLoadbalancingSSLSpec.Key, "key", "", "SSL private key (PEM)")

	// Common flags for other means to define parameters
	addParameterFileFlags(sslCreateCmd, false, assets.IploadbalancingOpenapiSchema, "/ipLoadbalancing/{serviceName}/ssl", "post", iploadbalancing.IPLoadbalancingSSLCreateExample, nil)
	addInteractiveEditorFlag(sslCreateCmd)
	markFlagsMutuallyExclusive(sslCreateCmd, "from-file", "editor")

	return sslCreateCmd
}
