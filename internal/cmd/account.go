package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/flags"
	"stash.ovh.net/api/ovh-cli/internal/services/account"
)

func init() {
	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Manage your account",
	}

	// Commands to manage SSH keys
	sshKeysCmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage your SSH keys",
	}
	accountCmd.AddCommand(sshKeysCmd)

	sshKeysCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "list",
		Short: "List your SSH keys",
		Run:   account.ListSSHKeys,
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
		Use:   "list",
		Short: "List your OAuth2 clients",
		Run:   account.ListOAuth2Clients,
	}))

	oauth2ClientCmd.AddCommand(&cobra.Command{
		Use:   "get <client_id>",
		Short: "Get details of an OAuth2 client",
		Args:  cobra.ExactArgs(1),
		Run:   account.GetOauth2Client,
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
	addInitParameterFileFlag(oauth2CreateCmd, account.MeOpenapiSchema, "/me/api/oauth2/client", "post", account.Oauth2ClientCreateSample, nil)
	oauth2CreateCmd.Flags().StringVar(&flags.ParametersFile, "from-file", "", "File containing creation parameters")
	oauth2CreateCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define creation parameters")
	oauth2CreateCmd.MarkFlagsMutuallyExclusive("from-file", "editor")
	oauth2ClientCmd.AddCommand(oauth2CreateCmd)

	oauth2ClientCmd.AddCommand(&cobra.Command{
		Use:   "delete <client_id>",
		Short: "Delete the given OAuth2 client",
		Args:  cobra.ExactArgs(1),
		Run:   account.DeleteOauth2Client,
	})

	oauth2ClientEditCmd := &cobra.Command{
		Use:   "edit <client_id>",
		Short: "Edit an OAuth2 client",
		Args:  cobra.ExactArgs(1),
		Run:   account.EditOauth2Client,
	}
	oauth2ClientEditCmd.Flags().StringArrayVar(&account.Oauth2ClientSpec.CallbackUrls, "callback-urls", nil, "Callback URLs for the OAuth2 client")
	oauth2ClientEditCmd.Flags().StringVar(&account.Oauth2ClientSpec.Description, "description", "", "Description of the OAuth2 client")
	oauth2ClientEditCmd.Flags().StringVar(&account.Oauth2ClientSpec.Name, "name", "", "Name of the OAuth2 client")
	oauth2ClientEditCmd.Flags().BoolVar(&flags.ParametersViaEditor, "editor", false, "Use a text editor to define creation parameters")
	oauth2ClientCmd.AddCommand(oauth2ClientEditCmd)

	rootCmd.AddCommand(accountCmd)
}
