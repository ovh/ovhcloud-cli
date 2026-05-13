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

func (ms *MockSuite) TestSupportTicketsCreateCmd(assert, require *td.T) {
	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/support/tickets/create",
		tdhttpmock.JSONBody(td.JSON(`
			{
				"body":        "Detailed description",
				"category":    "assistance",
				"product":     "publiccloud",
				"serviceName": "service-id",
				"subject":     "Short summary"
			}`,
		)),
		httpmock.NewStringResponder(200, `
		{
			"ticketId": 15647305,
			"ticketNumber": 9876543,
			"messageId": 42
		}`),
	)

	out, err := cmd.Execute("support-tickets", "create",
		"--subject", "Short summary",
		"--body", "Detailed description",
		"--category", "assistance",
		"--product", "publiccloud",
		"--service-name", "service-id",
	)
	require.CmpNoError(err)
	assert.String(out, `✅ Ticket 15647305 (number 9876543) created successfully`)
}

func (ms *MockSuite) TestSupportTicketsCreateCmdMissingMandatoryField(assert, require *td.T) {
	// No httpmock responder registered: the mandatory-field check in
	// common.CreateResource must fire before any HTTP call is attempted.
	out, _ := cmd.Execute("support-tickets", "create", "--body", "Detailed description")
	assert.Cmp(out, td.Contains(`🛑`))
	assert.Cmp(out, td.Contains(`mandatory field "subject"`))
}

func (ms *MockSuite) TestSupportTicketsCreateCmdMutuallyExclusiveFlags(assert, require *td.T) {
	// --from-file and --editor are mutually exclusive; cobra rejects the
	// combination before any HTTP call is attempted.
	_, err := cmd.Execute("support-tickets", "create",
		"--from-file", "/dev/null",
		"--editor",
	)
	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains(`[from-file editor]`))
}
