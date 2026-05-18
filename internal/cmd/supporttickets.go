// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/services/supporttickets"
	"github.com/spf13/cobra"
)

func init() {
	supportticketsCmd := &cobra.Command{
		Use:   "support-tickets",
		Short: "Retrieve information and manage your support tickets",
	}

	// Command to list tickets
	supportticketsListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List your support tickets",
		Run:     supporttickets.ListSupportTickets,
	}
	supportticketsCmd.AddCommand(withFilterFlag(supportticketsListCmd))

	// Command to get a single support ticket
	supportticketsCmd.AddCommand(&cobra.Command{
		Use:   "get <ticket_id>",
		Short: "Retrieve information of a specific support ticket",
		Args:  cobra.ExactArgs(1),
		Run:   supporttickets.GetSupportTickets,
	})

	// Command to list messages for a support ticket
	supportticketsMessagesCmd := &cobra.Command{
		Use:     "messages <ticket_id>",
		Aliases: []string{"msgs"},
		Short:   "List messages for a support ticket",
		Args:    cobra.ExactArgs(1),
		Run:     supporttickets.ListSupportTicketMessages,
	}
	supportticketsCmd.AddCommand(withFilterFlag(supportticketsMessagesCmd))

	// Command to reply to a support ticket
	supportticketsReplyCmd := &cobra.Command{
		Use:   "reply <ticket_id>",
		Short: "Reply to a support ticket",
		Args:  cobra.ExactArgs(1),
		Run:   supporttickets.ReplySupportTicket,
	}
	supportticketsReplyCmd.Flags().StringVar(&supporttickets.ReplySpec.Body, "body", "", "Text body of the ticket reply")
	_ = supportticketsReplyCmd.MarkFlagRequired("body")
	supportticketsCmd.AddCommand(supportticketsReplyCmd)

	supportticketsCmd.AddCommand(getSupportTicketCreateCmd())

	rootCmd.AddCommand(supportticketsCmd)
}

func getSupportTicketCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new support ticket",
		Long: `Use this command to create a new support ticket.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud support-tickets create --subject 'My issue' --body 'Detailed description of the issue' --product publiccloud --category assistance

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud support-tickets create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud support-tickets create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud support-tickets create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud support-tickets create --from-file ./params.json --subject 'Overridden subject'

3. Using your default text editor:

	ovhcloud support-tickets create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud support-tickets create --editor --subject 'Overridden subject'
`,
		Run:  supporttickets.CreateSupportTicket,
		Args: cobra.NoArgs,
	}

	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Subject, "subject", "", "Ticket subject (short summary)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Body, "body", "", "Ticket message body")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Category, "category", "", "Ticket category (assistance, billing, incident)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Product, "product", "", "Ticket product (adsl, cdn, dedicated, dedicated-billing, dedicated-other, dedicatedcloud, domain, exchange, fax, hosting, housing, iaas, mail, network, publiccloud, sms, ssl, storage, telecom-billing, telecom-other, vac, voip, vps, web-billing, web-other)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.ServiceName, "service-name", "", "Service name (resource identifier) the ticket is about")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Subcategory, "subcategory", "", "Ticket subcategory (alerts, autorenew, bill, down, inProgress, new, other, perfs, start, usage)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Impact, "impact", "", "Ticket impact - Business/Enterprise support only (low, medium, high)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Urgency, "urgency", "", "Ticket urgency - Business/Enterprise support only (low, medium, high)")
	createCmd.Flags().StringVar(&supporttickets.CreateSpec.Type, "type", "", "Ticket type - DEPRECATED (criticalIntervention, genericRequest)")
	_ = createCmd.Flags().MarkHidden("type")
	createCmd.Flags().StringSliceVar(&supporttickets.CreateSpec.Watchers, "watchers", nil, "Comma-separated list of e-mail addresses to notify on ticket updates (max. 10)")

	// Common flags for other means to define parameters
	addParameterFileFlags(createCmd, false, assets.SupportOpenapiSchema, "/support/tickets/create", "post", supporttickets.TicketCreateExample, nil)
	addInteractiveEditorFlag(createCmd)
	markFlagsMutuallyExclusive(createCmd, "from-file", "editor")

	return createCmd
}
