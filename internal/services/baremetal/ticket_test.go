// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"strings"
	"testing"
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
