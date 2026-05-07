// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudKubeListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube",
		httpmock.NewStringResponder(200, `["kube-12345"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345",
		httpmock.NewStringResponder(200, `{
			"id": "kube-12345",
			"name": "test-kube",
			"region": "GRA11",
			"plan": "free",
			"version": "1.21.5",
			"status": "INSTALLING",
			"createdAt": "2021-10-12T14:23:45+00:00"
		}`).Once())

	out, err := cmd.Execute("cloud", "managed-kubernetes", "ls", "--cloud-project", "fakeProjectID")

	require.CmpNoError(err)
	assert.String(out, `
┌────────────┬───────────┬────────┬──────┬─────────┬────────────┐
│     id     │   name    │ region │ plan │ version │   status   │
├────────────┼───────────┼────────┼──────┼─────────┼────────────┤
│ kube-12345 │ test-kube │ GRA11  │ free │ 1.21.5  │ INSTALLING │
└────────────┴───────────┴────────┴──────┴─────────┴────────────┘
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}
