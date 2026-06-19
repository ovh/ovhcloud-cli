// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/ovh/ovhcloud-cli/internal/services/cloud"
	"github.com/spf13/cobra"
)

func initCloudSavingsPlanCommand(cloudCmd *cobra.Command) {
	savingsPlanCmd := &cobra.Command{
		Use:     "savings-plan",
		Aliases: []string{"sp"},
		Short:   "Manage savings plans for your cloud project",
		Long: `Manage OVHcloud Savings Plans for your Public Cloud project.

Savings Plans allow you to commit to a consistent amount of usage (measured in $/hour) 
for a 1-month term, in exchange for discounted pricing on your cloud resources.

Available flavors include:
- Rancher: rancher, rancher_standard, rancher_ovhcloud_edition
- General purpose instances: b3-8, b3-16, b3-32, b3-64, b3-128, b3-256
- Compute optimized instances: c3-4, c3-8, c3-16, c3-32, c3-64, c3-128
- Memory optimized instances: r3-16, r3-32, r3-64, r3-128, r3-256, r3-512`,
	}
	savingsPlanCmd.PersistentFlags().StringVar(&cloud.CloudProject, "cloud-project", "", "Cloud project ID")

	// List subscribed savings plans
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List subscribed savings plans",
		Run:     cloud.ListSavingsPlans,
	}
	savingsPlanCmd.AddCommand(withFilterFlag(listCmd))

	// Get a specific savings plan
	savingsPlanCmd.AddCommand(&cobra.Command{
		Use:   "get <savings_plan_id>",
		Short: "Get details of a specific savings plan",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.GetSavingsPlan,
	})

	// List available offers
	listOffersCmd := &cobra.Command{
		Use:   "list-offers",
		Short: "List available savings plan offers to subscribe to",
		Long: `List available savings plan offers that can be subscribed to.

Use --product-code to filter by flavor (e.g., 'b3-8', 'rancher', 'c3-16').
Use --deployment-type to filter by availability zone configuration (1AZ or 3AZ).

Note: Rancher flavors only support 1AZ deployment.`,
		Run: cloud.ListSavingsPlanOffers,
	}
	listOffersCmd.Flags().StringVar(&cloud.SavingsPlanOffersFilter.ProductCode, "product-code", "", "Filter offers by product code (e.g., 'b3-8', 'rancher')")
	listOffersCmd.Flags().StringVar(&cloud.SavingsPlanOffersFilter.DeploymentType, "deployment-type", "1AZ", "Deployment type: 1AZ or 3AZ (default: 1AZ)")
	savingsPlanCmd.AddCommand(withFilterFlag(listOffersCmd))

	// Subscribe to a savings plan
	subscribeCmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to a new savings plan",
		Long: `Subscribe to a new OVHcloud Savings Plan.

You can subscribe in two ways:

1. Using an offer ID directly:
   ovhcloud cloud savings-plan subscribe --offer-id <offer_id> --display-name "My Plan" --size 2

2. Using flavor and deployment type (the CLI will find the matching offer):
   ovhcloud cloud savings-plan subscribe --flavor b3-8 --deployment-type 1AZ --display-name "My Plan" --size 2

Available flavors:
- Rancher: rancher, rancher_standard, rancher_ovhcloud_edition (1AZ only)
- General purpose: b3-8, b3-16, b3-32, b3-64, b3-128, b3-256
- Compute optimized: c3-4, c3-8, c3-16, c3-32, c3-64, c3-128
- Memory optimized: r3-16, r3-32, r3-64, r3-128, r3-256, r3-512

Deployment types:
- 1AZ: Single availability zone (default)
- 3AZ: Three availability zones (not available for Rancher)`,
		Run: cloud.SubscribeSavingsPlan,
	}
	subscribeCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.DisplayName, "display-name", "", "Custom display name (required)")
	subscribeCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.OfferID, "offer-id", "", "Offer ID from list-offers (alternative to --flavor)")
	subscribeCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.Flavor, "flavor", "", "Savings plan flavor (e.g., b3-8, rancher, c3-16)")
	subscribeCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.DeploymentType, "deployment-type", "1AZ", "Deployment type: 1AZ or 3AZ (default: 1AZ)")
	subscribeCmd.Flags().IntVar(&cloud.SavingsPlanSubscribeSpec.Size, "size", 0, "Size of the savings plan (required)")
	subscribeCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.StartDate, "start-date", "", "Start date (YYYY-MM-DD format, defaults to today)")
	savingsPlanCmd.AddCommand(subscribeCmd)

	// Simulate a subscription
	simulateCmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate a savings plan subscription",
		Long: `Simulate subscribing to an OVHcloud Savings Plan without actually subscribing.

This is useful to preview what the savings plan would look like before committing.
You can use either --offer-id or --flavor with --deployment-type.`,
		Run: cloud.SimulateSavingsPlanSubscription,
	}
	simulateCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.DisplayName, "display-name", "", "Custom display name (required)")
	simulateCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.OfferID, "offer-id", "", "Offer ID from list-offers (alternative to --flavor)")
	simulateCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.Flavor, "flavor", "", "Savings plan flavor (e.g., b3-8, rancher, c3-16)")
	simulateCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.DeploymentType, "deployment-type", "1AZ", "Deployment type: 1AZ or 3AZ (default: 1AZ)")
	simulateCmd.Flags().IntVar(&cloud.SavingsPlanSubscribeSpec.Size, "size", 0, "Size of the savings plan (required)")
	simulateCmd.Flags().StringVar(&cloud.SavingsPlanSubscribeSpec.StartDate, "start-date", "", "Start date (YYYY-MM-DD format, defaults to today)")
	savingsPlanCmd.AddCommand(simulateCmd)

	// Terminate a savings plan
	terminateCmd := &cobra.Command{
		Use:     "terminate <savings_plan_id>",
		Aliases: []string{"unsubscribe"},
		Short:   "Terminate/unsubscribe from a savings plan",
		Long: `Terminate an existing savings plan subscription.

By default, the savings plan will be terminated at the end of its current period.
You can specify a termination date using the --termination-date flag.`,
		Args: cobra.ExactArgs(1),
		Run:  cloud.TerminateSavingsPlan,
	}
	terminateCmd.Flags().String("termination-date", "", "Termination date (YYYY-MM-DD format, optional)")
	savingsPlanCmd.AddCommand(terminateCmd)

	// Change period end action
	changeEndActionCmd := &cobra.Command{
		Use:   "set-renewal <savings_plan_id>",
		Short: "Set the action at the end of the savings plan period",
		Long: `Set the action to be performed when the savings plan reaches the end of its period.

Available actions:
- REACTIVATE: Automatically renew the savings plan for another period
- TERMINATE: Terminate the savings plan at the end of the period`,
		Args: cobra.ExactArgs(1),
		Run:  cloud.ChangeSavingsPlanEndAction,
	}
	changeEndActionCmd.Flags().String("action", "", "Action at period end: REACTIVATE or TERMINATE (required)")
	changeEndActionCmd.MarkFlagRequired("action")
	savingsPlanCmd.AddCommand(changeEndActionCmd)

	// Change size
	changeSizeCmd := &cobra.Command{
		Use:   "resize <savings_plan_id>",
		Short: "Change the size of a savings plan",
		Long: `Change the size of an existing savings plan.

Note: You can only increase the size of a savings plan, not decrease it.`,
		Args: cobra.ExactArgs(1),
		Run:  cloud.ChangeSavingsPlanSize,
	}
	changeSizeCmd.Flags().Int("size", 0, "New size for the savings plan (required)")
	changeSizeCmd.MarkFlagRequired("size")
	savingsPlanCmd.AddCommand(changeSizeCmd)

	// Edit display name
	editCmd := &cobra.Command{
		Use:   "edit <savings_plan_id>",
		Short: "Edit a savings plan's display name",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.EditSavingsPlanDisplayName,
	}
	editCmd.Flags().String("display-name", "", "New display name (required)")
	editCmd.MarkFlagRequired("display-name")
	savingsPlanCmd.AddCommand(editCmd)

	// List periods history
	listPeriodsCmd := &cobra.Command{
		Use:   "list-periods <savings_plan_id>",
		Short: "List the period history of a savings plan",
		Args:  cobra.ExactArgs(1),
		Run:   cloud.ListSavingsPlanPeriods,
	}
	savingsPlanCmd.AddCommand(withFilterFlag(listPeriodsCmd))

	cloudCmd.AddCommand(savingsPlanCmd)
}
