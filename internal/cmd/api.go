// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/apicall"
	"github.com/spf13/cobra"
)

func init() {
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Call the OVHcloud API directly",
		Long: `Call any OVHcloud API endpoint with the credentials already configured.

The CLI covers a part of the API surface: this is the way out for the rest.
Requests are signed with the active profile and rendered with the usual
--output formats, so an endpoint with no dedicated command is still usable
from a script.

It is a raw passthrough. None of the confirmations, parameter validation or
safety checks the product commands provide apply here, so use --dry-run to
see what would be sent before sending it.`,
	}

	apiCallCmd := &cobra.Command{
		Use:   "call <method> <path>",
		Short: "Call any OVHcloud API path",
		Long: `Call any OVHcloud API path with the credentials already configured.

The path may be given with or without its leading slash and API version:
"dedicated/server", "/dedicated/server" and "/v1/dedicated/server" are the
same endpoint. A path already starting with /v1/ or /v2/ is left untouched.

A body for POST and PUT is read from --from-file or written in $EDITOR with
--editor.`,
		Example: `  # List the dedicated servers of the account
  ovhcloud api call GET /dedicated/server

  # Read one server, as JSON
  ovhcloud api call GET /dedicated/server/ns3168421.ip-51-77-12.eu -o json

  # Extract a single field
  ovhcloud api call GET /dedicated/server/ns3168421.ip-51-77-12.eu -o 'datacenter'

  # Reach a v2 endpoint
  ovhcloud api call GET /v2/publicCloud/project

  # Send a body, checking first what would leave
  ovhcloud api call PUT /dedicated/server/ns3168421.ip-51-77-12.eu --from-file body.json --dry-run`,
		Args: cobra.ExactArgs(2),
		Run:  apicall.Call,
	}

	addParameterFileFlags(apiCallCmd, true, nil, "", "", "", nil)
	addInteractiveEditorFlag(apiCallCmd)
	// Both together used to be accepted, with the file quietly winning. A
	// command whose whole promise is to send exactly what it was given must not
	// choose between two payloads without saying so.
	markFlagsMutuallyExclusive(apiCallCmd, "from-file", "editor")
	apiCallCmd.Flags().BoolVar(&apicall.DryRun, "dry-run", false,
		"Print the request that would be sent, without sending it")

	apiCmd.AddCommand(apiCallCmd)
	rootCmd.AddCommand(apiCmd)
}
