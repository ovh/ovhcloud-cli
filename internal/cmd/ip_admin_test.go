// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"encoding/json"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

const testRipeBlock = "https://eu.api.ovh.com/v1/ip/203.0.113.32%2F30"

// The API takes the whole RIPE object, so a request carrying only one field
// would publish an empty other one — in the public registry, and without
// anybody asking for it.
func (ms *MockSuite) TestIpRipeSetKeepsTheFieldItWasNotAskedToChange(assert, require *td.T) {
	httpmock.RegisterResponder("GET", testRipeBlock+"/ripe",
		httpmock.NewStringResponder(200, `{"netname":"OVH_251564576","description":"Failover Ips"}`))

	var sent map[string]any
	httpmock.RegisterResponder("PUT", testRipeBlock+"/ripe",
		func(req *http.Request) (*http.Response, error) {
			json.NewDecoder(req.Body).Decode(&sent)
			return httpmock.NewStringResponse(200, `null`), nil
		})

	_, err := cmd.Execute("ip", "ripe", "set", "203.0.113.32/30",
		"--description", "Renamed", "--yes")

	require.CmpNoError(err)
	assert.Cmp(sent["description"], "Renamed")
	assert.Cmp(sent["netname"], "OVH_251564576", "the netname must survive a description-only change")
}

func (ms *MockSuite) TestIpRipeSetRefusesToChangeNothing(assert, require *td.T) {
	_, err := cmd.Execute("ip", "ripe", "set", "203.0.113.32/30", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--netname"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0, "a change of nothing costs no request")
}

// The API accepts a change-contact body with no contact in it and answers an
// empty list of tasks: a command that looks like it worked and changed nothing.
func (ms *MockSuite) TestIpChangeContactRefusesAnEmptyRequest(assert, require *td.T) {
	_, err := cmd.Execute("ip", "service", "change-contact", "ip-192.0.2.1", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--admin"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// The preview names every field and withholds one value: the token is a
// single-use credential and --dry-run is the command whose output ends up in a
// terminal buffer or a pipeline log.
func (ms *MockSuite) TestIpConfirmTerminationFingerprintsTheToken(assert, require *td.T) {
	out, err := cmd.Execute("ip", "service", "confirm-termination", "ip-192.0.2.1",
		"SuperSecretToken42", "--reason", "TOO_EXPENSIVE", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("18 characters"))
	assert.Cmp(out, td.Contains("TOO_EXPENSIVE"))
	assert.Cmp(out, td.Not(td.Contains("SuperSecretToken42")), "the token itself must not be printed")
}

// cobra counts the arguments, it does not look at them: an empty token
// satisfies ExactArgs(2) and would travel all the way to a 400.
func (ms *MockSuite) TestIpConfirmTerminationRefusesAnEmptyToken(assert, require *td.T) {
	_, err := cmd.Execute("ip", "service", "confirm-termination", "ip-192.0.2.1", "  ", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("ip service terminate"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// The fourteen accepted reasons are read from the specification embedded in
// the binary, not transcribed: a list copied into Go drifts the day the API
// gains a fifteenth, silently, into a 400 nobody can explain.
func (ms *MockSuite) TestIpConfirmTerminationRejectsAnUnknownReason(assert, require *td.T) {
	_, err := cmd.Execute("ip", "service", "confirm-termination", "ip-192.0.2.1",
		"token", "--reason", "BECAUSE", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("TOO_EXPENSIVE"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// The eight licence routes are asked in one command, and an address carrying
// none says so rather than printing an empty table.
func (ms *MockSuite) TestIpLicensesAsksEveryProduct(assert, require *td.T) {
	base := "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/license/"
	for _, product := range []string{"cloudLinux", "cpanel", "directadmin", "plesk",
		"sqlserver", "virtuozzo", "windows", "worklight"} {
		body := `[]`
		if product == "plesk" {
			body = `["plesk-ca-75831"]`
		}
		httpmock.RegisterResponder("GET", base+product, httpmock.NewStringResponder(200, body))
	}

	out, err := cmd.Execute("ip", "licenses", "192.0.2.0/24")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("plesk-ca-75831"))
	assert.Cmp(httpmock.GetTotalCallCount(), 8, "one request per licence route")
}

func (ms *MockSuite) TestIpLicensesSaysWhenThereIsNone(assert, require *td.T) {
	base := "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/license/"
	for _, product := range []string{"cloudLinux", "cpanel", "directadmin", "plesk",
		"sqlserver", "virtuozzo", "windows", "worklight"} {
		httpmock.RegisterResponder("GET", base+product, httpmock.NewStringResponder(200, `[]`))
	}

	out, err := cmd.Execute("ip", "licenses", "192.0.2.0/24")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("No licence"))
}

// This route answers HTTP 500 for every IPv4 block. The error is reported as
// itself, with what the route says about its own scope.
func (ms *MockSuite) TestIpDelegationExplainsTheIPv4Failure(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/delegation",
		httpmock.NewStringResponder(500, `{"class":"Server::InternalServerError","message":"Internal server error"}`))

	_, err := cmd.Execute("ip", "delegation", "list", "192.0.2.0/24")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("IPv6 subnets"))
}

// The write needs a value only the preview route can give, so the refusal
// names that route rather than the flag alone.
func (ms *MockSuite) TestIpByoipAggregateNamesThePreviewCommand(assert, require *td.T) {
	_, err := cmd.Execute("ip", "byoip", "aggregate", "192.0.2.0/24", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("ovhcloud ip byoip aggregations"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestIpByoipSliceNamesThePreviewCommand(assert, require *td.T) {
	_, err := cmd.Execute("ip", "byoip", "slice", "192.0.2.0/24", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("ovhcloud ip byoip slices"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// The token is withheld in every output, including the one a pipeline reads.
func (ms *MockSuite) TestIpMigrationTokenIsWithheldUnderJson(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/migrationToken",
		httpmock.NewStringResponder(200, `{"customerId":"ab12345-ovh","token":"ExampleTokenValue123"}`))

	out, err := cmd.Execute("ip", "migration-token", "get", "192.0.2.0/24", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("ab12345-ovh"))
	assert.Cmp(out, td.Not(td.Contains("ExampleTokenValue123")), "the token must not reach a log")
}

func (ms *MockSuite) TestIpMigrationTokenCreateNeedsACustomer(assert, require *td.T) {
	_, err := cmd.Execute("ip", "migration-token", "create", "192.0.2.0/24", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--customer-id"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// Registering an IP to another organisation is published to the regional
// registry, so it takes the strongest guard the CLI has rather than a yes/no.
func (ms *MockSuite) TestIpChangeOrgPreviewsTheCall(assert, require *td.T) {
	out, err := cmd.Execute("ip", "change-org", "192.0.2.0/24", "RIPE_66451", "--dry-run")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains("/changeOrg"))
	for call := range httpmock.GetCallCountInfo() {
		assert.Cmp(call, td.Not(td.HasPrefix("POST")))
	}
}

// --size -1 satisfied "a value was given" and travelled to the API as a prefix
// length, through a confirmation offering to slice into /-1 blocks.
func (ms *MockSuite) TestIpByoipSliceRefusesAnImpossiblePrefix(assert, require *td.T) {
	_, err := cmd.Execute("ip", "byoip", "slice", "192.0.2.0/24", "--size", "-1", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("between 1 and 128"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// A blank value is not a value: --netname "  " passed the "nothing to change"
// guard and would have published a blank netname to the public registry.
func (ms *MockSuite) TestIpRipeSetRefusesABlankChange(assert, require *td.T) {
	_, err := cmd.Execute("ip", "ripe", "set", "203.0.113.32/30", "--netname", "   ", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("nothing to change"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

func (ms *MockSuite) TestIpMigrationTokenCreateRefusesABlankCustomer(assert, require *td.T) {
	_, err := cmd.Execute("ip", "migration-token", "create", "192.0.2.0/24",
		"--customer-id", "  ", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("--customer-id"))
	assert.Cmp(httpmock.GetTotalCallCount(), 0)
}

// Only a 404 means there is no token. Saying so after a 403 or a 500 sends the
// operator to `migration-token create`, which fails for the reason the message
// just hid.
func (ms *MockSuite) TestIpMigrationTokenGetDoesNotBlameAMissingTokenForEveryFailure(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/migrationToken",
		httpmock.NewStringResponder(403, `{"class":"Client::Forbidden","message":"This call has not been granted"}`))

	_, err := cmd.Execute("ip", "migration-token", "get", "192.0.2.0/24")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to read the migration token"))
	assert.Cmp(err.Error(), td.Not(td.Contains("migration-token create")), "the wrong remedy must not be offered")
}

func (ms *MockSuite) TestIpMigrationTokenGetStillSaysWhenThereIsNone(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/migrationToken",
		httpmock.NewStringResponder(404, `{"message":"The requested object (migrationToken) does not exist"}`))

	_, err := cmd.Execute("ip", "migration-token", "get", "192.0.2.0/24")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no migration token exists"))
	assert.Cmp(err.Error(), td.Contains("migration-token create"))
}

// The single-object route fails the same way as the list, and going through
// the generic helper meant the failure lost the sentence that explains it.
func (ms *MockSuite) TestIpDelegationGetExplainsTheFailureToo(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/ip/192.0.2.0%2F24/delegation/ns1.example.com",
		httpmock.NewStringResponder(500, `{"class":"Server::InternalServerError","message":"Internal server error"}`))

	_, err := cmd.Execute("ip", "delegation", "get", "192.0.2.0/24", "ns1.example.com")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("IPv6 subnets"))
}

// The same note serves four call sites, two of which are writes. Fixed to
// "read", a refused POST printed "failed to read the reverse delegation of X",
// which sends the operator to check a right they never used.
func (ms *MockSuite) TestIpDelegationAddReportsAWriteAsAWrite(assert, require *td.T) {
	httpmock.RegisterResponder("POST", "https://eu.api.ovh.com/v1/ip/2001:db8::%2F64/delegation",
		httpmock.NewStringResponder(403, `{"message":"This call has not been granted"}`))

	_, err := cmd.Execute("ip", "delegation", "add", "2001:db8::/64", "ns1.example", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Not(td.Contains("failed to read")), "a refused POST is not a failed read")
	assert.Cmp(err.Error(), td.Contains("failed to add to the reverse delegation"))
}

func (ms *MockSuite) TestIpDelegationRemoveReportsAWriteAsAWrite(assert, require *td.T) {
	httpmock.RegisterResponder("DELETE", `=~^https://eu\.api\.ovh\.com/v1/ip/2001:db8::%2F64/delegation/.+`,
		httpmock.NewStringResponder(403, `{"message":"This call has not been granted"}`))

	_, err := cmd.Execute("ip", "delegation", "remove", "2001:db8::/64", "ns1.example", "--yes")

	require.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("failed to remove from the reverse delegation"))
}
