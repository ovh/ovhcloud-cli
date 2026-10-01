// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/assets"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
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

	td.CmpNoError(t, waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu", 0))
}

// An empty destination is the parked state, and it has to be distinguishable
// from "still on the old service".
func TestWaitForRoutingRecognisesParked(t *testing.T) {
	withIpAPI(t, 5, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": ""}}`)

	td.CmpNoError(t, waitForRouting("1.2.3.4/32", "", 0))
}

// Giving up is not the move failing, and the message must not claim it is:
// the task may well still be running. Measured on a real move, the IP was
// still on its old service a minute after the request was accepted.
func TestWaitForRoutingTimesOutWithoutClaimingFailure(t *testing.T) {
	withIpAPI(t, 2, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000005.ip-203-0-113.eu"}}`)

	err := waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu", 0)

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

// The half of eb56fab that had no test: parking waits for exactly the empty
// string, so a read that fails must not be counted as one. Folded together, the
// wait ended on the very answer it was looking for and called the park done —
// the worst shape of wrong, because the operator is told the traffic stopped.
func TestWaitForParkedDoesNotReadAFailureAsArrival(t *testing.T) {
	withIpAPI(t, 2, `{}`)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32",
		httpmock.NewStringResponder(500, `{"message": "gateway is having a day"}`))

	err := waitForRouting("1.2.3.4/32", "", 0)

	td.Require(t).CmpError(err, "a park is never concluded from a failed read")
	td.Cmp(t, err.Error(), td.Contains("stopped waiting"))
}

// A task that ended in one of the three terminal failures leaves the IP where it
// was, so the state says "not there yet" for as long as anyone asks. The wait ran
// its full ten minutes and then said "not routed yet, follow it with ovhcloud ip
// tasks", which reads as still in progress.
//
// The task is asked second and only about failure: the state stays the authority
// on where the IP is, because a request here can create more than one task.
func TestWaitForRoutingStopsOnATerminalTaskFailure(t *testing.T) {
	withIpAPI(t, 120, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000005.ip-203-0-113.eu"}}`)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32/task/4242",
		httpmock.NewStringResponder(200,
			`{"taskId": 4242, "function": "genericMoveFloatingIp", "status": "customerError", "comment": "destination refused the route"}`))

	err := waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu", 4242)

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("customerError"), "the status is named")
	td.Cmp(t, err.Error(), td.Contains("destination refused the route"), "and so is the reason")
	td.Cmp(t, err.Error(), td.Not(td.Contains("stopped waiting")),
		"it stopped because the task failed, not because it ran out of patience")
}

// The state remains the authority, and this is the case that says so: the IP had
// not arrived when the round began, the task then reported a terminal failure,
// and the IP had arrived by the time it was read again.
//
// That ordering is the two-tasks-for-one-request shape measured on a vRack
// attach — the task we were handed can fail while the work completes elsewhere.
// So a failure is never announced on the strength of the task: the state is read
// once more first, and it wins.
func TestWaitForRoutingTrustsTheStateOverAFailedTask(t *testing.T) {
	withIpAPI(t, 2, `{}`)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32",
		httpmock.ResponderFromMultipleResponses([]*http.Response{
			httpmock.NewStringResponse(200, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000005.ip-203-0-113.eu"}}`),
			httpmock.NewStringResponse(200, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000006.ip-203-0-113.eu"}}`),
		}))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32/task/4242",
		httpmock.NewStringResponder(200, `{"taskId": 4242, "status": "ovhError"}`))

	td.CmpNoError(t, waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu", 4242),
		"the IP arrived; the failed task does not overrule it")
}

// A task that cannot be read is not a task that failed. Turning a 500 on the
// task route into "your move failed" would be worse than the wait it replaces.
func TestWaitForRoutingDoesNotFailOnAnUnreadableTask(t *testing.T) {
	withIpAPI(t, 2, `{"ip": "1.2.3.4/32", "routedTo": {"serviceName": "ns0000005.ip-203-0-113.eu"}}`)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/1.2.3.4%2F32/task/4242",
		httpmock.NewStringResponder(500, `{"message": "nope"}`))

	err := waitForRouting("1.2.3.4/32", "ns0000006.ip-203-0-113.eu", 4242)

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("stopped waiting"), "it waited, it did not conclude")
}

// The three statuses treated as terminal failures are retyped in Go, because the
// schema lists the seven statuses without saying which are failures — that
// classification is not in the document. This holds them against the enum, so a
// rename upstream shows up here instead of quietly turning the branch off.
func TestTerminalTaskFailuresAreAllDeclaredByTheSchema(t *testing.T) {
	declared, err := openapi.GetComponentEnum(assets.IpOpenapiSchema, "ip.TaskStatusEnum")
	td.Require(t).CmpNoError(err)
	td.Cmp(t, len(declared) > 3, true, "positive control: the enum was read")

	for _, status := range terminalTaskFailures {
		td.Cmp(t, slices.Contains(declared, status), true,
			"%q is treated as a terminal failure but the schema does not declare it", status)
	}
}

// go-ovh decodes with UseNumber, so a JSON integer arrives as json.Number. A
// type switch that only handles float64 is dead code — a mistake this repository
// has already made twice, once in baremetal.go and once in the vRack wait.
func TestTaskIDOfReadsAJsonNumber(t *testing.T) {
	var task map[string]any
	decoder := json.NewDecoder(strings.NewReader(`{"taskId": 559188894}`))
	decoder.UseNumber()
	td.Require(t).CmpNoError(decoder.Decode(&task))

	td.Cmp(t, taskIDOf(task), int64(559188894))
	td.Cmp(t, taskIDOf(map[string]any{}), int64(0), "and an absent task is zero, not a panic")
}
