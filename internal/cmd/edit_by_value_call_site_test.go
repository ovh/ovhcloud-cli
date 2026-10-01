// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// vps edit used to hand VpsSpec over by value. addExplicitlySetFlags matches a
// flag to its field by address, and a copy has none of the addresses the flags
// were bound to, so on that call site --sla-monitoring=false was still dropped
// by omitempty and restored from the fetched VPS. The baremetal test covers a
// call site that already passed a pointer; this one covers a call site the PR
// had to change.
func (ms *MockSuite) TestVpsEditDisablesSlaMonitoring(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/vps/fakeVps",
		httpmock.NewStringResponder(200, `{"name": "fakeVps", "displayName": "fakeVps", "slaMonitoring": true}`),
	)
	httpmock.RegisterMatcherResponder("PUT", "https://eu.api.ovh.com/v1/vps/fakeVps",
		tdhttpmock.JSONBody(td.JSONPointer("/slaMonitoring", false)),
		httpmock.NewStringResponder(200, `null`),
	)

	out, err := cmd.Execute("vps", "edit", "fakeVps", "--sla-monitoring=false")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("updated successfully"))
}
