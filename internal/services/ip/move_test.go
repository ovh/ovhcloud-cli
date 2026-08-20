// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

func sample() []destination {
	return []destination{
		{Family: "dedicatedServer", Service: "ns0000005.ip-203-0-113.eu"},
		{Family: "dedicatedServer", Service: "ns0000006.ip-203-0-113.eu"},
		{Family: "vps", Service: "vps-c924a68c.vps.ovh.net", Nexthop: []string{"1.2.3.4"}},
	}
}

// The service name is what the operator copies out of another command's
// output, and case is not something they should have to reproduce.
func TestPickDestinationIgnoresCase(t *testing.T) {
	chosen, ok := pickDestination(sample(), "NS0000006.IP-203-0-113.EU")

	td.Require(t).Cmp(ok, true)
	td.Cmp(t, chosen.Service, "ns0000006.ip-203-0-113.eu")
	td.Cmp(t, chosen.Family, "dedicatedServer")
}

func TestPickDestinationRefusesWhatIsNotThere(t *testing.T) {
	_, ok := pickDestination(sample(), "ns9999999.example")

	td.Cmp(t, ok, false)
}

// The refusal has to be useful without being a wall of text: a real account
// offered a hundred destinations for one address, so the families are counted
// and the command that lists them is given.
func TestUnknownDestinationCountsRatherThanLists(t *testing.T) {
	err := unknownDestination("203.0.113.80/32", "ns9999999.example", sample())

	td.Require(t).CmpError(err)
	message := err.Error()

	td.Cmp(t, message, td.Contains("ns9999999.example"))
	td.Cmp(t, message, td.Contains("2 dedicatedServer"))
	td.Cmp(t, message, td.Contains("1 vps"))
	td.Cmp(t, message, td.Contains("ovhcloud ip destinations 203.0.113.80/32"),
		"the way out has to be named")
	td.Cmp(t, message, td.Not(td.Contains("vps-c924a68c")),
		"individual services are not listed: there can be a hundred of them")
}

// The sentence before a move must name the service whose traffic stops, not
// only the one it is going to. That service is the thing that goes dark, and
// it is the only part the operator cannot see from the command they typed.
func TestMoveWarningNamesTheServiceThatLosesTheIp(t *testing.T) {
	warning := moveWarning("203.0.113.80/32", "ns0000005.ip-203-0-113.eu", sample()[1], false)

	td.Cmp(t, warning, td.Contains("ns0000005.ip-203-0-113.eu"), "what it leaves")
	td.Cmp(t, warning, td.Contains("ns0000006.ip-203-0-113.eu"), "where it goes")
	td.Cmp(t, warning, td.Contains("dedicatedServer"))
}

// An IP that serves nothing is a different sentence: there is no traffic to
// stop, and saying there is would be a warning about something that is not
// happening.
func TestMoveWarningSaysSoWhenNothingIsServed(t *testing.T) {
	warning := moveWarning("203.0.113.80/32", "", sample()[1], false)

	td.Cmp(t, warning, td.Contains("not routed to any service"))
	td.Cmp(t, strings.Contains(warning, "stops the traffic"), false,
		"nothing is being interrupted here")
}

// The API answers a map, so its iteration order changes from run to run. Two
// readings of the same command have to be comparable, which they are not if a
// hundred rows arrive shuffled.
func TestDestinationsAreSortedByFamilyThenService(t *testing.T) {
	destinations := []destination{
		{Family: "vps", Service: "b"},
		{Family: "dedicatedServer", Service: "z"},
		{Family: "vps", Service: "a"},
		{Family: "dedicatedServer", Service: "a"},
	}

	sortDestinations(destinations)

	td.Cmp(t, destinations, []destination{
		{Family: "dedicatedServer", Service: "a"},
		{Family: "dedicatedServer", Service: "z"},
		{Family: "vps", Service: "a"},
		{Family: "vps", Service: "b"},
	})
}

func TestSortedKeysAreOrdered(t *testing.T) {
	td.Cmp(t, sortedKeys(map[string]int{"vps": 1, "dedicatedServer": 2, "cloudProject": 3}),
		[]string{"cloudProject", "dedicatedServer", "vps"})
}

// withIpAPI points the shared client at httpmock and shrinks the poll so the
// timeout path is reachable in a test instead of in ten minutes.
func withIpAPI(t *testing.T, attempts int, block string) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	origInterval, origAttempts := movePollInterval, movePollAttempts
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = client
	movePollInterval, movePollAttempts = time.Millisecond, attempts

	t.Cleanup(func() {
		httpLib.Client = origClient
		movePollInterval, movePollAttempts = origInterval, origAttempts
	})

	// go-ovh computes its clock delta before the first signed call.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32",
		httpmock.NewStringResponder(200, block))
}

// The wait reads where the IP is, not whether a task finished. A vRack attach
// measured earlier in this CLI produced two tasks for one request, so a wait
// that trusts the task it was handed can report success while the other half
// is still running.
func TestWaitForRoutingReturnsWhenTheIpHasActuallyMoved(t *testing.T) {
	withIpAPI(t, 5, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000006.ip-203-0-113.eu"}}`)

	td.CmpNoError(t, waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu"))
}

// An empty destination is the parked state, and it has to be distinguishable
// from "still on the old service".
func TestWaitForRoutingRecognisesParked(t *testing.T) {
	withIpAPI(t, 5, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": ""}}`)

	td.CmpNoError(t, waitForRouting("1.2.3.4/32", ""))
}

// Giving up is not the move failing, and the message must not claim it is:
// the task may well still be running. Measured on a real move, the IP was
// still on its old service a minute after the request was accepted.
func TestWaitForRoutingTimesOutWithoutClaimingFailure(t *testing.T) {
	withIpAPI(t, 2, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000005.ip-203-0-113.eu"}}`)

	err := waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("stopped waiting"))
	td.Cmp(t, err.Error(), td.Contains("ovhcloud ip tasks"), "the way to follow it is named")
	td.Cmp(t, err.Error(), td.Not(td.Contains("failed")), "the move was not observed to fail")
}

// An unread routing is not an empty routing. The prompt is the last place a
// wrong move can still be stopped, and "not routed to any service" is a green
// light: it must never be what a failed read produces.
func TestAnUnreadRoutingIsNotAnEmptyOne(t *testing.T) {
	unknown := moveWarning("203.0.113.80/32", "", sample()[1], true)
	if strings.Contains(unknown, "is not routed to any service") {
		t.Fatalf("a failed read must not read as a free IP: %q", unknown)
	}
	if !strings.Contains(unknown, "Could not read") {
		t.Fatalf("the prompt has to say the reading failed, got %q", unknown)
	}
	free := moveWarning("203.0.113.80/32", "", sample()[1], false)
	if !strings.Contains(free, "is not routed to any service") {
		t.Fatalf("an IP really routed nowhere still says so, got %q", free)
	}
	if free == unknown {
		t.Fatal("the two cases produce the same sentence, so the prompt cannot tell them apart")
	}
}
