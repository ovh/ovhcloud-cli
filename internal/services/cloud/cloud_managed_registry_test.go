// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	"encoding/json"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/display"
)

// A quota that will not parse is one unreadable cell, not a reason to answer
// nothing. This ran display.OutputError, which does not return —
// OutputWithFormat finishes on ExitFunc(1), which is os.Exit — once per plan
// inside the listing loop, so a single malformed imageStorage killed the whole
// `registry plans` listing with the other plans already collected and never
// shown.
//
// Both halves are asserted, because the no-op ExitFunc the suite installs made
// the old code look like it merely printed a warning: execution fell through with
// imageStorage still zero and wrote "0" — a wrong quota wearing the shape of a
// real one.
func TestAPlanWithAnUnreadableQuotaDoesNotKillTheListing(t *testing.T) {
	saved := display.ExitFunc
	t.Cleanup(func() { display.ExitFunc = saved })

	exited := false
	display.ExitFunc = func(int) { exited = true }

	plan := map[string]any{
		"registryLimits": map[string]any{
			"imageStorage":    json.Number("not-a-number"),
			"parallelRequest": json.Number("42"),
		},
	}

	formatContainerRegistryPlans(plan)

	td.Cmp(t, exited, false, "one bad plan must not stop the command")
	td.Cmp(t, plan["imageStorage"], "not-a-number",
		"the cell says what the API said, rather than the 0 the old fall-through wrote")
	td.Cmp(t, plan["parallelRequest"], json.Number("42"), "and the rest of the plan is still read")
}

// Positive control: a quota that does parse is still formatted.
func TestAReadableQuotaIsStillFormatted(t *testing.T) {
	plan := map[string]any{
		"registryLimits": map[string]any{"imageStorage": json.Number("10737418240")},
	}

	formatContainerRegistryPlans(plan)

	td.Cmp(t, plan["imageStorage"], "10G")
}
