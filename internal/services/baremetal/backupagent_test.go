// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import "testing"

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
