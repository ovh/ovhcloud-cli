// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package backupservices

import (
	"strings"
	"testing"
)

// "ERROR" answers "error doing what?" with nothing. The type is the half that
// says which operation is stuck, and this account carries three
// BACKUP_VAULT_CREATE in ERROR that the status alone could not have named.
func TestATaskSaysWhatFailedAndNotOnlyThatItDid(t *testing.T) {
	summary := taskSummary([]currentTask{{Type: "BACKUP_VAULT_CREATE", Status: "ERROR"}})
	if summary != "BACKUP_VAULT_CREATE ERROR" {
		t.Fatalf("got %q", summary)
	}

	// The schema declares errors on a task and the API leaves the list empty
	// even on a failed one. When it is there, it is shown.
	withMessage := taskSummary([]currentTask{{
		Type:   "VSPC_AGENT_UPDATE",
		Status: "ERROR",
		Errors: []struct {
			Message string `json:"message"`
		}{{Message: "agent unreachable"}},
	}})
	if !strings.Contains(withMessage, "agent unreachable") {
		t.Fatalf("got %q", withMessage)
	}

	// A task with no type still reports its status rather than an empty cell.
	if bare := taskSummary([]currentTask{{Status: "PENDING"}}); bare != "PENDING" {
		t.Fatalf("got %q", bare)
	}
}

// No task running is the normal case, and it reads as a dash rather than as a
// blank somebody has to decide the meaning of.
func TestNoTaskReadsAsADash(t *testing.T) {
	if summary := taskSummary(nil); summary != "—" {
		t.Fatalf("got %q", summary)
	}
}

// Most resources of this API answer with resourceStatus, but an agent and a
// backup server answer with status instead. Reading one of the two would print
// an empty column on exactly the resources somebody is checking on.
func TestStatusIsReadFromEitherFieldTheApiUses(t *testing.T) {
	if got := statusOf(resource{ResourceStatus: "READY"}); got != "READY" {
		t.Fatalf("got %q", got)
	}
	if got := statusOf(resource{CurrentState: map[string]any{"status": "NOT_INSTALLED"}}); got != "NOT_INSTALLED" {
		t.Fatalf("got %q", got)
	}
	if got := statusOf(resource{}); got != "" {
		t.Fatalf("nothing to report must stay empty, got %q", got)
	}
}

// A vault, a tenant and a VSPC tenant are all named in their target spec and
// echoed in their current state; a tenant on this account is named after its
// own identifier, which is not a name but is what there is.
func TestAResourceIsNamedByWhatItCarries(t *testing.T) {
	both := resource{ID: "id-1",
		TargetSpec:   map[string]any{"name": "asked"},
		CurrentState: map[string]any{"name": "actual"}}
	if both.name() != "actual" {
		t.Fatalf("what is true now wins over what was asked: %q", both.name())
	}

	asked := resource{ID: "id-1", TargetSpec: map[string]any{"name": "asked"}}
	if asked.name() != "asked" {
		t.Fatalf("got %q", asked.name())
	}

	bare := resource{ID: "id-1"}
	if bare.name() != "id-1" {
		t.Fatalf("a resource with no name is named by its identifier, got %q", bare.name())
	}
}

// Three vaults whose names are generated are told apart by where they store.
func TestVaultRegionsAreDedupedAndSorted(t *testing.T) {
	vault := resource{CurrentState: map[string]any{"buckets": []any{
		map[string]any{"region": "eu-west-sbg"},
		map[string]any{"region": "ca-east-tor"},
		map[string]any{"region": "eu-west-sbg"},
		map[string]any{},
		"not an object",
	}}}

	regions := bucketRegions(vault)
	if strings.Join(regions, ",") != "ca-east-tor,eu-west-sbg" {
		t.Fatalf("got %v", regions)
	}
}

// A missing list is not a list of zero things somebody should read as a fact.
func TestCountOfAnAbsentListIsZeroAndNotAGuess(t *testing.T) {
	if countOf(nil) != 0 {
		t.Fatal("absent counts as none")
	}
	if countOf([]any{1, 2, 3}) != 3 {
		t.Fatal("three is three")
	}
	if countOf("not a list") != 0 {
		t.Fatal("something that is not a list counts nothing")
	}
}
