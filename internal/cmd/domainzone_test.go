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
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/1",
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

func (ms *MockSuite) TestDomainZoneListRecords(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/domain/zone/example.com/record",
		httpmock.NewStringResponder(200, `[1, 2]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/1",
		httpmock.NewStringResponder(200, `{
				"id": 1,
				"fieldType": "A",
				"subDomain": "www",
				"target": "127.0.0.1",
				"ttl": 3600,
				"zone": "example.com"
			}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/2",
		httpmock.NewStringResponder(200, `{
				"id": 2,
				"fieldType": "MX",
				"subDomain": "",
				"target": "mail.example.com",
				"ttl": 3600,
				"zone": "example.com"
			}`).Once())

	out, err := cmd.Execute("domain-zone", "record", "list", "example.com", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[{"id": 1, "fieldType": "A", "subDomain": "www", "target": "127.0.0.1", "ttl": 3600, "zone": "example.com"}, {"id": 2, "fieldType": "MX", "subDomain": "", "target": "mail.example.com", "ttl": 3600, "zone": "example.com"}]`))
}

func (ms *MockSuite) TestDomainZoneRefresh(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/domain/zone/example.com/refresh",
		httpmock.NewStringResponder(200, ``).Once())

	out, err := cmd.Execute("domain-zone", "refresh", "example.com")

	require.CmpNoError(err)
	assert.String(out, `✅ Zone example.com refreshed!`)
}

func (ms *MockSuite) TestDomainZoneCreateRecord(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST", "https://eu.api.ovh.com/v1/domain/zone/example.com/record",
		tdhttpmock.JSONBody(td.JSON(`{
				"fieldType": "A",
				"subDomain": "example-created",
				"target":    "127.0.0.1",
				"ttl":       0
			}`),
		),
		httpmock.NewStringResponder(200, `{
			"id": 1,
			"fieldType": "A",
			"subDomain": "example-created",
			"target":    "127.0.0.1",
			"ttl":       0
		}`),
	)

	out, err := cmd.Execute("domain-zone", "record", "create", "example.com", "--field-type", "A", "--sub-domain", "example-created", "--target", "127.0.0.1", "--ttl", "0")

	require.CmpNoError(err)
	assert.String(out, `✅ record 1 created in example.com, don't forget to refresh the associated zone!`)
}

func (ms *MockSuite) TestDomainZoneUpdateRecord(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/1",
		httpmock.NewStringResponder(200, `{
				"fieldType": "A",
				"id": 1,
				"subDomain": "example",
				"target": "127.0.0.1",
				"ttl": 60,
				"zone": "example.com"
			}`).Once())

	httpmock.RegisterMatcherResponder("PUT", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/1",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"subDomain": "example-updated",
				"target":    "127.0.0.2",
				"ttl":       0,
			}`),
		),
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("domain-zone", "record", "update", "example.com", "1", "--sub-domain", "example-updated", "--target", "127.0.0.2", "--ttl", "0")

	require.CmpNoError(err)
	assert.String(out, `✅ record 1 in example.com updated, don't forget to refresh the associated zone!`)
}

func (ms *MockSuite) TestDomainZoneDeleteRecord(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v1/domain/zone/example.com/record/1",
		httpmock.NewStringResponder(200, ``),
	)

	out, err := cmd.Execute("domain-zone", "record", "delete", "example.com", "1")
	require.CmpNoError(err)
	assert.String(out, `✅ record 1 deleted successfully from example.com`)
}
