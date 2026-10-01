// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/vrack"
	"github.com/spf13/cobra"
)

// The vRack commands exist twice: here, and under `ovhcloud vrack`. The network
// view is the right home for the object, and it is not where somebody holding a
// server looks — this CLI already has a command nobody finds for that reason,
// `ip reverse set`, which is invisible from `baremetal --help`.
//
// Only the argument order differs: you name what you have first. The work is
// done by the vrack package, so the two cannot drift apart.

// AttachBaremetalToVrack attaches this server to the named vRack.
func AttachBaremetalToVrack(_ *cobra.Command, args []string) {
	vrack.Attach(args[1], args[0])
}

// DetachBaremetalFromVrack removes this server from a vRack.
//
// The vRack is optional because the server is in at most one, and looking it up
// is what this CLI is for: making an operator find an identifier the API
// already knows is how a gesture becomes a chore.
func DetachBaremetalFromVrack(_ *cobra.Command, args []string) {
	target := ""
	if len(args) > 1 {
		target = args[1]
	}

	if target == "" {
		found, err := theVrackOf(args[0])
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
		// Half this account's servers are in no vRack, so this is an ordinary
		// answer rather than an edge case. Falling through with an empty name
		// would build `/v1/vrack//dedicatedServerInterfaceDetails` and report a
		// listing failure for a vRack that was never named.
		if found == "" {
			display.OutputError(&flags.OutputFormatConfig,
				"%s is not in any vRack, so there is nothing to detach", args[0])
			return
		}
		target = found
	}

	vrack.Detach(target, args[0])
}

// ShowBaremetalVrack says which vRack this server is in, if any.
func ShowBaremetalVrack(_ *cobra.Command, args []string) {
	server := args[0]

	// show reads every membership rather than the first: a server with two
	// vRack interfaces can be in two vRacks, and hiding the second would be a
	// listing that answers "which vRack is this in" with half the truth.
	names, err := vracksOf(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(names) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"server": server, "vrack": nil},
			"%s is not in any vRack.", server)
		return
	}

	for _, name := range names {
		vrack.ShowMembership(name, server)
	}
}

// vrackOf reads the vRack a server belongs to.
//
// It answers an empty name rather than an error when there is none: not being
// in a vRack is an ordinary state, and half of this account's servers are in it.
//
// The API answers an array, and it does so because a server with two vRack
// interfaces can sit in two vRacks at once. Picking the first would let
// `baremetal vrack detach ns1` cut a network nobody named, which is the same
// mistake this package already refuses to make when a server has several
// interfaces — so it is refused here too, in the same terms.
func vracksOf(server string) ([]string, error) {
	var vracks []string
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/vrack", url.PathEscape(server))

	if err := httpLib.Client.Get(endpoint, &vracks); err != nil {
		return nil, fmt.Errorf("failed to read the vRack of %s: %w", server, err)
	}

	return vracks, nil
}

// theVrackOf is vracksOf for the callers that must act on exactly one.
//
// Reading and deciding are separate on purpose: `show` wants every membership,
// while `detach` must not guess between them.
func theVrackOf(server string) (string, error) {
	vracks, err := vracksOf(server)
	if err != nil {
		return "", err
	}

	switch len(vracks) {
	case 0:
		return "", nil
	case 1:
		return vracks[0], nil
	default:
		return "", fmt.Errorf(
			"%s is in %d vRacks (%s); name the one to detach from",
			server, len(vracks), strings.Join(vracks, ", "))
	}
}
