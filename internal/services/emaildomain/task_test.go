// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package emaildomain

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

const redirectionTaskURL = "https://eu.api.ovh.com/1.0/email/domain/example.com/task/redirection/42"

// withTaskAPI points the shared client at httpmock and shortens the poll, so
// the waiting paths are reachable in a test instead of in ten minutes.
func withTaskAPI(t *testing.T) {
	t.Helper()
	httpmock.Activate(t)

	origClient, origInterval := httpLib.Client, PollInterval
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = httpLib.NewAPIClient(client)
	PollInterval = time.Millisecond

	t.Cleanup(func() {
		httpLib.Client, PollInterval = origClient, origInterval
	})

	// go-ovh computes its clock delta before the first signed call.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
}

// answer is one reply of a mocked endpoint.
type answer struct {
	code int
	body string
}

// respondInSequence answers each call with the next of the given answers,
// repeating the last one once they run out.
//
// The response is built on every call rather than prepared once: a response
// body is read to exhaustion by the client, so a reused one would hand the
// second caller an empty body — which is exactly the case these tests are here
// to exercise.
func respondInSequence(url string, answers ...answer) {
	var calls int
	httpmock.RegisterResponder("GET", url, func(*http.Request) (*http.Response, error) {
		a := answers[min(calls, len(answers)-1)]
		calls++
		return httpmock.NewStringResponse(a.code, a.body), nil
	})
}

// A redirection task carries no status: it is over once it has left its queue,
// and the API then answers 404. Waiting for a status that never comes would
// hang until the timeout on every successful operation.
func TestWaitForTask_LeavingTheQueueCountsAsDone(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42", "action": "add", "type": "forward"}`},
		answer{404, `{"message": "This task does not exist"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	td.CmpNoError(t, waitForTask(ctx, "example.com", taskFamilyRedirection, "42"))
}

// Other task families do send a status. Honouring it ends the wait on the
// platform's own word instead of on a disappearance.
func TestWaitForTask_HonoursAStatusWhenTheAPISendsOne(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42", "status": "doing"}`},
		answer{200, `{"id": "42", "status": "done"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	td.CmpNoError(t, waitForTask(ctx, "example.com", taskFamilyRedirection, "42"))
}

// A task that ended badly is the operation failing. Reporting it as anything
// else would tell the user the CLI did not understand the answer, when what
// happened is that their redirection was not created.
func TestWaitForTask_ReportsATerminalFailure(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42", "function": "add", "status": "customerError"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	err := waitForTask(ctx, "example.com", taskFamilyRedirection, "42")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("customerError"))
	td.Cmp(t, err.Error(), td.Contains("42"))
	td.Cmp(t, err.Error(), td.Contains("(add)"),
		"the operation is named, so the failure says what did not happen")
}

// Nothing forces the API to send a function, and a message built around one
// that is missing must not print "task 42 () ended".
func TestWaitForTask_ReportsATerminalFailureWithoutAFunction(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42", "status": "ovhError"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	err := waitForTask(ctx, "example.com", taskFamilyRedirection, "42")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains(`task 42 ended with status "ovhError"`))
}

// Running out of time is the CLI giving up, not the task failing and not the
// task being cancelled. The status quoted was read a whole interval ago, so it
// is named as the last thing seen.
func TestWaitForTask_TimeoutDoesNotClaimTheTaskFailed(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42", "status": "doing"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	err := waitForTask(ctx, "example.com", taskFamilyRedirection, "42")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("gave up waiting"))
	td.Cmp(t, err.Error(), td.Contains("last status seen: doing"),
		"the status is given as an observation, not as the present state")
	td.Cmp(t, err.Error(), td.Contains("was not cancelled"))
	td.Cmp(t, err.Error(), td.Not(td.Contains("failed")))
}

// The budget named in the message is the one the wait was actually given, not
// the ten minutes of the default.
func TestWaitForTask_TimeoutReportsTheBudgetItWasGiven(t *testing.T) {
	withTaskAPI(t)
	respondInSequence(redirectionTaskURL,
		answer{200, `{"id": "42"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	err := waitForTask(ctx, "example.com", taskFamilyRedirection, "42")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("gave up waiting after"))
	td.Cmp(t, err.Error(), td.Not(td.Contains("10m0s")))
}

// A deletion is confirmed on the object, not on the task: a 404 on a mistyped
// task path would otherwise read exactly like a task that completed.
func TestWaitUntilGone(t *testing.T) {
	withTaskAPI(t)
	const redirectionURL = "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345"
	respondInSequence(redirectionURL,
		answer{200, `{"id": "12345"}`},
		answer{404, `{"message": "This redirection does not exist"}`},
	)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	td.CmpNoError(t, waitUntilGone(ctx, "/email/domain/example.com/redirection/12345", "redirection 12345"))
}

func TestWaitUntilGone_TimeoutSaysWhatWasSeen(t *testing.T) {
	withTaskAPI(t)
	const redirectionURL = "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345"
	respondInSequence(redirectionURL, answer{200, `{"id": "12345"}`})

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	err := waitUntilGone(ctx, "/email/domain/example.com/redirection/12345", "redirection 12345")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("gave up waiting"))
	td.Cmp(t, err.Error(), td.Contains("still there on the last check"))
}

// The identifiers of this API are strings, but the client decodes numbers with
// UseNumber: both have to come out as the digits the user will type back, with
// no "%!d(json.Number=…)" and no exponent.
func TestTaskIdentifier(t *testing.T) {
	decode := func(body string) map[string]any {
		var task map[string]any
		decoder := json.NewDecoder(strings.NewReader(body))
		decoder.UseNumber()
		td.Require(t).CmpNoError(decoder.Decode(&task))
		return task
	}

	td.Cmp(t, taskIdentifier(decode(`{"id": "42"}`)), "42")
	td.Cmp(t, taskIdentifier(decode(`{"id": 156839472}`)), "156839472")
	td.Cmp(t, taskIdentifier(decode(`{"action": "add"}`)), "",
		"a response without an identifier must not invent one")
	td.Cmp(t, taskIdentifier(nil), "",
		"an empty body is a valid answer to a DELETE")
}
