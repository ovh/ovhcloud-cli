// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/browser"
	"github.com/spf13/cobra"
)

func init() {
	browserCmd := &cobra.Command{
		Use:   "browser",
		Short: "Launch a TUI simulating the OVHcloud Manager interface",
		Long: `Launch an interactive Terminal User Interface that simulates the 
OVHcloud Manager (https://manager.eu.ovhcloud.com/).

Navigate through your Public Cloud services using keyboard controls.
The browser makes direct API calls to fetch and display real data.

Features:
  - Real-time data fetching from OVHcloud API
  - Table views for projects, instances, and services
  - Hierarchical navigation through cloud resources
  - Web-like interface in your terminal

Navigate using:
  - ↑↓: Move through menus/tables
  - Enter: Select item or view details
  - ←/Esc: Go back
  - q: Quit`,
		Run: browser.StartBrowser,
	}

	rootCmd.AddCommand(browserCmd)
}
