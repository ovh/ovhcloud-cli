// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// A real order, recorded on 18 August 2026. The password and the two URLs are
// the API's own shape: it hands out an access token in a query string, which is
// why `list` never prints one and `get` says what the link carries.
const (
	realOrder = `{"orderId": 900000004, "date": "2026-08-16T02:17:31+02:00", "expirationDate": "2026-08-23T02:17:31+02:00", "retractionDate": null, "password": "Ex4mpleP4ss", "metadatas": [], "priceWithTax": {"text": "5.48 EUR", "value": 5.48, "currencyCode": "EUR"}, "priceWithoutTax": {"text": "5.48 EUR", "value": 5.48, "currencyCode": "EUR"}, "tax": {"text": "0.00 EUR", "value": 0, "currencyCode": "EUR"}, "url": "https://www.ovh.com/cgi-bin/order/display-order.cgi?orderId=900000004&orderPassword=Ex4mpleP4ss", "pdfUrl": "https://www.ovh.com/cgi-bin/order/display-order.cgi?orderId=900000004&orderPassword=Ex4mpleP4ss"}`

	// The API's own answer on an order it calls "delivered": four steps, all
	// TODO, no history. Which is why `follow` prints the overall status too.
	realFollowUp = `[{"step": "VALIDATING", "status": "TODO", "history": []}, {"step": "VALIDATED", "status": "TODO", "history": []}, {"step": "DELIVERING", "status": "TODO", "history": []}, {"step": "AVAILABLE", "status": "TODO", "history": []}]`
)

var orderBodiesB map[string][]map[string]any

func recordOrderBody(key, response string) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(raw, &body)
		}
		orderBodiesB[key] = append(orderBodiesB[key], body)
		return httpmock.NewStringResponse(200, response), nil
	}
}

// registerFollowUp wires one order in the given state.
func registerFollowUp(status, order string) {
	orderBodiesB = map[string][]map[string]any{}
	base := "https://eu.api.ovh.com/v1/me"

	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/me/order\?`,
		httpmock.NewStringResponder(200, `[900000004]`))
	httpmock.RegisterResponder("GET", base+"/order/900000004",
		httpmock.NewStringResponder(200, order))
	httpmock.RegisterResponder("GET", base+"/order/900000004/status",
		httpmock.NewStringResponder(200, `"`+status+`"`))
	httpmock.RegisterResponder("GET", base+"/order/900000004/followUp",
		httpmock.NewStringResponder(200, realFollowUp))
	httpmock.RegisterResponder("GET", base+"/payment/method?default=true",
		httpmock.NewStringResponder(200, `[900000010]`))
	httpmock.RegisterResponder("POST", base+"/order/900000004/pay",
		recordOrderBody("pay", `null`))
	httpmock.RegisterResponder("POST", base+"/order/900000004/waiveRetraction",
		recordOrderBody("waive", `null`))
}

func orderCallCount(suffix string) int {
	return httpmock.GetCallCountInfo()["POST https://eu.api.ovh.com/v1/me/order/900000004/"+suffix]
}

// The state of an order is the reason to run `list`, and the API keeps it in a
// second endpoint. A list that skipped it would be a list of prices.
func (ms *MockSuite) TestOrderListShowsTheState(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "list", "--days", "7")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("900000004"))
	assert.Cmp(out, td.Contains("delivered"), "the state is fetched, not left blank")
	assert.Cmp(out, td.Contains("kept"), "and so is whether the order can still be walked back")
}

// The order form's link carries an access token. `get` hands it over because
// that is what was asked for; `list` must not print fifty of them for orders
// nobody asked about.
func (ms *MockSuite) TestOrderListDoesNotPrintTheAccessToken(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "list", "--days", "7", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("Ex4mpleP4ss")), "no order password in a list")
	assert.Cmp(out, td.Not(td.Contains("orderPassword")), "and no URL carrying one")
}

func (ms *MockSuite) TestOrderGetShowsTheLinkAndWhatItCarries(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "get", "900000004", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("orderPassword=Ex4mpleP4ss"), "the link is handed over")
	assert.Cmp(out, td.Not(td.Contains(`"password"`)),
		"but not repeated under a name a log scanner greps for")
}

// The four steps read TODO on an order the API calls delivered. Printing the
// table alone would say nothing has happened.
func (ms *MockSuite) TestOrderFollowPrintsTheOverallStatus(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "follow", "900000004")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("VALIDATING"))
	assert.Cmp(out, td.Contains("DELIVERING"))
}

// The limit must keep the NEWEST orders, not an arbitrary slice: the API
// returns identifiers in ascending order, so taking the head of that list would
// show the three oldest orders on the account under a heading that says recent.
//
// The count of what was dropped is written to stderr, beside the table rather
// than inside the JSON a caller is parsing, so it is not asserted here; it was
// checked against a real account, where a thirty-day window holding 52 orders
// and --limit 3 printed "49 older orders in this window are not shown".
func (ms *MockSuite) TestOrderListKeepsTheNewestWhenItStopsAtTheLimit(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/me/order\?`,
		httpmock.NewStringResponder(200, `[900000004, 900000003, 900000002]`))

	_, err := cmd.Execute("order", "list", "--days", "7", "--limit", "1")

	require.CmpNoError(err)
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/me/order/900000004"], 1,
		"the newest order is the one read")
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/me/order/900000002"], 0,
		"and the oldest is not")
}

func (ms *MockSuite) TestOrderListRefusesAnImpossibleWindow(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	_, err := cmd.Execute("order", "list", "--days", "0")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--days must be at least 1"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing is sent")
}

func (ms *MockSuite) TestOrderListRefusesADateItCannotParse(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	_, err := cmd.Execute("order", "list", "--from", "last tuesday")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("2026-08-01T00:00:00Z"), "the shape it wants is shown")
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// Paying spends money: an unattended run that did not say --yes must not.
func (ms *MockSuite) TestOrderPayRefusesWithoutConfirmation(assert, require *td.T) {
	registerFollowUp("notPaid", realOrder)

	_, err := cmd.Execute("order", "pay", "900000004")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(orderCallCount("pay"), 0, "no payment must reach the API")
}

// The default payment method is found rather than demanded, and it travels in
// the shape the API asks for: {"paymentMethod": {"id": N}}.
func (ms *MockSuite) TestOrderPayUsesTheDefaultPaymentMethod(assert, require *td.T) {
	registerFollowUp("notPaid", realOrder)

	out, err := cmd.Execute("order", "pay", "900000004", "--yes")

	require.CmpNoError(err)
	require.Cmp(len(orderBodiesB["pay"]), 1)
	method, ok := orderBodiesB["pay"][0]["paymentMethod"].(map[string]any)
	require.True(ok, "the payment method is an object, not a bare number")
	assert.Cmp(method["id"], float64(900000010))
	assert.Cmp(out, td.Contains("5.48"), "and the amount is reported")
}

// An order that is not awaiting payment must be told apart here rather than by
// the API: the CLI has just read the state and can say so plainly.
func (ms *MockSuite) TestOrderPayRefusesAnOrderAlreadySettled(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	_, err := cmd.Execute("order", "pay", "900000004", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("not awaiting payment"))
	assert.Cmp(orderCallCount("pay"), 0)
}

func (ms *MockSuite) TestOrderPayDryRunSendsNothing(assert, require *td.T) {
	registerFollowUp("notPaid", realOrder)

	out, err := cmd.Execute("order", "pay", "900000004", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("/v1/me/order/900000004/pay"))
	assert.Cmp(orderCallCount("pay"), 0)
}

// Waiving cannot be undone, so it asks for the order number to be typed back
// rather than a yes.
func (ms *MockSuite) TestOrderWaiveRetractionRefusesWithoutConfirmation(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	_, err := cmd.Execute("order", "waive-retraction", "900000004")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("keeps its retraction period"))
	assert.Cmp(orderCallCount("waiveRetraction"), 0)
}

func (ms *MockSuite) TestOrderWaiveRetractionProceedsWithYes(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "waive-retraction", "900000004", "--yes")

	require.CmpNoError(err)
	assert.Cmp(orderCallCount("waiveRetraction"), 1)
	assert.Cmp(out, td.Contains("given up"))
}

// Waiving twice is not an error the API needs to answer: the date is already in
// the order the CLI just read.
func (ms *MockSuite) TestOrderWaiveRetractionRefusesWhenAlreadyWaived(assert, require *td.T) {
	registerFollowUp("delivered",
		`{"orderId": 900000004, "date": "2026-08-16T02:17:31+02:00", "retractionDate": "2026-08-17T09:00:00+02:00", "priceWithTax": {"text": "5.48 EUR", "value": 5.48, "currencyCode": "EUR"}, "priceWithoutTax": {"text": "5.48 EUR", "value": 5.48, "currencyCode": "EUR"}, "url": "https://example.invalid", "pdfUrl": "https://example.invalid"}`)

	_, err := cmd.Execute("order", "waive-retraction", "900000004", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("already waived on 2026-08-17"))
	assert.Cmp(orderCallCount("waiveRetraction"), 0)
}

// An argument that is not a number must be named as such, with where to find a
// real one — not sent to the API to come back as a 404.
func (ms *MockSuite) TestOrderGetRefusesSomethingThatIsNotAnOrderNumber(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	_, err := cmd.Execute("order", "get", "yesterday")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("ovhcloud order list"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// --filter is registered on `order follow`, so it has to reach the four step
// rows. The assertion that carries this test is the absence of the excluded
// steps: asserting only the kept one would pass with no filtering at all.
func (ms *MockSuite) TestOrderFollowIsFiltered(assert, require *td.T) {
	registerFollowUp("delivered", realOrder)

	out, err := cmd.Execute("order", "follow", "900000004", "--filter", `step=="DELIVERING"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("DELIVERING"))
	assert.Cmp(out, td.Not(td.Contains("VALIDATING")), "a step the filter excludes must not be printed")
}
