// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"fmt"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	serviceconfig "github.com/ovh/ovhcloud-cli/internal/services/config"
	"github.com/spf13/cobra"
)

func Login(_ *cobra.Command, _ []string) {
	selectedRegion := display.RunLoginPicker("Which OVHcloud API do you want to login to ?", []string{"EU", "CA", "US", "Custom endpoint"})

	if selectedRegion == "" {
		return
	}
	customEndpoint := selectedRegion == "Custom endpoint"

	credentials := display.RunLoginInput(customEndpoint)
	for k, v := range credentials {
		if v == "" {
			display.ExitWarning("no value provided for %q", k)
			return
		}
	}

	// If no configuration file could be loaded, choose the location to write a new one
	if flags.CliConfigPath == "" {
		choices := make(map[string]string, len(config.ConfigPaths))
		expandedPaths := config.ExpandConfigPaths()
		for idx, cfg := range config.ConfigPaths {
			choices[cfg] = expandedPaths[idx]
		}

		_, path, err := display.RunGenericChoicePicker("Please choose a location to store your configuration", choices, 0)
		if err != nil {
			display.ExitError("failed to select a config path: %s", err)
			return
		}

		if path == "" {
			display.ExitWarning("no config path selected, configuration not saved")
			return
		}

		flags.CliConfigPath = path
	}

	// Set API endpoint to use in config
	if customEndpoint {
		selectedRegion = credentials["endpoint"]
		delete(credentials, "endpoint")
	} else {
		selectedRegion = fmt.Sprintf("ovh-%s", strings.ToLower(selectedRegion))
	}
	serviceconfig.SetEndpoint(nil, []string{selectedRegion})

	// Set credentials in config
	for k, v := range credentials {
		if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, selectedRegion, k, v); err != nil {
			display.ExitError("failed to write configuration %q: %s", k, err)
			return
		}
	}
}
