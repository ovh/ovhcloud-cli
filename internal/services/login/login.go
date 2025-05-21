package login

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"stash.ovh.net/api/ovh-cli/internal/config"
	"stash.ovh.net/api/ovh-cli/internal/display"
	"stash.ovh.net/api/ovh-cli/internal/flags"
)

func Login(_ *cobra.Command, _ []string) {
	selectedRegion := display.RunLoginPicker("Which OVHcloud API do you want to login to ?", []string{"EU", "CA", "US"})

	if selectedRegion == "" {
		return
	}

	credentials := display.RunLoginInput()

	for k, v := range credentials {
		if v == "" {
			display.ExitError("no value provided for %q", k)
		}
	}

	regionConfigKey := fmt.Sprintf("ovh-%s", strings.ToLower(selectedRegion))

	if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", "endpoint", regionConfigKey); err != nil {
		display.ExitError("failed to write endpoint in configuration: %s", err)
	}

	for k, v := range credentials {
		if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, regionConfigKey, k, v); err != nil {
			display.ExitError("failed to write configuration %q: %s", k, err)
		}
	}
}
