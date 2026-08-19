// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

const (
	ticketCreatePath = "/support/tickets/create"
	ticketCreateURL  = "/v1" + ticketCreatePath

	// contextMarker opens the block this command appends. It is spelled out in
	// the message so the person reading the ticket knows which half a human
	// wrote.
	contextMarker = "--- Collected automatically by the OVHcloud CLI (ovhcloud baremetal ticket) ---"
)

// TicketSpec holds the parameters of the ticket command.
var TicketSpec struct {
	Subject     string
	Body        string
	Category    string
	Subcategory string
	Product     string
	Impact      string
	Urgency     string
	Watchers    []string
	NoContext   bool
}

var (
	ticketProducts = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.SupportOpenapiSchema, ticketCreatePath, "post", "product")
	})
	ticketCategories = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.SupportOpenapiSchema, ticketCreatePath, "post", "category")
	})
	ticketSubcategories = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.SupportOpenapiSchema, ticketCreatePath, "post", "subcategory")
	})
	ticketImpacts = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.SupportOpenapiSchema, ticketCreatePath, "post", "impact")
	})
)

func CompleteTicketProduct(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(ticketProducts)
}

func CompleteTicketCategory(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(ticketCategories)
}

func CompleteTicketSubcategory(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(ticketSubcategories)
}

func CompleteTicketImpact(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(ticketImpacts)
}

// contextFields are the identity of a machine as support reads it. Order is the
// order they appear in the ticket, so it goes from what names the server to
// what describes its current condition.
var contextFields = []struct{ label, key string }{
	{"Commercial range", "commercialRange"},
	{"Datacenter", "datacenter"},
	{"Rack", "rack"},
	{"Operating system", "os"},
	{"Main IP", "ip"},
	{"Reverse", "reverse"},
	{"State", "state"},
	{"Power state", "powerState"},
	{"Support level", "supportLevel"},
	{"Professional use", "professionalUse"},
	{"Monitoring", "monitoring"},
	{"No intervention", "noIntervention"},
}

// OpenTicketForServer creates a support ticket about one dedicated server, with
// the state of that server already in it.
//
// C8 of the audit asks for a ticket "pre-filled with machine context". The
// context is not decoration: a support agent's first three questions are which
// machine, in what state, and what is already running on it — and an operator
// answering them by hand answers them from memory.
func OpenTicketForServer(cmd *cobra.Command, args []string) {
	server := args[0]

	if strings.TrimSpace(TicketSpec.Subject) == "" {
		display.OutputError(&flags.OutputFormatConfig, "--subject cannot be blank: it is the line support sorts on")
		return
	}
	if strings.TrimSpace(TicketSpec.Body) == "" {
		display.OutputError(&flags.OutputFormatConfig, "--body cannot be blank: the collected context describes the machine, not the problem")
		return
	}

	for _, check := range []struct {
		name  string
		value string
		read  func() ([]string, error)
	}{
		{"product", TicketSpec.Product, ticketProducts},
		{"category", TicketSpec.Category, ticketCategories},
		{"subcategory", TicketSpec.Subcategory, ticketSubcategories},
		{"impact", TicketSpec.Impact, ticketImpacts},
		{"urgency", TicketSpec.Urgency, ticketImpacts},
	} {
		if err := common.CheckEnumFlag(check.name, check.value, check.read); err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	body := TicketSpec.Body
	if !TicketSpec.NoContext {
		body = fmt.Sprintf("%s\n\n%s\n%s", TicketSpec.Body, contextMarker, serverContext(server))
	}

	payload := map[string]any{
		"subject":     TicketSpec.Subject,
		"body":        body,
		"serviceName": server,
		"product":     TicketSpec.Product,
	}
	for key, value := range map[string]string{
		"category":    TicketSpec.Category,
		"subcategory": TicketSpec.Subcategory,
		"impact":      TicketSpec.Impact,
		"urgency":     TicketSpec.Urgency,
	} {
		if value != "" {
			payload[key] = value
		}
	}
	if len(TicketSpec.Watchers) > 0 {
		payload["watchers"] = TicketSpec.Watchers
	}

	preview := fmt.Sprintf("Subject: %s\nProduct: %s\nService: %s\n\n%s",
		TicketSpec.Subject, TicketSpec.Product, server, body)

	// This does not go through common.ReportDryRun, and the reason is the
	// payload: the thing an operator needs to read before agreeing is the
	// message that leaves the machine, not the path it leaves by. Reporting
	// both would mean two display calls, and under -o json two calls are two
	// JSON documents on one stdout.
	if flags.DryRun {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"calls": []map[string]any{{"method": http.MethodPost, "endpoint": ticketCreateURL}}, "payload": payload},
			"🔍 Dry run: nothing was sent. This would have been created:\n\n%s", preview)
		return
	}

	if !common.ConfirmAction(common.Disruptive, server,
		fmt.Sprintf("This opens a support ticket at OVHcloud, which a person will read:\n\n%s", preview)) {
		display.OutputError(&flags.OutputFormatConfig, "no ticket was created for %s", server)
		return
	}

	var created map[string]any
	if err := httpLib.Client.Post(ticketCreateURL, payload, &created); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create the ticket for %s: %s", server, err)
		return
	}

	ticketID, hasID := created["ticketId"]
	switch number, hasNumber := created["ticketNumber"]; {
	case hasID && hasNumber && ticketID != nil && number != nil:
		display.OutputInfo(&flags.OutputFormatConfig, created,
			"✅ Ticket %v (number %v) opened for %s. Follow it with: ovhcloud support-tickets get %v",
			ticketID, number, server, ticketID)
	case hasID && ticketID != nil:
		display.OutputInfo(&flags.OutputFormatConfig, created,
			"✅ Ticket %v opened for %s. Follow it with: ovhcloud support-tickets get %v", ticketID, server, ticketID)
	default:
		display.OutputInfo(&flags.OutputFormatConfig, created, "✅ Ticket opened for %s.", server)
	}
}

// identityLines renders the fields that say which machine this is.
//
// A field the API did not send is left out rather than printed empty: a support
// agent reading "Rack" with nothing after it cannot tell a machine with no rack
// from a CLI that failed to read one. A false is not an empty, though, and
// monitoring is exactly the field where false is the interesting value.
func identityLines(detail map[string]any) []string {
	lines := make([]string, 0, len(contextFields))
	for _, field := range contextFields {
		value, ok := detail[field.key]
		if !ok || value == nil || fmt.Sprint(value) == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%-20s %v", field.label, value))
	}
	return lines
}

// serverContext describes the machine the ticket is about.
//
// It never fails the command. A server that cannot be read is the very
// situation a ticket gets opened for, so the failure is written into the ticket
// instead of stopping it — the support agent learns more from "the API would
// not answer for this server, here is the error" than from a ticket that was
// never sent.
func serverContext(server string) string {
	escaped := url.PathEscape(server)
	lines := []string{fmt.Sprintf("%-20s %s", "Service name", server)}

	var detail map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/dedicated/server/%s", escaped), &detail); err != nil {
		return strings.Join(append(lines,
			fmt.Sprintf("\nThe CLI could not read this server from the API: %s", err)), "\n")
	}

	lines = append(lines, identityLines(detail)...)

	found := checkState(server, detail)
	found = append(found, checkMonitoring(server, detail)...)
	found = append(found, checkIntervention(server, detail)...)

	var boot map[string]any
	if bootID := bootIdentifier(detail); bootID != 0 {
		if err := httpLib.Client.Get(fmt.Sprintf("/v1/dedicated/server/%s/boot/%d", escaped, bootID), &boot); err == nil {
			lines = append(lines, fmt.Sprintf("%-20s #%d %s, kernel %s", "Active boot", bootID,
				stringValue(boot, "bootType"), stringValue(boot, "kernel")))
			found = append(found, checkBoot(server, boot)...)
		}
	}

	var running []int64
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/dedicated/server/%s/task?status=doing", escaped), &running); err == nil {
		lines = append(lines, fmt.Sprintf("%-20s %d", "Running tasks", len(running)))
		found = append(found, checkTasks(server, running)...)
	}

	var planned []int64
	if err := httpLib.Client.Get(fmt.Sprintf("/v1/dedicated/server/%s/plannedIntervention", escaped), &planned); err == nil {
		lines = append(lines, fmt.Sprintf("%-20s %d", "Planned maintenance", len(planned)))
		found = append(found, checkPlannedIntervention(server, planned)...)
	}

	// The renewal check is deliberately absent: it reads serviceInfos five times
	// because that field disagrees with itself between reads, and five requests
	// is not a reasonable price for a line of context in a technical ticket.
	sortFindings(found)
	lines = append(lines, "", "What ovhcloud baremetal doctor reports on this server:")
	if len(found) == 0 {
		lines = append(lines, "  nothing")
		return strings.Join(lines, "\n")
	}
	for _, f := range found {
		lines = append(lines, fmt.Sprintf("  [%s] %s — %s", f.Severity, f.Check, f.Detail))
	}

	return strings.Join(lines, "\n")
}
