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
	CatalogCommitment    string
	CatalogCountry       string
	CatalogAvailableOnly bool
	CatalogRefresh       bool
)

const (
	datacenterAvailabilitiesPath = "/dedicated/server/datacenter/availabilities"

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
	catalogCacheVersion = 3
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
	catalogCountries = sync.OnceValues(func() ([]string, error) {
		return openapi.GetParameterEnum(assets.BaremetalOpenapiSchema, "/dedicated/server/availabilities", "get", "country")
	})
)

func CompleteCatalogDatacenter(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return completeEnum(availabilityDatacenters)
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
	Recurring float64 `json:"recurring"`
	Interval  int     `json:"interval"`

	// Setup is the installation charge. It is a real charge, billed on top of
	// the first period rather than instead of it — see dueAtOrder below.
	Setup float64 `json:"setup"`

	// Promotion names the offer that lowered Setup or Recurring, empty when
	// none applies. A price that holds until the end of the month is not the
	// same fact as a price that holds, and an operator comparing providers
	// deserves to be told which one they are reading.
	Promotion string  `json:"promotion,omitempty"`
	ListPrice float64 `json:"listPrice,omitempty"`
}

type catalogPrices struct {
	Currency string                      `json:"currency"`
	Plans    map[string]map[string]price `json:"plans"`
	// Le nom sous lequel la gamme est vendue, par plan code. Le catalogue le
	// porte deja (`invoiceName`) : la table affichait « 24adv03-v3 » la ou une
	// facture, un devis et le site disent « ADVANCE-3 ». Un product manager
	// doit reconnaitre sa gamme avant de reconnaitre son code.
	Names map[string]string `json:"names"`
}

// commercialName rend « ADVANCE-3 2024 » depuis `invoiceName` et le plan code.
//
// `invoiceName` vaut « ADVANCE-3 | AMD EPYC 4464P » : le CPU est deja une
// colonne ailleurs, donc seule la partie de gauche sert. La casse n est PAS
// retouchee -- un Title-case naif ecrit « Rise-Xl » et « Rise-Gpu-1 », donc
// abime les sigles que la gamme porte vraiment.
//
// L annee vient des deux premiers chiffres du plan code (24adv03-v3 -> 2024).
// Elle desambigue : RISE-1 existe en 2024 et en 2025, sous deux codes.
// Mesure du 25/08 : 99 des 244 plans annonces par /availabilities portent un
// `invoiceName` -- le catalogue public eco ne couvre ni Scale, ni High Grade,
// ni SAP. Pour les 145 autres cette fonction rend "" et la colonne garde le
// plan code : inventer un nom serait pire que montrer le code.
func commercialName(planCode, invoiceName string) string {
	name := strings.TrimSpace(strings.SplitN(invoiceName, "|", 2)[0])
	if name == "" {
		return ""
	}
	if len(planCode) >= 2 && planCode[0] >= '0' && planCode[0] <= '9' &&
		planCode[1] >= '0' && planCode[1] <= '9' {
		return name + " 20" + planCode[:2]
	}
	return name
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
		//
		// Checked for that, and only for that. This used to be validated against
		// the schema's enum, which locks the command out of every account it was
		// meant to serve outside Europe: see checkSubsidiary.
		if err := checkSubsidiary(resolved); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"this account reports a subsidiary this command cannot use: %s", err)
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

	// Nothing narrowed the hardware yet: one plan code times its memory,
	// storage and datacenter combinations is a wall of rows a PM cannot read a
	// range out of. Show the range and its cheapest configuration first; the
	// same flags that narrow the API call also narrow this table, so asking
	// for --memory, --storage, --gpu or --plan-code is how the options show up.
	// -o json/yaml/custom keep every row: that reduction is for the table only.
	format := &flags.OutputFormatConfig
	narrowed := CatalogPlanCode != "" || CatalogMemory != "" || CatalogStorage != "" ||
		CatalogSystemStorage != "" || CatalogGPU != ""
	if !narrowed && format.CustomFormat() == "" && !format.IsInteractive() &&
		!format.IsYaml() && !format.IsJson() {
		filtered = defaultConfigPerRange(filtered)
		fmt.Println("💡 One row per range, its cheapest configuration — narrow with " +
			"--memory, --storage, --gpu or --plan-code to see the options.")
	}

	// Six columns, not eight, and the two widest in readable form: the table
	// was 225 characters across, so any terminal narrower than that wrapped
	// every line and the borders stopped lining up — which is exactly what
	// "les | font deriver l affichage" describes. `server` went because it is
	// the plan code minus its suffix, and `location` because the cheapest
	// configuration of a range says nothing about where the range is sold.
	// Both remain one -o json away.
	// Une gamme sans nom n est pas une gamme sans importance : Scale, High
	// Grade, HCI et SAP se vendent sur devis, et le catalogue public ne les
	// decrit pas. Le dire evite de lire le tiret comme un defaut d affichage.
	if unnamed := countUnnamed(filtered); unnamed > 0 && !format.IsJson() &&
		!format.IsYaml() && format.CustomFormat() == "" && !format.IsInteractive() {
		fmt.Printf("💡 %d of these rows carry no commercial name: the public catalogue "+
			"prices Eco ranges only, and the rest sell on quotation.\n", unnamed)
	}

	display.RenderTable(filtered,
		[]string{"name range", "planCode", "ram memory", "disks storage", "eta delivery",
			"monthly", "dueAtOrder"},
		&flags.OutputFormatConfig)
}

// countUnnamed compte les lignes que le catalogue public ne nomme pas.
func countUnnamed(rows []map[string]any) int {
	n := 0
	for _, row := range rows {
		if name, _ := row["name"].(string); name == "" || name == "-" {
			n++
		}
	}
	return n
}

// defaultConfigPerRange keeps one row per range ("server"): its cheapest
// priced configuration, or the first seen when none is priced (Scale, High
// Grade and SAP sell on quotation only). That is the first thing a PM needs —
// which ranges exist and what they start at — before narrowing to the RAM,
// storage or GPU options each one takes.
func defaultConfigPerRange(rows []map[string]any) []map[string]any {
	best := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		server, _ := row["server"].(string)
		if current, seen := best[server]; !seen || cheaperDefault(row, current) {
			best[server] = row
		}
	}

	winners := make([]map[string]any, 0, len(best))
	for _, row := range best {
		winners = append(winners, row)
	}
	// Same comparator as buildCatalogRows: soonest delivery first. Sorting the
	// reduced set by range name instead would drop the one thing an operator
	// reads this table for, just for the accounts with more than one range.
	sort.SliceStable(winners, func(i, j int) bool {
		left, _ := winners[i]["deliveryHours"].(int)
		right, _ := winners[j]["deliveryHours"].(int)
		if left != right {
			return left < right
		}
		leftCode, _ := winners[i]["planCode"].(string)
		rightCode, _ := winners[j]["planCode"].(string)
		return leftCode < rightCode
	})
	return winners
}

// cheaperDefault says whether candidate is the better range default than
// current: priced beats unpriced, then the lower monthly price, then the
// shorter delivery.
func cheaperDefault(candidate, current map[string]any) bool {
	candidatePrice, candidatePriced := candidate["monthlyValue"].(float64)
	currentPrice, currentPriced := current["monthlyValue"].(float64)
	if candidatePriced != currentPriced {
		return candidatePriced
	}
	if candidatePriced && candidatePrice != currentPrice {
		return candidatePrice < currentPrice
	}
	candidateHours, _ := candidate["deliveryHours"].(int)
	currentHours, _ := current["deliveryHours"].(int)
	return candidateHours < currentHours
}

func checkEnumFlags() error {
	if err := checkEnumValues("datacenter", CatalogDatacenters, availabilityDatacenters); err != nil {
		return err
	}
	if CatalogCountry != "" {
		return checkSubsidiary(CatalogCountry)
	}
	return nil
}

// checkSubsidiary accepts anything that can safely name a cache file, and
// leaves the question of which subsidiaries exist to the API.
//
// The enum this replaces came from the embedded schema, which is a snapshot of
// the EU API — the Makefile can only download from there. Measured against both
// endpoints, the accepted set is a property of the endpoint, not of the product:
//
//	eu.api.ovh.com  accepts CZ DE ES FI FR GB IE IT LT MA NL PL PT SN TN
//	ca.api.ovh.com  accepts CA QC ASIA AU SG IN WE WS
//
// So the enum was wrong in both directions. It carries EU, which the EU endpoint
// answers 400 to, and it carries none of the eight the CA endpoint accepts — so
// on a CA-configured profile the resolved subsidiary was refused locally, and
// --country, the way out, was refused by the same list. The command was
// unreachable for those accounts, which no amount of reading a better enum
// fixes: there is no local list that is right for both endpoints.
//
// An unknown value now costs one request and comes back as
// "invalid ovhSubsidiary", which names the problem. A wrong local gate costs the
// command.
func checkSubsidiary(value string) error {
	if value == "" {
		return fmt.Errorf("the subsidiary is empty")
	}
	// Two to eight ASCII letters covers every subsidiary either endpoint
	// accepts, and excludes anything that could walk out of the cache directory.
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return fmt.Errorf("%q is not a subsidiary code (letters only)", value)
		}
	}
	if len(value) < 2 || len(value) > 8 {
		return fmt.Errorf("%q is not a subsidiary code (two to eight letters)", value)
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
	if len(CatalogDatacenters) > 0 || CatalogPlanCode != "" || CatalogServer != "" {
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
}

// fetchAvailabilities asks the API for what it can filter, rather than
// downloading four megabytes and sieving them here: one plan code narrows the
// answer from 8973 entries to a handful.
func fetchAvailabilities() ([]availabilityEntry, error) {
	query := url.Values{}

	if len(CatalogDatacenters) > 0 {
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

	endpoint := "/v1" + datacenterAvailabilitiesPath
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
			// Le nom commercial, celui que porte la facture.
			InvoiceName string `json:"invoiceName"`
			Pricings []struct {
				Mode         string   `json:"mode"`
				Price        float64  `json:"price"`
				Interval     int      `json:"interval"`
				IntervalUnit string   `json:"intervalUnit"`
				Capacities   []string `json:"capacities"`
				Promotions   []struct {
					Name string `json:"name"`
					// Total is what the charge becomes once the promotion is
					// applied, in ucents like every other amount here. It is
					// read rather than recomputed from `discount`, because a
					// percentage rounded twice does not land on the cent the
					// checkout will charge.
					Total struct {
						Value float64 `json:"value"`
					} `json:"total"`
				} `json:"promotions"`
			} `json:"pricings"`
		} `json:"plans"`
	}

	endpoint := fmt.Sprintf("/v1/order/catalog/public/%s?ovhSubsidiary=%s",
		baremetalCatalogProduct, url.QueryEscape(country))
	if err := httpLib.Client.Get(endpoint, &raw); err != nil {
		return catalogPrices{}, err
	}

	prices := catalogPrices{Currency: raw.Locale.CurrencyCode,
		Plans: map[string]map[string]price{}, Names: map[string]string{}}
	for _, plan := range raw.Plans {
		// Le nom se retient meme quand le plan n a pas de prix dans ce mode :
		// une gamme vendue sur devis a droit a son nom comme les autres.
		if name := commercialName(plan.PlanCode, plan.InvoiceName); name != "" {
			prices.Names[plan.PlanCode] = name
		}
		byMode := map[string]price{}
		for _, pricing := range plan.Pricings {
			// The catalogue states prices in ucents: 8999000000 is 89.99. A
			// promotion replaces the charge outright rather than annotating it,
			// so the amount to keep is the promoted one — the checkout bills
			// that, and reporting the crossed-out price would overstate every
			// offer currently on sale.
			listed := pricing.Price / 1e8
			amount := listed
			promotion := ""
			if len(pricing.Promotions) > 0 {
				amount = pricing.Promotions[0].Total.Value / 1e8
				promotion = pricing.Promotions[0].Name
			}

			entry := byMode[pricing.Mode]
			switch {
			case containsString(pricing.Capacities, "renew"):
				entry.Recurring = amount
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
				entry.Setup = amount
			default:
				// A capacity this command does not price. Counting it into the
				// list price would make the crossed-out figure disagree with
				// the checkout's own.
				continue
			}
			entry.ListPrice += listed
			if promotion != "" {
				entry.Promotion = promotion
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
		locations := make([][2]string, 0, len(offer.Datacenters))
		for _, datacenter := range offer.Datacenters {
			locations = append(locations, [2]string{datacenter.Datacenter, datacenter.Availability})
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

				// Readable forms of the two widest columns. The raw addon
				// codes stay in the row: -o json still gives ram-128g-ecc-4800
				// and softraid-2x1920nvme, which is what a script matches on.
				"ram":   humaniseMemory(offer.Memory),
				"disks": humaniseStorage(offer.Storage),
				"eta":   shortDelay(location[1]),
			}

			// Le nom d abord, le code ensuite : le code reste ce qu on passe a
			// `baremetal order` et a --plan-code, donc il ne disparait pas -- il
			// cesse seulement d etre la seule facon de reconnaitre une gamme.
			if name, named := prices.Names[offer.PlanCode]; named {
				row["name"] = name
			} else {
				row["name"] = "-"
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

// humaniseMemory turns an addon code into what the row is actually offering.
//
// Every code in the catalogue follows one shape, ram-<size>g-<qualifiers>, so
// the size is the only part a reader needs: ram-128g-on-die-ecc-3600 and
// ram-128g-ecc-4800 are both "128 GB" as far as choosing a range goes, and the
// qualifiers cost 20 characters of table width to say what -o json says
// better. An unrecognised code is returned untouched rather than mangled.
func humaniseMemory(code string) string {
	rest, ok := strings.CutPrefix(code, "ram-")
	if !ok {
		return code
	}
	size, _, _ := strings.Cut(rest, "-")
	gigabytes, err := strconv.Atoi(strings.TrimSuffix(size, "g"))
	if err != nil {
		return code
	}
	if gigabytes >= 1024 && gigabytes%1024 == 0 {
		return fmt.Sprintf("%d TB", gigabytes/1024)
	}
	return fmt.Sprintf("%d GB", gigabytes)
}

// humaniseStorage does the same for the widest column of the table.
//
// softraid-2x1920nvme becomes "2x1.92 TB NVMe"; a hybrid layout keeps both of
// its groups. noraid-0 means the range carries no disk of its own — every one
// of them is sold on quotation — and an empty-looking cell reads as a bug, so
// it says so.
func humaniseStorage(code string) string {
	if code == "" {
		return ""
	}
	if strings.HasPrefix(code, "noraid") {
		return "on request"
	}

	rest := code
	for _, prefix := range []string{"hybridsoftraid-", "softraid-", "hardraid-", "raid-"} {
		if trimmed, ok := strings.CutPrefix(rest, prefix); ok {
			rest = trimmed
			break
		}
	}

	var groups []string
	for _, part := range strings.Split(rest, "-") {
		count, size, ok := strings.Cut(part, "x")
		if !ok {
			continue
		}
		if _, err := strconv.Atoi(count); err != nil {
			continue
		}
		digits := strings.TrimRight(size, "abcdefghijklmnopqrstuvwxyz")
		kind := strings.TrimPrefix(size, digits)
		gigabytes, err := strconv.Atoi(digits)
		if err != nil {
			continue
		}
		groups = append(groups, strings.TrimSpace(fmt.Sprintf("%sx%s %s", count,
			humaniseSize(gigabytes), diskKind(kind))))
	}
	if len(groups) == 0 {
		return code
	}
	return strings.Join(groups, " + ")
}

// humaniseSize keeps the number short without rounding it away: 1920 GB is
// 1.92 TB, and 960 GB stays 960 GB rather than becoming 0.96 TB.
func humaniseSize(gigabytes int) string {
	if gigabytes < 1000 {
		return fmt.Sprintf("%d GB", gigabytes)
	}
	terabytes := float64(gigabytes) / 1000
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", terabytes), "0"), ".") + " TB"
}

// diskKind spells out the suffixes the catalogue uses for the medium, in the
// casing the products are actually sold under.
func diskKind(suffix string) string {
	switch suffix {
	case "sa", "sata":
		return "SATA"
	case "ssd":
		return "SSD"
	case "nvme":
		return "NVMe"
	case "":
		return ""
	default:
		return strings.ToUpper(suffix)
	}
}

// shortDelay is humaniseDelay in the width a table can afford.
//
// The long form is what a sentence wants ("within the hour", "within 72
// hours"); a column wants "< 1 h". The stock warning is the one thing that
// cannot be dropped in the short form either — it is the only signal the API
// gives that the machine may be gone before the order lands.
func shortDelay(availability string) string {
	switch availability {
	case "comingSoon":
		return "soon"
	case "unavailable":
		return "—"
	case "", "unknown":
		return "?"
	}

	hours := deliveryHours(availability)
	suffix := ""
	if strings.HasSuffix(availability, "-low") {
		suffix = " (low)"
	}
	switch {
	case hours <= 1:
		return "< 1 h" + suffix
	case hours < 24:
		return fmt.Sprintf("%d h%s", hours, suffix)
	default:
		return fmt.Sprintf("%d d%s", hours/24, suffix)
	}
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

	// What the operator actually pays on the day they order: the installation
	// charge PLUS the first period. Both, not either.
	//
	// An earlier version of this reported the installation charge alone,
	// reasoning that since it equals the monthly price on every one of the 99
	// plans, it must BE the first month. Two numbers being equal is not a
	// reason, and the order API disagrees: a cart holding one 24adv01-v3 at
	// 89.99 a month quotes 179.98, itemised as one INSTALLATION line and one
	// DURATION line. The rule below was then checked against seven live quotes
	// — promoted and not, monthly and committed — and matched each to the cent.
	// Understating the price by half is the failure this command exists to
	// prevent, so it is worth the seven carts.
	//
	// On a commitment the installation charge is zero and the upfront payment
	// is the first renewal, so the same sum still holds.
	dueAtOrder := entry.Setup + entry.Recurring

	row["monthly"] = formatAmount(entry.Recurring/float64(months), prices.Currency)
	row["monthlyValue"] = entry.Recurring / float64(months)
	row["dueAtOrder"] = formatAmount(dueAtOrder, prices.Currency)
	row["dueAtOrderValue"] = dueAtOrder
	row["setup"] = formatAmount(entry.Setup, prices.Currency)
	row["recurring"] = formatAmount(entry.Recurring, prices.Currency)
	row["recurringMonths"] = months

	// A promotion is a fact with an expiry date, and the table shows the
	// promoted figure. Saying so — and saying what it was crossed out from —
	// is the difference between a price an operator can quote to their own
	// management and a price they have to go and check.
	if entry.Promotion != "" {
		row["promotion"] = entry.Promotion
		row["listPrice"] = formatAmount(entry.ListPrice, prices.Currency)
		// "(promo, was 129.98 EUR)" spent 24 characters of a table that did not
		// have them; the currency is already in the figure beside it.
		row["dueAtOrder"] = fmt.Sprintf("%s (was %s)",
			row["dueAtOrder"], formatAmount(entry.ListPrice, prices.Currency))
	}
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
