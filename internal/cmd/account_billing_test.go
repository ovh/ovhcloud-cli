// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const (
	billsURL    = "https://eu.api.ovh.com/v1/me/bill"
	refundsURL  = "https://eu.api.ovh.com/v1/me/refund"
	usageURL    = "https://eu.api.ovh.com/v1/me/consumption/usage/current"
	forecastURL = "https://eu.api.ovh.com/v1/me/consumption/usage/forecast"
	servicesURL = "https://eu.api.ovh.com/v1/services"
)

// captureQuery answers a collection and records the query it was asked with.
func captureQuery(url, body string, seen *string) {
	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			*seen = req.URL.RawQuery
			return httpmock.NewStringResponse(200, body), nil
		})
}

func registerOneBill(id string) {
	httpmock.RegisterResponder(http.MethodGet, billsURL+"/"+id,
		httpmock.NewStringResponder(200, fmt.Sprintf(`{"billId":%q,"date":"2026-08-01T08:19:43+02:00",
			"category":"autorenew","orderId":900000006,
			"priceWithTax":{"text":"104.99 €","value":104.99,"currencyCode":"EUR"},
			"priceWithoutTax":{"text":"104.99 €","value":104.99,"currencyCode":"EUR"},
			"tax":{"text":"0.00 €","value":0,"currencyCode":"EUR"},
			"password":"B1llP4ssEx",
			"url":"https://www.ovh.com/cgi-bin/order/facture.pdf?esign=SIGNATURE&reference=x&timestamp=1",
			"pdfUrl":"https://www.ovh.com/cgi-bin/order/facture.pdf?esign=SIGNATURE&reference=x&timestamp=1"}`, id)))
}

// Without --from the window is the current month: the account this was built
// against holds 2215 invoices, and each one listed is one request to detail it.
func (ms *MockSuite) TestBillListDefaultsToTheCurrentMonth(assert, require *td.T) {
	var query string
	captureQuery(billsURL, `[]`, &query)

	_, err := cmd.Execute("account", "bill", "list")

	require.CmpNoError(err)
	assert.Cmp(query, td.Contains("date.from="))
	assert.Cmp(query, td.Contains("-01T00%3A00%3A00Z"), "the window starts on the first of the month")
}

// The category is a server-side filter. Applying it after the fact would list
// everything first, which is what the window guard exists to prevent.
func (ms *MockSuite) TestBillListSendsTheCategoryToTheApi(assert, require *td.T) {
	var query string
	captureQuery(billsURL, `[]`, &query)

	_, err := cmd.Execute("account", "bill", "list", "--category", "purchase-servers")

	require.CmpNoError(err)
	assert.Cmp(query, td.Contains("category=purchase-servers"))
}

// The accepted values are read from the embedded schema, not retyped.
func (ms *MockSuite) TestBillListRefusesACategoryTheApiDoesNotKnow(assert, require *td.T) {
	captureQuery(billsURL, `[]`, new(string))

	_, err := cmd.Execute("account", "bill", "list", "--category", "purchase-unicorns")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("purchase-unicorns"))
	assert.Cmp(err.Error(), td.Contains("purchase-servers"), "the refusal lists what is accepted")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "nothing is asked of the API before the flag is checked")
}

// Expanding a listing is one request per invoice. Past the guard it says how
// many rather than making them.
func (ms *MockSuite) TestBillListRefusesAWindowTooWideToExpand(assert, require *td.T) {
	ids := make([]string, 0, 401)
	for i := range 401 {
		ids = append(ids, fmt.Sprintf("%q", fmt.Sprintf("PI_FR%d", i)))
	}
	captureQuery(billsURL, "["+strings.Join(ids, ",")+"]", new(string))

	_, err := cmd.Execute("account", "bill", "list", "--from", "2019-01-01")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("401 invoices match"))
	assert.Cmp(err.Error(), td.Contains("--from"))
	assert.Cmp(httpmock.GetTotalCallCount(), 1, "not one invoice is expanded once the count is known")
}

// The link returns the PDF with no API token at all. A mask that only covers
// the human-readable output covers nothing, so it is applied to the object.
func (ms *MockSuite) TestBillGetHidesTheLinkAndPasswordEvenInJson(assert, require *td.T) {
	registerOneBill("PI_FR1")

	out, err := cmd.Execute("account", "bill", "get", "PI_FR1", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("SIGNATURE")), "the signature must never reach stdout")
	assert.Cmp(out, td.Not(td.Contains("B1llP4ssEx")))
	assert.Cmp(out, td.Contains(`"hidden": true`))
}

func (ms *MockSuite) TestBillGetRevealsOnDemand(assert, require *td.T) {
	registerOneBill("PI_FR1")

	out, err := cmd.Execute("account", "bill", "get", "PI_FR1", "--reveal", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("SIGNATURE"))
	assert.Cmp(out, td.Contains("B1llP4ssEx"))
}

func (ms *MockSuite) TestBillDetailsListsWhatWasCharged(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, billsURL+"/PI_FR1/details",
		httpmock.NewStringResponder(200, `["PI_FR2"]`))
	httpmock.RegisterResponder(http.MethodGet, billsURL+"/PI_FR1/details/PI_FR2",
		httpmock.NewStringResponder(200, `{"billDetailId":"PI_FR2","domain":"ns1.example",
			"description":"ADVANCE-1 rental","quantity":"1",
			"totalPrice":{"text":"104.99 €","value":104.99,"currencyCode":"EUR"}}`))

	out, err := cmd.Execute("account", "bill", "details", "PI_FR1")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ns1.example"))
	assert.Cmp(out, td.Contains("104.99"))
}

func (ms *MockSuite) TestRefundListHidesItsLinksToo(assert, require *td.T) {
	captureQuery(refundsURL, `["API_FR1"]`, new(string))
	httpmock.RegisterResponder(http.MethodGet, refundsURL+"/API_FR1",
		httpmock.NewStringResponder(200, `{"refundId":"API_FR1","date":"2026-03-05T14:22:17+01:00",
			"originalBillId":"PI_FR9","orderId":900000005,
			"priceWithTax":{"text":"-71.43 €","value":-71.43,"currencyCode":"EUR"},
			"password":"SECRETPASS","pdfUrl":"https://www.ovh.com/x?esign=SIGNATURE"}`))

	out, err := cmd.Execute("account", "refund", "list", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("API_FR1"))
	assert.Cmp(out, td.Not(td.Contains("SIGNATURE")))
	assert.Cmp(out, td.Not(td.Contains("SECRETPASS")))
}

// The entry-level price is null on every current-usage entry measured; the
// amount lives in the elements. One row per entry would have shown a column
// of nothing.
func (ms *MockSuite) TestUsageShowsOneRowPerElement(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, usageURL,
		httpmock.NewStringResponder(200, `[{"serviceId":74756498,"price":null,
			"beginDate":"2026-08-06T00:00:00Z","endDate":"2026-08-19T00:00:00Z",
			"elements":[
			  {"planCode":"okms-servicekey-monthly-consumption","planFamily":"okms-servicekey",
			   "quantity":9,"price":{"text":"0.54 €"}},
			  {"planCode":"okms-secret-monthly-consumption","planFamily":"okms-secret",
			   "quantity":13,"price":{"text":"0.39 €"}}]}]`))

	out, err := cmd.Execute("account", "usage")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("okms-servicekey-monthly-consumption"))
	assert.Cmp(out, td.Contains("okms-secret-monthly-consumption"))
	assert.Cmp(out, td.Contains("0.54"))
	assert.Cmp(out, td.Contains("0.39"))
	assert.Cmp(out, td.Contains("2026-08-06 → 2026-08-19"))
}

// A service that reports no element still has a line, and its missing amount
// is a dash rather than a blank cell.
func (ms *MockSuite) TestUsageKeepsAServiceThatReportsNoElement(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, usageURL,
		httpmock.NewStringResponder(200, `[{"serviceId":73553170,"price":null,
			"beginDate":"2026-07-10T00:00:00Z","endDate":"2026-07-10T00:00:00Z","elements":[]}]`))

	out, err := cmd.Execute("account", "usage")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("73553170"))
	assert.Cmp(out, td.Contains("—"))
}

func (ms *MockSuite) TestUsageForecastReadsTheOtherRoute(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, forecastURL,
		httpmock.NewStringResponder(200, `[{"serviceId":134043784,
			"beginDate":"2026-06-01T00:00:00Z","endDate":"2026-08-19T00:00:00Z",
			"price":{"text":"0.12 €"},"elements":[]}]`))

	out, err := cmd.Execute("account", "usage", "--forecast")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("0.12"))
	assert.Cmp(httpmock.GetCallCountInfo()["GET "+usageURL], 0, "the current period is not read")
}

// --- baremetal cost ---

// A server bills as several services: the machine, and the components sold
// with it. The price of the machine alone would not be the price paid.
func registerServerServices(server string, bodies map[string]string) {
	ids := make([]string, 0, len(bodies))
	for id := range bodies {
		ids = append(ids, fmt.Sprintf("%q", id))
	}
	httpmock.RegisterResponder(http.MethodGet, servicesURL,
		func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("resourceName") != server {
				return httpmock.NewStringResponse(200, `[]`), nil
			}
			return httpmock.NewStringResponse(200, "["+strings.Join(ids, ",")+"]"), nil
		})
	for id, body := range bodies {
		httpmock.RegisterResponder(http.MethodGet, servicesURL+"/"+id,
			httpmock.NewStringResponder(200, body))
	}
}

func (ms *MockSuite) TestCostAddsUpTheMachineAndWhatIsIncluded(assert, require *td.T) {
	registerServerServices("ns1.example", map[string]string{
		"1": `{"serviceId":1,"parentServiceId":null,"route":{"path":"/dedicated/server/{serviceName}"},
			"resource":{"displayName":"ns1.example"},
			"billing":{"plan":{"invoiceName":"ADVANCE-1 | AMD EPYC 4245P"},
			  "pricing":{"description":"rental for 1 month","price":{"text":"104.99 €","value":104.99,"currencyCode":"EUR"}},
			  "renew":{"current":{"mode":"automatic","nextDate":"2026-09-01T10:29:33Z"}}}}`,
		"2": `{"serviceId":2,"route":null,"resource":{"product":{"description":"2x SSD NVMe 1.92TB"}},
			"billing":{"plan":{"invoiceName":"2x SSD NVMe 1.92TB Datacenter Class Soft RAID"},
			  "pricing":{"description":"rental for 1 month","price":{"text":"70.00 €","value":70,"currencyCode":"EUR"}}}}`,
	})

	out, err := cmd.Execute("baremetal", "cost", "ns1.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ADVANCE-1"))
	assert.Cmp(out, td.Contains("2x SSD NVMe 1.92TB"), "what is included is listed, not summed away")
	// A billed component is not hypothetical: three servers on the account
	// measured carry one, at 70, 10 and 8 euros. Counting only the machine
	// would under-report the bill by exactly that much.
	assert.Cmp(out, td.Contains("174.99 EUR"), "the total is the machine plus its components")
	assert.Cmp(out, td.Not(td.Contains("104.99 EUR")), "the machine alone is not the price paid")
	assert.Cmp(out, td.Contains("Renews automatic on 2026-09-01"))
}

// serviceInfos.renew.automatic contradicts itself on exactly these services
// (PUBM-55135). Naming the parent is true; deriving a boolean is not.
func (ms *MockSuite) TestCostSaysWhenTheParentCarriesTheRenewal(assert, require *td.T) {
	registerServerServices("ns2.example", map[string]string{
		"9": `{"serviceId":9,"parentServiceId":133558145,"route":{"path":"/dedicated/server/{serviceName}"},
			"resource":{"displayName":"ns2.example"},
			"billing":{"plan":{"invoiceName":"HGR-HCI-1"},
			  "pricing":{"description":"rental for 1 month","price":{"text":"1106.00 €","value":1106,"currencyCode":"EUR"}},
			  "renew":null}}`,
	})

	out, err := cmd.Execute("baremetal", "cost", "ns2.example")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("133558145"))
	assert.Cmp(out, td.Contains("carried by parent service"))
	assert.Cmp(out, td.Not(td.Contains("Renews automatic")))
}

func (ms *MockSuite) TestCostSaysWhereToLookWhenNothingIsBilled(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet, servicesURL,
		httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("baremetal", "cost", "ns-unknown.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("nothing is billed for ns-unknown.example"))
	assert.Cmp(err.Error(), td.Contains("baremetal list"))
}

// Services answered, but none is the machine. That is not "costs nothing".
func (ms *MockSuite) TestCostRefusesWhenNoServiceIsTheMachine(assert, require *td.T) {
	registerServerServices("ns3.example", map[string]string{
		"7": `{"serviceId":7,"route":null,"billing":{"plan":{"invoiceName":"32GB DDR5"},
			"pricing":{"price":{"text":"0.00 €","value":0,"currencyCode":"EUR"}}}}`,
	})

	_, err := cmd.Execute("baremetal", "cost", "ns3.example")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("none of which is the machine"))
	assert.Cmp(err.Error(), td.Contains("/dedicated/server/{serviceName}"))
}
