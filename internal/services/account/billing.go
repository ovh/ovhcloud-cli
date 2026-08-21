// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package account

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

const (
	billsPath    = "/v1/me/bill"
	refundsPath  = "/v1/me/refund"
	usagePath    = "/v1/me/consumption/usage/current"
	forecastPath = "/v1/me/consumption/usage/forecast"
)

var (
	// BillFrom and BillTo bound the window a listing covers. Empty means
	// the current month, which is a guard rather than a convenience: the
	// account measured for this work carries 2215 invoices, and listing
	// them expands one HTTP call per invoice.
	BillFrom string
	BillTo   string

	// BillCategory filters on billing.CategoryEnum, read from the schema.
	BillCategory string

	// BillOrderID keeps only what was billed for one order.
	BillOrderID int64

	// RevealBillSecrets prints the download link and the PDF password
	// instead of their fingerprints. Both are bearer secrets: the link
	// returns the PDF with no API token at all, measured.
	RevealBillSecrets bool

	// BillUsageForecast reads the forecast instead of the current usage.
	BillUsageForecast bool
)

// billTooManyToExpand is the point past which a listing is refused rather
// than turned into that many HTTP calls. 2215 invoices exist on the account
// this was measured on; a year of them is 396.
const billTooManyToExpand = 400

var billColumnsToDisplay = []string{"billId", "date", "category", "priceWithTax.text price", "orderId"}

var refundColumnsToDisplay = []string{"refundId", "date", "originalBillId", "priceWithTax.text price", "orderId"}

var billDetailColumns = []string{"billDetailId", "domain", "description", "quantity", "totalPrice.text total"}

var usageColumns = []string{"serviceId", "plan", "family", "quantity", "price", "period"}

// billCategories reads billing.CategoryEnum from the embedded schema rather
// than repeating it here. The game protocols of #256 were copied by hand and
// were already stale when they shipped.
var billCategories = sync.OnceValues(func() ([]string, error) {
	return openapi.GetComponentEnum(assets.MeOpenapiSchema, "billing.CategoryEnum")
})

// CompleteBillCategory offers the categories on <tab>.
func CompleteBillCategory(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(billCategories)
}

// billingWindow returns the from/to pair a listing runs on. An empty --from
// means the first day of the current month: without a window the call is
// answered in full, and the expansion that follows is one request per invoice.
func billingWindow(from, to string) (string, string, error) {
	if from == "" {
		now := time.Now().UTC()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	} else {
		parsed, err := parseWindowBound(from)
		if err != nil {
			return "", "", fmt.Errorf("--from is not a date: %w", err)
		}
		from = parsed
	}

	if to != "" {
		parsed, err := parseWindowBound(to)
		if err != nil {
			return "", "", fmt.Errorf("--to is not a date: %w", err)
		}
		to = parsed
	}

	if to != "" && to < from {
		return "", "", fmt.Errorf("--to (%s) is before --from (%s)", to, from)
	}

	return from, to, nil
}

// parseWindowBound accepts a plain day as well as a full timestamp, because
// nobody types an RFC3339 stamp to ask for last month.
func parseWindowBound(value string) (string, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC().Format(time.RFC3339), nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", fmt.Errorf("%q is neither YYYY-MM-DD nor a full RFC3339 timestamp", value)
	}

	return parsed.UTC().Format(time.RFC3339), nil
}

// billQuery builds the query string of a billing collection. The filters are
// applied by the API; nothing is filtered after the fact.
func billQuery(from, to, category string, orderID int64) (string, error) {
	values := url.Values{}
	values.Set("date.from", from)
	if to != "" {
		values.Set("date.to", to)
	}

	if category != "" {
		if err := common.CheckEnumFlag("category", category, billCategories); err != nil {
			return "", err
		}
		values.Set("category", category)
	}

	if orderID != 0 {
		values.Set("orderId", fmt.Sprint(orderID))
	}

	return "?" + values.Encode(), nil
}

// ListBills lists the invoices of a window.
func ListBills(_ *cobra.Command, _ []string) {
	from, to, err := billingWindow(BillFrom, BillTo)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	query, err := billQuery(from, to, BillCategory, BillOrderID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	ids, err := httpLib.FetchArray(billsPath+query, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list invoices: %s", err)
		return
	}

	if len(ids) > billTooManyToExpand {
		display.OutputError(&flags.OutputFormatConfig,
			"%d invoices match, which is %d requests to detail them.\n   Narrow the window with --from and --to, or a category with --category.",
			len(ids), len(ids))
		return
	}

	bills, err := httpLib.FetchObjectsParallel[map[string]any](billsPath+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read invoices: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(bills))
	for _, bill := range bills {
		// FetchObjectsParallel preallocates one slot per id and only writes the
		// ones that succeeded, so with --ignore-errors the failures stay in the
		// slice as nil maps. Kept, they became a blank row in the table and a
		// bare {} under -o json — one per failed read, with nothing saying how
		// many were missing, which is worse than a short list.
		// FetchExpandedArray in the same package already drops them; these two
		// call sites bypass it because they need a query string and a count
		// first.
		if bill == nil {
			continue
		}

		rows = append(rows, billSecretsView(bill))
	}

	common.RenderFilteredTable(rows, billColumnsToDisplay)
}

// GetBill shows one invoice.
func GetBill(_ *cobra.Command, args []string) {
	bill, err := fetchBillingObject(billsPath, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read invoice %s: %s", args[0], err)
		return
	}

	display.OutputObject(billSecretsView(bill), args[0], billTemplate, &flags.OutputFormatConfig)
}

// ListBillDetails shows what one invoice charges for, line by line.
func ListBillDetails(_ *cobra.Command, args []string) {
	common.ManageListRequest(
		fmt.Sprintf("%s/%s/details", billsPath, url.PathEscape(args[0])),
		"",
		billDetailColumns,
		flags.GenericFilters,
	)
}

// ListRefunds lists the refunds of a window.
func ListRefunds(_ *cobra.Command, _ []string) {
	from, to, err := billingWindow(BillFrom, BillTo)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	values := url.Values{}
	values.Set("date.from", from)
	if to != "" {
		values.Set("date.to", to)
	}
	if BillOrderID != 0 {
		values.Set("orderId", fmt.Sprint(BillOrderID))
	}

	ids, err := httpLib.FetchArray(refundsPath+"?"+values.Encode(), "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list refunds: %s", err)
		return
	}

	refunds, err := httpLib.FetchObjectsParallel[map[string]any](refundsPath+"/%s", ids, flags.IgnoreErrors)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read refunds: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(refunds))
	for _, refund := range refunds {
		if refund == nil {
			continue
		}

		rows = append(rows, billSecretsView(refund))
	}

	common.RenderFilteredTable(rows, refundColumnsToDisplay)
}

// GetRefund shows one refund.
func GetRefund(_ *cobra.Command, args []string) {
	refund, err := fetchBillingObject(refundsPath, args[0])
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read refund %s: %s", args[0], err)
		return
	}

	display.OutputObject(billSecretsView(refund), args[0], refundTemplate, &flags.OutputFormatConfig)
}

// usageEntry is one service's usage over the running period.
type usageEntry struct {
	ServiceID int64  `json:"serviceId"`
	BeginDate string `json:"beginDate"`
	EndDate   string `json:"endDate"`
	Price     *struct {
		Text string `json:"text"`
	} `json:"price"`
	Elements []struct {
		PlanCode   string `json:"planCode"`
		PlanFamily string `json:"planFamily"`
		Quantity   any    `json:"quantity"`
		Price      *struct {
			Text string `json:"text"`
		} `json:"price"`
	} `json:"elements"`
}

// ShowUsage reads what is running against the current invoice, or what it is
// forecast to cost.
//
// One row per element rather than one per service: on the current period the
// entry-level price is null on every entry measured, and the amount lives in
// the elements. A table built on the entry would have shown a column of
// nothing. The forecast fills both, so the entry price is used as the fallback
// for a service that reports no element.
func ShowUsage(_ *cobra.Command, _ []string) {
	path := usagePath
	if BillUsageForecast {
		path = forecastPath
	}

	entries, err := httpLib.FetchArray(path, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read usage: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(entries))
	for _, raw := range entries {
		entry, err := decodeUsageEntry(raw)
		if err != nil {
			continue
		}

		period := day(entry.BeginDate) + " → " + day(entry.EndDate)

		if len(entry.Elements) == 0 {
			rows = append(rows, map[string]any{
				"serviceId": entry.ServiceID,
				"plan":      "—",
				"family":    "—",
				"quantity":  "—",
				"price":     priceText(entry.Price),
				"period":    period,
			})
			continue
		}

		for _, element := range entry.Elements {
			rows = append(rows, map[string]any{
				"serviceId": entry.ServiceID,
				"plan":      element.PlanCode,
				"family":    element.PlanFamily,
				"quantity":  element.Quantity,
				"price":     priceText(element.Price),
				"period":    period,
			})
		}
	}

	common.RenderFilteredTable(rows, usageColumns)
}

// decodeUsageEntry turns one raw entry into the shape this command reads.
func decodeUsageEntry(raw any) (usageEntry, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return usageEntry{}, err
	}

	var entry usageEntry
	if err := json.Unmarshal(encoded, &entry); err != nil {
		return usageEntry{}, err
	}

	return entry, nil
}

// priceText renders an amount that the API may not have filled in.
func priceText(price *struct {
	Text string `json:"text"`
}) string {
	if price == nil || price.Text == "" {
		return "—"
	}

	return price.Text
}

// day keeps the date and drops the time.
func day(timestamp string) string {
	if before, _, found := strings.Cut(timestamp, "T"); found {
		return before
	}

	return timestamp
}

// fetchBillingObject reads one invoice or refund.
func fetchBillingObject(collection, id string) (map[string]any, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("no identifier given")
	}

	objects, err := httpLib.FetchObjectsParallel[map[string]any](
		collection+"/%s", []any{id}, false)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("nothing returned")
	}

	return objects[0], nil
}

// billSecretsView replaces the download links and the PDF password with their
// fingerprints unless --reveal.
//
// These are not decorative. Measured on a real invoice: the pdfUrl returns the
// PDF in full with no API token whatsoever (HTTP 200, application/pdf), and the
// same URL with its esign parameter altered returns HTML instead. The link is
// therefore a bearer credential, and -o json would put it in a pipeline log.
// The substitution happens on the object and not in the template, for the same
// reason: a mask that only covers the human-readable output covers nothing.
func billSecretsView(object map[string]any) map[string]any {
	view := map[string]any{}
	for key, value := range object {
		view[key] = value
	}

	if RevealBillSecrets {
		return view
	}

	hidden := false
	for _, field := range []string{"password", "url", "pdfUrl"} {
		if value, set := view[field]; set && fmt.Sprint(value) != "" {
			view[field] = common.Fingerprint(fmt.Sprint(value))
			hidden = true
		}
	}

	if hidden {
		view["hidden"] = true
	}

	return view
}
