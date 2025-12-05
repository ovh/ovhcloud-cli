// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestOauth2ClientCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/me/api/oauth2/client",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"callbackUrls": [
					"https://example.com/callback"
				],
				"description": "Test OAuth2 client",
				"flow": "AUTHORIZATION_CODE",
				"name": "test-client"
			}`),
		),
		httpmock.NewStringResponder(200, `{"clientId": "client-12345", "clientSecret": "sicrette"}`),
	)

	out, err := cmd.Execute("account", "api", "oauth2", "client", "create", "--name", "test-client",
		"--flow", "AUTHORIZATION_CODE", "--callback-urls", "https://example.com/callback",
		"--description", "Test OAuth2 client", "--json")
	require.CmpNoError(err)
	assert.Cmp(json.RawMessage(out), td.JSON(`
		{
			"message": "✅ OAuth2 client created successfully (client ID: client-12345, client secret: sicrette)",
			"details": {
				"clientId": "client-12345",
				"clientSecret": "sicrette"
			}
		}`),
	)
}
