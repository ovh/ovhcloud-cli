// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// Flags of `order`.
var (
	OrderDatacenter string
	OrderConfigs    []string
	OrderCommitment string
	OrderQuantity   int
	OrderQuoteOnly  bool
	OrderNoPay      bool
)

// The order is seven calls, and every one of them was walked against the live
// API on 18 August 2026 with the cart deleted afterwards. Four things that walk
// taught, none of which the specification says:
//
//  1. `requiredConfiguration` gets `required` wrong in both directions.
//     `dedicated_datacenter` is declared false and the checkout refuses without
//     it ("in customFields there must be a field with name
//     'dedicated_datacenter'"); `region` is declared true and the checkout is
//     happy without it. So the flag is not read, and the datacenter is asked
//     for because the checkout asks for it.
//
//  2. `POST .../configuration` validates nothing. `{"label":
//     "dedicated_datacenter", "value": "atlantide"}` answers 200 with an
//     identifier, and the refusal only arrives at the checkout, three calls
//     later. Every value is therefore checked here, against the allowed values
//     the API itself returned.
//
//  3. The vocabularies of the catalogue and of the order look alike and are
//     not. `region` at order is `europe` or `canada`, never the `eu-west-gra`
//     of the availability endpoints — and setting it to `canada` on a server
//     delivered in `gra` is accepted without a word, which is why this command
//     never guesses it.
//
//  4. `POST /assign` is not optional. Without it the checkout answers "This
//     cart hasn't been assigned yet".
const (
	orderCartPath    = "/v1/order/cart"
	baremetalProduct = "eco"
)

// orderDurations maps the commitment an operator asks for to the (duration,
// pricingMode) couple the cart takes. The couples are read back from the API
// before use — this table only says which one to look for.
var orderCommitmentModes = map[string]string{
	"default": "default",
	"12":      "upfront12",
	"24":      "upfront24",
}

// OrderBaremetal buys one of the plan codes `baremetal catalog` lists.
//
// It is a thin shell around order(), and deliberately so: display.OutputError
// ends the process with os.Exit(1), so a `defer` written above it never runs
// and every statement after it is dead. That was measured, not assumed — seven
// carts were left behind on the account by the refusal paths of an earlier
// version of this file, one per run, while the successful path cleaned up
// correctly. Anything that has to happen on the way out therefore happens in
// order(), which returns instead of exiting.
func OrderBaremetal(_ *cobra.Command, args []string) {
	outcome, err := order(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	if outcome == nil {
		// --dry-run has already printed the sequence.
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, outcome.details, "%s", outcome.message)
}

// outcome is what the command has to say once the cart is gone.
type outcome struct {
	message string
	details map[string]any
}

func order(planCode string) (*outcome, error) {
	mode, known := orderCommitmentModes[OrderCommitment]
	if !known {
		return nil, fmt.Errorf("--commitment does not accept %q; accepted values are: default, 12, 24", OrderCommitment)
	}
	if OrderQuantity < 1 {
		return nil, fmt.Errorf("--quantity must be at least 1, not %d", OrderQuantity)
	}

	configs, err := parseOrderConfigs(OrderConfigs, OrderDatacenter)
	if err != nil {
		return nil, err
	}

	// The whole sequence, described before any of it happens. `reboot-rescue`
	// established the rule in #242: a preview that shows the last call hides
	// the ones that made it possible. Here the hidden ones create a cart and
	// put a machine in it.
	if common.ReportDryRun(orderCalls(planCode)...) {
		return nil, nil
	}

	subsidiary, err := accountSubsidiary()
	if err != nil {
		return nil, fmt.Errorf("could not read the subsidiary of this account: %s", err)
	}

	// Before the cart, not after it. An account with no usable payment method
	// can walk the whole sequence, read a quote and only then be refused; the
	// check costs one call and moves that refusal to the first second.
	// --quote never pays, so it never asks.
	if !OrderQuoteOnly && !OrderNoPay {
		if err := checkPaymentMethod(); err != nil {
			return nil, err
		}
	}

	cartID, err := createCart(subsidiary)
	if err != nil {
		return nil, fmt.Errorf("failed to create an order cart: %s", err)
	}
	// A cart is a draft, not a state: it is created, filled, priced and
	// consumed inside one invocation, and no cart identifier is ever shown to
	// the operator. Deleting it on the way out keeps a failed attempt from
	// leaving anything behind — including the attempts that fail, which is why
	// nothing between here and the return may call display.OutputError.
	defer deleteCart(cartID)

	if err := assignCart(cartID); err != nil {
		return nil, fmt.Errorf("failed to assign the cart to this account: %s", err)
	}

	offer, err := readOffer(cartID, planCode)
	if err != nil {
		return nil, fmt.Errorf("failed to read the offer %s: %s", planCode, err)
	}

	duration, found := offer.durationFor(mode)
	if !found {
		return nil, fmt.Errorf("%s is not sold with the %s commitment; it is sold with: %s",
			planCode, OrderCommitment, strings.Join(offer.commitments(), ", "))
	}

	itemID, err := addItem(cartID, planCode, duration, mode, OrderQuantity)
	if err != nil {
		return nil, fmt.Errorf("failed to put %s in the cart: %s", planCode, err)
	}

	required, err := readRequiredConfiguration(cartID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to read what %s must be configured with: %s", planCode, err)
	}

	settled, err := settleConfiguration(required, configs)
	if err != nil {
		return nil, err
	}

	for _, label := range sortedLabels(settled) {
		if err := setConfiguration(cartID, itemID, label, settled[label]); err != nil {
			return nil, fmt.Errorf("failed to set %s=%s: %s", label, settled[label], err)
		}
	}

	quote, err := readQuote(cartID)
	if err != nil {
		return nil, fmt.Errorf("the order was refused before anything was bought: %s%s",
			err, hintForQuoteError(err, required, settled))
	}

	if OrderQuoteOnly {
		return &outcome{
			message: renderQuote(offer, settled, quote) + "\nNothing was ordered: --quote stops here.",
			details: quoteDetails(offer, settled, quote),
		}, nil
	}

	// On stderr, beside the confirmation prompt rather than in the middle of
	// the JSON a caller is parsing — and printed even under --yes, so a
	// pipeline's log still records what was bought.
	fmt.Fprintln(os.Stderr, renderQuote(offer, settled, quote))

	// The datacenter is the choice that is both irreversible and expensive to
	// get wrong: a server delivered on the wrong continent is terminated, not
	// moved. So it is what has to be typed back — not the plan code, which is
	// no longer on screen and which the operator has just typed anyway.
	confirmation := settled[datacenterLabel]
	warning := fmt.Sprintf("This orders %s in %s for %s and pays it. %s",
		offer.ProductName, confirmation, quote.totalWithTax(), retractionNote)
	if confirmation == "" {
		// No datacenter to type back. Fall back to the yes/no of #242 rather
		// than inventing a token out of something the operator did not choose.
		confirmation = offer.ProductName
		warning = fmt.Sprintf("This orders %s for %s and pays it. %s",
			offer.ProductName, quote.totalWithTax(), retractionNote)
	}
	if OrderNoPay {
		warning = strings.Replace(warning, " and pays it.", " without paying it.", 1)
	}

	if !common.ConfirmAction(common.Destructive, confirmation, warning) {
		return nil, fmt.Errorf("order of %s cancelled; nothing was bought", offer.ProductName)
	}

	placed, err := checkout(cartID, !OrderNoPay)
	if err != nil {
		return nil, fmt.Errorf("the order was refused: %s", err)
	}

	message := fmt.Sprintf("✅ Order %d placed for %s and paid. Follow it with:\n  ovhcloud order follow %d",
		placed.OrderID, offer.ProductName, placed.OrderID)
	if OrderNoPay {
		message = fmt.Sprintf("✅ Order %d placed for %s and NOT paid. Pay it with:\n  ovhcloud order pay %d\nor at %s",
			placed.OrderID, offer.ProductName, placed.OrderID, placed.URL)
	}

	return &outcome{
		message: message,
		details: map[string]any{
			"orderId":       placed.OrderID,
			"url":           placed.URL,
			"paid":          !OrderNoPay,
			"product":       offer.ProductName,
			"planCode":      planCode,
			"quantity":      OrderQuantity,
			"commitment":    OrderCommitment,
			"configuration": settled,
		},
	}, nil
}

// retractionNote is why waiveRetractationPeriod is never sent. The checkout
// body accepts it, and sending it would quietly sign away a right the operator
// did not mention. The API models waiving as its own later call, and this
// command follows that shape.
const retractionNote = "The retraction period is kept; waiving it is a separate, deliberate act."

const datacenterLabel = "dedicated_datacenter"

// orderCalls is the sequence, for --dry-run. The cart and item identifiers do
// not exist yet, so they are named rather than filled: the point is to let an
// operator read what would be created before it is.
func orderCalls(planCode string) []common.Call {
	calls := []common.Call{
		{Method: "GET", Endpoint: "/v1/me"},
	}
	if !OrderQuoteOnly && !OrderNoPay {
		calls = append(calls, common.Call{Method: "GET", Endpoint: "/v1/me/payment/method?default=true"})
	}
	calls = append(calls,
		common.Call{Method: "POST", Endpoint: orderCartPath},
		common.Call{Method: "POST", Endpoint: orderCartPath + "/{cartId}/assign"},
		common.Call{Method: "GET", Endpoint: fmt.Sprintf("%s/{cartId}/%s?planCode=%s", orderCartPath, baremetalProduct, planCode)},
		common.Call{Method: "POST", Endpoint: fmt.Sprintf("%s/{cartId}/%s", orderCartPath, baremetalProduct)},
		common.Call{Method: "GET", Endpoint: orderCartPath + "/{cartId}/item/{itemId}/requiredConfiguration"},
		common.Call{Method: "POST", Endpoint: orderCartPath + "/{cartId}/item/{itemId}/configuration"},
		common.Call{Method: "GET", Endpoint: orderCartPath + "/{cartId}/checkout"},
	)
	if !OrderQuoteOnly {
		calls = append(calls, common.Call{Method: "POST", Endpoint: orderCartPath + "/{cartId}/checkout"})
	}
	calls = append(calls, common.Call{Method: "DELETE", Endpoint: orderCartPath + "/{cartId}"})
	return calls
}

func parseOrderConfigs(pairs []string, datacenter string) (map[string]string, error) {
	configs := map[string]string{}
	for _, pair := range pairs {
		label, value, found := strings.Cut(pair, "=")
		if !found || label == "" {
			return nil, fmt.Errorf("--config takes label=value, not %q", pair)
		}
		configs[label] = value
	}
	if datacenter != "" {
		if existing, clash := configs[datacenterLabel]; clash && existing != datacenter {
			return nil, fmt.Errorf("--datacenter says %q and --config %s says %q; they cannot both be right",
				datacenter, datacenterLabel, existing)
		}
		configs[datacenterLabel] = datacenter
	}
	return configs, nil
}

func sortedLabels(configs map[string]string) []string {
	labels := make([]string, 0, len(configs))
	for label := range configs {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}

// requiredField is one configuration the product declares.
type requiredField struct {
	Label         string   `json:"label"`
	Required      bool     `json:"required"`
	AllowedValues []string `json:"allowedValues"`
}

// settleConfiguration decides what will be sent, and refuses rather than guess.
//
// Three rules, in order:
//
//   - A label the API allows exactly one value for is a formality, not a
//     choice: it is filled here and never exposed as a flag. `dedicated_os` is
//     the one that matters — every eco plan allows `none_64.en` and nothing
//     else, because a dedicated server is delivered bare and the system is
//     installed afterwards with `baremetal reinstall`. Asking an operator to
//     name an operating system at order time would be asking them to pick from
//     a list of one.
//   - A value the operator gave is checked against the allowed values the API
//     returned, because the configuration endpoint accepts anything and the
//     refusal only arrives at the checkout.
//   - Everything else is left alone. `required` is not trusted (see the
//     comment at the top of this file), so a label is not demanded here on its
//     say-so; the checkout is the authority, and its refusal is relayed with
//     the accepted values attached.
func settleConfiguration(required []requiredField, given map[string]string) (map[string]string, error) {
	settled := map[string]string{}
	known := map[string]requiredField{}

	for _, field := range required {
		known[field.Label] = field
		if len(field.AllowedValues) == 1 {
			settled[field.Label] = field.AllowedValues[0]
		}
	}

	for label, value := range given {
		field, declared := known[label]
		if !declared {
			return nil, fmt.Errorf("this product has no configuration called %q; it takes: %s",
				label, strings.Join(configurableLabels(required), ", "))
		}
		if len(field.AllowedValues) > 0 && !containsString(field.AllowedValues, value) {
			return nil, fmt.Errorf("%s does not accept %q; accepted values are: %s",
				label, value, strings.Join(field.AllowedValues, ", "))
		}
		settled[label] = value
	}

	return settled, nil
}

// configurableLabels lists what an operator may set, which is not what the
// product declares: a label with a single allowed value is filled without them.
func configurableLabels(required []requiredField) []string {
	labels := make([]string, 0, len(required))
	for _, field := range required {
		if len(field.AllowedValues) == 1 {
			continue
		}
		labels = append(labels, field.Label)
	}
	sort.Strings(labels)
	return labels
}

// hintForQuoteError turns the checkout's terse refusal into something
// actionable. "in customFields there must be a field with name
// 'dedicated_datacenter'" names a label and stops; the accepted values for that
// label are already in hand, and the operator needs them to fix the command.
// flagFor names the dedicated flag when one exists. Telling an operator to
// write `--config dedicated_datacenter=gra` when `--datacenter gra` is right
// there is technically true and practically unhelpful.
func flagFor(label string) string {
	if label == datacenterLabel {
		return "--datacenter <value>"
	}
	return fmt.Sprintf("--config %s=<value>", label)
}

func hintForQuoteError(err error, required []requiredField, settled map[string]string) string {
	message := err.Error()
	for _, field := range required {
		if !strings.Contains(message, field.Label) {
			continue
		}
		if _, alreadySet := settled[field.Label]; alreadySet {
			continue
		}
		if len(field.AllowedValues) > 0 {
			return fmt.Sprintf("\nAdd %s. Accepted values: %s",
				flagFor(field.Label), strings.Join(field.AllowedValues, ", "))
		}
		return fmt.Sprintf("\nAdd %s.", flagFor(field.Label))
	}
	return ""
}

// offer is what the cart says about a plan code before anything is added.
type offer struct {
	PlanCode    string `json:"planCode"`
	ProductName string `json:"productName"`
	Prices      []struct {
		Duration    string   `json:"duration"`
		PricingMode string   `json:"pricingMode"`
		Capacities  []string `json:"capacities"`
	} `json:"prices"`
}

// durationFor finds the period the cart wants for a pricing mode instead of
// hardcoding it. The couples are P1M/default, P1Y/upfront12 and P2Y/upfront24
// today; a table in Go would be a copy that stops being true without saying so,
// and the answer is one call away.
//
// Only the renewal price is looked at: every mode also carries a P0D
// installation price, and P0D as a duration is refused by the cart.
func (o offer) durationFor(mode string) (string, bool) {
	for _, price := range o.Prices {
		if price.PricingMode == mode && containsString(price.Capacities, "renew") {
			return price.Duration, true
		}
	}
	return "", false
}

func (o offer) commitments() []string {
	seen := map[string]bool{}
	names := []string{}
	for _, price := range o.Prices {
		if !containsString(price.Capacities, "renew") || seen[price.PricingMode] {
			continue
		}
		seen[price.PricingMode] = true
		for asked, mode := range orderCommitmentModes {
			if mode == price.PricingMode {
				names = append(names, asked)
			}
		}
	}
	sort.Strings(names)
	return names
}

// amount is one price as the order API states it: a number, a currency and the
// string it wants shown. The text is used, not rebuilt, so the operator reads
// the same figure the invoice will carry.
type amount struct {
	CurrencyCode string  `json:"currencyCode"`
	Text         string  `json:"text"`
	Value        float64 `json:"value"`
}

type quote struct {
	OrderID int    `json:"orderId"`
	URL     string `json:"url"`
	Prices  struct {
		WithoutTax         amount `json:"withoutTax"`
		WithTax            amount `json:"withTax"`
		Tax                amount `json:"tax"`
		OriginalWithoutTax amount `json:"originalWithoutTax"`
		Reduction          amount `json:"reduction"`
	} `json:"prices"`
	Contracts []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"contracts"`
	Details []struct {
		DetailType  string `json:"detailType"`
		Description string `json:"description"`
		// A number, not a string — checked against the API's own answer after
		// a first version of this struct read the OpenAPI type and got a
		// "cannot unmarshal number into Go struct field" on the first live
		// quote.
		Quantity   int    `json:"quantity"`
		TotalPrice amount `json:"totalPrice"`
	} `json:"details"`
}

func (q quote) totalWithTax() string {
	if q.Prices.WithTax.Text != "" {
		return q.Prices.WithTax.Text
	}
	return q.Prices.WithoutTax.Text
}

// renderQuote is what an operator reads before agreeing to spend money.
//
// It shows the commercial reference and the datacenter rather than the plan
// code: `24adv01-v3` says nothing about what is being bought, and the operator
// typed it a moment ago anyway. Every configuration line is one the API
// returned and this command actually set — nothing is illustrated.
func renderQuote(o offer, settled map[string]string, q quote) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\n  Offer        %s\n", o.ProductName)
	for _, label := range sortedLabels(settled) {
		// The system is a formality with one accepted value, filled by this
		// command; showing it would suggest a choice was made.
		if label == "dedicated_os" {
			continue
		}
		fmt.Fprintf(&b, "  %-12s %s\n", prettyLabel(label), settled[label])
	}
	fmt.Fprintf(&b, "  Quantity     %d\n", OrderQuantity)
	fmt.Fprintf(&b, "  Commitment   %s\n", prettyCommitment(OrderCommitment))
	fmt.Fprintf(&b, "  Due now      %s excl. tax", q.Prices.WithoutTax.Text)
	if q.Prices.WithTax.Text != "" && q.Prices.WithTax.Value != q.Prices.WithoutTax.Value {
		fmt.Fprintf(&b, "     %s incl. tax", q.Prices.WithTax.Text)
	}
	b.WriteString("\n")
	if q.Prices.Reduction.Value > 0 {
		fmt.Fprintf(&b, "  Reduction    %s off %s\n",
			q.Prices.Reduction.Text, q.Prices.OriginalWithoutTax.Text)
	}
	if len(q.Contracts) > 0 {
		names := make([]string, 0, len(q.Contracts))
		for _, c := range q.Contracts {
			// Trimmed: the API returns at least one name with a trailing tab,
			// and it lands mid-sentence in the summary.
			names = append(names, strings.TrimSpace(c.Name))
		}
		fmt.Fprintf(&b, "  Contracts    %d to accept: %s\n", len(names), strings.Join(names, "; "))
	}

	return b.String()
}

func quoteDetails(o offer, settled map[string]string, q quote) map[string]any {
	contracts := make([]map[string]any, 0, len(q.Contracts))
	for _, c := range q.Contracts {
		contracts = append(contracts, map[string]any{"name": strings.TrimSpace(c.Name), "url": c.URL})
	}
	lines := make([]map[string]any, 0, len(q.Details))
	for _, d := range q.Details {
		lines = append(lines, map[string]any{
			"type": d.DetailType, "description": d.Description,
			"quantity": d.Quantity, "total": d.TotalPrice.Text,
		})
	}
	return map[string]any{
		"product":       o.ProductName,
		"planCode":      o.PlanCode,
		"configuration": settled,
		"quantity":      OrderQuantity,
		"commitment":    OrderCommitment,
		"withoutTax":    q.Prices.WithoutTax.Value,
		"withTax":       q.Prices.WithTax.Value,
		"currency":      q.Prices.WithoutTax.CurrencyCode,
		"contracts":     contracts,
		"details":       lines,
	}
}

func prettyLabel(label string) string {
	switch label {
	case datacenterLabel:
		return "Datacenter"
	case "region":
		return "Region"
	case "enable-backup":
		return "Backup"
	default:
		return label
	}
}

func prettyCommitment(commitment string) string {
	switch commitment {
	case "12":
		return "12 months, paid upfront"
	case "24":
		return "24 months, paid upfront"
	default:
		return "monthly"
	}
}

// --- the seven calls ---------------------------------------------------------

// checkPaymentMethod refuses early when nothing can pay. The list endpoint
// takes `default=true`, so this is one call and not a list followed by a detail
// per entry.
func checkPaymentMethod() error {
	var ids []int
	if err := httpLib.Client.Get("/v1/me/payment/method?default=true", &ids); err != nil {
		return fmt.Errorf("could not read the payment methods of this account: %s", err)
	}
	if len(ids) == 0 {
		return fmt.Errorf("this account has no default payment method, so an order cannot be paid; add one, or use --no-pay to place the order and pay it later")
	}

	var method struct {
		Status string `json:"status"`
		Label  string `json:"label"`
	}
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/me/payment/method/%d", ids[0]), &method); err != nil {
		return fmt.Errorf("could not read the default payment method: %s", err)
	}
	if method.Status != "VALID" {
		return fmt.Errorf("the default payment method (%s) is %s, not VALID, so an order cannot be paid; fix it, or use --no-pay",
			method.Label, method.Status)
	}
	return nil
}

func createCart(subsidiary string) (string, error) {
	var cart struct {
		CartID string `json:"cartId"`
	}
	body := map[string]any{"ovhSubsidiary": subsidiary}
	if err := httpLib.Client.Post(orderCartPath, body, &cart); err != nil {
		return "", err
	}
	if cart.CartID == "" {
		return "", fmt.Errorf("the API returned a cart with no identifier")
	}
	return cart.CartID, nil
}

// assignCart attaches the cart to the account. It is mandatory: without it the
// checkout answers "This cart hasn't been assigned yet", four calls later.
func assignCart(cartID string) error {
	var ignored any
	return httpLib.Client.Post(fmt.Sprintf("%s/%s/assign", orderCartPath, url.PathEscape(cartID)), nil, &ignored)
}

// deleteCart tidies up. Its failure is not worth a message: the cart expires by
// itself, and this runs on the way out of a command that has already said
// whatever it had to say.
func deleteCart(cartID string) {
	var ignored any
	_ = httpLib.Client.Delete(fmt.Sprintf("%s/%s", orderCartPath, url.PathEscape(cartID)), &ignored)
}

func readOffer(cartID, planCode string) (offer, error) {
	var offers []offer
	endpoint := fmt.Sprintf("%s/%s/%s?planCode=%s",
		orderCartPath, url.PathEscape(cartID), baremetalProduct, url.QueryEscape(planCode))
	if err := httpLib.Client.Get(endpoint, &offers); err != nil {
		return offer{}, err
	}
	if len(offers) == 0 {
		return offer{}, fmt.Errorf("no such plan code in the %s catalogue; `ovhcloud baremetal catalog` lists the ones there are", baremetalProduct)
	}
	return offers[0], nil
}

func addItem(cartID, planCode, duration, mode string, quantity int) (int, error) {
	var item struct {
		ItemID int `json:"itemId"`
	}
	body := map[string]any{
		"planCode":    planCode,
		"duration":    duration,
		"pricingMode": mode,
		"quantity":    quantity,
	}
	endpoint := fmt.Sprintf("%s/%s/%s", orderCartPath, url.PathEscape(cartID), baremetalProduct)
	if err := httpLib.Client.Post(endpoint, body, &item); err != nil {
		return 0, err
	}
	if item.ItemID == 0 {
		return 0, fmt.Errorf("the API returned an item with no identifier")
	}
	return item.ItemID, nil
}

func readRequiredConfiguration(cartID string, itemID int) ([]requiredField, error) {
	var fields []requiredField
	endpoint := fmt.Sprintf("%s/%s/item/%d/requiredConfiguration", orderCartPath, url.PathEscape(cartID), itemID)
	if err := httpLib.Client.Get(endpoint, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func setConfiguration(cartID string, itemID int, label, value string) error {
	var ignored any
	endpoint := fmt.Sprintf("%s/%s/item/%d/configuration", orderCartPath, url.PathEscape(cartID), itemID)
	return httpLib.Client.Post(endpoint, map[string]any{"label": label, "value": value}, &ignored)
}

func readQuote(cartID string) (quote, error) {
	var q quote
	endpoint := fmt.Sprintf("%s/%s/checkout", orderCartPath, url.PathEscape(cartID))
	if err := httpLib.Client.Get(endpoint, &q); err != nil {
		return quote{}, err
	}
	return q, nil
}

// checkout is the only call in this file that spends money.
//
// `waiveRetractationPeriod` is accepted by this body and is never sent. The API
// exposes waiving as POST /me/order/{id}/waiveRetraction, a separate and later
// act; sending it here would sign away a right on behalf of an operator who
// never mentioned it, inside a flag they used for something else.
func checkout(cartID string, pay bool) (quote, error) {
	var placed quote
	endpoint := fmt.Sprintf("%s/%s/checkout", orderCartPath, url.PathEscape(cartID))
	body := map[string]any{"autoPayWithPreferredPaymentMethod": pay}
	if err := httpLib.Client.Post(endpoint, body, &placed); err != nil {
		return quote{}, err
	}
	return placed, nil
}
