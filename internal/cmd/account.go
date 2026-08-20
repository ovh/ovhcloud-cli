// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/completion"
	"github.com/ovh/ovhcloud-cli/internal/services/account"
	"github.com/spf13/cobra"
)

func init() {
	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Manage your account",
	}

	accountCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Retrieve basic personal information",
		Run:   account.GetMe,
	})

	// Commands to manage SSH keys
	sshKeysCmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage your SSH keys",
	}
	accountCmd.AddCommand(sshKeysCmd)

	sshKeysCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your SSH keys",
		Run:     account.ListSSHKeys,
	}))

	// API commands
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Manage your API credentials",
	}
	accountCmd.AddCommand(apiCmd)

	oauth2Cmd := &cobra.Command{
		Use:   "oauth2",
		Short: "Manage your OAuth2 clients",
	}
	apiCmd.AddCommand(oauth2Cmd)

	oauth2ClientCmd := &cobra.Command{
		Use:   "client",
		Short: "Manage your OAuth2 clients",
	}
	oauth2Cmd.AddCommand(oauth2ClientCmd)

	oauth2ClientCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your OAuth2 clients",
		Run:     account.ListOAuth2Clients,
	}))

	oauth2ClientCmd.AddCommand(&cobra.Command{
		Use:               "get <client_id>",
		Short:             "Get details of an OAuth2 client",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/me/api/oauth2/client"),
		Run:               account.GetOauth2Client,
	})

	oauth2CreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new OAuth2 client",
		Run:   account.CreateOAuth2Client,
	}
	oauth2CreateCmd.Flags().StringArrayVar(&account.Oauth2ClientSpec.CallbackUrls, "callback-urls", nil, "Callback URLs for the OAuth2 client")
	oauth2CreateCmd.Flags().StringVar(&account.Oauth2ClientSpec.Description, "description", "", "Description of the OAuth2 client")
	oauth2CreateCmd.Flags().StringVar(&account.Oauth2ClientSpec.Flow, "flow", "AUTHORIZATION_CODE", "OAuth2 flow type (default: AUTHORIZATION_CODE)")
	oauth2CreateCmd.Flags().StringVar(&account.Oauth2ClientSpec.Name, "name", "", "Name of the OAuth2 client")
	addParameterFileFlags(oauth2CreateCmd, false, assets.MeOpenapiSchema, "/me/api/oauth2/client", "post", account.Oauth2ClientCreateSample, nil)
	addInteractiveEditorFlag(oauth2CreateCmd)
	markFlagsMutuallyExclusive(oauth2CreateCmd, "from-file", "editor")

	oauth2ClientCmd.AddCommand(oauth2CreateCmd)

	oauth2ClientCmd.AddCommand(&cobra.Command{
		Use:               "delete <client_id>",
		Short:             "Delete the given OAuth2 client",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/me/api/oauth2/client"),
		Run:               account.DeleteOauth2Client,
	})

	oauth2ClientEditCmd := &cobra.Command{
		Use:               "edit <client_id>",
		Short:             "Edit an OAuth2 client",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.ServiceList("/v1/me/api/oauth2/client"),
		Run:               account.EditOauth2Client,
	}
	oauth2ClientEditCmd.Flags().StringArrayVar(&account.Oauth2ClientSpec.CallbackUrls, "callback-urls", nil, "Callback URLs for the OAuth2 client")
	oauth2ClientEditCmd.Flags().StringVar(&account.Oauth2ClientSpec.Description, "description", "", "Description of the OAuth2 client")
	oauth2ClientEditCmd.Flags().StringVar(&account.Oauth2ClientSpec.Name, "name", "", "Name of the OAuth2 client")
	addInteractiveEditorFlag(oauth2ClientEditCmd)
	oauth2ClientCmd.AddCommand(oauth2ClientEditCmd)

	// Billing commands
	billCmd := &cobra.Command{
		Use:   "bill",
		Short: "Read your invoices",
	}
	accountCmd.AddCommand(billCmd)

	billListCmd := withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your invoices",
		Long: "List your invoices.\n\n" +
			"Without --from, the window is the current month. That is a guard, not a\n" +
			"default for comfort: an account can hold thousands of invoices, and each\n" +
			"one listed is one request to detail it.",
		Run: account.ListBills,
	})
	addBillingWindowFlags(billListCmd)
	billListCmd.Flags().StringVar(&account.BillCategory, "category", "", "Keep only one category of invoice")
	billListCmd.Flags().Int64Var(&account.BillOrderID, "order-id", 0, "Keep only what was billed for this order")
	billListCmd.RegisterFlagCompletionFunc("category", account.CompleteBillCategory)
	billListCmd.Flags().BoolVar(&account.RevealBillSecrets, "reveal", false,
		"Print the download links instead of their fingerprints")
	billCmd.AddCommand(billListCmd)

	billGetCmd := &cobra.Command{
		Use:   "get <bill_id>",
		Short: "Get one invoice",
		Args:  cobra.ExactArgs(1),
		Run:   account.GetBill,
	}
	billGetCmd.Flags().BoolVar(&account.RevealBillSecrets, "reveal", false,
		"Print the download link and the PDF password instead of their fingerprints")
	billCmd.AddCommand(billGetCmd)

	billCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "details <bill_id>",
		Short: "List what one invoice charges for",
		Args:  cobra.ExactArgs(1),
		Run:   account.ListBillDetails,
	}))

	// Refunds
	refundCmd := &cobra.Command{
		Use:   "refund",
		Short: "Read your refunds",
	}
	accountCmd.AddCommand(refundCmd)

	refundListCmd := withFilterFlag(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your refunds",
		Run:     account.ListRefunds,
	})
	addBillingWindowFlags(refundListCmd)
	refundListCmd.Flags().Int64Var(&account.BillOrderID, "order-id", 0, "Keep only what was refunded for this order")
	refundListCmd.Flags().BoolVar(&account.RevealBillSecrets, "reveal", false,
		"Print the download links instead of their fingerprints")
	refundCmd.AddCommand(refundListCmd)

	refundGetCmd := &cobra.Command{
		Use:   "get <refund_id>",
		Short: "Get one refund",
		Args:  cobra.ExactArgs(1),
		Run:   account.GetRefund,
	}
	refundGetCmd.Flags().BoolVar(&account.RevealBillSecrets, "reveal", false,
		"Print the download link and the PDF password instead of their fingerprints")
	refundCmd.AddCommand(refundGetCmd)

	// Usage running against the next invoice
	usageCmd := withFilterFlag(&cobra.Command{
		Use:   "usage",
		Short: "Show what is running against your next invoice",
		Long: "Show what is running against your next invoice.\n\n" +
			"Dedicated servers do not appear here: they are billed at a flat rate, not\n" +
			"per usage. What a machine costs is answered by: ovhcloud baremetal cost <server>",
		Run: account.ShowUsage,
	})
	usageCmd.Flags().BoolVar(&account.BillUsageForecast, "forecast", false,
		"Show the forecast for the period instead of the usage so far")
	accountCmd.AddCommand(usageCmd)

	rootCmd.AddCommand(accountCmd)
}

// addBillingWindowFlags gives a listing its window. Both bounds accept a plain
// YYYY-MM-DD as well as a full RFC3339 timestamp.
func addBillingWindowFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&account.BillFrom, "from", "",
		"Start of the window, YYYY-MM-DD or RFC3339 (default: first day of the current month)")
	cmd.Flags().StringVar(&account.BillTo, "to", "",
		"End of the window, YYYY-MM-DD or RFC3339")
}
