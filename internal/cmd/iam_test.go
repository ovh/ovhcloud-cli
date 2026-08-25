// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"net/http"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

func (ms *MockSuite) TestIAMTokenListCmd(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/identity/user/fakeUser/token",
		httpmock.NewStringResponder(200, `["token1","token2"]`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/identity/user/fakeUser/token/token1",
		httpmock.NewStringResponder(200, `{
			"name": "token1",
			"description": "First token",
			"expiresAt": "2025-01-01T00:00:00Z"
		}`),
	)

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/identity/user/fakeUser/token/token2",
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
💡 Use option -o json or -o yaml to get the raw output with all information`[1:])
}

func (ms *MockSuite) TestIAMPolicyDeleteCmd(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", "https://eu.api.ovh.com/v2/iam/policy/policy-1234",
		httpmock.NewStringResponder(204, ``),
	)

	out, err := cmd.Execute("iam", "policy", "delete", "policy-1234")
	require.CmpNoError(err)
	assert.String(out, `✅ IAM policy policy-1234 deleted successfully`)
}

func (ms *MockSuite) TestIAMPolicyCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v2/iam/policy",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"identities": [
					"urn:v1:eu:identity:account:aa1-ovh"
				],
				"name": "MyPolicy",
				"permissions": {
					"allow": [
						{
							"action": "domain:apiovh:get"
						}
					]
				},
				"resources": [
					{
						"urn": "urn:v1:eu:resource:domain:*"
					}
				]
			}`,
		)),
		httpmock.NewStringResponder(200, `
		{
			"id": "policy-1234",
			"identities": [
				"urn:v1:eu:identity:account:aa1-ovh"
			],
			"name": "MyPolicy",
			"permissions": {
				"allow": [
					{
						"action": "domain:apiovh:get"
					}
				]
			},
			"resources": [
				{
					"urn": "urn:v1:eu:resource:domain:*"
				}
			]
		}`),
	)

	out, err := cmd.Execute("iam", "policy", "create",
		"--name", "MyPolicy",
		"--allow", "domain:apiovh:get",
		"--identity", "urn:v1:eu:identity:account:aa1-ovh",
		"--resource", "urn:v1:eu:resource:domain:*",
	)
	require.CmpNoError(err)
	assert.String(out, `✅ IAM policy policy-1234 created successfully`)
}

// `--tag env=` est la tentative naturelle pour retirer un tag, et c est le
// pire resultat possible : l API l accepte, stocke `{"env": ""}`, et le tag a
// l air parti alors qu une politique ecrite sur sa CLE le trouve toujours.
//
// Ce que ce test tient, c est qu il n atteint pas l API du tout : le refus doit
// tomber avant le GET que l edition fait pour fusionner.
func (ms *MockSuite) TestIAMResourceEditRefusesAnEmptyTagValue(assert, require *td.T) {
	urn := "urn:v1:eu:resource:dedicatedServer:ns1.example"
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/iam/resource/"+urn,
		httpmock.NewStringResponder(200, `{"urn":"`+urn+`","tags":{"env":"prod"}}`),
	)
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v2/iam/resource/"+urn,
		httpmock.NewStringResponder(200, `null`),
	)

	_, err := cmd.Execute("iam", "resource", "edit", urn, "--tag", "env=")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("iam resource tag remove"), "il dit par quoi remplacer")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "et rien n atteint l API")
}

// Le controle negatif : une valeur non vide passe, sinon la garde interdirait
// de poser un tag.
func (ms *MockSuite) TestIAMResourceEditAcceptsARealTagValue(assert, require *td.T) {
	urn := "urn:v1:eu:resource:dedicatedServer:ns1.example"
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/iam/resource/"+urn,
		httpmock.NewStringResponder(200, `{"urn":"`+urn+`","tags":{"env":"prod"}}`),
	)
	httpmock.RegisterMatcherResponder("PUT", "https://eu.api.ovh.com/v2/iam/resource/"+urn,
		tdhttpmock.JSONBody(td.JSONPointer("/tags/env", "staging")),
		httpmock.NewStringResponder(200, `null`),
	)

	_, err := cmd.Execute("iam", "resource", "edit", urn, "--tag", "env=staging")

	require.CmpNoError(err)
	puts := 0
	for endpoint, count := range httpmock.GetCallCountInfo() {
		if strings.HasPrefix(endpoint, "PUT ") {
			puts += count
		}
	}
	assert.Cmp(puts, 1, "l edition va bien jusqu au bout")
}
