// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ovh/go-ovh/ovh"
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
		switch err := httpLib.Client.Post("/auth/logout", nil, nil); {
		case err == nil:
			display.OutputInfo(&flags.OutputFormatConfig, nil, "🔒 API credentials revoked")
		case isInvalidCredentialError(err):
			display.OutputWarning(&flags.OutputFormatConfig, "credentials were already invalid or revoked, skipping remote revocation")
		default:
			display.OutputWarning(&flags.OutputFormatConfig, "could not revoke credentials via the API: %s", err)
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

// confirmLogout asks the user to confirm the logout. The default (an empty
// answer, i.e. just pressing Enter) is "no": it returns true only if the user
// explicitly answers "y"/"yes". The prompt can be skipped entirely with the
// -y/--yes flag.
func confirmLogout(section string) bool {
	fmt.Fprintf(os.Stderr,
		"⚠️  This will revoke the current API credentials and remove them from section %q of your configuration.\n"+
			"Continue? [y/N] (press Enter to cancel; use -y/--yes to skip this prompt): ",
		section)

	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	return answer == "y" || answer == "yes"
}

// isInvalidCredentialError reports whether err is an OVHcloud API error caused
// by credentials that are already invalid or revoked (HTTP 401/403).
func isInvalidCredentialError(err error) bool {
	var apiErr *ovh.APIError
	return errors.As(err, &apiErr) &&
		(apiErr.Code == http.StatusUnauthorized || apiErr.Code == http.StatusForbidden)
}
