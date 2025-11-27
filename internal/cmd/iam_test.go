// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestIAMTokenListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/me/identity/user/fakeUser/token",
		httpmock.NewStringResponder(200, `["token1","token2"]`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/me/identity/user/fakeUser/token/token1",
		httpmock.NewStringResponder(200, `{
			"name": "token1",
			"description": "First token",
			"expiresAt": "2025-01-01T00:00:00Z"
		}`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/me/identity/user/fakeUser/token/token2",
		httpmock.NewStringResponder(200, `{
			"name": "token2",
			"description": "Second token",
			"expiresAt": "2025-01-01T00:00:00Z"
		}`),
	)

	out, err := cmd.Execute("iam", "user", "token", "list", "fakeUser")
	require.CmpNoError(err)
	assert.String(out, `
┌────────┬──────────────┬──────────────────────┐
│  name  │ description  │      expiresAt       │
├────────┼──────────────┼──────────────────────┤
│ token1 │ First token  │ 2025-01-01T00:00:00Z │
│ token2 │ Second token │ 2025-01-01T00:00:00Z │
└────────┴──────────────┴──────────────────────┘
💡 Use option --json or --yaml to get the raw output with all information`[1:])
}
