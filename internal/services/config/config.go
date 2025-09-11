// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/spf13/cobra"
)

var (
	validAPIRegions = []string{"EU", "CA", "US"}
)

func ShowConfig(_ *cobra.Command, _ []string) {
	display.RenderConfigTable(flags.CliConfig)
}

func SetConfig(_ *cobra.Command, args []string) {
	if _, ok := config.ConfigurableFields[args[0]]; !ok {
		allowedKeys := slices.Collect(maps.Keys(config.ConfigurableFields))
		display.ExitError("unknown configuration field %q, customizable fields are: %s", args[0], allowedKeys)
		return
	}
	if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", args[0], args[1]); err != nil {
		display.ExitError("failed to set configuration: %s", err)
		return
	}
}

func SetEndpoint(_ *cobra.Command, args []string) {
	var endpoint string

	if slices.Contains(validAPIRegions, args[0]) {
		endpoint = fmt.Sprintf("ovh-%s", strings.ToLower(args[0]))
	} else {
		// Check if given value is a valid URL
		url, err := url.Parse(args[0])
		if err != nil {
			display.ExitError("invalid API endpoint %q, valid values are [EU, CA, US] or a valid URL", args[0])
			return
		}

		if url.Scheme != "https" && url.Scheme != "http" {
			display.ExitError(`given url has an invalid scheme, only "http" and "https" are allowed`)
			return
		}

		endpoint = args[0]
	}

	if err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", "endpoint", endpoint); err != nil {
		display.ExitError("failed to set API endpoint configuration: %s", err)
		return
	}
}
