// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// withTaskAPI points the shared client at httpmock and shortens the poll so the
// timeout path is reachable in a test instead of in fifty minutes.
func withTaskAPI(t *testing.T, attempts int, task string) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	origInterval, origAttempts := taskPollInterval, taskPollAttempts
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = client
	taskPollInterval, taskPollAttempts = time.Millisecond, attempts

	t.Cleanup(func() {
		httpLib.Client = origClient
		taskPollInterval, taskPollAttempts = origInterval, origAttempts
	})

	// go-ovh computes its clock delta before the first signed call.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server/srv/task/156839472",
		httpmock.NewStringResponder(200, task))
}

// Giving up on the wait is not the task failing, and it is not the CLI knowing
// the task is running now: the loop sleeps one interval before giving up, so
// the status it quotes is already old. The message must claim neither.
func TestWaitForTask_TimeoutDoesNotClaimTheTaskFailed(t *testing.T) {
	withTaskAPI(t, 2, `{"taskId": 156839472, "function": "reinstallServer", "status": "doing"}`)

	err := waitForDedicatedServerTask("srv", "156839472")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("stopped waiting"))
	td.Cmp(t, err.Error(), td.Contains("last status seen: doing"),
		"the status is given as an observation, not as the present state")
	td.Cmp(t, err.Error(), td.Contains("reinstallServer"),
		"the operation is named here too, not only on the failure paths")
	td.Cmp(t, err.Error(), td.Not(td.Contains("it is still running")),
		"the CLI stopped looking a whole interval ago and cannot assert this")
	td.Cmp(t, err.Error(), td.Contains("list-tasks srv"))
}

// The identifier must survive into the timeout message too: it arrives as a
// json.Number, which %d renders as %!d(json.Number=…).
func TestWaitForTask_TimeoutPrintsAReadableIdentifier(t *testing.T) {
	withTaskAPI(t, 1, `{"taskId": 156839472, "status": "todo"}`)

	err := waitForDedicatedServerTask("srv", "156839472")

	td.Require(t).CmpError(err)
	td.Cmp(t, err.Error(), td.Contains("156839472"))
	td.Cmp(t, err.Error(), td.Not(td.Contains("%!")))
}
