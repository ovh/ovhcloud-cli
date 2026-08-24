// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package backupservices

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// What the backup product costs is not in the backup API. The v2 resources
// carry no price, no billing date and no renewal mode; those live in the
// account's service router, keyed by the very same identifiers.
//
// Measured on 20 August 2026: /v1/services?resourceName=<uuid> answers with
// exactly one service for the backup tenant, for the VSPC tenant and for each
// vault, and each of those carries its plan code, its price, its period, its
// next billing date and its renewal mode. So the join is one lookup per
// resource rather than a sweep of the 826 services on this account.

// billedResource is one line of the billing view.
type billedResource struct {
	kind string
	id   string
	name string
}

// ShowBilling shows what every piece of the backup product costs.
func ShowBilling(_ *cobra.Command, _ []string) {
	tenant, err := ResolveTenant()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	resources, err := billableResources(tenant)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Consumption is one half of the answer and the prices are the other. Losing
	// the half that failed is better than answering nothing, so the table is
	// still printed.
	//
	// But it is printed as ONE document. This used to emit an OutputInfo and then
	// the table, which under -o json put two JSON documents on stdout back to
	// back — something no parser accepts, so a script asking for the prices got a
	// decode error precisely on the runs where half the data was missing. The
	// sentence goes to the log, which is stderr, and the fact goes where a script
	// will find it: in the consumption cell of every row.
	usage, err := currentUsage()
	if err != nil {
		log.Printf("🟠 Current consumption could not be read (%s); the prices below are still what is billed.", err)
		usage = nil
	}
	usageReadable := err == nil

	rows := make([]map[string]any, len(resources))
	group := new(errgroup.Group)
	group.SetLimit(10)

	for index, item := range resources {
		group.Go(func() error {
			row := map[string]any{
				"kind": item.kind,
				"id":   item.id,
				"name": item.name,
			}

			service, err := serviceOf(item.id)
			switch {
			case err != nil:
				// A resource with no billable service behind it is a fact, not
				// a failure of the command: it is what an included component
				// looks like.
				row["plan"] = "—"
				row["note"] = err.Error()

			default:
				billing := service.Billing
				row["serviceId"] = service.ServiceID
				row["plan"] = billing.Plan.Code
				row["invoiceName"] = billing.Plan.InvoiceName
				row["price"] = billing.Pricing.Price.Text
				row["priceValue"] = billing.Pricing.Price.Value
				row["currency"] = billing.Pricing.Price.CurrencyCode
				row["period"] = billing.Pricing.Duration
				row["nextBillingDate"] = billing.NextBillingDate
				row["renew"] = billing.Renew.Current.Mode
				row["state"] = billing.Lifecycle.Current.State
				row["consumption"] = usageOf(usage, service.ServiceID, usageReadable)
			}

			rows[index] = row

			return nil
		})
	}
	if err := group.Wait(); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	common.RenderFilteredTable(rows,
		[]string{"kind", "name", "plan", "price", "period", "consumption", "renew", "nextBillingDate", "state"})
}

// billableResources lists everything of the backup product that has a price.
func billableResources(tenant string) ([]billedResource, error) {
	resources := []billedResource{{kind: "tenant", id: tenant, name: tenant}}

	vspcs, err := listVspc(tenant)
	if err != nil {
		return nil, err
	}
	for _, vspc := range vspcs {
		resources = append(resources, billedResource{kind: "vspc", id: vspc.ID, name: vspc.name()})
	}

	var vaults []resource
	path := fmt.Sprintf("%s/%s/vault", TenantsPath, url.PathEscape(tenant))
	if err := httpLib.Client.Get(path, &vaults); err != nil {
		return nil, fmt.Errorf("failed to list the vaults of %s: %w", tenant, err)
	}
	for _, vault := range vaults {
		resources = append(resources, billedResource{kind: "vault", id: vault.ID, name: vault.name()})
	}

	return resources, nil
}

// service is the part of a service router entry this command reads.
type service struct {
	ServiceID int `json:"serviceId"`
	Billing   struct {
		NextBillingDate string `json:"nextBillingDate"`
		Plan            struct {
			Code        string `json:"code"`
			InvoiceName string `json:"invoiceName"`
		} `json:"plan"`
		Pricing struct {
			Duration string `json:"duration"`
			Price    struct {
				CurrencyCode string  `json:"currencyCode"`
				Text         string  `json:"text"`
				Value        float64 `json:"value"`
			} `json:"price"`
		} `json:"pricing"`
		Renew struct {
			Current struct {
				Mode string `json:"mode"`
			} `json:"current"`
		} `json:"renew"`
		Lifecycle struct {
			Current struct {
				State string `json:"state"`
			} `json:"current"`
		} `json:"lifecycle"`
	} `json:"billing"`
}

// serviceOf finds the billable service behind one backup identifier.
//
// The service router is queried by resource name rather than swept: this
// account carries 826 services and reading them all to find five would be a
// minute of requests for an answer one lookup gives.
func serviceOf(resourceName string) (service, error) {
	var ids []int

	path := "/v1/services?resourceName=" + url.QueryEscape(resourceName)
	if err := httpLib.Client.Get(path, &ids); err != nil {
		return service{}, fmt.Errorf("no billable service found for %s: %w", resourceName, err)
	}

	if len(ids) == 0 {
		return service{}, fmt.Errorf("no billable service is registered under %s", resourceName)
	}

	var found service
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/services/%d", ids[0]), &found); err != nil {
		return service{}, fmt.Errorf("failed to read service %d: %w", ids[0], err)
	}

	return found, nil
}

// usageEntry is one line of the account's current consumption.
type usageEntry struct {
	ServiceID int `json:"serviceId"`
	Elements  []struct {
		PlanCode string `json:"planCode"`
		Details  []struct {
			Quantity float64 `json:"quantity"`
			UniqueID string  `json:"uniqueId"`
		} `json:"details"`
	} `json:"elements"`
	Price *struct {
		Text string `json:"text"`
	} `json:"price"`
}

// currentUsage reads this month's consumption.
//
// A plain function, not a sync.OnceValues. It was one, and a once-value over a
// live API read is a process-wide cache over data that changes: the first call
// decided the answer for the life of the process, so a transient 500 on that
// first read made every later `backup-services billing` in the same process
// report the consumption as unreadable — for as long as the process lived. The
// once-values elsewhere in this repository all wrap the embedded schema, which
// cannot change; this one wraps an account.
//
// It is also the only reason a test could pass alone and fail in its suite, which
// is how this was found.
func currentUsage() ([]usageEntry, error) {
	var entries []usageEntry
	if err := httpLib.Client.Get("/v1/me/consumption/usage/current", &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// usageOf says what a service has consumed so far this month.
//
// Pay-as-you-go storage that nothing has written to yet reports no line at all,
// which is the state of every backup service on this account — all nine agents
// are NOT_INSTALLED, so nothing has been stored. "none yet" says that; an empty
// cell would read as a column that failed to fill.
// The last argument tells the two empty answers apart: consumption that could
// not be read, and consumption that is legitimately absent. "unknown" for both
// would have made a failed read indistinguishable from a service nobody has used
// yet — and the second is the normal state of a freshly ordered vault.
func usageOf(entries []usageEntry, serviceID int, readable bool) string {
	if !readable {
		return "unreadable"
	}
	if entries == nil {
		return "none yet"
	}

	var parts []string
	for _, entry := range entries {
		if entry.ServiceID != serviceID {
			continue
		}

		for _, element := range entry.Elements {
			quantity := 0.0
			for _, detail := range element.Details {
				quantity += detail.Quantity
			}
			parts = append(parts, fmt.Sprintf("%s×%g", element.PlanCode, quantity))
		}

		if entry.Price != nil && entry.Price.Text != "" {
			parts = append(parts, entry.Price.Text)
		}
	}

	if len(parts) == 0 {
		return "none yet"
	}

	sort.Strings(parts)

	return strings.Join(parts, ", ")
}
