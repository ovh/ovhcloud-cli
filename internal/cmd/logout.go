// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httplib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/login"
	"github.com/spf13/cobra"
)

func init() {
	logoutCmd := &cobra.Command{
		Use:     "logout",
		Short:   "Revoke your API credentials and remove them from the configuration",
		Example: "ovhcloud logout\novhcloud logout --yes\novhcloud logout --profile work",
		Run:     login.Logout,
	}

	logoutCmd.Flags().BoolVarP(&login.LogoutAssumeYes, "yes", "y", false, "Do not ask for confirmation")

	// Initialize the API client if possible, but — unlike other commands — do
	// not abort when it cannot be initialized: logout must still be able to
	// clean up local credentials even when they are already invalid or missing.
	logoutCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		config.ActiveProfileOverride = flags.Profile
		if httplib.Client == nil {
			httplib.InitClientWithProfile(flags.CliConfig, flags.Profile)
		}
	}

	rootCmd.AddCommand(logoutCmd)
}
