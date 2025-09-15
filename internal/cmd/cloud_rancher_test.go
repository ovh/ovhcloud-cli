// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestCloudRancherCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/publicCloud/project/fakeProjectID/rancher",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"targetSpec": {
					"name": "test-rancher",
					"plan": "OVHCLOUD_EDITION",
					"version": "2.11.3"
				}
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "rancher-12345"}`),
	)

	out, err := cmd.Execute("cloud", "rancher", "create", "--cloud-project", "fakeProjectID", "--name", "test-rancher", "--plan", "OVHCLOUD_EDITION", "--version", "2.11.3")

	require.CmpNoError(err)
	assert.String(out, `✅ Rancher test-rancher created successfully (id: rancher-12345)`)
}
