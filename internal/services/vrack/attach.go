// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	// VrackInterface names the interface to act on when a server has several.
	VrackInterface string

	// VrackWait keeps the command running until the vRack reports the change.
	VrackWait bool

	// VrackTimeout bounds that wait.
	VrackTimeout = 10 * time.Minute

	taskPollInterval = 5 * time.Second
)

// AttachToVrack puts one dedicated server interface into a vRack.
func AttachToVrack(_ *cobra.Command, args []string) {
	attachToVrack(args[0], args[1])
}

// DetachFromVrack takes one dedicated server interface out of a vRack.
func DetachFromVrack(_ *cobra.Command, args []string) {
	detachFromVrack(args[0], args[1])
}

// Attach and Detach are the same work seen from the baremetal side, where the
// operator names the server first. They exist so the two command trees share
// one implementation rather than one being a copy that stops matching.
func Attach(vrack, target string) { attachToVrack(vrack, target) }

// Detach removes an interface from a vRack.
func Detach(vrack, target string) { detachFromVrack(vrack, target) }

// ShowMembership prints how one server sits in one vRack.
func ShowMembership(vrackName, server string) {
	attached, err := attachedInterfaces(vrackName)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	owned := interfacesOf(attached, server)
	if len(owned) == 0 {
		// The server API named this vRack and the vRack API does not know the
		// server. Saying so beats printing nothing: the two sides disagreeing
		// is information, and inventing an empty answer would hide it.
		display.OutputError(&flags.OutputFormatConfig,
			"%s reports being in %s, but %s does not list any of its interfaces",
			server, vrackName, vrackName)
		return
	}

	message := fmt.Sprintf("%s — %s\n  in %s via interface %s",
		server, owned[0].label(), vrackName, owned[0].Name)
	for _, itf := range owned[1:] {
		message += fmt.Sprintf("\n  in %s via interface %s", vrackName, itf.Name)
	}

	rows := make([]map[string]any, 0, len(owned))
	for _, itf := range owned {
		rows = append(rows, map[string]any{
			"interface": itf.UUID, "interfaceName": itf.Name,
		})
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"server": server, "vrack": vrackName, "interfaces": rows},
		"%s", message)
}

func attachToVrack(vrack, target string) {
	itf, err := resolveAttachable(vrack, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/vrack/%s/dedicatedServerInterface", url.PathEscape(vrack))

	if !common.ConfirmAction(common.Disruptive, vrack, attachWarning(vrack, itf)) {
		display.OutputError(&flags.OutputFormatConfig, "attachment to %s cancelled", vrack)
		return
	}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint,
		map[string]any{"dedicatedServerInterface": itf.UUID}, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to attach interface %s to %s: %s", itf.UUID, vrack, err)
		return
	}

	details := map[string]any{
		"vrack": vrack, "server": itf.Server, "interface": itf.UUID, "interfaceName": itf.Name,
	}

	if !VrackWait {
		display.OutputInfo(&flags.OutputFormatConfig, details,
			"⚡️ Attaching %s to %s. Follow it with:\n  ovhcloud vrack get %s",
			itf.label(), vrack, vrack)
		return
	}

	if err := waitForAttachment(vrack, itf.UUID, true, task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, details,
		"✅ %s is in %s, through interface %s.", itf.label(), vrack, itf.Name)
}

func detachFromVrack(vrack, target string) {
	itf, err := resolveAttached(vrack, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/vrack/%s/dedicatedServerInterface/%s",
		url.PathEscape(vrack), url.PathEscape(itf.UUID))

	if !common.ConfirmAction(common.Disruptive, vrack, detachWarning(vrack, itf)) {
		display.OutputError(&flags.OutputFormatConfig, "detachment from %s cancelled", vrack)
		return
	}

	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Delete(endpoint, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to detach interface %s from %s: %s", itf.UUID, vrack, err)
		return
	}

	details := map[string]any{
		"vrack": vrack, "server": itf.Server, "interface": itf.UUID, "interfaceName": itf.Name,
	}

	if !VrackWait {
		display.OutputInfo(&flags.OutputFormatConfig, details,
			"⚡️ Detaching %s from %s. Follow it with:\n  ovhcloud vrack get %s",
			itf.label(), vrack, vrack)
		return
	}

	if err := waitForAttachment(vrack, itf.UUID, false, task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, details,
		"✅ %s is out of %s.", itf.label(), vrack)
}

// attachWarning and detachWarning are what somebody reads in the second before
// they agree.
//
// They name the machine the way its owner does — "Yaniv - RISE-1 - LIM", not
// ns3141022.ip-51-77-67.eu — because the prompt is the last place a wrong
// target can still be noticed, and a hostname is an address, not a name. The
// interface is named too: a server can hold several, and cutting the wrong one
// is a different outage from cutting the machine.
func attachWarning(vrack string, itf serverInterface) string {
	return fmt.Sprintf("Attaching %s (interface %s) to %s changes how that machine reaches the network.",
		itf.label(), itf.Name, vrack)
}

func detachWarning(vrack string, itf serverInterface) string {
	return fmt.Sprintf("Detaching %s (interface %s) from %s cuts that machine off the private network.",
		itf.label(), itf.Name, vrack)
}

// resolveAttachable turns what the operator typed into the interface to attach.
//
// An operator names a machine; the API takes a UUID nobody knows by heart. The
// whole difficulty of this command is in between, and it is where the two
// failure modes that look identical have to be told apart: a server absent from
// allowedServices either has no vRack interface at all, or has one that is
// already attached elsewhere. Both produce the same silence, and they call for
// opposite actions.
func resolveAttachable(vrack, target string) (serverInterface, error) {
	attachable, err := attachableInterfaces(vrack)
	if err != nil {
		return serverInterface{}, err
	}

	if uuid := explicitUUID(target); uuid != "" {
		if VrackInterface != "" && !strings.EqualFold(VrackInterface, uuid) {
			return serverInterface{}, fmt.Errorf(
				"two different interfaces were named: %s as the target and %s with --interface", uuid, VrackInterface)
		}
		for _, itf := range attachable {
			if strings.EqualFold(itf.UUID, uuid) {
				return itf, nil
			}
		}
		// Absent from allowedServices has two causes, and one of them is that
		// the work is already done. Re-running an attach is what a script does
		// after a timeout, and telling it the interface is "attached to another
		// one" would send somebody looking for a vRack that does not exist.
		if attached, err := attachedInterfaces(vrack); err == nil {
			for _, itf := range attached {
				if strings.EqualFold(itf.UUID, uuid) {
					return serverInterface{}, fmt.Errorf(
						"interface %s is already in %s: nothing to do", itf.Name, vrack)
				}
			}
		}
		return serverInterface{}, fmt.Errorf(
			"interface %s cannot be attached to %s: it is not in the list of allowed services.\nList what can be attached with `ovhcloud vrack get %s`", uuid, vrack, vrack)
	}

	return pick(interfacesOf(attachable, target), target, func() error {
		return whyNotAttachable(vrack, target)
	})
}

// resolveAttached does the same for what is already in the vRack.
func resolveAttached(vrack, target string) (serverInterface, error) {
	attached, err := attachedInterfaces(vrack)
	if err != nil {
		return serverInterface{}, err
	}

	if uuid := explicitUUID(target); uuid != "" {
		for _, itf := range attached {
			if strings.EqualFold(itf.UUID, uuid) {
				return itf, nil
			}
		}
		return serverInterface{}, fmt.Errorf("interface %s is not attached to %s", uuid, vrack)
	}

	return pick(interfacesOf(attached, target), target, func() error {
		return fmt.Errorf(
			"%s has no interface in %s.\nSee what is attached with `ovhcloud vrack get %s`", target, vrack, vrack)
	})
}

// pick chooses among the interfaces a server owns.
//
// --interface is consulted first and checked against that same list rather than
// trusted: a UUID belonging to another machine would otherwise be sent to the
// API, which would accept it and attach something nobody named.
func pick(owned []serverInterface, server string, none func() error) (serverInterface, error) {
	if VrackInterface != "" {
		for _, itf := range owned {
			if strings.EqualFold(itf.UUID, VrackInterface) {
				return itf, nil
			}
		}
		switch len(owned) {
		case 0:
			return serverInterface{}, none()
		case 1:
			// Not a choice to make: there is one interface and the flag names a
			// different one. Saying "name the one to use" would send somebody
			// looking for an option they do not have.
			return serverInterface{}, fmt.Errorf(
				"interface %s does not belong to %s; its only interface is %s (%s)",
				VrackInterface, server, owned[0].Name, owned[0].UUID)
		default:
			return serverInterface{}, fmt.Errorf(
				"%s\n\ninterface %s is not one of them", ambiguous(server, owned), VrackInterface)
		}
	}

	// A display name held by two machines matches both. Refusing here is the
	// same refusal as for two interfaces of one server, for the same reason:
	// the alternative is cutting a network nobody named.
	if distinctServers(owned) > 1 {
		return serverInterface{}, ambiguousServers(server, owned)
	}

	switch len(owned) {
	case 1:
		return owned[0], nil
	case 0:
		return serverInterface{}, none()
	default:
		return serverInterface{}, ambiguous(server, owned)
	}
}

// ambiguousServers refuses a name that several machines answer to.
func ambiguousServers(name string, owned []serverInterface) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%q is the display name of several servers; use the hostname instead:", name)
	seen := map[string]bool{}
	for _, itf := range owned {
		if seen[itf.Server] {
			continue
		}
		seen[itf.Server] = true
		fmt.Fprintf(&b, "\n  %s", itf.Server)
	}
	return errors.New(b.String())
}

// explicitUUID reports the target as a UUID when that is what it is.
//
// Both forms have to be accepted: an operator types a server name, and a script
// reading `-o json` from a previous command has a UUID in hand. Telling them
// apart on the shape of the string is enough — a dedicated server name is a
// hostname and never parses as one of these.
func explicitUUID(target string) string {
	parts := strings.Split(target, "-")
	if len(parts) != 5 {
		return ""
	}
	for i, want := range []int{8, 4, 4, 4, 12} {
		if len(parts[i]) != want {
			return ""
		}
		// Hexadecimal, not merely "the right shape": a hostname that happened
		// to be dashed into those lengths would otherwise take the UUID path
		// and be reported as an unattachable interface rather than as a server
		// nobody could find.
		for _, c := range parts[i] {
			isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
			if !isHex {
				return ""
			}
		}
	}
	return target
}

// whyNotAttachable explains an absence from allowedServices.
//
// Measured on a real account: 7 of 35 dedicated servers have no virtual network
// interface whatsoever, so this is the likeliest error this command produces,
// not a defensive branch. The server's own interface list separates the two
// causes with no ambiguity, which is why it is worth the extra call — it is
// only ever made on the path that has already failed.
func whyNotAttachable(vrack, server string) error {
	var interfaces []string
	endpoint := fmt.Sprintf("/v1/dedicated/server/%s/virtualNetworkInterface", url.PathEscape(server))

	if err := httpLib.Client.Get(endpoint, &interfaces); err != nil {
		// A typo is likelier than a broken server, and the two read very
		// differently: "no such server" sends somebody to their shell history,
		// "its interfaces could not be read" sends them to a support ticket.
		var apiErr *ovh.APIError
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			return fmt.Errorf("no dedicated server named %s in this account", server)
		}
		return fmt.Errorf(
			"%s has no interface that can be attached to %s, and its interfaces could not be read to say why: %w",
			server, vrack, err)
	}

	if len(interfaces) == 0 {
		return fmt.Errorf(
			"%s has no virtual network interface, so it cannot join a vRack.\nThis is a property of the server, not of %s: list them with `ovhcloud baremetal vni list %s`",
			server, vrack, server)
	}

	return fmt.Errorf(
		"%s has no interface available for %s.\nA vRack interface belongs to one vRack at a time, so it is likely already attached to another one.\nFind out which with `ovhcloud baremetal vrack show %s`",
		server, vrack, server)
}

// ambiguous refuses to choose between several interfaces of one server.
//
// Picking the first would work most of the time and cut the wrong network the
// rest of the time. The names are listed because that is what makes the choice
// possible: an aggregation and a plain interface do not carry the same traffic.
func ambiguous(server string, owned []serverInterface) error {
	var b strings.Builder
	fmt.Fprintf(&b, "%s has %d interfaces; name the one to use with --interface:", server, len(owned))
	for _, itf := range owned {
		fmt.Fprintf(&b, "\n  %s  %s", itf.UUID, itf.Name)
	}
	return errors.New(b.String())
}
