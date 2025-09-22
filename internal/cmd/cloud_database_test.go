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

func (ms *MockSuite) TestCloudDatabaseCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/1.0/cloud/project/fakeProjectID/database/mysql",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"nodesList": [
					{
						"flavor": "db1-4",
						"region": "DE"
					}
				],
				"plan": "essential",
				"version": "8"
			}`),
		),
		httpmock.NewStringResponder(200, `{"id": "0f0c43f0-979a-11f0-94fd-0050568ce122"}`),
	)

	out, err := cmd.Execute("cloud", "database", "create", "--cloud-project", "fakeProjectID", "--engine", "mysql", "--version", "8", "--plan", "essential", "--nodes-list", "db1-4:DE")

	require.CmpNoError(err)
	assert.String(out, `✅ Database created successfully (id: 0f0c43f0-979a-11f0-94fd-0050568ce122)`)
}
