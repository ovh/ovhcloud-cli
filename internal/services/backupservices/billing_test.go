// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package backupservices

import (
	"strings"
	"testing"
)

// Pay-as-you-go storage nothing has written to yet reports no line at all, and
// that is the state of every backup service on this account. An empty cell
// would read as a column that failed to fill; "none yet" is the fact.
func TestNoConsumptionYetIsSaidAndNotLeftBlank(t *testing.T) {
	if got := usageOf([]usageEntry{}, 69737222); got != "none yet" {
		t.Fatalf("got %q", got)
	}
}

// Consumption that could not be read is not consumption of zero. Printing
// "none yet" there would answer a question the command failed to ask.
func TestUnreadableConsumptionIsNotZeroConsumption(t *testing.T) {
	if got := usageOf(nil, 69737222); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}

// A service that has consumed something says what, and how much.
func TestConsumptionIsReportedPerPlan(t *testing.T) {
	entries := []usageEntry{
		{ServiceID: 42, Elements: []struct {
			PlanCode string `json:"planCode"`
			Details  []struct {
				Quantity float64 `json:"quantity"`
				UniqueID string  `json:"uniqueId"`
			} `json:"details"`
		}{{
			PlanCode: "backup-vault-paygo",
			Details: []struct {
				Quantity float64 `json:"quantity"`
				UniqueID string  `json:"uniqueId"`
			}{{Quantity: 1.5}, {Quantity: 2.5}},
		}}},
		// Another service's usage must not leak into this line.
		{ServiceID: 99, Elements: []struct {
			PlanCode string `json:"planCode"`
			Details  []struct {
				Quantity float64 `json:"quantity"`
				UniqueID string  `json:"uniqueId"`
			} `json:"details"`
		}{{PlanCode: "somebody-else"}}},
	}

	got := usageOf(entries, 42)
	if !strings.Contains(got, "backup-vault-paygo×4") {
		t.Fatalf("the quantities of one plan are summed, got %q", got)
	}
	if strings.Contains(got, "somebody-else") {
		t.Fatalf("another service's usage leaked in: %q", got)
	}
}
