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

const ticketURL = "https://eu.api.ovh.com/v1/support/tickets"

// registerOpenTicket answers the read every lifecycle command makes to name the
// ticket it is about to act on.
func registerOpenTicket() {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305",
		httpmock.NewStringResponder(200, `{"ticketId":15647305,"ticketNumber":9876543,"state":"open","canBeClosed":true,"subject":"Disk replacement"}`))
}

func (ms *MockSuite) TestSupportTicketsCloseCmd(assert, require *td.T) {
	registerOpenTicket()
	httpmock.RegisterResponder(http.MethodPost, ticketURL+"/15647305/close",
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("support-tickets", "close", "15647305", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("is closed"))
	assert.Cmp(out, td.Contains("9876543"), "the ticket is named by what the operator recognises")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketURL+"/15647305/close"], 1)
}

// An already closed ticket is not an error and not a second close: the API
// would accept the call and nothing would happen.
func (ms *MockSuite) TestSupportTicketsCloseCmdOnAClosedTicket(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305",
		httpmock.NewStringResponder(200, `{"ticketId":15647305,"state":"closed","canBeClosed":false,"subject":"Disk replacement"}`))

	out, _ := cmd.Execute("support-tickets", "close", "15647305", "--yes")

	assert.Cmp(out, td.Contains("already closed"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketURL+"/15647305/close"], 0,
		"nothing is sent for a ticket that is already closed")
}

// canBeClosed is a field, not a route: the ticket already says whether support
// is still working on it.
func (ms *MockSuite) TestSupportTicketsCloseCmdWhenTheApiSaysItCannot(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305",
		httpmock.NewStringResponder(200, `{"ticketId":15647305,"state":"open","canBeClosed":false,"subject":"Disk replacement"}`))

	out, _ := cmd.Execute("support-tickets", "close", "15647305", "--yes")

	assert.Cmp(out, td.Contains("cannot be closed"))
	assert.Cmp(out, td.Contains("support-tickets reply"), "the refusal names what to do instead")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketURL+"/15647305/close"], 0)
}

func (ms *MockSuite) TestSupportTicketsCloseCmdDryRun(assert, require *td.T) {
	registerOpenTicket()

	out, err := cmd.Execute("support-tickets", "close", "15647305", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	assert.Cmp(out, td.Contains("/support/tickets/15647305/close"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketURL+"/15647305/close"], 0)
}

func (ms *MockSuite) TestSupportTicketsCloseCmdOnAnUnknownTicket(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/1",
		httpmock.NewStringResponder(404, `{"message":"This ticket does not exist"}`))

	out, _ := cmd.Execute("support-tickets", "close", "1", "--yes")

	assert.Cmp(out, td.Contains("no support ticket 1"))
}

// A reason made of spaces satisfies MarkFlagRequired and tells support nothing.
func (ms *MockSuite) TestSupportTicketsReopenCmdRefusesABlankReason(assert, require *td.T) {
	// No responder: the check must fire before the ticket is even read.
	out, _ := cmd.Execute("support-tickets", "reopen", "15647305", "--reason", "   ", "--yes")

	assert.Cmp(out, td.Contains("cannot be blank"))
}

func (ms *MockSuite) TestSupportTicketsReopenCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305",
		httpmock.NewStringResponder(200, `{"ticketId":15647305,"state":"closed","subject":"Disk replacement"}`))
	httpmock.RegisterMatcherResponder(http.MethodPost, ticketURL+"/15647305/reopen",
		tdhttpmock.JSONBody(td.JSON(`{"body":"The disk failed again"}`)),
		httpmock.NewStringResponder(200, ``))

	out, err := cmd.Execute("support-tickets", "reopen", "15647305", "--reason", "The disk failed again", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("is open again"))
}

func (ms *MockSuite) TestSupportTicketsReopenCmdOnAnOpenTicket(assert, require *td.T) {
	registerOpenTicket()

	out, _ := cmd.Execute("support-tickets", "reopen", "15647305", "--reason", "still broken", "--yes")

	assert.Cmp(out, td.Contains("already open"))
	assert.Cmp(out, td.Contains("support-tickets reply"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketURL+"/15647305/reopen"], 0)
}

func (ms *MockSuite) TestSupportTicketsCanBeScoredCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305/canBeScored",
		httpmock.NewStringResponder(200, `true`))

	out, err := cmd.Execute("support-tickets", "can-be-scored", "15647305")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("can be scored"))
}

func (ms *MockSuite) TestSupportTicketsCanBeScoredCmdWhenItCannot(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305/canBeScored",
		httpmock.NewStringResponder(200, `false`))

	out, err := cmd.Execute("support-tickets", "can-be-scored", "15647305")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("cannot be scored"))
}

// The filter belongs to the collection call and to nothing else. If the query
// were carried onto the expansion path, the request would go to
// "/support/tickets?status=open/15647305" — a URL no responder answers, so this
// test fails rather than passing on a filter that silently did nothing.
func (ms *MockSuite) TestSupportTicketsListCmdKeepsTheFilterOffTheExpansion(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"?status=open",
		httpmock.NewStringResponder(200, `[15647305]`))
	httpmock.RegisterResponder(http.MethodGet, ticketURL+"/15647305",
		httpmock.NewStringResponder(200, `{"ticketId":15647305,"serviceName":"ns1.example","state":"open","category":"assistance"}`))

	out, err := cmd.Execute("support-tickets", "list", "--status", "open")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("15647305"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+ticketURL+"?status=open"], 1)
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+ticketURL+"/15647305"], 1)
}

func (ms *MockSuite) TestSupportTicketsListCmdRefusesAnImpossibleStatus(assert, require *td.T) {
	out, _ := cmd.Execute("support-tickets", "list", "--status", "pending")

	assert.Cmp(out, td.Contains("--status does not accept"))
	assert.Cmp(out, td.Contains("open"), "the refusal names the accepted values")
}
