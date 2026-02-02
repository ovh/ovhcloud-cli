// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
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

	// Redirection subcommand
	emaildomainRedirectionCmd := &cobra.Command{
		Use:   "redirection",
		Short: "Manage email redirections for your domain",
	}
	emaildomainCmd.AddCommand(emaildomainRedirectionCmd)

	// List redirections
	emaildomainRedirectionListCmd := &cobra.Command{
		Use:     "list <service_name>",
		Aliases: []string{"ls"},
		Short:   "List all email redirections for a domain",
		Args:    cobra.ExactArgs(1),
		Run:     emaildomain.ListRedirections,
	}
	emaildomainRedirectionCmd.AddCommand(withFilterFlag(emaildomainRedirectionListCmd))

	// Get a specific redirection
	emaildomainRedirectionCmd.AddCommand(&cobra.Command{
		Use:   "get <service_name> <redirection_id>",
		Short: "Get details of a specific email redirection",
		Args:  cobra.ExactArgs(2),
		Run:   emaildomain.GetRedirection,
	})

	// Create redirection
	createRedirectionCmd := &cobra.Command{
		Use:   "create <service_name>",
		Short: "Create a new email redirection",
		Args:  cobra.ExactArgs(1),
		Run:   emaildomain.CreateRedirection,
	}
	createRedirectionCmd.Flags().StringVar(&emaildomain.RedirectionSpec.From, "from", "", "Source email address (e.g., alias@domain.com)")
	createRedirectionCmd.Flags().StringVar(&emaildomain.RedirectionSpec.To, "to", "", "Destination email address")
	createRedirectionCmd.Flags().BoolVar(&emaildomain.RedirectionSpec.LocalCopy, "local-copy", false, "Keep a local copy of the email")

	addInitParameterFileFlag(createRedirectionCmd, assets.EmaildomainOpenapiSchema, "/email/domain/{serviceName}/redirection", "post", emaildomain.RedirectionCreateExample, nil)
	addInteractiveEditorFlag(createRedirectionCmd)
	addFromFileFlag(createRedirectionCmd)
	createRedirectionCmd.MarkFlagsMutuallyExclusive("from-file", "editor")

	emaildomainRedirectionCmd.AddCommand(createRedirectionCmd)

	// Delete redirection
	emaildomainRedirectionCmd.AddCommand(&cobra.Command{
		Use:   "delete <service_name> <redirection_id>",
		Short: "Delete an email redirection",
		Args:  cobra.ExactArgs(2),
		Run:   emaildomain.DeleteRedirection,
	})

	rootCmd.AddCommand(emaildomainCmd)
}
