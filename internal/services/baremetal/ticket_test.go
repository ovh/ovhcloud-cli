// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"strings"
	"testing"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
)

// A field the API did not send is left out; a field it sent as false is not.
// Monitoring is the case that matters: false is the whole reason a support
// agent needs to see it.
func TestAFalseIsNotAnAbsence(t *testing.T) {
	lines := strings.Join(identityLines(map[string]any{
		"commercialRange": "ADVANCE-1",
		"monitoring":      false,
		"datacenter":      "rbx8",
	}), "\n")

	if !strings.Contains(lines, "Monitoring") || !strings.Contains(lines, "false") {
		t.Fatalf("monitoring off must be stated, got:\n%s", lines)
	}
	if strings.Contains(lines, "Rack") {
		t.Fatalf("a field the API did not send must not appear empty, got:\n%s", lines)
	}
	if !strings.Contains(lines, "ADVANCE-1") || !strings.Contains(lines, "rbx8") {
		t.Fatalf("the identity of the machine is missing, got:\n%s", lines)
	}
}

// The order is the order support reads: what names the machine before what
// describes its condition.
func TestIdentityGoesBeforeCondition(t *testing.T) {
	lines := identityLines(map[string]any{
		"commercialRange": "ADVANCE-1",
		"state":           "ok",
		"monitoring":      true,
	})
	if len(lines) != 3 {
		t.Fatalf("expected three lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "ADVANCE-1") {
		t.Fatalf("the commercial range names the machine and comes first, got %q", lines[0])
	}
}

func TestAnEmptyDetailProducesNoLines(t *testing.T) {
	if lines := identityLines(map[string]any{}); len(lines) != 0 {
		t.Fatalf("nothing read is nothing to say, got %v", lines)
	}
	// A null is what the API sends for a field it has no value for, and it is
	// not the string "<nil>".
	if lines := identityLines(map[string]any{"rack": nil, "os": ""}); len(lines) != 0 {
		t.Fatalf("a null and an empty string are both absences, got %v", lines)
	}
}

// --urgency was validated against the impact list. The two carry the same
// three values today, so nothing failed and nothing would have — which is the
// whole difficulty: the defect is invisible until the schema moves.
func TestUrgencyIsReadFromTheUrgencyField(t *testing.T) {
	fromSchema, err := openapi.GetRequestFieldEnum(
		assets.SupportOpenapiSchema, ticketCreatePath, "post", "urgency")
	if err != nil {
		t.Fatalf("the schema does not describe an urgency field: %s", err)
	}
	got, err := ticketUrgencies()
	if err != nil {
		t.Fatalf("reading the accepted urgencies: %s", err)
	}
	if strings.Join(got, ",") != strings.Join(fromSchema, ",") {
		t.Fatalf("urgency accepts %v, the schema says %v", got, fromSchema)
	}
}

// The canary for the test above, and the honest part of it: while impact and
// urgency carry identical values, that test passes with either wiring, so it
// proves nothing on its own. This one goes red the day the two lists diverge
// — which is exactly the day the wiring starts to matter — and says what to do
// about it. A test that cannot fail is not a test; a test that says when it
// starts being one is worth keeping.
func TestImpactAndUrgencyStillCarryTheSameValues(t *testing.T) {
	impacts, err := ticketImpacts()
	if err != nil {
		t.Fatalf("reading the accepted impacts: %s", err)
	}
	urgencies, err := ticketUrgencies()
	if err != nil {
		t.Fatalf("reading the accepted urgencies: %s", err)
	}
	if strings.Join(impacts, ",") != strings.Join(urgencies, ",") {
		t.Fatalf("impact %v and urgency %v have diverged: "+
			"TestUrgencyIsReadFromTheUrgencyField now discriminates, and every "+
			"place that reads one list for both fields has to be checked",
			impacts, urgencies)
	}
}
