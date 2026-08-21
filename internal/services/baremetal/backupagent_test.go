// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/ovh/go-ovh/ovh"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// All nine agents of this account carry an empty policy while two retention
// policies exist beside them. An empty cell reads as "not shown"; the fact is
// that these agents retain nothing.
func TestAnAgentWithoutAPolicySaysSo(t *testing.T) {
	if policyOrNone("") != "none" {
		t.Fatal("no policy has to be readable as no policy")
	}
	if policyOrNone("14d_retention") != "14d_retention" {
		t.Fatal("a policy is its own name")
	}
}

// The agent names the machine it protects in its current state, which is the
// only thing joining a UUID to a server somebody can type.
func TestAnAgentNamesTheMachineItProtects(t *testing.T) {
	agent := backupAgent{CurrentState: map[string]any{"productResourceName": "ns1.example"}}
	if agent.protects() != "ns1.example" {
		t.Fatalf("got %q", agent.protects())
	}

	if (backupAgent{}).protects() != "" {
		t.Fatal("an agent that says nothing protects nothing this command can name")
	}
}

// What is true now wins over what was asked for, so an edit that has not been
// applied yet does not read as applied.
func TestAnAgentPolicyIsReadFromWhatIsTrueNow(t *testing.T) {
	agent := backupAgent{
		TargetSpec:   map[string]any{"policy": "30d_retention"},
		CurrentState: map[string]any{"policy": "14d_retention"},
	}
	if agent.policy() != "14d_retention" {
		t.Fatalf("got %q", agent.policy())
	}

	pending := backupAgent{TargetSpec: map[string]any{"policy": "30d_retention"}}
	if pending.policy() != "30d_retention" {
		t.Fatalf("got %q", pending.policy())
	}
}

// CREATING and UPDATING are transitions; the other four statuses are places an
// agent stops. NOT_INSTALLED is where a freshly created one stops, and treating
// it as a transition would make --wait time out on a success.
func TestOnlyTwoAgentStatusesAreTransitions(t *testing.T) {
	for _, transient := range []string{"CREATING", "UPDATING"} {
		if !backupAgentTransient[transient] {
			t.Fatalf("%s is a transition", transient)
		}
	}
	for _, settled := range []string{"NOT_INSTALLED", "NOT_CONFIGURED", "ENABLED", "DISABLED"} {
		if backupAgentTransient[settled] {
			t.Fatalf("%s is where an agent stops, not a transition", settled)
		}
	}
}

// withAgentAPI points the shared client at httpmock and shortens the poll, so
// the timeout path is reachable in a test rather than in five minutes.
func withAgentAPI(t *testing.T, agents string) {
	t.Helper()
	httpmock.Activate(t)

	origClient := httpLib.Client
	origInterval, origAttempts := backupAgentPollInterval, backupAgentPollAttempts
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	if err != nil {
		t.Fatalf("could not build a client: %s", err)
	}
	httpLib.Client = client
	backupAgentPollInterval, backupAgentPollAttempts = time.Millisecond, 2

	t.Cleanup(func() {
		httpLib.Client = origClient
		backupAgentPollInterval, backupAgentPollAttempts = origInterval, origAttempts
	})

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v2/backupServices/tenant/t-1/vspc/s-1/backupAgent",
		httpmock.NewStringResponder(200, agents))
}

// An empty status is not a finished one. backupAgentTransient has no entry for
// "", so an answer carrying no status at all counted as settled and was returned
// as a successfully created agent — the wait ending on the absence of an answer.
func TestWaitForAgentDoesNotTakeABlankStatusForDone(t *testing.T) {
	withAgentAPI(t, `[{"id":"a-1","status":"",
		"currentState":{"productResourceName":"ns1.example"},"currentTasks":[]}]`)

	_, err := waitForBackupAgent("t-1", "s-1", "ns1.example", true)

	if err == nil {
		t.Fatal("a blank status must be waited on, never taken for an outcome")
	}
	if !strings.Contains(err.Error(), "stopped waiting") {
		t.Fatalf("got %q", err)
	}
}

// The status enumeration has no failed value and there is no task route, so
// currentTasks is the only place a failure appears. Measured on this account:
// three BACKUP_VAULT_CREATE and one VSPC_AGENT_UPDATE sat in ERROR for five
// minutes on resources every status field called READY.
func TestWaitForAgentStopsOnAFailedTask(t *testing.T) {
	withAgentAPI(t, `[{"id":"a-1","status":"NOT_INSTALLED",
		"currentState":{"productResourceName":"ns1.example"},
		"currentTasks":[{"id":"tk-1","type":"VSPC_AGENT_CREATE","status":"ERROR","errors":[]}]}]`)

	_, err := waitForBackupAgent("t-1", "s-1", "ns1.example", true)

	if err == nil {
		t.Fatal("a failed task is not a created agent")
	}
	if !strings.Contains(err.Error(), "VSPC_AGENT_CREATE ERROR") {
		t.Fatalf("the operation has to be named, ERROR alone says nothing: %q", err)
	}
	if strings.Contains(err.Error(), "stopped waiting") {
		t.Fatalf("it stopped because the task failed, not out of patience: %q", err)
	}
}

// WAITING_USER_INPUT is not an error, and it is just as terminal for a command
// whose only move is to wait: only a person can unblock it, so holding the
// terminal open until the timeout tells nobody anything.
func TestWaitForAgentStopsWhenOnlyAPersonCanUnblockIt(t *testing.T) {
	withAgentAPI(t, `[{"id":"a-1","status":"CREATING",
		"currentState":{"productResourceName":"ns1.example"},
		"currentTasks":[{"id":"tk-1","type":"VSPC_AGENT_CREATE","status":"WAITING_USER_INPUT","errors":[]}]}]`)

	_, err := waitForBackupAgent("t-1", "s-1", "ns1.example", true)

	if err == nil || !strings.Contains(err.Error(), "WAITING_USER_INPUT") {
		t.Fatalf("got %v", err)
	}
}

// A removal that failed leaves the agent in place, so the loop would otherwise
// poll a corpse for its whole timeout and then blame its own patience.
func TestWaitForAgentRemovalStopsOnAFailedTask(t *testing.T) {
	withAgentAPI(t, `[{"id":"a-1","status":"DELETING",
		"currentState":{"productResourceName":"ns1.example"},
		"currentTasks":[{"id":"tk-1","type":"VSPC_AGENT_DELETE","status":"ERROR",
			"errors":[{"message":"the vault still holds restore points"}]}]}]`)

	_, err := waitForBackupAgent("t-1", "s-1", "ns1.example", false)

	if err == nil {
		t.Fatal("a failed removal is not a removal")
	}
	if !strings.Contains(err.Error(), "the vault still holds restore points") {
		t.Fatalf("the reason is printed when the API gives one: %q", err)
	}
}
