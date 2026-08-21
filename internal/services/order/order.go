// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

// Package order reads and settles what has already been bought.
//
// `baremetal order` places an order and returns a number. Without this package
// that number is a dead end: the operator has to open a browser to learn
// whether the order was paid, where it is in delivery, or how to pay one that
// was placed with --no-pay. Closing that loop is the point.
package order

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// Flags.
var (
	ListDays         int
	ListFrom         string
	ListTo           string
	ListLimit        int
	PayPaymentMethod int
)

// DefaultListDays is the window `order list` reads when nothing says otherwise.
const DefaultListDays = 30

// price is how the billing API states an amount: a number, a currency and the
// string it wants shown. The text is used rather than rebuilt, so the figure on
// screen is the one on the invoice.
type price struct {
	CurrencyCode string  `json:"currencyCode"`
	Text         string  `json:"text"`
	Value        float64 `json:"value"`
}

type order struct {
	OrderID         int    `json:"orderId"`
	Date            string `json:"date"`
	ExpirationDate  string `json:"expirationDate"`
	RetractionDate  string `json:"retractionDate"`
	PriceWithTax    price  `json:"priceWithTax"`
	PriceWithoutTax price  `json:"priceWithoutTax"`
	Tax             price  `json:"tax"`
	URL             string `json:"url"`
	PDFURL          string `json:"pdfUrl"`

	// Password is the order form's access token. It is also embedded in URL and
	// PDFURL, so it is never printed on its own: a field labelled "password" is
	// what a log scanner finds, and repeating it under that name would add a
	// second copy without adding anything an operator can use.
	Password string `json:"password"`
}

// ListOrders shows what was ordered recently, and how far each order has got.
//
// Two calls per order — the order itself and its status, which the API keeps
// apart — so the window is bounded by default and the bound is stated rather
// than applied in silence. On an account with 1103 orders, a year's window
// would be two thousand requests for a command someone ran to check on
// yesterday's purchase.
func ListOrders(_ *cobra.Command, _ []string) {
	from, to, err := listWindow()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	query := url.Values{}
	query.Set("date.from", from)
	if to != "" {
		query.Set("date.to", to)
	}

	var ids []int
	if err := httpLib.Client.Get("/v1/me/order?"+query.Encode(), &ids); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list orders: %s", err)
		return
	}

	if len(ids) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"orders": []any{}},
			"No order placed since %s.", from)
		return
	}

	// Order identifiers grow with time, so the newest are the largest. Sorting
	// here rather than trusting the API's order means the limit below always
	// keeps the most recent rather than an arbitrary slice.
	sort.Sort(sort.Reverse(sort.IntSlice(ids)))

	truncated := 0
	if ListLimit > 0 && len(ids) > ListLimit {
		truncated = len(ids) - ListLimit
		ids = ids[:ListLimit]
	}

	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		// Sequential on purpose. go-ovh writes to the shared http.Client on
		// every request (ovh.go:387), so the parallel helper in this repo races
		// on it — shown by the race detector while building the catalogue.
		one, err := readOrder(id)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to read order %d: %s", id, err)
			return
		}
		status, err := readStatus(id)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "failed to read the status of order %d: %s", id, err)
			return
		}
		rows = append(rows, orderRow(one, status))
	}

	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	// Said, not silently applied: a table that stops at 25 with no word looks
	// like an account with 25 orders. Printed before the table, because
	// OutputWarning ends the process.
	if truncated > 0 {
		fmt.Fprintf(os.Stderr,
			"🟠 %d older orders in this window are not shown. Raise --limit, or narrow --days.\n", truncated)
	}

	display.RenderTable(filtered,
		[]string{"orderId", "date", "status", "priceWithTax", "retraction"},
		&flags.OutputFormatConfig)
}

// listWindow turns the flags into the window the API takes. --days is the one
// an operator reaches for; --from and --to are for a precise question.
func listWindow() (string, string, error) {
	if ListFrom != "" && ListDays != DefaultListDays {
		return "", "", fmt.Errorf("--from and --days both set the start of the window; keep one")
	}

	if ListTo != "" {
		if _, err := time.Parse(time.RFC3339, ListTo); err != nil {
			return "", "", fmt.Errorf("--to must be a date like 2026-08-31T00:00:00Z, not %q", ListTo)
		}
	}

	if ListFrom != "" {
		if _, err := time.Parse(time.RFC3339, ListFrom); err != nil {
			return "", "", fmt.Errorf("--from must be a date like 2026-08-01T00:00:00Z, not %q", ListFrom)
		}
		return ListFrom, ListTo, nil
	}

	if ListDays < 1 {
		return "", "", fmt.Errorf("--days must be at least 1, not %d", ListDays)
	}
	return time.Now().UTC().AddDate(0, 0, -ListDays).Format(time.RFC3339), ListTo, nil
}

func orderRow(o order, status string) map[string]any {
	row := map[string]any{
		"orderId":         o.OrderID,
		"date":            shortDate(o.Date),
		"status":          status,
		"priceWithTax":    o.PriceWithTax.Text,
		"priceWithoutTax": o.PriceWithoutTax.Text,
		"value":           o.PriceWithTax.Value,
		"currency":        o.PriceWithTax.CurrencyCode,
		"expiration":      shortDate(o.ExpirationDate),
		"retraction":      "kept",
	}
	// A retraction date means the right was given up on that day. Empty means
	// it still stands, which is the answer to the question this column exists
	// for: can this still be walked back?
	if o.RetractionDate != "" {
		row["retraction"] = "waived " + shortDate(o.RetractionDate)
	}
	return row
}

// shortDate keeps the day and drops the rest. The API answers
// "2026-08-16T02:17:31+02:00", and a table of those is unreadable.
func shortDate(value string) string {
	if len(value) < 10 {
		return value
	}
	return value[:10]
}

// GetOrder shows one order in full: what it cost, where it is, and the link to
// its order form.
func GetOrder(_ *cobra.Command, args []string) {
	id, err := orderID(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	one, err := readOrder(id)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read order %d: %s", id, err)
		return
	}
	status, err := readStatus(id)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the status of order %d: %s", id, err)
		return
	}

	details := orderRow(one, status)
	// The link is the reason to run this rather than open a browser, so it is
	// shown — and it carries the order form's access token, which is why the
	// output says so. `list` does not show it: printing a token per row for
	// fifty orders nobody asked about is a different act from handing over the
	// one that was asked for.
	details["url"] = one.URL
	// The API answers the same string twice on most orders; showing it twice
	// only makes the reader check whether they differ.
	if one.PDFURL != one.URL {
		details["pdfUrl"] = one.PDFURL
	}

	display.OutputObject(details, args[0], orderTemplate, &flags.OutputFormatConfig)
}

//go:embed templates/order.tmpl
var orderTemplate string

// FollowOrder shows where an order is in delivery, step by step.
func FollowOrder(_ *cobra.Command, args []string) {
	id, err := orderID(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var steps []struct {
		Step    string `json:"step"`
		Status  string `json:"status"`
		History []struct {
			Date   string `json:"date"`
			Status string `json:"status"`
		} `json:"history"`
	}
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/me/order/%d/followUp", id), &steps); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to follow order %d: %s", id, err)
		return
	}

	status, err := readStatus(id)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the status of order %d: %s", id, err)
		return
	}

	rows := make([]map[string]any, 0, len(steps))
	for _, step := range steps {
		latest := ""
		if n := len(step.History); n > 0 {
			latest = fmt.Sprintf("%s %s", shortDate(step.History[n-1].Date), step.History[n-1].Status)
		}
		rows = append(rows, map[string]any{
			"step":   step.Step,
			"status": step.Status,
			"latest": latest,
			// Carried on every row, and deliberately absent from the column
			// list below: RenderTable projects only the named columns for the
			// human table but hands the whole row map to the JSON and YAML
			// encoders. So this is the one place the overall status can live
			// without widening the table — and it has to live somewhere, because
			// the line printed to stderr further down never reaches -o json,
			// and the overall status is the single thing `follow` adds over
			// `order get`.
			"orderStatus": status,
		})
	}

	// The four steps can all read TODO on an order the API elsewhere calls
	// delivered — measured on a real delivered order. Printing the overall
	// status alongside is what keeps the table from reading as "nothing has
	// happened yet".
	fmt.Fprintf(os.Stderr, "Order %d is %s.\n", id, status)

	// --filter is registered on `order follow`, so it has to reach the rows.
	// RenderTable does not filter; withFilterFlag only binds the flag.
	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered, []string{"step", "status", "latest"}, &flags.OutputFormatConfig)
}

func orderID(value string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%q is not an order number; `ovhcloud order list` shows the recent ones", value)
	}
	return id, nil
}

func readOrder(id int) (order, error) {
	var one order
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/me/order/%d", id), &one); err != nil {
		return order{}, err
	}
	return one, nil
}

// readStatus is a call of its own because the API keeps it apart from the order
// itself: GET /me/order/{id} carries the price and the dates and no state.
func readStatus(id int) (string, error) {
	var status string
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/me/order/%d/status", id), &status); err != nil {
		return "", err
	}
	return status, nil
}

// PayOrder settles an order placed without payment.
//
// `baremetal order --no-pay` exists so a purchase can be reviewed by whoever
// signs for it before the money leaves; this is the other half. It spends
// money, so it is guarded like anything else that does — with a yes/no rather
// than a typed name: the order was already agreed to once, at the point where
// the summary and the price were on screen, and the sum has not changed since.
func PayOrder(_ *cobra.Command, args []string) {
	id, err := orderID(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	one, status, err := readOrderAndStatus(id)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// An order that is already paid is not an error worth a stack trace, but
	// saying nothing and sending the call anyway would let the API answer
	// something confusing about a state the CLI could see.
	if status != "notPaid" && status != "documentsRequested" {
		display.OutputError(&flags.OutputFormatConfig,
			"order %d is %s, not awaiting payment; nothing to pay", id, status)
		return
	}

	method := PayPaymentMethod
	if method == 0 {
		method, err = defaultPaymentMethod()
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	endpoint := fmt.Sprintf("/v1/me/order/%d/pay", id)

	if !common.ConfirmAction(common.Disruptive, args[0], fmt.Sprintf(
		"This pays order %d: %s incl. tax, with payment method %d.",
		id, one.PriceWithTax.Text, method)) {
		display.OutputError(&flags.OutputFormatConfig, "payment of order %d cancelled", id)
		return
	}

	body := map[string]any{"paymentMethod": map[string]any{"id": method}}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var ignored any
	if err := httpLib.Client.Post(endpoint, body, &ignored); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to pay order %d: %s", id, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"orderId": id, "paymentMethod": method, "amount": one.PriceWithTax.Value},
		"✅ Order %d paid: %s incl. tax.", id, one.PriceWithTax.Text)
}

// WaiveOrderRetraction gives up the right to walk an order back.
//
// It exists as a command of its own precisely so that it is never a side effect
// of buying. `baremetal order` will not send waiveRetractationPeriod even
// though the checkout body accepts it; giving up a legal right belongs in a
// sentence an operator typed on purpose. Which is also why this is guarded by
// the order number rather than a yes: it cannot be undone, and there is no
// second chance to read the summary.
func WaiveOrderRetraction(_ *cobra.Command, args []string) {
	id, err := orderID(args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	one, _, err := readOrderAndStatus(id)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if one.RetractionDate != "" {
		display.OutputError(&flags.OutputFormatConfig,
			"the retraction period of order %d was already waived on %s", id, shortDate(one.RetractionDate))
		return
	}

	endpoint := fmt.Sprintf("/v1/me/order/%d/waiveRetraction", id)

	if !common.ConfirmAction(common.Destructive, args[0], fmt.Sprintf(
		"This gives up the retraction period on order %d (%s incl. tax). It cannot be undone, and delivery of some services will not start before it is given up.",
		id, one.PriceWithTax.Text)) {
		display.OutputError(&flags.OutputFormatConfig, "order %d keeps its retraction period", id)
		return
	}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var ignored any
	if err := httpLib.Client.Post(endpoint, nil, &ignored); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to waive the retraction period of order %d: %s", id, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"orderId": id, "retraction": "waived"},
		"⚡️ Retraction period of order %d given up.", id)
}

func readOrderAndStatus(id int) (order, string, error) {
	one, err := readOrder(id)
	if err != nil {
		return order{}, "", fmt.Errorf("failed to read order %d: %s", id, err)
	}
	status, err := readStatus(id)
	if err != nil {
		return order{}, "", fmt.Errorf("failed to read the status of order %d: %s", id, err)
	}
	return one, status, nil
}

// defaultPaymentMethod finds the one the account pays with, so that the common
// case needs no flag. The list endpoint takes `default=true`, so this is one
// call rather than a list followed by a detail per entry.
func defaultPaymentMethod() (int, error) {
	var ids []int
	if err := httpLib.Client.Get("/v1/me/payment/method?default=true", &ids); err != nil {
		return 0, fmt.Errorf("could not read the payment methods of this account: %s", err)
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("this account has no default payment method; name one with --payment-method, or add one first")
	}
	return ids[0], nil
}
