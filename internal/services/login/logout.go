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
	//
	// Its outcome is recorded and reported at the end rather than printed here.
	// It used to be printed here, by display.OutputWarning — and OutputWarning
	// does not return: OutputWithFormat finishes on ExitFunc(0), which is
	// os.Exit. So on every path but the happy one, `ovhcloud logout` stopped
	// right here, exit 0, having printed "skipping remote revocation" — a
	// sentence that reads as "carrying on" — while step 2 never ran and the
	// credentials stayed in the configuration file. Reproduced against the built
	// binary with a revoked key: exit 0, and consumer_key still on disk.
	//
	// The comment above already said the opposite of what the code did.
	revocation := "not attempted"
	if httpLib.Client != nil {
		switch err := httpLib.Client.Post("/auth/logout", nil, nil); {
		case err == nil:
			revocation = "revoked via the API"
		case isInvalidCredentialError(err):
			revocation = "already invalid or revoked, so nothing to revoke remotely"
		default:
			revocation = fmt.Sprintf("could not be revoked via the API: %s", err)
		}
	} else {
		revocation = "not revoked remotely: the API client was not initialised"
	}

	// 2. Remove the credentials from the local configuration.
	//
	// This is the half that matters, and the half that used to be skipped: the
	// remote revocation is a courtesy, taking the key off this disk is the
	// command.
	if err := config.DeleteCredentials(cfg, path, section); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to remove credentials from %s: %s\n   The remote revocation %s, so the key on this disk is what is left to deal with.",
			path, err, revocation)
		return
	}

	// One document, and it carries both halves. Under -o json the outcome of the
	// revocation is a field rather than a separate document printed before this
	// one — two JSON documents on one stdout is something no parser accepts.
	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"configFile": path, "section": section, "remoteRevocation": revocation},
		"✅ Logged out: credentials removed from %s (remote revocation: %s)", path, revocation)
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
