// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/maxatome/tdhttpmock"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// Mock responses
var (
	mockServiceInfos = `{"serviceId": 12345}`

	mockSavingsPlansList = `[
		{
			"id": "sp-001",
			"displayName": "My Savings Plan",
			"flavor": "b3-8",
			"size": 2,
			"status": "ACTIVE",
			"periodEndDate": "2027-01-31",
			"periodEndAction": "REACTIVATE"
		},
		{
			"id": "sp-002",
			"displayName": "Rancher Plan",
			"flavor": "rancher standard",
			"size": 1,
			"status": "ACTIVE",
			"periodEndDate": "2026-12-31",
			"periodEndAction": "TERMINATE"
		}
	]`

	mockSavingsPlanGet = `{
		"id": "sp-001",
		"displayName": "My Savings Plan",
		"flavor": "b3-8",
		"size": 2,
		"status": "ACTIVE",
		"periodStartDate": "2026-01-31",
		"periodEndDate": "2027-01-31",
		"periodEndAction": "REACTIVATE",
		"startDate": "2026-01-31"
	}`

	mockSavingsPlanOffers = `[
		{"offerId": "offer-b3-8-1az-12m"},
		{"offerId": "offer-b3-8-1az-24m"}
	]`

	mockSavingsPlanOffers3AZ = `[
		{"offerId": "offer-b3-8-3az-12m"},
		{"offerId": "offer-b3-8-3az-24m"}
	]`

	mockSubscribeResult = `{
		"id": "sp-new-001",
		"displayName": "New Savings Plan",
		"flavor": "b3-8",
		"size": 2,
		"status": "PENDING"
	}`

	mockSimulateResult = `{
		"orderId": "order-12345",
		"prices": [
			{"duration": "P12M", "price": {"value": 100.00, "currencyCode": "EUR"}}
		]
	}`

	mockPeriods = `[
		{
			"id": "period-001",
			"periodStartDate": "2026-01-31",
			"periodEndDate": "2027-01-31",
			"size": 2,
			"status": "ACTIVE"
		}
	]`
)

func (ms *MockSuite) TestCloudSavingsPlanListCmd(assert, require *td.T) {
	// Mock service info endpoint
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	// Mock savings plans list endpoint
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed",
		httpmock.NewStringResponder(200, mockSavingsPlansList),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Contains(out, "sp-001")
	assert.Contains(out, "My Savings Plan")
	assert.Contains(out, "sp-002")
	assert.Contains(out, "Rancher Plan")
}

func (ms *MockSuite) TestCloudSavingsPlanListCmdJSONFormat(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed",
		httpmock.NewStringResponder(200, mockSavingsPlansList),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list", "--cloud-project", "fakeProjectID", "-o", "json")
	require.CmpNoError(err)

	var result []map[string]any
	require.CmpNoError(json.Unmarshal([]byte(out), &result))
	assert.Cmp(len(result), 2)
	assert.Cmp(result[0]["id"], "sp-001")
}

func (ms *MockSuite) TestCloudSavingsPlanGetCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001",
		httpmock.NewStringResponder(200, mockSavingsPlanGet),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "get", "--cloud-project", "fakeProjectID", "sp-001")
	require.CmpNoError(err)
	assert.Contains(out, "sp-001")
	assert.Contains(out, "My Savings Plan")
	assert.Contains(out, "b3-8")
}

func (ms *MockSuite) TestCloudSavingsPlanListOffersCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list-offers", "--cloud-project", "fakeProjectID")
	require.CmpNoError(err)
	assert.Contains(out, "offer-b3-8-1az-12m")
}

func (ms *MockSuite) TestCloudSavingsPlanListOffersWithProductCodeCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers),
	)

	// Mock 3AZ endpoint for filtering
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8+3AZ",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers3AZ),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list-offers", "--cloud-project", "fakeProjectID", "--product-code", "b3-8")
	require.CmpNoError(err)
	assert.Contains(out, "offer-b3-8-1az-12m")
}

func (ms *MockSuite) TestCloudSavingsPlanListOffers3AZCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8+3AZ",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers3AZ),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list-offers", "--cloud-project", "fakeProjectID", "--product-code", "b3-8", "--deployment-type", "3AZ")
	require.CmpNoError(err)
	assert.Contains(out, "offer-b3-8-3az-12m")
}

func (ms *MockSuite) TestCloudSavingsPlanSubscribeWithOfferIdCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribe/execute",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"displayName": "New Savings Plan",
			"offerId": "offer-b3-8-1az-12m",
			"size": 2
		}`)),
		httpmock.NewStringResponder(200, mockSubscribeResult),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "subscribe", "--cloud-project", "fakeProjectID",
		"--offer-id", "offer-b3-8-1az-12m",
		"--display-name", "New Savings Plan",
		"--size", "2")
	require.CmpNoError(err)
	assert.Contains(out, "Successfully subscribed")
	assert.Contains(out, "sp-new-001")
}

func (ms *MockSuite) TestCloudSavingsPlanSubscribeWithFlavorCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	// Mock offer lookup for 1AZ
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers),
	)

	// Mock 3AZ offers for filtering
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8+3AZ",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers3AZ),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribe/execute",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"displayName": "Flavor Based Plan",
			"offerId": "offer-b3-8-1az-12m",
			"size": 3
		}`)),
		httpmock.NewStringResponder(200, mockSubscribeResult),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "subscribe", "--cloud-project", "fakeProjectID",
		"--flavor", "b3-8",
		"--deployment-type", "1AZ",
		"--display-name", "Flavor Based Plan",
		"--size", "3")
	require.CmpNoError(err)
	assert.Contains(out, "Successfully subscribed")
}

func (ms *MockSuite) TestCloudSavingsPlanSubscribeWith3AZCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	// Mock offer lookup for 3AZ
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribable?productCode=b3-8+3AZ",
		httpmock.NewStringResponder(200, mockSavingsPlanOffers3AZ),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribe/execute",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"displayName": "3AZ Savings Plan",
			"offerId": "offer-b3-8-3az-12m",
			"size": 2
		}`)),
		httpmock.NewStringResponder(200, mockSubscribeResult),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "subscribe", "--cloud-project", "fakeProjectID",
		"--flavor", "b3-8",
		"--deployment-type", "3AZ",
		"--display-name", "3AZ Savings Plan",
		"--size", "2")
	require.CmpNoError(err)
	assert.Contains(out, "Successfully subscribed")
}

func (ms *MockSuite) TestCloudSavingsPlanSimulateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribe/simulate",
		tdhttpmock.JSONBody(td.SuperJSONOf(`{
			"displayName": "Simulated Plan",
			"offerId": "offer-b3-8-1az-12m",
			"size": 2
		}`)),
		httpmock.NewStringResponder(200, mockSimulateResult),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "simulate", "--cloud-project", "fakeProjectID",
		"--offer-id", "offer-b3-8-1az-12m",
		"--display-name", "Simulated Plan",
		"--size", "2")
	require.CmpNoError(err)
	assert.Contains(out, "order-12345")
	assert.Contains(out, "orderId")
}

func (ms *MockSuite) TestCloudSavingsPlanTerminateCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001/terminate",
		tdhttpmock.JSONBody(td.JSON(`{}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "terminate", "--cloud-project", "fakeProjectID", "sp-001")
	require.CmpNoError(err)
	assert.Contains(out, "scheduled for termination")
}

func (ms *MockSuite) TestCloudSavingsPlanSetRenewalCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001/changePeriodEndAction",
		tdhttpmock.JSONBody(td.JSON(`{"periodEndAction": "REACTIVATE"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "set-renewal", "--cloud-project", "fakeProjectID", "sp-001", "--action", "REACTIVATE")
	require.CmpNoError(err)
	assert.Contains(out, "Period end action changed")
}

func (ms *MockSuite) TestCloudSavingsPlanResizeCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPost,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001/changeSize",
		tdhttpmock.JSONBody(td.JSON(`{"size": 5}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "resize", "--cloud-project", "fakeProjectID", "sp-001", "--size", "5")
	require.CmpNoError(err)
	assert.Contains(out, "size changed")
}

func (ms *MockSuite) TestCloudSavingsPlanEditCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterMatcherResponder(http.MethodPut,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001",
		tdhttpmock.JSONBody(td.JSON(`{"displayName": "Updated Name"}`)),
		httpmock.NewStringResponder(200, `{}`),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "edit", "--cloud-project", "fakeProjectID", "sp-001", "--display-name", "Updated Name")
	require.CmpNoError(err)
	assert.Contains(out, "Updated")
}

func (ms *MockSuite) TestCloudSavingsPlanListPeriodsCmd(assert, require *td.T) {
	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/serviceInfos",
		httpmock.NewStringResponder(200, mockServiceInfos),
	)

	httpmock.RegisterResponder(http.MethodGet,
		"https://eu.api.ovh.com/v1/services/12345/savingsPlans/subscribed/sp-001/periods",
		httpmock.NewStringResponder(200, mockPeriods),
	)

	out, err := cmd.Execute("cloud", "savings-plan", "list-periods", "--cloud-project", "fakeProjectID", "sp-001")
	require.CmpNoError(err)
	assert.Contains(out, "2026-01-31")
	assert.Contains(out, "2027-01-31")
}
