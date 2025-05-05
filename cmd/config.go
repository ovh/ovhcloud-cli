package cmd

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
)

func showConfig(_ *cobra.Command, _ []string) {
	display.RenderConfigTable(cliConfig)
}

func setConfig(_ *cobra.Command, args []string) {
	if _, ok := config.ConfigurableFields[args[0]]; !ok {
		allowedKeys := slices.Collect(maps.Keys(config.ConfigurableFields))
		display.ExitError("unknown configuration field %q, customizable fields are: %s", args[0], allowedKeys)
	}
	if err := config.SetConfigValue(cliConfig, cliConfigPath, "", args[0], args[1]); err != nil {
		display.ExitError("failed to set configuration: %s", err)
	}
}

func setRegion(_ *cobra.Command, args []string) {
	if args[0] != "EU" && args[0] != "CA" && args[0] != "US" {
		display.ExitError("invalid region %q, valid values are [EU, CA, US]", args[0])
	}

	if err := config.SetConfigValue(cliConfig, cliConfigPath, "", "endpoint", fmt.Sprintf("ovh-%s", strings.ToLower(args[0]))); err != nil {
		display.ExitError("failed to set region configuration: %s", err)
	}
}

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
		Run:   showConfig,
	})

	configCmd.AddCommand(&cobra.Command{
		Example:               "ovh-cli config set default_cloud_project xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		Use:                   "set <configuration key> <configuration value>",
		Short:                 "Set a value in the CLI configuration",
		Run:                   setConfig,
		Args:                  cobra.ExactArgs(2),
		DisableFlagsInUseLine: true,
	})

	configCmd.AddCommand(&cobra.Command{
		Example:               "ovh-cli config set-region EU",
		Use:                   "set-region <region>",
		Short:                 "Configure CLI to use the given API region (EU, CA, US)",
		Run:                   setRegion,
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
	})

	rootCmd.AddCommand(configCmd)
}
