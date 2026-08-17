// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// LogoutAssumeYes skips the confirmation prompt when set (via -y/--yes).
var LogoutAssumeYes bool

// Logout revokes the current API credentials server-side and removes them from
// the local configuration file.
func Logout(_ *cobra.Command, _ []string) {
	cfg := flags.CliConfig
	path := flags.CliConfigPath

	if cfg == nil || path == "" {
		display.OutputError(&flags.OutputFormatConfig, "no configuration file found, nothing to log out from")
		return
	}
	
	var section string
	if profileName := config.GetActiveProfileName(cfg, flags.Profile); profileName != "" && !config.IsDefaultProfile(profileName) {
		section = config.ProfileSectionName(profileName)
	} else {
		endpoint, err := config.GetConfigValue(cfg, "default", "endpoint")
		if err != nil || endpoint == "" {
			display.OutputError(&flags.OutputFormatConfig, "no endpoint configured, nothing to log out from")
			return
		}
		section = endpoint
	}

	if !LogoutAssumeYes && !confirmLogout(section) {
		display.OutputInfo(&flags.OutputFormatConfig, nil, "Logout cancelled")
		return
	}

	// 1. Revoke the current credentials server-side (best effort): the
	// credentials may already be invalid, in which case we still want to clean
	// up the local configuration.
	if httpLib.Client != nil {
		if err := httpLib.Client.Post("/auth/logout", nil, nil); err != nil {
			display.OutputWarning(&flags.OutputFormatConfig, "could not revoke credentials via the API (they may already be invalid): %s", err)
		} else {
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🔒 API credentials revoked")
		}
	} else {
		display.OutputWarning(&flags.OutputFormatConfig, "API client not initialized, skipping remote revocation")
	}

	// 2. Remove the credentials from the local configuration.
	if err := config.DeleteCredentials(cfg, path, section); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to remove credentials from configuration: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Logged out successfully (credentials removed from %s)", path)
}

// confirmLogout asks the user to confirm the logout. It returns true only if
// the user explicitly answers yes.
func confirmLogout(section string) bool {
	fmt.Fprintf(os.Stderr, "⚠️  This will revoke the current API credentials and remove them from section %q of your configuration.\nContinue? [y/N]: ", section)

	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	return answer == "y" || answer == "yes"
}
