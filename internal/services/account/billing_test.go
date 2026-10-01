// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package account

import (
	"strings"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
)

// Without a window the account measured for this returns 2215 invoices, and
// listing them is one request each. The default is a guard.
func TestAnEmptyWindowStartsAtTheFirstOfTheMonth(t *testing.T) {
	assert := td.Assert(t)

	from, to, err := billingWindow("", "")

	assert.CmpNoError(err)
	assert.Cmp(to, "")

	parsed, err := time.Parse(time.RFC3339, from)
	assert.CmpNoError(err)
	assert.Cmp(parsed.Day(), 1)
	assert.Cmp(parsed.Hour(), 0)
	assert.Cmp(parsed.Month(), time.Now().UTC().Month())
}

// Nobody types an RFC3339 stamp to ask for last month.
func TestAPlainDayIsAcceptedAsAWindowBound(t *testing.T) {
	assert := td.Assert(t)

	from, to, err := billingWindow("2026-01-01", "2026-02-01T12:30:00Z")

	assert.CmpNoError(err)
	assert.Cmp(from, "2026-01-01T00:00:00Z")
	assert.Cmp(to, "2026-02-01T12:30:00Z")
}

func TestAWindowThatEndsBeforeItStartsIsRefused(t *testing.T) {
	assert := td.Assert(t)

	_, _, err := billingWindow("2026-06-01", "2026-01-01")

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("is before"))
}

func TestSomethingThatIsNotADateIsRefused(t *testing.T) {
	assert := td.Assert(t)

	_, _, err := billingWindow("last month", "")

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--from is not a date"))
}

// The filters are applied by the API. Nothing is filtered after the fact.
func TestTheQueryCarriesEveryFilterGiven(t *testing.T) {
	assert := td.Assert(t)

	query, err := billQuery("2026-08-01T00:00:00Z", "2026-08-31T00:00:00Z", "purchase-servers", 4242)

	assert.CmpNoError(err)
	assert.Cmp(strings.HasPrefix(query, "?"), true)
	assert.Cmp(query, td.Contains("category=purchase-servers"))
	assert.Cmp(query, td.Contains("orderId=4242"))
	assert.Cmp(query, td.Contains("date.from=2026-08-01T00%3A00%3A00Z"))
	assert.Cmp(query, td.Contains("date.to="))
	// A colon left raw would end the value at the hour.
	assert.Cmp(strings.Contains(query, "00:00:00"), false)
}

// An absent bound is absent from the query, not sent empty.
func TestAnAbsentUpperBoundIsNotSent(t *testing.T) {
	assert := td.Assert(t)

	query, err := billQuery("2026-08-01T00:00:00Z", "", "", 0)

	assert.CmpNoError(err)
	assert.Cmp(strings.Contains(query, "date.to"), false)
	assert.Cmp(strings.Contains(query, "category"), false)
	assert.Cmp(strings.Contains(query, "orderId"), false)
}

// The link returns the PDF with no API token at all, measured. -o json would
// put it in a pipeline log, so the substitution happens on the object.
func TestTheDownloadLinkAndThePasswordAreHiddenByDefault(t *testing.T) {
	assert := td.Assert(t)
	RevealBillSecrets = false
	defer func() { RevealBillSecrets = false }()

	view := billSecretsView(map[string]any{
		"billId":   "PI_FR1",
		"password": "B1llP4ssEx",
		"url":      "https://www.ovh.com/cgi-bin/order/facture.pdf?esign=SECRET",
		"pdfUrl":   "https://www.ovh.com/cgi-bin/order/facture.pdf?esign=SECRET",
	})

	assert.Cmp(view["billId"], "PI_FR1")
	assert.Cmp(view["hidden"], true)
	for _, field := range []string{"password", "url", "pdfUrl"} {
		got := view[field].(string)
		assert.Cmp(strings.Contains(got, "esign"), false, field+" must not leak the signature")
		assert.Cmp(got, td.Not("B1llP4ssEx"))
	}
}

func TestRevealPrintsTheSecretsAndDropsTheHiddenMarker(t *testing.T) {
	assert := td.Assert(t)
	RevealBillSecrets = true
	defer func() { RevealBillSecrets = false }()

	view := billSecretsView(map[string]any{
		"password": "B1llP4ssEx",
		"pdfUrl":   "https://www.ovh.com/x?esign=SECRET",
	})

	assert.Cmp(view["password"], "B1llP4ssEx")
	assert.Cmp(view["pdfUrl"], "https://www.ovh.com/x?esign=SECRET")
	assert.Cmp(view["hidden"], nil)
}

// Masking must not invent fields the API did not send.
func TestAnObjectWithoutSecretsIsNotMarkedHidden(t *testing.T) {
	assert := td.Assert(t)
	RevealBillSecrets = false

	view := billSecretsView(map[string]any{"billId": "PI_FR1"})

	assert.Cmp(view["hidden"], nil)
	assert.Cmp(len(view), 1)
}

// The entry-level price is null on every current-usage entry measured.
func TestAMissingAmountIsSaidWithADashNotAnEmptyCell(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(priceText(nil), "—")
	assert.Cmp(priceText(&struct {
		Text string `json:"text"`
	}{Text: ""}), "—")
	assert.Cmp(priceText(&struct {
		Text string `json:"text"`
	}{Text: "320.00 €"}), "320.00 €")
}

func TestADateIsShownWithoutItsTime(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(day("2026-09-01T10:29:33Z"), "2026-09-01")
	assert.Cmp(day("2026-09-01"), "2026-09-01")
	assert.Cmp(day(""), "")
}
