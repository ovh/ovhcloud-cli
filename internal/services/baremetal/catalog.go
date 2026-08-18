// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/cache"
	"github.com/ovh/ovhcloud-cli/internal/display"
	filtersLib "github.com/ovh/ovhcloud-cli/internal/filters"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/spf13/cobra"
)

// Flags of `catalog`.
var (
	CatalogPlanCode      string
	CatalogServer        string
	CatalogMemory        string
	CatalogStorage       string
	CatalogSystemStorage string
	CatalogGPU           string
	CatalogDatacenters   []string
	CatalogRegions       []string
	CatalogCommitment    string
	CatalogCountry       string
	CatalogAvailableOnly bool
	CatalogRefresh       bool
)

const (
	datacenterAvailabilitiesPath = "/dedicated/server/datacenter/availabilities"
	regionAvailabilitiesPath     = "/dedicated/server/region/availabilities"

	// The public catalogue is fetched whole — it takes no filter — and weighs
	// about twelve megabytes for a hundred plans. What is kept here is the
	// distilled price table, a few kilobytes, so a repeat call reads that
	// instead of the download. Prices move about once a month; stock moves by
	// the hour, which is why availability is never cached.
	catalogCacheNamespace = "catalog"
	catalogCacheTTL       = 24 * time.Hour

	// The only public catalogue carrying dedicated servers. It is not just the
	// entry range despite the name: Rise, Advance, Kimsufi and So you Start are
	// all in it. Scale, High Grade and SAP are not sold from a public price
	// list at all.
	baremetalCatalogProduct = "eco"

	// Bumped whenever catalogPrices changes shape.
	catalogCacheVersion = 1
)

// commitmentModes maps what an operator asks for to what the catalogue calls it.
var commitmentModes = map[string]string{
	"default": "default",
	"12":      "upfront12",
	"24":      "upfront24",
}

var (
	availabilityDatacenters = sync.OnceValues(func() ([]string, error) {
		return openapi.GetComponentEnum(assets.BaremetalOpenapiSchema, "dedicated.AvailabilityDatacenterEnum")
	})
	availabilityRegions = sync.OnceValues(func() ([]string, error) {
		return openapi.GetParameterEnum(assets.BaremetalOpenapiSchema, regionAvailabilitiesPath, "get", "regions")
	})
	catalogCountries = sync.OnceValues(func() ([]string, error) {
		return openapi.GetParameterEnum(assets.BaremetalOpenapiSchema, "/dedicated/server/availabilities", "get", "country")
	})
)

func CompleteCatalogDatacenter(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(availabilityDatacenters)
}

func CompleteCatalogRegion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(availabilityRegions)
}

func CompleteCatalogCountry(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(catalogCountries)
}

func CompleteCatalogCommitment(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"default", "12", "24"}, cobra.ShellCompDirectiveNoFileComp
}

// price is what one plan costs under one commitment.
type price struct {
	// Recurring is the amount billed at every renewal, and Interval how many
	// months that covers. A twelve-month commitment renews once a year, so
	// comparing it to a monthly price means dividing.
	Recurring  float64 `json:"recurring"`
	Interval   int     `json:"interval"`
	DueAtOrder float64 `json:"dueAtOrder"`
}

type catalogPrices struct {
	Currency string                      `json:"currency"`
	Plans    map[string]map[string]price `json:"plans"`
}

// GetBaremetalCatalog lists what can be ordered, where it can be delivered, how
// long delivery takes and what it costs.
func GetBaremetalCatalog(_ *cobra.Command, _ []string) {
	if _, found := commitmentModes[CatalogCommitment]; !found {
		display.OutputError(&flags.OutputFormatConfig,
			"--commitment does not accept %q; accepted values are: default, 12, 24", CatalogCommitment)
		return
	}

	if err := checkEnumFlags(); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	country := CatalogCountry
	if country == "" {
		resolved, err := accountSubsidiary()
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"could not read the subsidiary of this account (%s); name it with --country", err)
			return
		}
		// Checked like an operator-supplied value, because it goes on to name a
		// file in the cache directory. It arrives from the API rather than from
		// a keyboard, which makes it trusted by habit rather than by argument.
		if err := checkEnumFlag("country", resolved, catalogCountries); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"this account reports an unknown subsidiary: %s", err)
			return
		}
		country = resolved
	}

	// These two calls used to run concurrently, which looked free: the catalogue
	// takes seconds and the availabilities under one. The race detector says
	// otherwise — go-ovh writes to the shared http.Client on every request
	// (`c.Client.Timeout = c.Timeout`, ovh.go:387), so two requests through one
	// client race on it. The parallelism also bought almost nothing: the
	// catalogue is cached, so from the second run of the day there is only one
	// call left to make.
	offers, err := fetchAvailabilities()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch availabilities: %s", err)
		return
	}

	prices, err := fetchCatalogPrices(country)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to fetch the %s catalogue: %s", country, err)
		return
	}

	rows := buildCatalogRows(offers, prices, commitmentModes[CatalogCommitment])

	if len(rows) == 0 {
		// A bare empty table reads as a broken command. Say which of the two
		// plausible reasons it is.
		display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"offers": []any{}},
			"No offer matches. %s", emptyReason())
		return
	}

	filtered, err := filtersLib.FilterLines(rows, flags.GenericFilters)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to filter results: %s", err)
		return
	}

	display.RenderTable(filtered,
		[]string{"planCode", "server", "memory", "storage", "location", "delivery", "monthly", "dueAtOrder"},
		&flags.OutputFormatConfig)
}

func checkEnumFlags() error {
	if err := checkEnumValues("datacenter", CatalogDatacenters, availabilityDatacenters); err != nil {
		return err
	}
	if err := checkEnumValues("region", CatalogRegions, availabilityRegions); err != nil {
		return err
	}
	if CatalogCountry != "" {
		return checkEnumFlag("country", CatalogCountry, catalogCountries)
	}
	return nil
}

func checkEnumValues(name string, values []string, read func() ([]string, error)) error {
	for _, value := range values {
		if err := checkEnumFlag(name, value, read); err != nil {
			return err
		}
	}
	return nil
}

func emptyReason() string {
	if len(CatalogDatacenters) > 0 || len(CatalogRegions) > 0 || CatalogPlanCode != "" || CatalogServer != "" {
		return "Widen the filters, or drop --available-only if it is set."
	}
	return "This is unexpected with no filter set: the catalogue should never be empty."
}

// availabilityEntry is one hardware configuration and where it can be delivered.
type availabilityEntry struct {
	PlanCode      string `json:"planCode"`
	Server        string `json:"server"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
	SystemStorage string `json:"systemStorage"`
	GPU           string `json:"gpu"`
	FQN           string `json:"fqn"`

	Datacenters []struct {
		Datacenter   string `json:"datacenter"`
		Availability string `json:"availability"`
	} `json:"datacenters"`

	Regions []struct {
		Region       string `json:"region"`
		Availability string `json:"availability"`
	} `json:"regions"`
}

// fetchAvailabilities asks the API for what it can filter, rather than
// downloading four megabytes and sieving them here: one plan code narrows the
// answer from 8973 entries to a handful.
func fetchAvailabilities() ([]availabilityEntry, error) {
	path := datacenterAvailabilitiesPath
	query := url.Values{}

	if len(CatalogRegions) > 0 {
		path = regionAvailabilitiesPath
		for _, region := range CatalogRegions {
			query.Add("regions", region)
		}
	} else if len(CatalogDatacenters) > 0 {
		query.Set("datacenters", strings.Join(CatalogDatacenters, ","))
	}

	for name, value := range map[string]string{
		"planCode":      CatalogPlanCode,
		"server":        CatalogServer,
		"memory":        CatalogMemory,
		"storage":       CatalogStorage,
		"systemStorage": CatalogSystemStorage,
		"gpu":           CatalogGPU,
	} {
		if value != "" {
			query.Set(name, value)
		}
	}

	endpoint := "/v1" + path
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	var entries []availabilityEntry
	if err := httpLib.Client.Get(endpoint, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// fetchCatalogPrices returns the price table for a subsidiary, from the cache
// when it is fresh enough.
func fetchCatalogPrices(country string) (catalogPrices, error) {
	// The version travels in the name: an entry written by an older build has a
	// different shape, and one that still unmarshals into the current struct
	// would be read as fresh while missing half its fields.
	key := fmt.Sprintf("eco-%s-v%d.json", strings.ToLower(country), catalogCacheVersion)

	if CatalogRefresh {
		cache.Write(catalogCacheNamespace, key, nil, 0)
	} else if data, found := cache.Read(catalogCacheNamespace, key, catalogCacheTTL); found {
		var cached catalogPrices
		if err := json.Unmarshal(data, &cached); err == nil && len(cached.Plans) > 0 {
			return cached, nil
		}
		// A cache entry that no longer parses is a cache entry from an older
		// version of this command: fall through and fetch.
	}

	var raw struct {
		Locale struct {
			CurrencyCode string `json:"currencyCode"`
		} `json:"locale"`
		Plans []struct {
			PlanCode string `json:"planCode"`
			Pricings []struct {
				Mode         string   `json:"mode"`
				Price        float64  `json:"price"`
				Interval     int      `json:"interval"`
				IntervalUnit string   `json:"intervalUnit"`
				Capacities   []string `json:"capacities"`
			} `json:"pricings"`
		} `json:"plans"`
	}

	endpoint := fmt.Sprintf("/v1/order/catalog/public/%s?ovhSubsidiary=%s",
		baremetalCatalogProduct, url.QueryEscape(country))
	if err := httpLib.Client.Get(endpoint, &raw); err != nil {
		return catalogPrices{}, err
	}

	prices := catalogPrices{Currency: raw.Locale.CurrencyCode, Plans: map[string]map[string]price{}}
	for _, plan := range raw.Plans {
		byMode := map[string]price{}
		for _, pricing := range plan.Pricings {
			entry := byMode[pricing.Mode]
			switch {
			case containsString(pricing.Capacities, "renew"):
				// The catalogue states prices in ucents: 8999000000 is 89.99.
				entry.Recurring = pricing.Price / 1e8
				// Every interval on this catalogue is counted in months today.
				// Dividing a period priced in days by its number as if they
				// were months would understate the offer without a word, so an
				// unexpected unit is treated as a single period instead.
				if pricing.IntervalUnit == "month" {
					entry.Interval = pricing.Interval
				} else {
					entry.Interval = 1
				}
			case containsString(pricing.Capacities, "installation"):
				entry.DueAtOrder = pricing.Price / 1e8
			}
			byMode[pricing.Mode] = entry
		}
		if len(byMode) > 0 {
			prices.Plans[plan.PlanCode] = byMode
		}
	}

	if encoded, err := json.Marshal(prices); err == nil {
		cache.Write(catalogCacheNamespace, key, encoded, catalogCacheTTL)
	}

	return prices, nil
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// buildCatalogRows flattens one entry per configuration and location, because
// the same plan is delivered at different speeds in different places, and that
// difference is the reason to run this command.
func buildCatalogRows(offers []availabilityEntry, prices catalogPrices, mode string) []map[string]any {
	rows := make([]map[string]any, 0, len(offers))

	for _, offer := range offers {
		locations := make([][2]string, 0, len(offer.Datacenters)+len(offer.Regions))
		for _, datacenter := range offer.Datacenters {
			locations = append(locations, [2]string{datacenter.Datacenter, datacenter.Availability})
		}
		for _, region := range offer.Regions {
			locations = append(locations, [2]string{region.Region, region.Availability})
		}

		for _, location := range locations {
			if CatalogAvailableOnly && !isDeliverable(location[1]) {
				continue
			}

			row := map[string]any{
				"planCode":      offer.PlanCode,
				"server":        offer.Server,
				"memory":        offer.Memory,
				"storage":       offer.Storage,
				"systemStorage": offer.SystemStorage,
				"gpu":           offer.GPU,
				"fqn":           offer.FQN,
				"location":      location[0],
				"availability":  location[1],
				"delivery":      humaniseDelay(location[1]),
				"deliveryHours": deliveryHours(location[1]),
			}

			applyPrice(row, prices, offer.PlanCode, mode)
			rows = append(rows, row)
		}
	}

	// Soonest first. Sorting on the raw code would put 1440H before 24H, which
	// is the opposite of what the column means.
	sort.SliceStable(rows, func(i, j int) bool {
		// A comparator that panics takes the whole command with it, and every
		// row here is built a few lines above: the checked assertion costs
		// nothing and turns a future mistake into a wrong order rather than a
		// crash.
		left, _ := rows[i]["deliveryHours"].(int)
		right, _ := rows[j]["deliveryHours"].(int)
		if left != right {
			return left < right
		}
		leftCode, _ := rows[i]["planCode"].(string)
		rightCode, _ := rows[j]["planCode"].(string)
		return leftCode < rightCode
	})

	return rows
}

// applyPrice attaches what a plan costs, or says why it has no price. A blank
// cell reads as a bug; a hundred plans are priced publicly and a hundred and
// forty-five — Scale, High Grade, SAP — are sold on quotation.
func applyPrice(row map[string]any, prices catalogPrices, planCode, mode string) {
	byMode, priced := prices.Plans[planCode]
	if !priced {
		row["monthly"] = "on quotation"
		row["dueAtOrder"] = "on quotation"
		return
	}

	entry, hasMode := byMode[mode]
	if !hasMode {
		row["monthly"] = "not sold with this commitment"
		row["dueAtOrder"] = "-"
		return
	}

	months := entry.Interval
	if months < 1 {
		months = 1
	}

	// What the operator actually pays on the day they order.
	//
	// The catalogue prices one charge as `installation` and one as `renew`. On
	// the monthly offer they are equal — the installation charge IS the first
	// month, which is why this column is not headed "setup fee": there is no
	// setup fee, and adding it to the monthly price would report twice the
	// truth. On a commitment the installation charge is zero, because the
	// upfront payment is the first renewal: reporting that zero as the amount
	// due would say a twelve-month commitment costs nothing to start.
	dueAtOrder := entry.DueAtOrder
	if dueAtOrder == 0 && entry.Recurring > 0 {
		dueAtOrder = entry.Recurring
	}

	row["monthly"] = formatAmount(entry.Recurring/float64(months), prices.Currency)
	row["monthlyValue"] = entry.Recurring / float64(months)
	row["dueAtOrder"] = formatAmount(dueAtOrder, prices.Currency)
	row["recurring"] = formatAmount(entry.Recurring, prices.Currency)
	row["recurringMonths"] = months
}

func formatAmount(amount float64, currency string) string {
	if currency == "" {
		return strconv.FormatFloat(amount, 'f', 2, 64)
	}
	return fmt.Sprintf("%.2f %s", amount, currency)
}

// isDeliverable says whether an availability code means a machine can actually
// be ordered now.
func isDeliverable(availability string) bool {
	switch availability {
	case "unavailable", "comingSoon", "unknown", "":
		return false
	}
	return true
}

// deliveryHours turns an availability code into hours, for sorting. Codes that
// are not a delay sort last, in the order an operator would rank them.
func deliveryHours(availability string) int {
	switch availability {
	case "comingSoon":
		return 1 << 28
	case "unavailable":
		return 1 << 29
	case "", "unknown":
		return 1 << 30
	}

	// Codes look like 24H, 240H, or 1H-low and 1H-high for the two grades of
	// same-hour delivery.
	digits := strings.TrimSuffix(strings.SplitN(availability, "-", 2)[0], "H")
	hours, err := strconv.Atoi(digits)
	if err != nil {
		return 1 << 30
	}
	if strings.HasSuffix(availability, "-low") {
		// Same delay, less stock: rank it just after its -high sibling.
		hours = hours*2 + 1
	}
	return hours
}

// humaniseDelay renders the delay an operator is really being told about. The
// raw codes count in hours up to 2160, and reading "2160H" as three months is
// arithmetic nobody should do at a prompt.
func humaniseDelay(availability string) string {
	switch availability {
	case "comingSoon":
		return "coming soon"
	case "unavailable":
		return "unavailable"
	case "", "unknown":
		return "unknown"
	}

	hours := deliveryHours(availability)
	// "-low" and "-high" are the same delay with different stock behind it.
	// Rendering both as "within the hour" throws away the only warning the API
	// gives that the machine may be gone by the time the order goes through.
	suffix := ""
	if strings.HasSuffix(availability, "-low") {
		suffix = " (low stock)"
	}

	switch {
	case hours <= 3:
		return "within the hour" + suffix
	case hours < 48:
		return fmt.Sprintf("%dh", hours) + suffix
	case hours < 720:
		return fmt.Sprintf("%dd", hours/24) + suffix
	default:
		return fmt.Sprintf("%dmo", hours/720) + suffix
	}
}

// accountSubsidiary reads which price list this account buys from, so the
// operator does not have to name it.
func accountSubsidiary() (string, error) {
	var me struct {
		OvhSubsidiary string `json:"ovhSubsidiary"`
	}
	if err := httpLib.Client.Get("/v1/me", &me); err != nil {
		return "", err
	}
	if me.OvhSubsidiary == "" {
		return "", fmt.Errorf("the account declares no subsidiary")
	}
	return me.OvhSubsidiary, nil
}
