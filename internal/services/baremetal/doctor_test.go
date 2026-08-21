// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"strings"
	"testing"
)

// The obvious test — bootType == "rescue" — misses half of them. On the fleet
// measured, six servers of thirty-five were booted on a rescue system: three
// carried bootType "rescue", and three carried bootType "internal" on an entry
// whose kernel is "rescue-customer". The type is the API's, and it is wrong for
// that entry; the kernel is not.
func TestARescueBootIsRecognisedByItsKernelToo(t *testing.T) {
	if !isRescueBoot("rescue", "rescue12-customer") {
		t.Fatal("the current rescue is typed rescue")
	}
	if !isRescueBoot("internal", "rescue-customer") {
		t.Fatal("the retired rescue is typed internal and would otherwise be missed")
	}
	if isRescueBoot("harddisk", "hd") {
		t.Fatal("booting from the disk is not a rescue")
	}
	if isRescueBoot("internal", "something-else") {
		t.Fatal("internal alone does not make a rescue")
	}
}

func TestBootFromDiskReportsNothing(t *testing.T) {
	if found := checkBoot("ns1", map[string]any{
		"bootType": "harddisk", "kernel": "hd", "description": "Boot to disk",
	}); len(found) != 0 {
		t.Fatalf("a server booting from its disk is healthy, got %d finding(s)", len(found))
	}
}

// A boot entry of type "power" with the poweroff kernel exists in the list every
// server is offered. A server sitting on it does not come back up.
func TestABootThatPowersOffIsCritical(t *testing.T) {
	found := checkBoot("ns1", map[string]any{
		"bootType": "power", "kernel": "poweroff", "description": "Power-off server",
	})
	if len(found) != 1 || found[0].Severity != critical {
		t.Fatalf("expected one critical finding, got %+v", found)
	}
}

// Three servers of the fleet sat on a rescue image whose removal had been
// announced fourteen months earlier.
func TestARetiredRescueImageIsReportedSeparately(t *testing.T) {
	found := checkBoot("ns1", map[string]any{
		"bootType":    "internal",
		"kernel":      "rescue-customer",
		"description": "Customer rescue system (Debian-10-based)[REMOVAL ON 2025-06-23]",
	})

	if len(found) != 2 {
		t.Fatalf("expected the rescue and its retired image, got %+v", found)
	}
	checks := found[0].Check + "," + found[1].Check
	if !strings.Contains(checks, "boot-image") {
		t.Fatalf("the retired image should be its own finding, got %s", checks)
	}
}

// `renew.automatic` comes back differently between consecutive reads of the
// same object — measured 10/10, 9/11, 8/12, 6/14 and 5/15 over twenty reads of
// five servers. Reporting "renewal is off" from readings that disagree is a
// coin toss; what is certain is that nobody can tell, and that is the finding.
func TestDisagreeingReadingsAreReportedAsSuch(t *testing.T) {
	readings := []map[string]any{
		{"renew": map[string]any{"automatic": true}, "expiration": "2026-09-01"},
		{"renew": map[string]any{"automatic": false}, "expiration": "2026-09-01"},
		{"renew": map[string]any{"automatic": true}, "expiration": "2026-09-01"},
	}

	found := checkRenewal("ns1", readings)
	if len(found) == 0 {
		t.Fatal("readings that disagree must not be silently dropped")
	}
	if !strings.Contains(found[0].Detail, "different answers") {
		t.Fatalf("the finding should say the API disagreed with itself, got %q", found[0].Detail)
	}
	if strings.Contains(found[0].Detail, "is off") {
		t.Fatalf("it must not assert a value it could not establish, got %q", found[0].Detail)
	}
}

func TestAgreedReadingsDecide(t *testing.T) {
	off := []map[string]any{
		{"renew": map[string]any{"automatic": false}},
		{"renew": map[string]any{"automatic": false}},
	}
	if value, verdict := agreedRenewal(off); verdict != renewalAgreed || value {
		t.Fatalf("readings that agree on false decide false, got value=%v verdict=%v", value, verdict)
	}

	on := []map[string]any{
		{"renew": map[string]any{"automatic": true}},
		{"renew": map[string]any{"automatic": true}},
	}
	if value, verdict := agreedRenewal(on); verdict != renewalAgreed || !value {
		t.Fatalf("readings that agree on true decide true, got value=%v verdict=%v", value, verdict)
	}

	// Three outcomes, not two: an absent field is neither an agreement nor a
	// disagreement, and calling it the latter made the finding describe an
	// instability that had not happened.
	if _, verdict := agreedRenewal([]map[string]any{{"renew": map[string]any{}}}); verdict != renewalAbsent {
		t.Fatalf("an absent field is absent, not %v", verdict)
	}

	mixed := []map[string]any{
		{"renew": map[string]any{"automatic": true}},
		{"renew": map[string]any{"automatic": false}},
	}
	if _, verdict := agreedRenewal(mixed); verdict != renewalDisagreed {
		t.Fatalf("readings that differ disagree, got %v", verdict)
	}

	if _, verdict := agreedRenewal(nil); verdict != renewalAbsent {
		t.Fatalf("no reading at all establishes nothing, got %v", verdict)
	}
}

// A server whose renewal reads as off, five times, is still reported with what
// established it: this route cannot give certainty, and the finding must not
// pretend otherwise.
func TestAnAgreedOffReadingSaysHowItWasEstablished(t *testing.T) {
	readings := make([]map[string]any, 5)
	for i := range readings {
		readings[i] = map[string]any{
			"renew": map[string]any{"automatic": false}, "expiration": "2026-09-01",
		}
	}

	found := checkRenewal("ns1", readings)
	if len(found) == 0 {
		t.Fatal("a renewal that reads as off must be reported")
	}
	if !strings.Contains(found[0].Detail, "5 times out of 5") {
		t.Fatalf("the finding should say how it was established, got %q", found[0].Detail)
	}
	if !strings.Contains(found[0].Detail, "confirm before acting") {
		t.Fatalf("the finding should carry its own uncertainty, got %q", found[0].Detail)
	}
}

// A server set to be deleted at expiry is the one renewal finding that is not
// in doubt: that field was stable across every reading.
func TestDeletionAtExpiryIsCritical(t *testing.T) {
	found := checkRenewal("ns1", []map[string]any{
		{"renew": map[string]any{"automatic": true, "deleteAtExpiration": true},
			"expiration": "2026-09-01"},
	})

	var seen bool
	for _, f := range found {
		if f.Severity == critical && strings.Contains(f.Detail, "deleted") {
			seen = true
		}
	}
	if !seen {
		t.Fatalf("expected a critical finding about the deletion, got %+v", found)
	}
}

func TestMonitoringOffIsReported(t *testing.T) {
	if found := checkMonitoring("ns1", map[string]any{"monitoring": false}); len(found) != 1 {
		t.Fatal("monitoring off means nobody is called when the server dies")
	}
	if found := checkMonitoring("ns1", map[string]any{"monitoring": true}); len(found) != 0 {
		t.Fatal("monitoring on is not a finding")
	}
	// An absent field is not a false one: it would report every server of an
	// account whose API stopped sending it.
	if found := checkMonitoring("ns1", map[string]any{}); len(found) != 0 {
		t.Fatal("an absent field must not be read as off")
	}
}

func TestFindingsAreOrderedWorstFirst(t *testing.T) {
	findings := []finding{
		{Server: "b", Severity: note, Check: "tasks"},
		{Server: "a", Severity: warning, Check: "boot"},
		{Server: "c", Severity: critical, Check: "power"},
	}
	sortFindings(findings)

	if findings[0].Severity != critical || findings[2].Severity != note {
		t.Fatalf("expected critical, warning, note — got %v, %v, %v",
			findings[0].Severity, findings[1].Severity, findings[2].Severity)
	}
}

// An unreadable date reports "no" rather than zero: zero would mean "expires
// today" and raise an alarm about a field nobody managed to parse.
func TestAnUnreadableDateIsNotToday(t *testing.T) {
	if _, ok := daysUntil(""); ok {
		t.Fatal("an empty date cannot be counted from")
	}
	if _, ok := daysUntil("next tuesday"); ok {
		t.Fatal("an unparseable date cannot be counted from")
	}
	if _, ok := daysUntil("2026-09-01"); !ok {
		t.Fatal("the format this API uses must be read")
	}
}
