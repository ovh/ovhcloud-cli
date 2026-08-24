// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	_ "embed"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// dedicatedServerRoute is how the billing catalogue names a dedicated server.
// A server resolves to four or five billable services — itself plus the
// components sold with it — and only the machine carries this route. The
// others carry none, so this is what separates the server from its options.
const dedicatedServerRoute = "/dedicated/server/{serviceName}"

//go:embed templates/cost.tmpl
var costTemplate string

// billableService is the part of /v1/services this command reads.
type billableService struct {
	ServiceID       int64  `json:"serviceId"`
	ParentServiceID *int64 `json:"parentServiceId"`
	Route           *struct {
		Path string `json:"path"`
	} `json:"route"`
	Resource struct {
		DisplayName string `json:"displayName"`
		State       string `json:"state"`
		Product     struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"product"`
	} `json:"resource"`
	Billing struct {
		NextBillingDate string `json:"nextBillingDate"`
		ExpirationDate  string `json:"expirationDate"`
		Plan            struct {
			Code        string `json:"code"`
			InvoiceName string `json:"invoiceName"`
		} `json:"plan"`
		Pricing struct {
			Description string `json:"description"`
			Duration    string `json:"duration"`
			Price       struct {
				CurrencyCode string  `json:"currencyCode"`
				Text         string  `json:"text"`
				Value        float64 `json:"value"`
			} `json:"price"`
		} `json:"pricing"`
		Renew *struct {
			Current *struct {
				Mode     string `json:"mode"`
				NextDate string `json:"nextDate"`
				Period   string `json:"period"`
			} `json:"current"`
		} `json:"renew"`
		Engagement *struct {
			EndDate string `json:"endDate"`
		} `json:"engagement"`
	} `json:"billing"`
}

// isTheMachine says whether this service is the server rather than one of the
// components billed alongside it.
func (s billableService) isTheMachine() bool {
	return s.Route != nil && s.Route.Path == dedicatedServerRoute
}

// label names a line the way an invoice would.
func (s billableService) label() string {
	if s.Billing.Plan.InvoiceName != "" {
		return s.Billing.Plan.InvoiceName
	}
	if s.Resource.Product.Description != "" {
		return s.Resource.Product.Description
	}

	return s.Billing.Plan.Code
}

// ShowBaremetalCost says what a server costs, what is included in that price,
// and when it renews.
func ShowBaremetalCost(_ *cobra.Command, args []string) {
	server := args[0]

	services, err := servicesOf(server)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	var machine *billableService
	var included []billableService
	var total float64
	currency := ""

	for _, service := range services {
		if service.isTheMachine() {
			copied := service
			machine = &copied
		} else {
			included = append(included, service)
		}
		total += service.Billing.Pricing.Price.Value
		if service.Billing.Pricing.Price.CurrencyCode != "" {
			currency = service.Billing.Pricing.Price.CurrencyCode
		}
	}

	if machine == nil {
		display.OutputError(&flags.OutputFormatConfig,
			"%d billable service(s) answer for %s, none of which is the machine itself.\n   The billing catalogue reached them, but none carries the route %s.",
			len(services), server, dedicatedServerRoute)
		return
	}

	sort.Slice(included, func(i, j int) bool {
		return included[i].label() < included[j].label()
	})

	lines := make([]map[string]any, 0, len(services))
	lines = append(lines, costRow(*machine, "server"))
	for _, service := range included {
		lines = append(lines, costRow(service, "included"))
	}

	// One rendering, not two: a table followed by a summary would put two
	// documents on one stdout, and -o json would carry only the last.
	display.OutputObject(map[string]any{
		"server":      server,
		"plan":        machine.label(),
		"total":       money(total, currency),
		"totalValue":  total,
		"currency":    currency,
		"services":    len(services),
		"renewal":     renewalPhrase(*machine),
		"nextBilling": machine.Billing.NextBillingDate,
		"expiration":  machine.Billing.ExpirationDate,
		"lines":       lines,
	}, server, costTemplate, &flags.OutputFormatConfig)
}

// costRow is one line of the table.
func costRow(service billableService, kind string) map[string]any {
	price := service.Billing.Pricing.Price.Text
	if price == "" {
		price = "—"
	}

	return map[string]any{
		"item":      service.label(),
		"kind":      kind,
		"price":     price,
		"billed":    service.Billing.Pricing.Description,
		"serviceId": service.ServiceID,
	}
}

// renewalPhrase says how the machine renews, and says it differently when the
// renewal is not the machine's own.
//
// A child service — one with a parentServiceId — carries no renewal block: the
// parent holds it. Reporting "unknown" there would be wrong, and reading
// serviceInfos.renew.automatic instead would be worse: on the six child
// services of the account measured for this, that field returns contradictory
// values on consecutive reads (PUBM-55135). So this says what is true — the
// parent decides — and names the parent.
func renewalPhrase(machine billableService) string {
	if machine.Billing.Renew != nil && machine.Billing.Renew.Current != nil {
		current := machine.Billing.Renew.Current
		phrase := fmt.Sprintf("Renews %s", current.Mode)
		if current.NextDate != "" {
			phrase += " on " + day(current.NextDate)
		}
		if machine.Billing.Engagement != nil && machine.Billing.Engagement.EndDate != "" {
			phrase += fmt.Sprintf(", committed until %s", day(machine.Billing.Engagement.EndDate))
		}

		return phrase + "."
	}

	if machine.ParentServiceID != nil {
		return fmt.Sprintf("Renewal is carried by parent service %d, not by this machine.", *machine.ParentServiceID)
	}

	return "This service declares no renewal."
}

// servicesOf resolves a server into the services billed for it.
func servicesOf(server string) ([]billableService, error) {
	ids, err := httpLib.FetchArray("/v1/services?resourceName="+url.QueryEscape(server), "")
	if err != nil {
		return nil, fmt.Errorf("failed to find what is billed for %s: %w", server, err)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("nothing is billed for %s.\n   Check the name with: ovhcloud baremetal list", server)
	}

	services, err := httpLib.FetchObjectsParallel[billableService]("/v1/services/%s", ids, flags.IgnoreErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to read the services of %s: %w", server, err)
	}

	return services, nil
}

// money renders a total the way the API renders a single price.
func money(value float64, currency string) string {
	if currency == "" {
		return fmt.Sprintf("%.2f", value)
	}

	return fmt.Sprintf("%.2f %s", value, currency)
}

// day keeps the date and drops the time, which no billing question needs.
func day(timestamp string) string {
	if before, _, found := strings.Cut(timestamp, "T"); found {
		return before
	}

	return timestamp
}
