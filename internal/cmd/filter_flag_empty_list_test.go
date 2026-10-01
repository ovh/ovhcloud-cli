// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// Without --filter, routing a list through FilteredRows must hand it back as it
// came. FilterLines builds its result from a nil slice, so an empty answer used
// to come back nil, and `-o json` printed "entries": null where the API had said
// [] — a change of shape for every script reading the field, on a run that asked
// for no filtering at all.
func (ms *MockSuite) TestAnEmptyListStaysAnEmptyListWithoutAFilter(assert, require *td.T) {
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/hosting/web/fakeHosting/ovhConfigCapabilities",
		httpmock.NewStringResponder(200, `[]`))

	out, err := cmd.Execute("webhosting", "ovh-config", "capabilities", "fakeHosting", "-o", "json")

	require.CmpNoError(err)
	assert.Cmp(out, td.Contains(`"entries": []`))
	assert.Cmp(out, td.Not(td.Contains("null")))
}
