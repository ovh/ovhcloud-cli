// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestDomainZoneGetRecord(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/domain/zone/example.com/record/1",
		httpmock.NewStringResponder(200, `{
				"fieldType": "A",
				"id": 1,
				"subDomain": "example",
				"target": "127.0.0.1",
				"ttl": 60,
				"zone": "example.com"
			}`).Once())

	out, err := cmd.Execute("domain-zone", "record", "get", "example.com", "1")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"fieldType": "A",
		"id": 1,
		"subDomain": "example",
		"target": "127.0.0.1",
		"ttl": 60,
		"zone": "example.com"
	}`))
}

func (ms *MockSuite) TestDomainZoneRefresh(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/1.0/domain/zone/example.com/refresh",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("domain-zone", "refresh", "example.com")

	require.CmpNoError(err)
	assert.String(out, `✅ Zone example.com refreshed!`)
}

func (ms *MockSuite) TestDomainZoneUpdateRecord(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/domain/zone/example.com/record/1",
		httpmock.NewStringResponder(200, `{
				"fieldType": "A",
				"id": 1,
				"subDomain": "example",
				"target": "127.0.0.1",
				"ttl": 60,
				"zone": "example.com"
			}`).Once())

	httpmock.RegisterMatcherResponder("PUT", "https://eu.api.ovh.com/1.0/domain/zone/example.com/record/1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"subDomain": "example-updated",
				"target":    "127.0.0.2",
				"ttl":       0,
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("domain-zone", "record", "update", "example.com", "1", "--subdomain", "example-updated", "--target", "127.0.0.2", "--ttl", "0")

	require.CmpNoError(err)
	assert.String(out, `✅ record 1 in example.com updated, don't forget to refresh the associated zone!`)
}
