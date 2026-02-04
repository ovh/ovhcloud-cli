// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cloud

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

var (
	cloudSavingsPlanColumnsToDisplay = []string{"id", "displayName name", "flavor", "size", "status", "periodEndDate endDate", "periodEndAction"}

	//go:embed templates/cloud_savings_plan.tmpl
	cloudSavingsPlanTemplate string

	// Subscription parameters set via CLI flags
	SavingsPlanSubscribeSpec struct {
		DisplayName    string `json:"displayName,omitempty"`
		OfferID        string `json:"offerId,omitempty"`
		Size           int    `json:"size,omitempty"`
		StartDate      string `json:"startDate,omitempty"`
		Flavor         string `json:"-"` // Used for offer lookup, not sent to API
		DeploymentType string `json:"-"` // 1AZ or 3AZ, used for offer filtering
	}

	// Filter parameters for listing offers
	SavingsPlanOffersFilter struct {
		ProductCode    string
		DeploymentType string // 1AZ or 3AZ
	}
)

// savingsPlanOffer represents a subscribable savings plan offer
type savingsPlanOffer struct {
	OfferID string `json:"offerId"`
}

// serviceResponse represents the response from /v1/services endpoint
type serviceResponse struct {
	ServiceID int `json:"serviceId"`
}

// getServiceID retrieves the service ID for a given cloud project
func getServiceID(projectID string) (int, error) {
	endpoint := fmt.Sprintf("/v1/cloud/project/%s/serviceInfos", projectID)

	var serviceInfo serviceResponse
	if err := httpLib.Client.Get(endpoint, &serviceInfo); err != nil {
		return 0, fmt.Errorf("failed to get service ID: %w", err)
	}

	return serviceInfo.ServiceID, nil
}

// ListSavingsPlans lists all subscribed savings plans for a cloud project
func ListSavingsPlans(_ *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed", serviceID)
	common.ManageListRequestNoExpand(endpoint, cloudSavingsPlanColumnsToDisplay, flags.GenericFilters)
}

// GetSavingsPlan retrieves a specific savings plan by ID
func GetSavingsPlan(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s", serviceID, url.PathEscape(args[0]))

	var savingsPlan map[string]any
	if err := httpLib.Client.Get(endpoint, &savingsPlan); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to get savings plan: %s", err)
		return
	}

	display.OutputObject(savingsPlan, args[0], cloudSavingsPlanTemplate, &flags.OutputFormatConfig)
}

// ListSavingsPlanOffers lists available savings plan offers that can be subscribed
func ListSavingsPlanOffers(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Handle deployment type and product code filtering
	deploymentType := strings.ToUpper(SavingsPlanOffersFilter.DeploymentType)
	if deploymentType == "" {
		deploymentType = "1AZ" // Default to 1AZ
	}
	if deploymentType != "1AZ" && deploymentType != "3AZ" {
		display.OutputError(&flags.OutputFormatConfig, "invalid deployment type: must be 1AZ or 3AZ")
		return
	}

	productCode := SavingsPlanOffersFilter.ProductCode
	if productCode != "" {
		// Normalize product code (replace underscores with spaces)
		productCode = strings.ReplaceAll(strings.ToLower(productCode), "_", " ")
		if productCode == "rancher" {
			productCode = "rancher standard"
		}

		// Check if rancher with 3AZ (not supported)
		if strings.HasPrefix(productCode, "rancher") && deploymentType == "3AZ" {
			display.OutputError(&flags.OutputFormatConfig, "3AZ deployment is not supported for Rancher flavors")
			return
		}

		// Append 3AZ to product code if needed
		if deploymentType == "3AZ" {
			productCode += " 3AZ"
		}
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribable", serviceID)
	if productCode != "" {
		endpoint = fmt.Sprintf("%s?productCode=%s", endpoint, url.QueryEscape(productCode))
	}

	offers, err := httpLib.FetchArray(endpoint, "")
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to list savings plan offers: %s", err)
		return
	}

	// For 1AZ without specific product code, we need to filter out 3AZ offers
	var filteredOffers []any
	if deploymentType == "1AZ" && SavingsPlanOffersFilter.ProductCode == "" {
		// We can't easily filter without knowing all 3AZ offers, so show all
		// User should use --product-code for specific filtering
		filteredOffers = offers
	} else if deploymentType == "1AZ" && productCode != "" && !strings.HasPrefix(productCode, "rancher") {
		// Fetch 3AZ offers to filter them out
		threeAZEndpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribable?productCode=%s",
			serviceID, url.QueryEscape(productCode+" 3AZ"))
		threeAZOffers, _ := httpLib.FetchArray(threeAZEndpoint, "")

		// Build set of 3AZ offer IDs
		threeAZIDs := make(map[string]struct{})
		for _, offer := range threeAZOffers {
			if offerMap, ok := offer.(map[string]any); ok {
				if offerId, ok := offerMap["offerId"].(string); ok {
					threeAZIDs[offerId] = struct{}{}
				}
			}
		}

		// Filter out 3AZ offers
		for _, offer := range offers {
			if offerMap, ok := offer.(map[string]any); ok {
				if offerId, ok := offerMap["offerId"].(string); ok {
					if _, is3AZ := threeAZIDs[offerId]; !is3AZ {
						filteredOffers = append(filteredOffers, offer)
					}
				}
			}
		}
	} else {
		filteredOffers = offers
	}

	// Convert offers to the expected format for display
	var offerList []map[string]any
	for _, offer := range filteredOffers {
		if offerMap, ok := offer.(map[string]any); ok {
			offerId := offerMap["offerId"].(string)
			offerList = append(offerList, map[string]any{
				"offerId": offerId,
			})
		}
	}

	display.RenderTable(offerList, []string{"offerId"}, &flags.OutputFormatConfig)
}

// findMatchingOffer finds the appropriate offer ID based on flavor and deployment type
func findMatchingOffer(serviceID int, flavor, deploymentType string) (string, error) {
	// Normalize flavor
	flavor = strings.ReplaceAll(strings.ToLower(flavor), "_", " ")
	if flavor == "rancher" {
		flavor = "rancher standard"
	}

	// Validate deployment type for rancher
	if strings.HasPrefix(flavor, "rancher") && deploymentType == "3AZ" {
		return "", fmt.Errorf("3AZ deployment is not supported for Rancher flavors")
	}

	// Build full flavor name
	fullFlavor := flavor
	if deploymentType == "3AZ" {
		fullFlavor += " 3AZ"
	}

	// Fetch subscribable offers
	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribable?productCode=%s", serviceID, url.QueryEscape(fullFlavor))
	var offers []savingsPlanOffer
	if err := httpLib.Client.Get(endpoint, &offers); err != nil {
		return "", fmt.Errorf("failed to fetch offers: %w", err)
	}

	if len(offers) == 0 {
		return "", fmt.Errorf("no offers found for flavor %q with deployment type %s", flavor, deploymentType)
	}

	// For 1AZ (non-rancher), we need to filter out 3AZ offers
	if deploymentType == "1AZ" && !strings.HasPrefix(flavor, "rancher") {
		// Fetch 3AZ offers to exclude them
		threeAZEndpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribable?productCode=%s", serviceID, url.QueryEscape(flavor+" 3AZ"))
		var threeAZOffers []savingsPlanOffer
		_ = httpLib.Client.Get(threeAZEndpoint, &threeAZOffers) // Ignore error, just means no 3AZ offers

		// Build set of 3AZ offer IDs
		threeAZIDs := make(map[string]struct{})
		for _, offer := range threeAZOffers {
			threeAZIDs[offer.OfferID] = struct{}{}
		}

		// Filter to keep only 1AZ offers
		var oneAZOffers []savingsPlanOffer
		for _, offer := range offers {
			if _, is3AZ := threeAZIDs[offer.OfferID]; !is3AZ {
				oneAZOffers = append(oneAZOffers, offer)
			}
		}
		offers = oneAZOffers
	}

	if len(offers) == 0 {
		return "", fmt.Errorf("no %s offers found for flavor %q", deploymentType, flavor)
	}

	// Return the first matching offer
	return offers[0].OfferID, nil
}

// SubscribeSavingsPlan subscribes to a new savings plan
func SubscribeSavingsPlan(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Validate required parameters
	if SavingsPlanSubscribeSpec.DisplayName == "" {
		display.OutputError(&flags.OutputFormatConfig, "display name is required (use --display-name)")
		return
	}
	if SavingsPlanSubscribeSpec.Size <= 0 {
		display.OutputError(&flags.OutputFormatConfig, "size must be greater than 0 (use --size)")
		return
	}

	// Determine offer ID
	offerID := SavingsPlanSubscribeSpec.OfferID
	if offerID == "" {
		// If no offer ID provided, try to find one based on flavor and deployment type
		if SavingsPlanSubscribeSpec.Flavor == "" {
			display.OutputError(&flags.OutputFormatConfig, "either --offer-id or --flavor is required")
			return
		}

		deploymentType := strings.ToUpper(SavingsPlanSubscribeSpec.DeploymentType)
		if deploymentType == "" {
			deploymentType = "1AZ"
		} else if deploymentType != "1AZ" && deploymentType != "3AZ" {
			display.OutputError(&flags.OutputFormatConfig, "invalid deployment type: must be 1AZ or 3AZ")
			return
		}

		var err error
		offerID, err = findMatchingOffer(serviceID, SavingsPlanSubscribeSpec.Flavor, deploymentType)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	// Build subscription request
	request := map[string]any{
		"displayName": SavingsPlanSubscribeSpec.DisplayName,
		"offerId":     offerID,
		"size":        SavingsPlanSubscribeSpec.Size,
	}

	// Add start date if provided, otherwise default to today
	if SavingsPlanSubscribeSpec.StartDate != "" {
		request["startDate"] = SavingsPlanSubscribeSpec.StartDate
	} else {
		request["startDate"] = time.Now().Format("2006-01-02")
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribe/execute", serviceID)

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, request, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to subscribe to savings plan: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Successfully subscribed to savings plan (ID: %s)", result["id"])
}

// SimulateSavingsPlanSubscription simulates a savings plan subscription
func SimulateSavingsPlanSubscription(cmd *cobra.Command, _ []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// Validate required parameters
	if SavingsPlanSubscribeSpec.DisplayName == "" {
		display.OutputError(&flags.OutputFormatConfig, "display name is required (use --display-name)")
		return
	}
	if SavingsPlanSubscribeSpec.Size <= 0 {
		display.OutputError(&flags.OutputFormatConfig, "size must be greater than 0 (use --size)")
		return
	}

	// Determine offer ID
	offerID := SavingsPlanSubscribeSpec.OfferID
	if offerID == "" {
		// If no offer ID provided, try to find one based on flavor and deployment type
		if SavingsPlanSubscribeSpec.Flavor == "" {
			display.OutputError(&flags.OutputFormatConfig, "either --offer-id or --flavor is required")
			return
		}

		deploymentType := "1AZ" // Default to 1AZ
		// Only use deployment type when it's 3AZ
		if strings.ToUpper(SavingsPlanSubscribeSpec.DeploymentType) == "3AZ" {
			deploymentType = "3AZ"
		}

		var err error
		offerID, err = findMatchingOffer(serviceID, SavingsPlanSubscribeSpec.Flavor, deploymentType)
		if err != nil {
			display.OutputError(&flags.OutputFormatConfig, "%s", err)
			return
		}
	}

	// Build simulation request
	request := map[string]any{
		"displayName": SavingsPlanSubscribeSpec.DisplayName,
		"offerId":     offerID,
		"size":        SavingsPlanSubscribeSpec.Size,
	}

	// Add start date if provided
	if SavingsPlanSubscribeSpec.StartDate != "" {
		request["startDate"] = SavingsPlanSubscribeSpec.StartDate
	} else {
		request["startDate"] = time.Now().Format("2006-01-02")
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribe/simulate", serviceID)

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, request, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to simulate savings plan subscription: %s", err)
		return
	}

	display.OutputObject(result, "simulation", "", &flags.OutputFormatConfig)
}

// TerminateSavingsPlan terminates an existing savings plan
func TerminateSavingsPlan(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	savingsPlanID := args[0]

	// Get termination date from flag or use nil for immediate termination
	terminationDate, _ := cmd.Flags().GetString("termination-date")

	request := map[string]any{}
	if terminationDate != "" {
		request["terminationDate"] = terminationDate
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s/terminate", serviceID, url.PathEscape(savingsPlanID))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, request, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to terminate savings plan: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Savings plan %s has been scheduled for termination", savingsPlanID)
}

// ChangeSavingsPlanEndAction changes the action performed at the end of the savings plan period
func ChangeSavingsPlanEndAction(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	savingsPlanID := args[0]
	action, _ := cmd.Flags().GetString("action")

	// Validate action
	action = strings.ToUpper(action)
	if action != "REACTIVATE" && action != "TERMINATE" {
		display.OutputError(&flags.OutputFormatConfig, "invalid action: must be REACTIVATE or TERMINATE")
		return
	}

	request := map[string]string{
		"periodEndAction": action,
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s/changePeriodEndAction", serviceID, url.PathEscape(savingsPlanID))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, request, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change period end action: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Period end action changed to %s for savings plan %s", action, savingsPlanID)
}

// ChangeSavingsPlanSize changes the size of an existing savings plan
func ChangeSavingsPlanSize(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	savingsPlanID := args[0]
	newSize, _ := cmd.Flags().GetInt("size")

	if newSize <= 0 {
		display.OutputError(&flags.OutputFormatConfig, "size must be greater than 0")
		return
	}

	request := map[string]int{
		"size": newSize,
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s/changeSize", serviceID, url.PathEscape(savingsPlanID))

	var result map[string]any
	if err := httpLib.Client.Post(endpoint, request, &result); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to change savings plan size: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, result, "✅ Savings plan %s size changed to %d", savingsPlanID, newSize)
}

// EditSavingsPlanDisplayName updates the display name of a savings plan
func EditSavingsPlanDisplayName(cmd *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	savingsPlanID := args[0]
	displayName, _ := cmd.Flags().GetString("display-name")

	if displayName == "" {
		display.OutputError(&flags.OutputFormatConfig, "display name is required (use --display-name)")
		return
	}

	request := map[string]string{
		"displayName": displayName,
	}

	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s", serviceID, url.PathEscape(savingsPlanID))

	if err := httpLib.Client.Put(endpoint, request, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to update savings plan: %s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, nil, "✅ Savings plan %s display name updated to %q", savingsPlanID, displayName)
}

// ListSavingsPlanPeriods lists the period history of a savings plan
func ListSavingsPlanPeriods(_ *cobra.Command, args []string) {
	projectID, err := getConfiguredCloudProject()
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	serviceID, err := getServiceID(projectID)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	savingsPlanID := args[0]
	endpoint := fmt.Sprintf("/v1/services/%d/savingsPlans/subscribed/%s/periods", serviceID, url.PathEscape(savingsPlanID))

	periodColumns := []string{"id", "periodStartDate startDate", "periodEndDate endDate", "size", "status"}
	common.ManageListRequestNoExpand(endpoint, periodColumns, flags.GenericFilters)
}
