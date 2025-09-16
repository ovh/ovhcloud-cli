// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudReferenceRancherVersionsListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/rancher/version",
		httpmock.NewStringResponder(200, `[
			{
				"cause": "END_OF_SUPPORT",
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.9.4",
				"message": "This Rancher version is no more supported, creations and updates to this version have been disabled.",
				"name": "2.9.4",
				"status": "UNAVAILABLE"
			},
			{
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.10.4",
				"name": "2.10.4",
				"status": "AVAILABLE"
			},
			{
				"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.11.3",
				"name": "2.11.3",
				"status": "AVAILABLE"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-versions", "--json", "--cloud-project", "fakeProjectID", "--filter", `status=="AVAILABLE"`)

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{
			"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.10.4",
			"name": "2.10.4",
			"status": "AVAILABLE"
		},
		{
			"changelogUrl": "https://github.com/rancher/rancher/releases/tag/v2.11.3",
			"name": "2.11.3",
			"status": "AVAILABLE"
		}
	]`))
}

func (ms *MockSuite) TestCloudReferenceRancherPlansListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/reference/rancher/plan",
		httpmock.NewStringResponder(200, `[
			{
				"name": "OVHCLOUD_EDITION",
				"status": "AVAILABLE"
			},
			{
				"name": "STANDARD",
				"status": "AVAILABLE"
			}
		]`).Once())

	out, err := cmd.Execute("cloud", "reference", "rancher", "list-plans", "--cloud-project", "fakeProjectID", "--format", "name")

	require.CmpNoError(err)
	assert.String(out, `"OVHCLOUD_EDITION"
"STANDARD"
`)
}
