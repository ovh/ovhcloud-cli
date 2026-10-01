// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const ticketCreateURL = "https://eu.api.ovh.com/v1/support/tickets/create"

// captureTicket answers the creation and keeps what was sent, because what this
// command is judged on is the payload, not the call.
func captureTicket(sent *map[string]any) {
	httpmock.RegisterResponder(http.MethodPost, ticketCreateURL,
		func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(sent); err != nil {
				return nil, err
			}
			return httpmock.NewStringResponse(200, `{"ticketId":15647305,"ticketNumber":9876543}`), nil
		})
}

func (ms *MockSuite) TestBaremetalTicketCarriesTheMachineWithIt(assert, require *td.T) {
	registerHealthyServer()
	var sent map[string]any
	captureTicket(&sent)

	out, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Disk is making noise", "--body", "The second disk clicks.", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("15647305"))
	assert.Cmp(sent["serviceName"], "ns1.example", "the server names itself, the operator does not retype it")
	assert.Cmp(sent["product"], "dedicated")
	assert.Cmp(sent["subject"], "Disk is making noise")

	body, _ := sent["body"].(string)
	assert.Cmp(body, td.Contains("The second disk clicks."), "the operator's words come first")
	assert.Cmp(body, td.Contains("Collected automatically"), "and the machine state is marked as collected")
	assert.Cmp(body, td.Contains("ns1.example"))
}

// A server booted on a rescue system is exactly what a support agent needs to
// read before answering, and the operator will not think to mention it.
func (ms *MockSuite) TestBaremetalTicketReportsWhatDoctorFinds(assert, require *td.T) {
	registerHealthyServer()
	httpmock.RegisterResponder(http.MethodGet, doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":false,"bootId":46371}`))
	httpmock.RegisterResponder(http.MethodGet, doctorServer+"/boot/46371",
		httpmock.NewStringResponder(200, `{"bootId":46371,"bootType":"internal","kernel":"rescue-customer","description":"Customer rescue system"}`))
	var sent map[string]any
	captureTicket(&sent)

	_, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Not reachable", "--body", "No SSH since this morning.", "--yes")

	require.CmpNoError(err)
	body, _ := sent["body"].(string)

	// Asserted on the findings block itself, not on substrings the identity
	// block supplies anyway. "rescue" appears in the Active boot line and
	// "Monitoring" is an identity field label, so the two assertions this
	// replaces passed with the doctor section empty — the one thing this feature
	// exists to add was the one thing untested.
	assert.Cmp(body, td.Contains("[warning] boot — booted on the rescue system"),
		"the boot finding, with its severity, is in the ticket")
	assert.Cmp(body, td.Contains("[warning] monitoring — monitoring is off"),
		"and so is the monitoring finding")
	assert.Cmp(body, td.Not(td.Contains("  nothing")), "the block is not empty")

	// The heading must not claim to be doctor's full verdict: the renewal check
	// is not run here.
	assert.Cmp(body, td.Contains("except its renewal check"))
}

// "nothing" in the findings block is read as a clean bill of health, so it has
// to mean "nothing found" and not "nothing looked". A sub-read that failed is
// named instead of being silently dropped — the same defect doctor itself was
// fixed for.
func (ms *MockSuite) TestBaremetalTicketSaysWhichChecksCouldNotRun(assert, require *td.T) {
	registerHealthyServer()
	httpmock.RegisterResponder(http.MethodGet, doctorServer,
		httpmock.NewStringResponder(200, `{"state":"ok","powerState":"poweron","monitoring":true,"bootId":1}`))
	httpmock.RegisterResponder(http.MethodGet, doctorServer+"/boot/1",
		httpmock.NewStringResponder(500, `{"message":"internal server error"}`))
	httpmock.RegisterResponder(http.MethodGet, doctorServer+"/task?status=doing",
		httpmock.NewStringResponder(500, `{"message":"internal server error"}`))
	var sent map[string]any
	captureTicket(&sent)

	_, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Disk noise", "--body", "It clicks.", "--yes")

	require.CmpNoError(err)
	body, _ := sent["body"].(string)
	assert.Cmp(body, td.Contains("not checked, the API did not answer"))
	assert.Cmp(body, td.Contains("boot"))
	assert.Cmp(body, td.Contains("tasks"))
	assert.Cmp(body, td.Not(td.Re(`(?m)^  nothing$`)),
		"a bare \"nothing\" would read as a clean bill of health over two checks that never ran")
}

// A machine the API will not answer for is the very reason somebody opens a
// ticket. The failure belongs in the ticket, not in the way of it.
func (ms *MockSuite) TestBaremetalTicketSurvivesAServerItCannotRead(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, doctorServer,
		httpmock.NewStringResponder(500, `{"message":"internal server error"}`))
	var sent map[string]any
	captureTicket(&sent)

	out, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Server unreachable", "--body", "Nothing answers.", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("15647305"), "the ticket is still created")
	body, _ := sent["body"].(string)
	assert.Cmp(body, td.Contains("could not read this server"), "and it says the CLI could not read the server")
}

func (ms *MockSuite) TestBaremetalTicketNoContextSendsTheDescriptionAlone(assert, require *td.T) {
	registerHealthyServer()
	var sent map[string]any
	captureTicket(&sent)

	_, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Billing question", "--body", "Why two invoices?", "--no-context", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["body"], "Why two invoices?")
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+doctorServer], 0,
		"--no-context reads nothing about the machine")
}

func (ms *MockSuite) TestBaremetalTicketDryRunShowsTheWholeMessage(assert, require *td.T) {
	registerHealthyServer()

	out, err := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Disk is making noise", "--body", "The second disk clicks.", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("Dry run"))
	assert.Cmp(out, td.Contains("The second disk clicks."), "a preview that hides the message previews nothing")
	assert.Cmp(out, td.Contains("Collected automatically"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketCreateURL], 0)
}

// A blank body passes MarkFlagRequired and leaves support with a subject and a
// machine dump — the one thing the collected block cannot supply is what is
// wrong from the operator's side.
func (ms *MockSuite) TestBaremetalTicketRefusesABlankDescription(assert, require *td.T) {
	out, _ := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Disk is making noise", "--body", "   ", "--yes")

	assert.Cmp(out, td.Contains("--body cannot be blank"))
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketCreateURL], 0)
}

func (ms *MockSuite) TestBaremetalTicketRefusesAProductTheApiWouldReject(assert, require *td.T) {
	out, _ := cmd.Execute("baremetal", "ticket", "ns1.example",
		"--subject", "Disk", "--body", "It clicks.", "--product", "toaster", "--yes")

	assert.Cmp(out, td.Contains("--product does not accept"))
	assert.Cmp(out, td.Contains("dedicated"), "the refusal names the accepted values")
	assert.Cmp(httpmock.GetCallCountInfo()["POST "+ticketCreateURL], 0)
}
