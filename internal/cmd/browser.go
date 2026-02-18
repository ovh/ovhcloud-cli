// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/flags"
	"github.com/ovh/ovhcloud-cli/internal/services/browser"
	"github.com/spf13/cobra"
)

func init() {
	browserCmd := &cobra.Command{
		Use:   "browser",
		Short: "Launch a TUI for the OVHcloud Manager - Public Cloud universe only [EXPERIMENTAL]",
		Long: `Launch an interactive Terminal User Interface that simulates the
OVHcloud Manager (https://manager.eu.ovhcloud.com/#/public-cloud/) - Public Cloud universe only.

⚠️  EXPERIMENTAL FEATURE - This navigation is experimental and may contain bugs.
If you encounter any issues, please report them at:
https://github.com/ovh/ovhcloud-cli/issues

Navigate through your Public Cloud services using keyboard controls.
The browser makes direct API calls to fetch and display real data.

Features:
  - Real-time data fetching from OVHcloud API
  - Table views for projects, instances, and services
  - Hierarchical navigation through cloud resources
  - Web-like interface in your terminal
  - Debug mode to view API requests and request IDs

Navigate using:
  - ↑↓: Move through menus/tables
  - Enter: Select item or view details
  - ←/Esc: Go back
  - d: Toggle debug panel (show API requests)
  - q: Quit`,
		Run: browser.StartBrowser,
	}

	browserCmd.Flags().BoolVar(&flags.Debug, "debug", false, "Enable debug mode to view API requests and request IDs")

	rootCmd.AddCommand(browserCmd)
}
