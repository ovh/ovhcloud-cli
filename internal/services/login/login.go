// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"fmt"
	"os"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	serviceconfig "github.com/ovh/ovhcloud-cli/internal/services/config"
	"github.com/spf13/cobra"
)

// LoginProfileFlag is set by the --profile flag on the login command.
var LoginProfileFlag string

// gettingStartedMessage returns a short onboarding banner listing the first
// commands a freshly-authenticated user is likely to need.
func gettingStartedMessage() string {
	return strings.Join([]string{
		"",
		"👉 A few commands to get started:",
		"",
		"  • List your public cloud projects:  ovhcloud cloud project list",
		"  • Set a default cloud project:      ovhcloud config set default_cloud_project <id>",
		"  • Browse a service:                 ovhcloud <service> --help",
		"  • List all available commands:      ovhcloud --help",
	}, "\n")
}

func Login(_ *cobra.Command, _ []string) {
	selectedRegion := display.RunLoginPicker("Which OVHcloud API do you want to login to ?", []string{"EU", "CA", "US", "Custom endpoint"})

	if selectedRegion == "" {
		return
	}
	customEndpoint := selectedRegion == "Custom endpoint"

	credentials := display.RunLoginInput(customEndpoint)
	for k, v := range credentials {
		if v == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "no value provided for %q", k)
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
			display.OutputError(&flags.OutputFormatConfig, "failed to select a config path: %s", err)
			return
		}

		if path == "" {
			display.OutputWarning(&flags.OutputFormatConfig, "no config path selected, configuration not saved")
			return
		}

		flags.CliConfigPath = path

		defer func() {
			if err := os.Chmod(flags.CliConfigPath, 0600); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to set permissions on config file: %s", err)
			}
		}()
	}

	// Determine the endpoint value
	endpoint := fmt.Sprintf("ovh-%s", strings.ToLower(selectedRegion))
	if customEndpoint {
		endpoint = credentials["endpoint"]
		delete(credentials, "endpoint")
	}

	// If a profile name is provided, store credentials in the profile section
	if LoginProfileFlag != "" {
		if config.IsDefaultProfile(LoginProfileFlag) {
			display.OutputError(&flags.OutputFormatConfig, "%q is a reserved profile name, please choose a different name", config.DefaultProfileName)
			return
		}

		// Store endpoint in the profile section
		if err := config.SetProfileConfigValue(flags.CliConfig, flags.CliConfigPath, LoginProfileFlag, "endpoint", endpoint); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to write profile endpoint: %s", err)
			return
		}

		// Store credentials in the profile section
		for k, v := range credentials {
			if err := config.SetProfileConfigValue(flags.CliConfig, flags.CliConfigPath, LoginProfileFlag, k, v); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to write profile configuration %q: %s", k, err)
				return
			}
		}

		// If no active profile is set yet, set this one as active
		if !config.IsProfileMode(flags.CliConfig, "") {
			if err := config.SetActiveProfile(flags.CliConfig, flags.CliConfigPath, LoginProfileFlag); err != nil {
				display.OutputError(&flags.OutputFormatConfig, "failed to set active profile: %s", err)
				return
			}
		}

		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Credentials saved to profile %q\n%s", LoginProfileFlag, gettingStartedMessage())
		return
	}

	// Legacy mode: store credentials in endpoint section
	configSection := endpoint
	if !customEndpoint {
		selectedRegion = strings.ToUpper(selectedRegion)
	}
	serviceconfig.SetEndpoint(nil, []string{selectedRegion})

	for k, v := range credentials {
		if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, configSection, k, v); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to write configuration %q: %s", k, err)
			return
		}
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Credentials saved successfully\n%s", gettingStartedMessage())
}
