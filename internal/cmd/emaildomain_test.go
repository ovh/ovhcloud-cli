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

func (ms *MockSuite) TestEmailDomainList(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain",
		httpmock.NewStringResponder(200, `["example.com", "test.com"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain/example.com",
		httpmock.NewStringResponder(200, `{
			"domain": "example.com",
			"status": "ok",
			"offer": "email-domain-pack"
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/email/domain/test.com",
		httpmock.NewStringResponder(200, `{
			"domain": "test.com",
			"status": "ok",
			"offer": "email-domain-pack"
		}`).Once())

	out, err := cmd.Execute("email-domain", "list", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"domain": "example.com", "status": "ok", "offer": "email-domain-pack"},
		{"domain": "test.com", "status": "ok", "offer": "email-domain-pack"}
	]`))
}

func (ms *MockSuite) TestEmailDomainRedirectionList(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		httpmock.NewStringResponder(200, `["12345", "67890"]`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{
			"id": "12345",
			"from": "alias@example.com",
			"to": "destination@example.com",
			"localCopy": false
		}`).Once())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/67890",
		httpmock.NewStringResponder(200, `{
			"id": "67890",
			"from": "test@example.com",
			"to": "admin@example.com",
			"localCopy": true
		}`).Once())

	out, err := cmd.Execute("email-domain", "redirection", "list", "example.com", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`[
		{"id": "12345", "from": "alias@example.com", "to": "destination@example.com", "localCopy": false},
		{"id": "67890", "from": "test@example.com", "to": "admin@example.com", "localCopy": true}
	]`))
}

func (ms *MockSuite) TestEmailDomainRedirectionGet(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, `{
			"id": "12345",
			"from": "alias@example.com",
			"to": "destination@example.com",
			"localCopy": false
		}`).Once())

	out, err := cmd.Execute("email-domain", "redirection", "get", "example.com", "12345", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`{
		"id": "12345",
		"from": "alias@example.com",
		"to": "destination@example.com",
		"localCopy": false
	}`))
}

func (ms *MockSuite) TestEmailDomainRedirectionCreate(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		tdhttpmock.JSONBody(td.JSON(`{
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "99999",
			"from": "new-alias@example.com",
			"to": "admin@example.com",
			"localCopy": false
		}`),
	)

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "new-alias@example.com",
		"--to", "admin@example.com")

	require.CmpNoError(err)
	assert.String(out, `✅ Email redirection created successfully (ID: 99999)`)
}

func (ms *MockSuite) TestEmailDomainRedirectionCreateWithLocalCopy(assert, require *td.T) {
	httpmock.RegisterMatcherResponder("POST", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection",
		tdhttpmock.JSONBody(td.JSON(`{
			"from": "backup@example.com",
			"to": "external@otherdomain.com",
			"localCopy": true
		}`)),
		httpmock.NewStringResponder(200, `{
			"id": "88888",
			"from": "backup@example.com",
			"to": "external@otherdomain.com",
			"localCopy": true
		}`),
	)

	out, err := cmd.Execute("email-domain", "redirection", "create", "example.com",
		"--from", "backup@example.com",
		"--to", "external@otherdomain.com",
		"--local-copy")

	require.CmpNoError(err)
	assert.String(out, `✅ Email redirection created successfully (ID: 88888)`)
}

func (ms *MockSuite) TestEmailDomainRedirectionDelete(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/1.0/email/domain/example.com/redirection/12345",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("email-domain", "redirection", "delete", "example.com", "12345")

	require.CmpNoError(err)
	assert.String(out, `✅ Email redirection 12345 deleted successfully`)
}
