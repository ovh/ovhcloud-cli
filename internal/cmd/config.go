package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/config"
	"github.com/spf13/cobra"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage your CLI configuration",
	}

	// Disable parent pre-run that verifies if the API client is correctly initialized
	configCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	// Command to show the full config
	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show CLI configuration",
		Run:   config.ShowConfig,
	})

	configCmd.AddCommand(&cobra.Command{
		Example:               "ovhcloud config set default_cloud_project xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Use:                   "set <configuration key> <configuration value>",
		Short:                 "Set a value in the CLI configuration",
		Run:                   config.SetConfig,
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
	})

	configCmd.AddCommand(&cobra.Command{
		Example:               "ovhcloud config set-endpoint EU",
		Use:                   "set-endpoint <region>",
		Short:                 "Configure CLI to use the given API endpoint (EU, CA, US), or a specific URL (e.g. https://eu.api.ovh.com/v1)",
		Run:                   config.SetEndpoint,
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
	})

	rootCmd.AddCommand(configCmd)
}
