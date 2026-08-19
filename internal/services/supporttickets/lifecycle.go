// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package supporttickets

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// TicketsPath is the collection every command in this package works from.
const TicketsPath = "/v1/support/tickets"

// ReopenSpec holds the parameters for the reopen command.
var ReopenSpec struct {
	Body string `json:"body,omitempty"`
}

// readTicket fetches one ticket so a command can say what it is about to act
// on. Naming the subject in a confirmation is the difference between agreeing
// to close "12345" and agreeing to close "12345 \"Disk replacement on
// ns3118333\"": the identifier alone is not something an operator can check.
func readTicket(id string) (map[string]any, error) {
	var ticket map[string]any
	if err := httpLib.Client.Get(fmt.Sprintf("%s/%s", TicketsPath, url.PathEscape(id)), &ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// ticketLabel identifies a ticket by everything the operator recognises it by.
func ticketLabel(id string, ticket map[string]any) string {
	label := id
	if number, ok := ticket["ticketNumber"]; ok && number != nil {
		label = fmt.Sprintf("%s (number %v)", label, number)
	}
	if subject := stringField(ticket, "subject"); subject != "" {
		label = fmt.Sprintf("%s %q", label, subject)
	}
	return label
}

// CloseSupportTicket ends a support conversation.
//
// Whether a ticket may be closed is a field on the ticket, not a route: the API
// answers "canBeScored" with a request of its own but carries "canBeClosed"
// inline. So the check costs the read this command already makes to name the
// ticket in its prompt.
func CloseSupportTicket(_ *cobra.Command, args []string) {
	id := args[0]
	endpoint := fmt.Sprintf("%s/%s/close", TicketsPath, url.PathEscape(id))

	ticket, err := readTicket(id)
	if err != nil {
		if common.IsNotFound(err) {
			display.OutputError(&flags.OutputFormatConfig, "no support ticket %s on this account", id)
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to read ticket %s: %s", id, err)
		return
	}
	label := ticketLabel(id, ticket)

	if state := stringField(ticket, "state"); state == "closed" {
		display.OutputWarning(&flags.OutputFormatConfig, "ticket %s is already closed", label)
		return
	}
	if closable, known := boolField(ticket, "canBeClosed"); known && !closable {
		display.OutputError(&flags.OutputFormatConfig,
			"ticket %s cannot be closed: the API reports canBeClosed=false, which is how it says support is still working on it. Reply to it instead with: ovhcloud support-tickets reply %s --body '...'",
			label, id)
		return
	}

	if common.ReportDryRun(common.Call{Method: http.MethodPost, Endpoint: endpoint}) {
		return
	}

	if !common.ConfirmAction(common.Disruptive, id,
		fmt.Sprintf("This closes support ticket %s. Somebody at OVHcloud support sees it close.", label)) {
		display.OutputError(&flags.OutputFormatConfig, "ticket %s was not closed", label)
		return
	}

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to close ticket %s: %s", label, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil,
		"✅ Ticket %s is closed. Reopen it with: ovhcloud support-tickets reopen %s --reason '...'", label, id)
}

// ReopenSupportTicket restarts a closed conversation. The reason is not
// optional decoration: the API requires it, and it is the message the support
// agent reads first.
func ReopenSupportTicket(_ *cobra.Command, args []string) {
	id := args[0]
	endpoint := fmt.Sprintf("%s/%s/reopen", TicketsPath, url.PathEscape(id))

	// A reason made of spaces satisfies MarkFlagRequired and says nothing. The
	// same hole was found by review in three guards of the ip lot.
	if strings.TrimSpace(ReopenSpec.Body) == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"--reason cannot be blank: it is the message support reads when the ticket comes back")
		return
	}

	ticket, err := readTicket(id)
	if err != nil {
		if common.IsNotFound(err) {
			display.OutputError(&flags.OutputFormatConfig, "no support ticket %s on this account", id)
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to read ticket %s: %s", id, err)
		return
	}
	label := ticketLabel(id, ticket)

	// "unknown" is a value of this enum, so only a definite "open" is a refusal.
	if state := stringField(ticket, "state"); state == "open" {
		display.OutputWarning(&flags.OutputFormatConfig,
			"ticket %s is already open; add to it with: ovhcloud support-tickets reply %s --body '...'", label, id)
		return
	}

	// The reason is not echoed here: it came from this very command line, and
	// common.Call carries no field for a payload on this branch.
	if common.ReportDryRun(common.Call{Method: http.MethodPost, Endpoint: endpoint}) {
		return
	}

	if !common.ConfirmAction(common.Disruptive, id,
		fmt.Sprintf("This reopens support ticket %s and notifies OVHcloud support.", label)) {
		display.OutputError(&flags.OutputFormatConfig, "ticket %s was not reopened", label)
		return
	}

	if err := httpLib.Client.Post(endpoint, ReopenSpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reopen ticket %s: %s", label, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Ticket %s is open again.", label)
}

// SupportTicketCanBeScored answers whether a ticket is eligible for the
// satisfaction survey.
//
// The route returns a bare boolean, which is not a table and not an object: it
// is rendered as the sentence it is, with the value kept in the details so
// -o json stays a machine answer.
func SupportTicketCanBeScored(_ *cobra.Command, args []string) {
	id := args[0]
	endpoint := fmt.Sprintf("%s/%s/canBeScored", TicketsPath, url.PathEscape(id))

	var scorable bool
	if err := httpLib.Client.Get(endpoint, &scorable); err != nil {
		if common.IsNotFound(err) {
			display.OutputError(&flags.OutputFormatConfig, "no support ticket %s on this account", id)
			return
		}
		display.OutputError(&flags.OutputFormatConfig, "failed to ask whether ticket %s can be scored: %s", id, err)
		return
	}

	details := map[string]any{"ticketId": id, "canBeScored": scorable}
	if scorable {
		display.OutputInfo(&flags.OutputFormatConfig, details,
			"Ticket %s can be scored. Scoring is not available from this CLI; it is in the OVHcloud control panel.", id)
		return
	}
	display.OutputInfo(&flags.OutputFormatConfig, details, "Ticket %s cannot be scored.", id)
}

func stringField(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}

// boolField reports the value and whether the field was a boolean at all, so a
// caller can tell "the API said false" from "the API said nothing".
func boolField(object map[string]any, key string) (bool, bool) {
	value, ok := object[key].(bool)
	return value, ok
}
