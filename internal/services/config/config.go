package config

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
)

func ShowConfig(_ *cobra.Command, _ []string) {
	display.RenderConfigTable(flags.CliConfig)
}

func SetConfig(_ *cobra.Command, args []string) {
	if _, ok := config.ConfigurableFields[args[0]]; !ok {
		allowedKeys := slices.Collect(maps.Keys(config.ConfigurableFields))
		display.ExitError("unknown configuration field %q, customizable fields are: %s", args[0], allowedKeys)
	}
	if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", args[0], args[1]); err != nil {
		display.ExitError("failed to set configuration: %s", err)
	}
}

func SetRegion(_ *cobra.Command, args []string) {
	if args[0] != "EU" && args[0] != "CA" && args[0] != "US" {
		display.ExitError("invalid region %q, valid values are [EU, CA, US]", args[0])
	}

	if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", "endpoint", fmt.Sprintf("ovh-%s", strings.ToLower(args[0]))); err != nil {
		display.ExitError("failed to set region configuration: %s", err)
	}
}
