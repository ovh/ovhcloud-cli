// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// serverInterface is one dedicated server interface, as both the vRack API and
// this CLI think of it: the UUID the API takes, the name a human reads, and the
// server it belongs to.
//
// The API returns this shape from two different endpoints — allowedServices for
// what may still be attached, dedicatedServerInterfaceDetails for what already
// is — which is why the type is shared rather than duplicated per caller.
type serverInterface struct {
	UUID   string `json:"dedicatedServerInterface"`
	Name   string `json:"name"`
	Server string `json:"dedicatedServer"`

	// DisplayName is the name the customer gave the server, when they gave it
	// one. It comes from IAM, not from the vRack API, and is empty when the
	// lookup was skipped or failed.
	DisplayName string `json:"displayName,omitempty"`
}

// label names the server both ways at once.
//
// The hostname is an address and the display name is how its owner knows it,
// and neither replaces the other. "Mail relay - Paris" is what says whether
// this is the right machine; ns0000002.ip-203-0-113.eu is what makes it the only
// one, because display names are not unique and are read from a cache that can
// be an hour old. A prompt that showed one of them would be exactly as wrong
// whichever one it dropped.
//
// Measured on the account this was written against: 23 of 35 servers carry a
// display name, so the parenthesis is the common case, not the exception.
func (s serverInterface) label() string {
	if s.DisplayName != "" && s.DisplayName != s.Server {
		return fmt.Sprintf("%s (%s)", s.DisplayName, s.Server)
	}
	return s.Server
}

// attachedInterfaces lists what is currently in the vRack, resolved.
//
// The plain dedicatedServerInterface endpoint answers with bare UUIDs, which
// name nothing to anybody. The Details variant answers with the same set plus
// the server and interface names, so it is the one used here.
func attachedInterfaces(vrack string) ([]serverInterface, error) {
	var interfaces []serverInterface
	endpoint := fmt.Sprintf("/v1/vrack/%s/dedicatedServerInterfaceDetails", url.PathEscape(vrack))

	if err := httpLib.Client.Get(endpoint, &interfaces); err != nil {
		return nil, fmt.Errorf("failed to list the interfaces attached to %s: %w", vrack, err)
	}

	return withDisplayNames(interfaces), nil
}

// attachableInterfaces lists what may still be attached to the vRack.
//
// It reads allowedServices and not eligibleServices. Both claim to answer this
// question; only one of them answers it now. eligibleServices is an
// asynchronous job whose stored result was a day old when this was measured,
// and reported no attachable interface at all while allowedServices listed
// eleven — with the owning server's name, which is what lets this CLI accept a
// server rather than a UUID.
func attachableInterfaces(vrack string) ([]serverInterface, error) {
	var allowed struct {
		DedicatedServerInterface []serverInterface `json:"dedicatedServerInterface"`
	}
	endpoint := fmt.Sprintf("/v1/vrack/%s/allowedServices", url.PathEscape(vrack))

	if err := httpLib.Client.Get(endpoint, &allowed); err != nil {
		return nil, fmt.Errorf("failed to list what can be attached to %s: %w", vrack, err)
	}

	return withDisplayNames(allowed.DedicatedServerInterface), nil
}

// withDisplayNames fills in the customer-chosen server names, and gives up
// quietly when it cannot.
//
// A display name makes the output readable; it is never what the command acts
// on. So a failed or empty lookup leaves the interfaces exactly as the vRack
// API returned them, and the caller still prints its table. Refusing to show a
// network because a cosmetic lookup failed would be the wrong trade.
func withDisplayNames(interfaces []serverInterface) []serverInterface {
	if len(interfaces) == 0 {
		return interfaces
	}

	names := serverDisplayNames()
	if len(names) == 0 {
		return interfaces
	}

	for i, itf := range interfaces {
		interfaces[i].DisplayName = names[itf.Server]
	}
	return interfaces
}

// serverDisplayNames maps every dedicated server of the account to the name its
// owner gave it, in one call.
//
// Two things about this endpoint are not obvious and were both measured. The
// resourceType filter is mandatory in practice: without it the answer is capped
// at a hundred resources and, on the account this was written against, held no
// dedicated server at all — only contacts. And the display name lives here
// rather than on the server itself: /dedicated/server/{n} has a `name` field,
// but it repeats the hostname on all 35 servers and is not what the manager
// writes when somebody renames a machine.
//
// Nothing is cached, deliberately. These names end up in the sentence an
// operator reads before cutting a machine off its network, and a cache would
// buy one saved request at the price of two ways to name the wrong machine: a
// rename made minutes ago would not show, and a second account used through
// --profile on the same laptop would read the first one's names. One GET is
// cheaper than either.
//
// It returns an empty map on any failure. Callers treat that as "no names
// available", never as an error.
func serverDisplayNames() map[string]string {
	var resources []struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}

	if err := httpLib.Client.Get("/v2/iam/resource?resourceType=dedicatedServer", &resources); err != nil {
		return nil
	}

	names := make(map[string]string, len(resources))
	for _, resource := range resources {
		if resource.DisplayName != "" && resource.DisplayName != resource.Name {
			names[resource.Name] = resource.DisplayName
		}
	}

	return names
}

// interfacesOf returns the entries belonging to one server, in a stable order.
//
// It matches the display name as well as the hostname, because this CLI prints
// the display name everywhere and an operator who copies what they were just
// shown should not be told their machine has no interface. Matching is
// case-insensitive for the same reason: these names are typed by people.
//
// Display names are not unique, so a name held by two servers matches both and
// the caller refuses rather than picking — the hostname remains the handle that
// always resolves to one machine.
func interfacesOf(interfaces []serverInterface, target string) []serverInterface {
	var owned []serverInterface
	for _, itf := range interfaces {
		if strings.EqualFold(itf.Server, target) ||
			(itf.DisplayName != "" && strings.EqualFold(itf.DisplayName, target)) {
			owned = append(owned, itf)
		}
	}

	sort.Slice(owned, func(i, j int) bool { return owned[i].Name < owned[j].Name })
	return owned
}

// distinctServers counts the machines a set of interfaces spans.
func distinctServers(interfaces []serverInterface) int {
	seen := map[string]bool{}
	for _, itf := range interfaces {
		seen[itf.Server] = true
	}
	return len(seen)
}

// otherContents lists everything else the vRack holds, by type.
//
// This command only attaches and detaches dedicated servers, and it lists all
// eleven attachable types anyway. Showing two of them because those are the two
// it can act on would make `vrack get` answer "what is in this vRack" with a
// filtered truth — and the account this was written against has 32 vRacks whose
// only content is a cloud project, every one of which would have looked empty.
//
// The identifiers are shown raw. Resolving them would mean a lookup per type,
// each with its own shape and its own failure; naming them is enough to say
// what is there.
func otherContents(vrack string) ([]map[string]any, int) {
	unreadable := 0
	// Ordered rather than ranged over a map: a listing that reshuffles itself
	// between two runs cannot be diffed, and these are read side by side.
	types := []struct {
		path  string
		label string
		one   string
	}{
		{"cloudProject", "Public Cloud projects", "Public Cloud project"},
		{"ip", "IP blocks", "IP block"},
		{"ipv6", "IPv6 blocks", "IPv6 block"},
		{"ipLoadbalancing", "Load balancers", "load balancer"},
		{"dedicatedCloud", "Hosted Private Cloud", "Hosted Private Cloud"},
		{"dedicatedConnect", "Dedicated Connect", "Dedicated Connect"},
		{"ovhCloudConnect", "OVHcloud Connect", "OVHcloud Connect"},
		{"vrackServices", "vRack Services", "vRack Service"},
		{"vmwareCloudDirectorVirtualDataCenter", "VMware Cloud Director", "VMware Cloud Director"},
		{"legacyVrack", "Legacy vRacks", "legacy vRack"},
	}

	var sections []map[string]any
	for _, t := range types {
		var ids []string
		endpoint := fmt.Sprintf("/v1/vrack/%s/%s", url.PathEscape(vrack), t.path)

		// A type this account cannot hold answers with an error rather than an
		// empty list, so an error is usually just a section with nothing to
		// show. Usually — a rate limit, a 5xx or a dropped connection produce
		// the same error, and ten of these run in a row. They are counted so
		// the summary can stop short of claiming the vRack is empty when what
		// happened is that nobody could read it.
		if err := httpLib.Client.Get(endpoint, &ids); err != nil {
			unreadable++
			continue
		}
		if len(ids) == 0 {
			continue
		}

		sections = append(sections, map[string]any{
			"label":    t.label,
			"singular": t.one,
			"ids":      ids,
			"count":    len(ids),
		})
	}

	return sections, unreadable
}

// summarise writes the one line that says what is in the vRack.
//
// It is the first thing printed because it is the thing being asked. Reading a
// count off a table means counting rows, and the answer to "is anything in
// here" should not require that.
func summarise(servers int, others []map[string]any, unreadable int) string {
	parts := make([]string, 0, len(others)+1)

	if servers > 0 {
		parts = append(parts, plural(servers, "dedicated server", "dedicated servers"))
	}
	for _, section := range others {
		count := section["count"].(int)
		parts = append(parts, plural(count,
			section["singular"].(string), section["label"].(string)))
	}

	if len(parts) == 0 {
		if unreadable > 0 {
			// "Empty" is a claim. When nothing could be read, the honest answer
			// is that nothing is known — a vRack holding thirty cloud projects
			// must never print as empty because ten GETs happened to fail.
			return fmt.Sprintf("Nothing could be read: %d of the content types failed to list.", unreadable)
		}
		return "This vRack is empty."
	}

	summary := strings.Join(parts, " · ")
	if unreadable > 0 {
		summary += fmt.Sprintf(" · %d content types could not be read", unreadable)
	}
	return summary
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
