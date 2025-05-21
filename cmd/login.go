package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/login"
)

func init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your OVHcloud account to create API credentials",
		Run:   login.Login,
	}

	// Disable parent pre-run that verifies if the API client is correctly initialized
	loginCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	rootCmd.AddCommand(loginCmd)
}
