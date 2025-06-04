package cmd

import (
	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/services/config"
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
		Example:               "ovh-cli config set default_cloud_project xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Use:                   "set <configuration key> <configuration value>",
		Short:                 "Set a value in the CLI configuration",
		Run:                   config.SetConfig,
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
	})

	configCmd.AddCommand(&cobra.Command{
		Example:               "ovh-cli config set-region EU",
		Use:                   "set-region <region>",
		Short:                 "Configure CLI to use the given API region (EU, CA, US)",
		Run:                   config.SetRegion,
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
	})

	rootCmd.AddCommand(configCmd)
}
