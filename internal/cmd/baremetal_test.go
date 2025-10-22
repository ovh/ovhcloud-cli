// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestBaremetalListCompatibleOSCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/dedicated/server/fakeBaremetal/install/compatibleTemplates",
		httpmock.NewStringResponder(200, `{
			"ovh": [
				"alma8-cpanel-latest_64",
				"alma8-plesk18_64",
				"alma8_64",
				"alma9-cpanel-latest_64",
				"alma9-plesk18_64",
				"alma9_64",
				"byoi_64",
				"byolinux_64"
			]
		}`),
	)

	out, err := cmd.Execute("baremetal", "list-compatible-os", "fakeBaremetal")

	require.CmpNoError(err)
	assert.String(out, `
┌────────┬────────────────────────┐
│ source │          name          │
├────────┼────────────────────────┤
│ ovh    │ alma8-cpanel-latest_64 │
│ ovh    │ alma8-plesk18_64       │
│ ovh    │ alma8_64               │
│ ovh    │ alma9-cpanel-latest_64 │
│ ovh    │ alma9-plesk18_64       │
│ ovh    │ alma9_64               │
│ ovh    │ byoi_64                │
│ ovh    │ byolinux_64            │
└────────┴────────────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}
