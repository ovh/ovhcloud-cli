// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/order"
	"github.com/spf13/cobra"
)

func init() {
	orderCmd := &cobra.Command{
		Use:   "order",
		Short: "Follow up and settle what you have ordered",
		Long: `Follow up and settle what you have ordered.

"ovhcloud baremetal order" places an order and returns a number. These commands
are what that number is for: what state the order is in, where it is in
delivery, how to pay one placed with --no-pay, and how to give up the
retraction period — which no other command does on your behalf.`,
	}

	orderListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List recent orders, their state and their price",
		Long: `List recent orders, their state and their price.

The state of an order is a separate call from the order itself, so this costs
two requests per order. That is why the window is thirty days and the number of
orders is capped by default, and why both bounds are printed when they bite
rather than applied in silence.`,
		Args: cobra.NoArgs,
		Run:  order.ListOrders,
	}
	orderListCmd.Flags().IntVar(&order.ListDays, "days", order.DefaultListDays, "How far back to look")
	orderListCmd.Flags().StringVar(&order.ListFrom, "from", "", "Start of the window, as 2026-08-01T00:00:00Z (instead of --days)")
	orderListCmd.Flags().StringVar(&order.ListTo, "to", "", "End of the window, as 2026-08-31T00:00:00Z")
	orderListCmd.Flags().IntVar(&order.ListLimit, "limit", 25, "How many of the most recent orders to show (0 for all of them)")
	markFlagsMutuallyExclusive(orderListCmd, "days", "from")
	orderCmd.AddCommand(withFilterFlag(orderListCmd))

	orderCmd.AddCommand(&cobra.Command{
		Use:   "get <order_id>",
		Short: "Retrieve one order, with the link to its order form",
		Args:  cobra.ExactArgs(1),
		Run:   order.GetOrder,
	})

	orderCmd.AddCommand(withFilterFlag(&cobra.Command{
		Use:   "follow <order_id>",
		Short: "Show where an order is in delivery, step by step",
		Args:  cobra.ExactArgs(1),
		Run:   order.FollowOrder,
	}))

	orderPayCmd := &cobra.Command{
		Use:   "pay <order_id>",
		Short: "Pay an order that was placed without payment",
		Long: `Pay an order that was placed without payment.

This is the other half of "ovhcloud baremetal order --no-pay", which exists so
that a purchase can be reviewed by whoever signs for it before the money leaves.`,
		Args: cobra.ExactArgs(1),
		Run:  order.PayOrder,
	}
	orderPayCmd.Flags().IntVar(&order.PayPaymentMethod, "payment-method", 0,
		"Which payment method to use (default: the account's default one)")
	addConfirmationFlags(orderPayCmd, "Print the call that would be made without making it")
	orderCmd.AddCommand(orderPayCmd)

	orderWaiveCmd := &cobra.Command{
		Use:   "waive-retraction <order_id>",
		Short: "Give up the retraction period on an order",
		Long: `Give up the retraction period on an order.

This cannot be undone. It is a command of its own so that it is never a side
effect of buying: "ovhcloud baremetal order" will not waive the retraction
period even though the order API accepts the field, because giving up a legal
right belongs in a sentence you typed on purpose.`,
		Args: cobra.ExactArgs(1),
		Run:  order.WaiveOrderRetraction,
	}
	addConfirmationFlags(orderWaiveCmd, "Print the call that would be made without making it")
	orderCmd.AddCommand(orderWaiveCmd)

	rootCmd.AddCommand(orderCmd)
}
