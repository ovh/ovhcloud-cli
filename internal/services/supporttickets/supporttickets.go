// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package supporttickets

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	supportticketsColumnsToDisplay = []string{"ticketId", "serviceName", "type", "category", "state"}
	messagesColumnsToDisplay       = []string{"messageId", "from", "creationDate", "body"}

	//go:embed templates/supporttickets.tmpl
	supportticketsTemplate string

	//go:embed parameter-samples/ticket-create.json
	TicketCreateExample string

	// ReplySpec holds the parameters for the reply command.
	ReplySpec struct {
		Body string `json:"body,omitempty"`
	}

	// CreateSpec holds the parameters for the create command.
	CreateSpec struct {
		Body        string   `json:"body,omitempty"`
		Category    string   `json:"category,omitempty"`
		Impact      string   `json:"impact,omitempty"`
		Product     string   `json:"product,omitempty"`
		ServiceName string   `json:"serviceName,omitempty"`
		Subcategory string   `json:"subcategory,omitempty"`
		Subject     string   `json:"subject,omitempty"`
		Type        string   `json:"type,omitempty"`
		Urgency     string   `json:"urgency,omitempty"`
		Watchers    []string `json:"watchers,omitempty"`
	}
)

func ListSupportTickets(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/support/tickets", "", supportticketsColumnsToDisplay, flags.GenericFilters)
}

func GetSupportTickets(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/support/tickets", args[0], supportticketsTemplate)
}

func ListSupportTicketMessages(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/support/tickets/%s/messages", url.PathEscape(args[0]))
	common.ManageListRequestNoExpand(endpoint, messagesColumnsToDisplay, flags.GenericFilters)
}

func ReplySupportTicket(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/support/tickets/%s/reply", url.PathEscape(args[0]))

	if err := httpLib.Client.Post(endpoint, ReplySpec, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to reply to ticket: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Reply sent successfully")
}

func CreateSupportTicket(cmd *cobra.Command, _ []string) {
	createdResource, err := common.CreateResource(
		cmd,
		"/support/tickets/create",
		"/v1/support/tickets/create",
		TicketCreateExample,
		CreateSpec,
		assets.SupportOpenapiSchema,
		[]string{"body", "subject"})
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to create ticket: %s", err)
		return
	}

	ticketID, hasID := createdResource["ticketId"]
	hasID = hasID && ticketID != nil
	ticketNumber, hasNumber := createdResource["ticketNumber"]
	hasNumber = hasNumber && ticketNumber != nil
	switch {
	case hasID && hasNumber:
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Ticket %v (number %v) created successfully", ticketID, ticketNumber)
	case hasID:
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Ticket %v created successfully", ticketID)
	default:
		display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Ticket created successfully")
	}
}
