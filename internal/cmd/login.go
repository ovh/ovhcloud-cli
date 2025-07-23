package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/login"
	"github.com/spf13/cobra"
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
