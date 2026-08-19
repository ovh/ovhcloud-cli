// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// The confirmation is the last moment a wrong target can be noticed, and
// ns3141022.ip-51-77-67.eu tells nobody which machine that is. 23 of the 35
// servers measured carry a name their owner chose; that is the one that has to
// be in the sentence.
func TestDetachWarningNamesTheServerItsOwnerWouldRecognise(t *testing.T) {
	assert := td.Assert(t)

	itf := serverInterface{
		UUID:        "f644bc65-4a61-45f4-9bed-90a0059d35e4",
		Name:        "d0:50:99:d7:55:0b",
		Server:      "ns3141022.ip-51-77-67.eu",
		DisplayName: "Yaniv - RISE-1 - LIM",
	}

	warning := detachWarning("pn-1066983", itf)

	assert.Contains(warning, "Yaniv - RISE-1 - LIM")
	assert.Contains(warning, "ns3141022.ip-51-77-67.eu",
		"and the hostname, because display names are neither unique nor freshly read")
	assert.Contains(warning, "d0:50:99:d7:55:0b", "and the interface, since a server can hold several")
	assert.Contains(warning, "pn-1066983")
}

// Twelve of the 35 servers have no chosen name. Printing the hostname twice
// would fill the sentence without adding anything to it.
func TestLabelFallsBackToTheHostnameWithoutRepeatingIt(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(serverInterface{Server: "ns3070493.ip-57-129-37.eu"}.label(),
		"ns3070493.ip-57-129-37.eu")

	// The API answers the hostname as the display name when nothing was set.
	same := serverInterface{
		Server:      "ns3070493.ip-57-129-37.eu",
		DisplayName: "ns3070493.ip-57-129-37.eu",
	}
	assert.Cmp(same.label(), "ns3070493.ip-57-129-37.eu")
}

// A script holding a UUID from a previous -o json call must be able to pass it
// straight back; an operator types a hostname. Both arrive as one string.
func TestExplicitUUIDTellsAUUIDFromAHostname(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(explicitUUID("f644bc65-4a61-45f4-9bed-90a0059d35e4"),
		"f644bc65-4a61-45f4-9bed-90a0059d35e4")
	assert.Cmp(explicitUUID("ns3141022.ip-51-77-67.eu"), "")
	assert.Cmp(explicitUUID("ns31695342.ip-162-19-43.eu"), "")
	assert.Cmp(explicitUUID("vrack_aggregation"), "")
}

// The task identifier has to survive the decoder the client actually uses.
//
// go-ovh decodes responses with UseNumber, so the id arrives as a json.Number
// and not as a float64 — a distinction a type switch does not forgive. This
// repository has already paid for it once: baremetal.go carries a comment about
// %d rendering "%!d(json.Number=156839472)" in the one message written to
// explain a failure. Here the cost would be quieter: every real task would fall
// through, the cancelled-task check would never fire, and a cancelled attach
// would poll for the full ten minutes before reporting that it may still be
// running.
func TestTaskIDSurvivesTheDecoderTheClientUses(t *testing.T) {
	assert := td.Assert(t)

	// Decoded the way go-ovh decodes it, rather than typed by hand.
	var decoded map[string]any
	dec := json.NewDecoder(strings.NewReader(`{"id": 559188894, "status": "todo"}`))
	dec.UseNumber()
	assert.CmpNoError(dec.Decode(&decoded))

	id, ok := taskID(decoded)
	assert.True(ok, "the id of a real response is usable")
	assert.Cmp(id, "559188894")

	_, ok = taskID(map[string]any{})
	assert.False(ok, "a body without an id yields no task to follow")
}

// A count of one printed as "1 Public Cloud projects" reads as a bug in the
// listing, which is exactly what somebody is trying to trust here.
func TestSummaryCountsInTheSingular(t *testing.T) {
	assert := td.Assert(t)

	one := []map[string]any{{
		"label": "Public Cloud projects", "singular": "Public Cloud project", "count": 1,
	}}
	assert.Cmp(summarise(1, one, 0), "1 dedicated server · 1 Public Cloud project")

	two := []map[string]any{{
		"label": "IP blocks", "singular": "IP block", "count": 4,
	}}
	assert.Cmp(summarise(2, two, 0), "2 dedicated servers · 4 IP blocks")

	assert.Cmp(summarise(0, nil, 0), "This vRack is empty.")
}

// "Empty" is a claim, and ten serial GETs give a rate limit ten chances to make
// it a false one. A vRack holding thirty cloud projects printing as empty is
// worse than a vRack that admits it could not be read.
func TestSummaryDoesNotCallAVrackEmptyWhenItCouldNotBeRead(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(summarise(0, nil, 10),
		"Nothing could be read: 10 of the content types failed to list.")

	partial := []map[string]any{{
		"label": "IP blocks", "singular": "IP block", "count": 2,
	}}
	assert.Cmp(summarise(0, partial, 3),
		"2 IP blocks · 3 content types could not be read")
}

// A name the CLI prints must be a name the CLI accepts. This command shows
// "Yaniv - RISE-1 - LIM" in every prompt and every table; an operator copying
// it back must not be told the machine has no interface.
func TestInterfacesOfAcceptsTheNameItPrints(t *testing.T) {
	assert := td.Assert(t)

	fleet := []serverInterface{
		{UUID: "aaa", Name: "d0:50:99:d7:55:0b",
			Server: "ns3141022.ip-51-77-67.eu", DisplayName: "Yaniv - RISE-1 - LIM"},
		{UUID: "bbb", Name: "9c:6b:00:22:03:32", Server: "ns3070493.ip-57-129-37.eu"},
	}

	assert.Cmp(len(interfacesOf(fleet, "ns3141022.ip-51-77-67.eu")), 1, "by hostname")
	assert.Cmp(len(interfacesOf(fleet, "Yaniv - RISE-1 - LIM")), 1, "and by the name its owner gave it")
	assert.Cmp(len(interfacesOf(fleet, "yaniv - rise-1 - lim")), 1, "typed by a human, so case-insensitively")
	assert.Cmp(len(interfacesOf(fleet, "nothing")), 0)
}

// Display names are not unique. Two machines answering to one name is the same
// problem as one machine with two interfaces, and gets the same refusal.
func TestPickRefusesADisplayNameHeldByTwoServers(t *testing.T) {
	assert := td.Assert(t)
	t.Cleanup(func() { VrackInterface = "" })

	fleet := []serverInterface{
		{UUID: "aaa", Name: "mac-a", Server: "ns1.example.net", DisplayName: "prod-db"},
		{UUID: "bbb", Name: "mac-b", Server: "ns2.example.net", DisplayName: "prod-db"},
	}

	_, err := pick(interfacesOf(fleet, "prod-db"), "prod-db",
		func() error { return errors.New("unused") })

	assert.CmpError(err)
	assert.Contains(err.Error(), "ns1.example.net")
	assert.Contains(err.Error(), "ns2.example.net")
	assert.Contains(err.Error(), "hostname", "and it says which handle does resolve")
}

// With one interface and a --interface naming another, there is nothing to
// choose between: "has 1 interfaces; name the one to use" would send somebody
// looking for an option they do not have.
func TestPickSaysTheFlagIsWrongRatherThanAskingForAChoice(t *testing.T) {
	assert := td.Assert(t)
	VrackInterface = "ffffffff-1111-2222-3333-444444444444"
	t.Cleanup(func() { VrackInterface = "" })

	owned := []serverInterface{{UUID: "aaa", Name: "d0:50:99:d7:55:0b", Server: "ns1.example.net"}}

	_, err := pick(owned, "ns1.example.net", func() error { return errors.New("unused") })

	assert.CmpError(err)
	assert.Contains(err.Error(), "only interface")
	assert.Not(err.Error(), td.Contains("1 interfaces"))
}

// A hostname that happens to be dashed into 8-4-4-4-12 must not enter the UUID
// path: it would be reported as an unattachable interface rather than as a
// server nobody could find.
func TestExplicitUUIDRequiresHexadecimal(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(explicitUUID("F644BC65-4A61-45F4-9BED-90A0059D35E4"),
		"F644BC65-4A61-45F4-9BED-90A0059D35E4", "uppercase is still a UUID")
	assert.Cmp(explicitUUID("zzzzzzzz-4a61-45f4-9bed-90a0059d35e4"), "",
		"the right shape is not enough")
}
