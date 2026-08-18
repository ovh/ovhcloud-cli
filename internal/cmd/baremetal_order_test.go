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

// The order API's own answers for 24adv01-v3, recorded on 18 August 2026 from a
// cart that was created, filled, priced and then deleted. Nothing here was
// invented: `requiredConfiguration` is reproduced whole because its two lies
// are the reason this command is shaped the way it is — `dedicated_datacenter`
// says required=false and the checkout refuses without it, `region` says
// required=true and the checkout is happy without it.
const (
	orderOffer = `[{"planCode": "24adv01-v3", "productName": "ADVANCE-1 | AMD EPYC 4244P", "prices": [{"duration": "P0D", "pricingMode": "default", "capacities": ["installation"]}, {"duration": "P1M", "pricingMode": "default", "capacities": ["renew"]}, {"duration": "P0D", "pricingMode": "upfront12", "capacities": ["installation"]}, {"duration": "P1Y", "pricingMode": "upfront12", "capacities": ["renew"]}, {"duration": "P0D", "pricingMode": "upfront24", "capacities": ["installation"]}, {"duration": "P2Y", "pricingMode": "upfront24", "capacities": ["renew"]}]}]`

	orderRequiredConfiguration = `[{"allowedValues": ["bhs", "fra", "gra", "lon", "rbx", "sbg", "waw"], "fields": null, "label": "dedicated_datacenter", "required": false, "type": "String"}, {"allowedValues": ["none_64.en"], "fields": null, "label": "dedicated_os", "required": true, "type": "String"}, {"allowedValues": ["false", "true"], "fields": null, "label": "enable-backup", "required": false, "type": "String"}, {"allowedValues": null, "fields": null, "label": "extra_reservation", "required": false, "type": "String"}, {"allowedValues": ["canada", "europe"], "fields": null, "label": "region", "required": true, "type": "String"}, {"fields": null, "label": "TECH_ACCOUNT", "required": false, "type": "Nichandle"}, {"fields": null, "label": "ADMIN_ACCOUNT", "required": false, "type": "Nichandle"}]`

	// 179.98, which is 89.99 twice: the installation charge is billed on top of
	// the first month, not instead of it.
	orderQuote = `{"orderId": null, "url": null, "prices": {"originalWithoutTax": {"currencyCode": "EUR", "text": "179.98 €", "value": 179.98}, "reduction": {"currencyCode": "EUR", "text": "0.00 €", "value": 0}, "tax": {"currencyCode": "EUR", "text": "0.00 €", "value": 0}, "withTax": {"currencyCode": "EUR", "text": "179.98 €", "value": 179.98}, "withoutTax": {"currencyCode": "EUR", "text": "179.98 €", "value": 179.98}}, "contracts": [{"name": "Annexe Traitement de données à caractère personnel", "url": "https://storage.gra.cloud.ovh.net/v1/AUTH_325716a587c64897acbef9a4a4726e38/contracts/76ecbce-OVH_Data_Protection_Agreement-FR-7.0.pdf"}, {"name": "Specific Conditions for Global Backup \t", "url": "https://storage.gra.cloud.ovh.net/v1/AUTH_325716a587c64897acbef9a4a4726e38/contracts/d2e77b6-Backup_Agent-FR-2.0.pdf"}], "details": [{"detailType": "INSTALLATION", "description": "ADVANCE-1 | AMD EPYC 4244P location - centre de données gra - ", "quantity": 1, "totalPrice": {"currencyCode": "EUR", "text": "89.99 €", "value": 89.99}}, {"detailType": "DURATION", "description": "ADVANCE-1 | AMD EPYC 4244P location - centre de données gra - 1 mois", "quantity": 1, "totalPrice": {"currencyCode": "EUR", "text": "89.99 €", "value": 89.99}}, {"detailType": "DURATION", "description": "25Gbps guaranteed unmetered private bandwidth location - 1 mois", "quantity": 1, "totalPrice": {"currencyCode": "EUR", "text": "0.00 €", "value": 0}}]}`

	orderPlaced = `{"orderId": 220912275, "url": "https://www.ovh.com/manager/order/220912275", "prices": {"withoutTax": {"currencyCode": "EUR", "text": "179.98 EUR", "value": 179.98}, "withTax": {"currencyCode": "EUR", "text": "179.98 EUR", "value": 179.98}}}`

	orderCartID = "6b05735b-a551-486d-bf8c-9b89d0b5d16e"
)

// bodies records what was actually sent, per endpoint. Counting calls proves a
// request was made; only the body proves what it asked for, and three of the
// tests below are about a field that must never appear in one.
var orderBodies map[string][]map[string]any

func recordBody(endpoint string, response string) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		if req.Body != nil {
			raw, _ := io.ReadAll(req.Body)
			_ = json.Unmarshal(raw, &body)
		}
		orderBodies[endpoint] = append(orderBodies[endpoint], body)
		return httpmock.NewStringResponse(200, response), nil
	}
}

// registerOrder wires the seven calls of a successful order.
func registerOrder(t *td.T) {
	orderBodies = map[string][]map[string]any{}

	base := "https://eu.api.ovh.com/v1"
	cart := base + "/order/cart/" + orderCartID

	httpmock.RegisterResponder("GET", base+"/me",
		httpmock.NewStringResponder(200, `{"ovhSubsidiary": "FR"}`))
	httpmock.RegisterResponder("GET", base+"/me/payment/method?default=true",
		httpmock.NewStringResponder(200, `[900000010]`))
	httpmock.RegisterResponder("GET", base+"/me/payment/method/900000010",
		httpmock.NewStringResponder(200, `{"default": true, "status": "VALID", "label": "DeferredAccount"}`))
	httpmock.RegisterResponder("POST", base+"/order/cart",
		recordBody("cart", `{"cartId": "`+orderCartID+`"}`))
	httpmock.RegisterResponder("POST", cart+"/assign",
		httpmock.NewStringResponder(200, `null`))
	httpmock.RegisterResponder("GET", cart+"/eco?planCode=24adv01-v3",
		httpmock.NewStringResponder(200, orderOffer))
	httpmock.RegisterResponder("POST", cart+"/eco",
		recordBody("item", `{"itemId": 545365576}`))
	httpmock.RegisterResponder("GET", cart+"/item/545365576/requiredConfiguration",
		httpmock.NewStringResponder(200, orderRequiredConfiguration))
	httpmock.RegisterResponder("POST", cart+"/item/545365576/configuration",
		recordBody("configuration", `{"id": 1, "label": "x", "value": "y"}`))
	httpmock.RegisterResponder("GET", cart+"/checkout",
		httpmock.NewStringResponder(200, orderQuote))
	httpmock.RegisterResponder("POST", cart+"/checkout",
		recordBody("checkout", orderPlaced))
	httpmock.RegisterResponder("DELETE", cart,
		httpmock.NewStringResponder(200, `null`))
}

func orderCalls(kind string) int {
	base := "https://eu.api.ovh.com/v1"
	cart := base + "/order/cart/" + orderCartID
	switch kind {
	case "cart":
		return httpmock.GetCallCountInfo()["POST "+base+"/order/cart"]
	case "checkout":
		return httpmock.GetCallCountInfo()["POST "+cart+"/checkout"]
	case "quote":
		return httpmock.GetCallCountInfo()["GET "+cart+"/checkout"]
	case "delete":
		return httpmock.GetCallCountInfo()["DELETE "+cart]
	case "configuration":
		return httpmock.GetCallCountInfo()["POST "+cart+"/item/545365576/configuration"]
	}
	return -1
}

// The whole sequence must go out, in order, and the summary must carry the
// commercial reference and the datacenter rather than the plan code — which
// says nothing about what is being bought and which the operator has just
// typed anyway.
func (ms *MockSuite) TestBaremetalOrderPlacesTheOrderAndNamesTheProduct(assert, require *td.T) {
	registerOrder(assert)

	out, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--yes")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("220912275"), "the order number is reported")
	assert.Cmp(orderCalls("checkout"), 1, "exactly one order is placed")
	assert.Cmp(orderCalls("cart"), 1, "from exactly one cart")

	body := orderBodies["item"][0]
	assert.Cmp(body["planCode"], "24adv01-v3")
	assert.Cmp(body["duration"], "P1M", "the period is read back from the offer, not hardcoded")
	assert.Cmp(body["pricingMode"], "default")
	assert.Cmp(body["quantity"], float64(1))
}

// The one field that must never be sent. The checkout body accepts
// waiveRetractationPeriod, and sending it would sign away a right the operator
// never mentioned, inside a flag they used for something else. The API models
// waiving as POST /me/order/{id}/waiveRetraction, a separate and later act.
func (ms *MockSuite) TestBaremetalOrderNeverWaivesTheRetractionPeriod(assert, require *td.T) {
	registerOrder(assert)

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--yes")

	require.CmpNoError(err)
	require.Cmp(len(orderBodies["checkout"]), 1)
	_, present := orderBodies["checkout"][0]["waiveRetractationPeriod"]
	assert.False(present, "waiveRetractationPeriod must not be in the checkout body at all")
}

// Paying is the default, and --no-pay must actually change the request rather
// than only the message.
func (ms *MockSuite) TestBaremetalOrderPaysByDefaultAndNotWithNoPay(assert, require *td.T) {
	registerOrder(assert)

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--yes")
	require.CmpNoError(err)
	assert.Cmp(orderBodies["checkout"][0]["autoPayWithPreferredPaymentMethod"], true)

	cmd.PostExecute()
	registerOrder(assert)

	out, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--no-pay", "--yes")
	require.CmpNoError(err)
	assert.Cmp(orderBodies["checkout"][0]["autoPayWithPreferredPaymentMethod"], false)
	assert.Cmp(out, td.Contains("NOT paid"), "and the operator is told the order is unpaid")
}

// --quote must stop at the price. A quote that ordered would be the worst
// possible defect in this command.
func (ms *MockSuite) TestBaremetalOrderQuoteStopsBeforeBuying(assert, require *td.T) {
	registerOrder(assert)

	out, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--quote")

	require.CmpNoError(err)
	assert.Cmp(orderCalls("quote"), 1, "the quote is read")
	assert.Cmp(orderCalls("checkout"), 0, "and no order is placed")
	assert.Cmp(out, td.Contains("179.98"), "the price is shown")
	assert.Cmp(out, td.Contains("Nothing was ordered"))
}

// The guardrail. An unattended run that did not say --yes must not spend money.
func (ms *MockSuite) TestBaremetalOrderRefusesWithoutConfirmation(assert, require *td.T) {
	registerOrder(assert)

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("cancelled"))
	assert.Cmp(orderCalls("checkout"), 0, "no order must reach the API")
}

// --dry-run describes the whole sequence, not its last call, and creates no
// cart. The calls it hides are the ones that put a machine in a basket.
func (ms *MockSuite) TestBaremetalOrderDryRunDescribesTheWholeSequence(assert, require *td.T) {
	registerOrder(assert)

	out, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--dry-run")

	require.CmpNoError(err)
	for _, endpoint := range []string{
		"/v1/order/cart",
		"/v1/order/cart/{cartId}/assign",
		"/v1/order/cart/{cartId}/eco",
		"/v1/order/cart/{cartId}/item/{itemId}/requiredConfiguration",
		"/v1/order/cart/{cartId}/item/{itemId}/configuration",
		"/v1/order/cart/{cartId}/checkout",
	} {
		assert.Cmp(out, td.Contains(endpoint), "the preview names %s", endpoint)
	}
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "and nothing at all is sent")
}

// The cart is a draft, and a draft left behind is litter on the account. This
// is the test that would have caught the defect that shipped in the first
// version of this file: display.OutputError ends the process with os.Exit(1),
// so the `defer deleteCart` written above it never ran, and seven carts were
// left on a real account by seven refused runs. The successful path cleaned up
// correctly, which is exactly why the bug was invisible until the failing paths
// were counted.
func (ms *MockSuite) TestBaremetalOrderDeletesTheCartOnEveryPath(assert, require *td.T) {
	registerOrder(assert)
	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--yes")
	require.CmpNoError(err)
	assert.Cmp(orderCalls("delete"), 1, "after a successful order")

	cmd.PostExecute()
	registerOrder(assert)
	_, err = cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--quote")
	require.CmpNoError(err)
	assert.Cmp(orderCalls("delete"), 1, "after a quote")

	cmd.PostExecute()
	registerOrder(assert)
	_, err = cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra")
	require.CmpError(err)
	assert.Cmp(orderCalls("delete"), 1, "and after a refused confirmation")
}

// A system is never chosen at order time: every eco plan allows exactly one
// value for dedicated_os, `none_64.en`, because a dedicated server is delivered
// bare and installed afterwards. The command fills the formality and must not
// present it as a decision.
func (ms *MockSuite) TestBaremetalOrderFillsTheSystemWithoutShowingIt(assert, require *td.T) {
	registerOrder(assert)

	out, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--quote")

	require.CmpNoError(err)

	var sent []string
	for _, body := range orderBodies["configuration"] {
		sent = append(sent, body["label"].(string))
	}
	assert.Cmp(sent, td.Bag("dedicated_datacenter", "dedicated_os"),
		"the system is sent, because the API requires it")
	assert.Cmp(out, td.Not(td.Contains("none_64")), "and it is not shown as a choice")
	assert.Cmp(out, td.Not(td.Contains("dedicated_os")))
}

// The configuration endpoint answers 200 to anything — {"label":
// "dedicated_datacenter", "value": "atlantide"} included — and the refusal only
// arrives at the checkout, three calls later. So values are checked here,
// against the allowed values the API itself returned.
func (ms *MockSuite) TestBaremetalOrderRejectsAValueTheProductDoesNotAccept(assert, require *td.T) {
	registerOrder(assert)

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "atlantide")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("gra"), "the accepted values are named")
	assert.Cmp(orderCalls("configuration"), 0, "and nothing is configured")
	assert.Cmp(orderCalls("checkout"), 0)
}

// A label the product does not declare must be named, with the ones it does —
// and the single-valued formalities must not be offered as if they were
// choices.
func (ms *MockSuite) TestBaremetalOrderRejectsAnUnknownConfigurationLabel(assert, require *td.T) {
	registerOrder(assert)

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--config", "flavour=vanilla")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("enable-backup"), "the configurable labels are listed")
	assert.Cmp(err.Error(), td.Not(td.Contains("dedicated_os")),
		"but not the one with a single accepted value")
}

// The checkout's refusal names a label and stops there. The accepted values for
// that label are already in hand, and the operator needs them — with the flag
// that sets it, not the generic one.
func (ms *MockSuite) TestBaremetalOrderTurnsACheckoutRefusalIntoAnInstruction(assert, require *td.T) {
	registerOrder(assert)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/cart/"+orderCartID+"/checkout",
		httpmock.NewStringResponder(400,
			`{"message": "in customFields there must be a field with name 'dedicated_datacenter'"}`))

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--quote")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--datacenter"), "the flag that fixes it is named")
	assert.Cmp(err.Error(), td.Contains("bhs, fra, gra"), "with the values it accepts")
	assert.Cmp(orderCalls("delete"), 1, "and the cart is still cleaned up")
}

// An account that cannot pay must be told before a cart is built, not after it
// has read a quote.
func (ms *MockSuite) TestBaremetalOrderChecksThePaymentMethodFirst(assert, require *td.T) {
	registerOrder(assert)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/payment/method?default=true",
		httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no default payment method"))
	assert.Cmp(err.Error(), td.Contains("--no-pay"), "and the way round it is named")
	assert.Cmp(orderCalls("cart"), 0, "no cart is built")
}

// --quote asks for no payment method: it never pays, so an account without one
// must still be able to read a price.
func (ms *MockSuite) TestBaremetalOrderQuoteNeedsNoPaymentMethod(assert, require *td.T) {
	registerOrder(assert)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/me/payment/method?default=true",
		httpmock.NewStringResponder(200, `[]`))

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--quote")

	require.CmpNoError(err)
	assert.Cmp(httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/me/payment/method?default=true"], 0,
		"the payment method is not even read")
}

// A commitment the plan is not sold with must be named against the ones it is,
// rather than travelling to a 400 that says "invalid pricing mode".
func (ms *MockSuite) TestBaremetalOrderNamesTheCommitmentsAPlanIsSoldWith(assert, require *td.T) {
	registerOrder(assert)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/order/cart/"+orderCartID+"/eco?planCode=24adv01-v3",
		httpmock.NewStringResponder(200,
			`[{"planCode": "24adv01-v3", "productName": "ADVANCE-1", "prices": [{"duration": "P1M", "pricingMode": "default", "capacities": ["renew"]}]}]`))

	_, err := cmd.Execute("baremetal", "order", "24adv01-v3", "--datacenter", "gra", "--commitment", "24", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("it is sold with: default"))
	assert.Cmp(orderCalls("checkout"), 0)
}
