// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// These two fixtures are the API's own answers, recorded on 18 August 2026 and
// trimmed to what a test needs. Inventing them would have tested the fixture.
const (
	realAvailabilities = `[{"fqn": "24adv01-v3.ram-128g-on-die-ecc-3600.hybridsoftraid-2x960nvme-pcie-gen4-2x1920nvme-pcie-gen4", "memory": "ram-128g-on-die-ecc-3600", "planCode": "24adv01-v3", "server": "24adv01", "storage": "hybridsoftraid-2x960nvme-pcie-gen4-2x1920nvme-pcie-gen4", "datacenters": [{"availability": "unavailable", "datacenter": "bhs"}, {"availability": "unavailable", "datacenter": "fra"}, {"availability": "1H-high", "datacenter": "gra"}, {"availability": "unavailable", "datacenter": "lon"}]}, {"fqn": "23scaleamd01-v3.ram-256g", "planCode": "23scaleamd01-v3", "server": "23scaleamd01", "memory": "ram-256g-ecc", "storage": "softraid-2x960nvme", "datacenters": [{"availability": "240H", "datacenter": "gra"}]}]`
	// RISE-1 on 18 August 2026, carrying a live promotion that waives its setup
	// charge. Trimmed to the fields the command reads; the amounts and the
	// promotion name are the API's own.
	promotedAvailabilities = `[{"fqn": "24rise01-v1.ram-32g-ecc-3200.softraid-2x512nvme", "memory": "ram-32g-ecc-3200", "planCode": "24rise01-v1", "server": "24rise01", "storage": "softraid-2x512nvme", "datacenters": [{"availability": "72H", "datacenter": "gra"}]}]`
	promotedCatalog        = `{"locale": {"currencyCode": "EUR"}, "plans": [{"planCode": "24rise01-v1", "pricings": [{"mode": "default", "price": 5699000000, "interval": 0, "intervalUnit": "none", "capacities": ["installation"], "promotions": [{"name": "FLASHSALE_WW_RISE_1_3_AND_GAME_T1", "total": {"value": 0}}]}, {"mode": "default", "price": 5699000000, "interval": 1, "intervalUnit": "month", "capacities": ["renew"], "promotions": []}]}]}`
	realCatalog            = `{"locale": {"currencyCode": "EUR", "subsidiary": "FR", "taxRate": 20}, "plans": [{"planCode": "24adv01-v3", "pricings": [{"phase": 0, "capacities": ["installation"], "commitment": 0, "description": "rental (only applicable for 1 time)", "interval": 0, "intervalUnit": "none", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": 1}, "price": 8999000000, "formattedPrice": "89.99 \u20ac", "tax": 1799800000, "mode": "default", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": null}, {"phase": 1, "capacities": ["renew"], "commitment": 0, "description": "rental for 1 month", "interval": 1, "intervalUnit": "month", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": null}, "price": 8999000000, "formattedPrice": "89.99 \u20ac", "tax": 1799800000, "mode": "default", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": null}, {"phase": 0, "capacities": ["installation"], "commitment": 0, "description": "rental (only applicable for 1 time)", "interval": 0, "intervalUnit": "none", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": 1}, "price": 0, "formattedPrice": "0.00 \u20ac", "tax": 0, "mode": "upfront12", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": null}, {"phase": 1, "capacities": ["renew"], "commitment": 12, "description": "rental for 12 months", "interval": 12, "intervalUnit": "month", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": null}, "price": 102589000000, "formattedPrice": "1025.89 \u20ac", "tax": 20517800000, "mode": "upfront12", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": {"defaultEndAction": "REACTIVATE_ENGAGEMENT", "duration": "P12M", "type": "upfront"}}, {"phase": 0, "capacities": ["installation"], "commitment": 0, "description": "rental (only applicable for 1 time)", "interval": 0, "intervalUnit": "none", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": 1}, "price": 0, "formattedPrice": "0.00 \u20ac", "tax": 0, "mode": "upfront24", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": null}, {"phase": 1, "capacities": ["renew"], "commitment": 24, "description": "rental for 24 months", "interval": 24, "intervalUnit": "month", "quantity": {"min": 1, "max": 10}, "repeat": {"min": 1, "max": null}, "price": 205177000000, "formattedPrice": "2051.77 \u20ac", "tax": 41035400000, "mode": "upfront24", "strategy": "tiered", "mustBeCompleted": false, "type": "rental", "promotions": [], "engagementConfiguration": {"defaultEndAction": "REACTIVATE_ENGAGEMENT", "duration": "P24M", "type": "upfront"}}]}]}`
)

// registerCatalog wires the API answers and, in the same breath, gives this test
// a cache of its own.
//
// The two belong together: the catalogue is cached on disk between runs, so a
// suite sharing one cache directory would have its first test fill it and every
// later test read from it instead of the responder it just registered. They
// would still pass — the cached values come from the same fixture — while
// measuring something else entirely. Isolating it here rather than in the
// caller means no test can forget.
func registerCatalog(t *td.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me",
		httpmock.NewStringResponder(200, `{"ovhSubsidiary": "FR"}`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=FR",
		httpmock.NewStringResponder(200, realCatalog))
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/dedicated/server/datacenter/availabilities`,
		httpmock.NewStringResponder(200, realAvailabilities))
}

// registerPromotedCatalog wires the same two endpoints to the RISE-1 answers.
func registerPromotedCatalog(t *td.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me",
		httpmock.NewStringResponder(200, `{"ovhSubsidiary": "FR"}`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=FR",
		httpmock.NewStringResponder(200, promotedCatalog))
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/dedicated/server/datacenter/availabilities`,
		httpmock.NewStringResponder(200, promotedAvailabilities))
}

// A promotion replaces the charge rather than annotating it, so the promoted
// figure is what the checkout bills — 56.99 for RISE-1, whose 56.99 setup
// charge is waived, against a 113.98 list price. Reporting the crossed-out
// price would overstate every offer currently on sale; reporting the promoted
// one without saying so would let an operator quote a price that expires on
// 31 August. Both halves are asserted here.
func (ms *MockSuite) TestBaremetalCatalogAppliesAndNamesAPromotion(assert, require *td.T) {
	registerPromotedCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"dueAtOrderValue": 56.99`), "the waived setup charge is not billed")
	assert.Cmp(out, td.Not(td.Contains(`"dueAtOrderValue": 113.98`)), "and the list price is not what is due")
	assert.Cmp(out, td.Contains("FLASHSALE_WW_RISE_1_3_AND_GAME_T1"), "the promotion is named")
	assert.Cmp(out, td.Contains("113.98 EUR"), "and what it was crossed out from is shown")
}

// The numbers below are the ones the API really returns for 24adv01-v3, so this
// test fails if the ucent conversion, the join or the commitment arithmetic
// drifts — not merely if the code changes.
func (ms *MockSuite) TestBaremetalCatalogJoinsPriceToAvailability(assert, require *td.T) {
	registerCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"planCode": "24adv01-v3"`))
	assert.Cmp(out, td.Contains(`"monthly": "89.99 EUR"`), "8999000000 ucents is 89.99")
	// 179.98, not 89.99. The order API quotes this exact cart at 179.98,
	// itemised as one INSTALLATION line and one DURATION line, and the number
	// below is that quote rather than an arithmetic of my own. Half the price
	// is the one error this column must never make.
	assert.Cmp(out, td.Contains(`"dueAtOrder": "179.98 EUR"`), "installation and first month are both charged")
	assert.Cmp(out, td.Not(td.Contains(`"dueAtOrder": "89.99 EUR"`)), "the setup charge is not the first month")
}

// A plan with no public price must say so. An empty cell reads as a bug, and a
// hundred and forty-five plan codes are sold on quotation.
func (ms *MockSuite) TestBaremetalCatalogNamesPlansSoldOnQuotation(assert, require *td.T) {
	registerCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("23scaleamd01-v3"), "the plan is listed")
	assert.Cmp(out, td.Contains("on quotation"), "and its price is explained rather than blank")
}

// Committing twelve months costs 1025.89 once instead of 89.99 twelve times,
// which is 85.49 a month. Two ways to misreport that, and this test refuses
// both: putting the period total in a column headed "monthly" inflates the
// offer twelvefold, and repeating the catalogue's zero installation charge as
// the amount due would tell an operator that a twelve-month commitment costs
// nothing to start. The mode is named upfront: the period is what is paid.
func (ms *MockSuite) TestBaremetalCatalogDividesAnUpfrontCommitment(assert, require *td.T) {
	registerCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "--commitment", "12", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"monthly": "85.49 EUR"`), "1025.89 over twelve months")
	assert.Cmp(out, td.Contains(`"dueAtOrder": "1025.89 EUR"`), "the whole period is paid at order, and its setup charge is zero")
	assert.Cmp(out, td.Not(td.Contains(`"monthly": "1025.89 EUR"`)), "never the period total in a monthly column")
	assert.Cmp(out, td.Not(td.Contains(`"dueAtOrder": "0.00 EUR"`)), "and never nothing")
}

// The availability code is a delay. Sorting it as text puts 1440H before 24H,
// which inverts the one thing this table is read for.
func (ms *MockSuite) TestBaremetalCatalogSortsBySoonestDelivery(assert, require *td.T) {
	registerCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("within the hour"), "1H-high is rendered as a delay, not a code")
	assert.Cmp(out, td.Contains("10d"), "and 240H as ten days")

	firstHour := indexOf(out, "within the hour")
	firstTenDays := indexOf(out, "10d")
	assert.Cmp(firstHour < firstTenDays, true, "the soonest delivery comes first")
}

// --available-only must drop what cannot be ordered, and keep what can.
func (ms *MockSuite) TestBaremetalCatalogAvailableOnlyDropsTheUnavailable(assert, require *td.T) {
	registerCatalog(assert)

	out, err := cmd.Execute("baremetal", "catalog", "--available-only")

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("unavailable")), "nothing undeliverable is listed")
	assert.Cmp(out, td.Contains("within the hour"), "and what is deliverable stays")
}

// A value the API would reject is named here, with the accepted ones, rather
// than travelling to a 400.
func (ms *MockSuite) TestBaremetalCatalogRejectsAnUnknownDatacenter(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog", "--datacenter", "atlantis")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("gra"), "the accepted values are listed")
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing is sent")
}

// The hardware filters must reach the API, not be applied after downloading
// four megabytes: a filtered call is 18 kB instead of 4.3 MB.
func (ms *MockSuite) TestBaremetalCatalogFiltersServerSide(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog", "--plan-code", "24adv01-v3")

	require.CmpNoError(err)
	// Skip the regexp responder's own key: httpmock records a served call under
	// both the pattern and the URL that matched, and map iteration order would
	// otherwise decide which one this assertion reads.
	var called string
	for call := range httpmock.GetCallCountInfo() {
		if indexOf(call, "datacenter/availabilities") >= 0 && indexOf(call, "=~") < 0 {
			called = call
		}
	}
	require.Not(called, "", "the availabilities endpoint must have been called")
	assert.Cmp(called, td.Contains("planCode=24adv01-v3"), "the filter travels as a query parameter")
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// The catalogue is twelve megabytes and takes seconds; the point of caching it
// is that the second call does not pay for it. The point of NOT caching
// availability is that stock moves by the hour.
func (ms *MockSuite) TestBaremetalCatalogCachesPricesButNotStock(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog")
	require.CmpNoError(err)
	cmd.PostExecute()
	_, err = cmd.Execute("baremetal", "catalog")
	require.CmpNoError(err)

	counts := httpmock.GetCallCountInfo()
	catalogCalls := counts["GET https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=FR"]
	// httpmock records a call twice when a regexp responder served it: once
	// under the pattern, once under the URL that matched. Counting only the
	// concrete URLs counts calls rather than bookkeeping entries.
	var availabilityCalls int
	for call, n := range counts {
		if indexOf(call, "datacenter/availabilities") >= 0 && indexOf(call, "=~") < 0 {
			availabilityCalls += n
		}
	}

	assert.Cmp(catalogCalls, 1, "the price list is downloaded once and reused")
	assert.Cmp(availabilityCalls, 2, "availability is asked again every time")
}

// "-low" and "-high" are the same delay with different stock behind it. Losing
// the distinction throws away the API's only warning that the machine may be
// gone by the time an order goes through.
func (ms *MockSuite) TestBaremetalCatalogKeepsTheLowStockWarning(assert, require *td.T) {
	registerCatalog(assert)
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/dedicated/server/datacenter/availabilities`,
		httpmock.NewStringResponder(200, `[{"planCode":"24adv01-v3","server":"24adv01","memory":"m","storage":"s",
			"datacenters":[{"availability":"1H-low","datacenter":"gra"}]}]`))

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("low stock"), "the grade of the stock survives the rendering")
}

// This command once carried a --region flag as well, reading
// /dedicated/server/region/availabilities. That endpoint answers 200 and
// returns all 8973 entries with an empty regions[] on every one: it neither
// honours the filter nor carries the data. The flag could not return a single
// row, and eleven tests passed anyway because the fixture recorded the
// datacenter endpoint and the region path was never exercised against the
// service. So the assertion below is on the surviving location filter, and it
// is on the wire rather than on the output: a filter is only applied if the
// API was asked to apply it.
func (ms *MockSuite) TestBaremetalCatalogSendsTheDatacenterFilter(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog", "--datacenter", "gra")

	require.CmpNoError(err)
	var called string
	for call := range httpmock.GetCallCountInfo() {
		if indexOf(call, "datacenter/availabilities") >= 0 && indexOf(call, "=~") < 0 {
			called = call
		}
	}
	require.Not(called, "", "the availabilities endpoint must have been called")
	assert.Cmp(called, td.Contains("datacenters=gra"), "the datacenter travels as a query parameter")
}

// --refresh must reach the API even when a fresh entry is sitting in the cache,
// otherwise the flag says something the command does not do.
func (ms *MockSuite) TestBaremetalCatalogRefreshIgnoresTheCache(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog")
	require.CmpNoError(err)
	cmd.PostExecute()
	_, err = cmd.Execute("baremetal", "catalog", "--refresh")
	require.CmpNoError(err)

	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=FR"], 2,
		"--refresh downloads the price list again")
}

// The subsidiary is not a closed list, and the command must not pretend it is.
//
// The set of accepted values is a property of the endpoint, measured on both:
// eu.api.ovh.com takes CZ DE ES FI FR GB IE IT LT MA NL PL PT SN TN, and
// ca.api.ovh.com takes CA QC ASIA AU SG IN WE WS. The embedded schema is a
// snapshot of the EU one, so validating against its enum locked eight
// subsidiaries out of the command — and `--country`, the documented way out, was
// checked against the very same list. QC stands in for all eight here.
func (ms *MockSuite) TestBaremetalCatalogServesASubsidiaryOutsideTheSchemaEnum(assert, require *td.T) {
	assert.Setenv("XDG_CACHE_HOME", assert.TempDir())
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me",
		httpmock.NewStringResponder(200, `{"ovhSubsidiary": "QC"}`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=QC",
		httpmock.NewStringResponder(200, realCatalog))
	httpmock.RegisterResponder("GET", `=~^https://eu\.api\.ovh\.com/v1/dedicated/server/datacenter/availabilities`,
		httpmock.NewStringResponder(200, realAvailabilities))

	out, err := cmd.Execute("baremetal", "catalog")

	require.CmpNoError(err)
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/order/catalog/public/eco?ovhSubsidiary=QC"], 1,
		"the subsidiary the account reports is the one asked for")
	assert.Cmp(out, td.Contains("24adv01-v3"), "and the offers come back")
}

// What replaces the enum still has to hold: the value names a file in the cache
// directory, so it is checked for being a subsidiary code and nothing else. This
// is the half of the old check that was doing real work.
func (ms *MockSuite) TestBaremetalCatalogRefusesASubsidiaryThatCouldNameAnyFile(assert, require *td.T) {
	registerCatalog(assert)

	_, err := cmd.Execute("baremetal", "catalog", "--country", "../../etc/passwd")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("letters only"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing is sent")
}
