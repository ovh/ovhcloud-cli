// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// A server resolves to four or five billable services and only the machine
// carries the route. The components carry none.
func TestOnlyTheMachineCarriesTheDedicatedServerRoute(t *testing.T) {
	assert := td.Assert(t)

	machine := billableService{}
	machine.Route = &struct {
		Path string `json:"path"`
	}{Path: dedicatedServerRoute}

	component := billableService{}

	other := billableService{}
	other.Route = &struct {
		Path string `json:"path"`
	}{Path: "/vps/{serviceName}"}

	assert.Cmp(machine.isTheMachine(), true)
	assert.Cmp(component.isTheMachine(), false, "a component carries no route at all")
	assert.Cmp(other.isTheMachine(), false, "another product's route is not this machine")
}

func TestALineIsNamedTheWayAnInvoiceNamesIt(t *testing.T) {
	assert := td.Assert(t)

	var withInvoiceName billableService
	withInvoiceName.Billing.Plan.InvoiceName = "ADVANCE-1 | AMD EPYC 4245P"
	withInvoiceName.Billing.Plan.Code = "26adv01-v1014"
	assert.Cmp(withInvoiceName.label(), "ADVANCE-1 | AMD EPYC 4245P")

	var withProductOnly billableService
	withProductOnly.Billing.Plan.Code = "softraid-2x960nvme"
	withProductOnly.Resource.Product.Description = "2x SSD NVMe 960GB"
	assert.Cmp(withProductOnly.label(), "2x SSD NVMe 960GB")

	var codeOnly billableService
	codeOnly.Billing.Plan.Code = "26adv01-v1014"
	assert.Cmp(codeOnly.label(), "26adv01-v1014")
}

func TestAMachineWithItsOwnRenewalSaysWhenAndHow(t *testing.T) {
	assert := td.Assert(t)

	var machine billableService
	machine.Billing.Renew = &struct {
		Current *struct {
			Mode     string `json:"mode"`
			NextDate string `json:"nextDate"`
			Period   string `json:"period"`
		} `json:"current"`
	}{Current: &struct {
		Mode     string `json:"mode"`
		NextDate string `json:"nextDate"`
		Period   string `json:"period"`
	}{Mode: "automatic", NextDate: "2026-09-01T10:29:33Z"}}

	assert.Cmp(renewalPhrase(machine), "Renews automatic on 2026-09-01.")
}

// The six child services of the account measured carry no renewal block, and
// serviceInfos.renew.automatic contradicts itself on them (PUBM-55135). Saying
// the parent decides is true; deriving a boolean would not be.
func TestAChildServiceSaysItsParentCarriesTheRenewal(t *testing.T) {
	assert := td.Assert(t)

	parent := int64(133558145)
	machine := billableService{ParentServiceID: &parent}

	phrase := renewalPhrase(machine)

	assert.Cmp(phrase, td.Contains("133558145"))
	assert.Cmp(phrase, td.Contains("carried by parent service"))
	assert.Cmp(phrase, td.Not(td.Contains("automatic")), "no renewal mode may be invented here")
}

// No renewal block and no parent either: that is a third thing, and it is not
// "renews automatically".
func TestAServiceWithNeitherRenewalNorParentSaysSo(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(renewalPhrase(billableService{}), "This service declares no renewal.")
}

func TestACommitmentIsNamedWithItsEndDate(t *testing.T) {
	assert := td.Assert(t)

	var machine billableService
	machine.Billing.Renew = &struct {
		Current *struct {
			Mode     string `json:"mode"`
			NextDate string `json:"nextDate"`
			Period   string `json:"period"`
		} `json:"current"`
	}{Current: &struct {
		Mode     string `json:"mode"`
		NextDate string `json:"nextDate"`
		Period   string `json:"period"`
	}{Mode: "automatic", NextDate: "2026-09-01T00:00:00Z"}}
	machine.Billing.Engagement = &struct {
		EndDate string `json:"endDate"`
	}{EndDate: "2027-02-01T00:00:00Z"}

	assert.Cmp(renewalPhrase(machine), td.Contains("committed until 2027-02-01"))
}

func TestATotalKeepsItsCurrency(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(money(104.99, "EUR"), "104.99 EUR")
	assert.Cmp(money(1106, "EUR"), "1106.00 EUR")
	assert.Cmp(money(12.5, ""), "12.50")
}
