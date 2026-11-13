// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/cmd"
)

// withFilterFlag only binds --filter to flags.GenericFilters; the filtering
// itself happens in ManageListRequest and ManageListRequestNoExpand. A command
// that assembles its own rows and calls display.RenderTable therefore accepts
// the flag, documents it, offers completion for it, and ignores it.
//
// Every assertion below is on the row the filter is supposed to REMOVE.
// Asserting the kept row is present would pass just as well with no filtering
// at all, which is how twenty of these went unnoticed.

func (ms *MockSuite) TestVpsRestorePointsAreFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/vps/vps-12345/automatedBackup/restorePoints?state=available",
		httpmock.NewStringResponder(200, `["2026-08-01T00:00:00Z","2026-08-20T00:00:00Z"]`))

	out, err := cmd.Execute("vps", "automated-backup", "list-restore-points", "vps-12345",
		"--filter", `restorePoint=~"2026-08-20"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("2026-08-01")), "the excluded restore point must be gone")
	assert.Cmp(out, td.Contains("2026-08-20"))
}

func (ms *MockSuite) TestKubeIPRestrictionsAreFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/cloud/project/fakeProjectID/kube/kube-12345/ipRestrictions",
		httpmock.NewStringResponder(200, `["203.0.113.0/24","198.51.100.7/32"]`))

	out, err := cmd.Execute("cloud", "managed-kubernetes", "ip-restrictions", "list", "kube-12345",
		"--cloud-project", "fakeProjectID", "--filter", `ip=~"^203\\."`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("198.51.100.7")), "the excluded restriction must be gone")
	assert.Cmp(out, td.Contains("203.0.113.0"))
}

func (ms *MockSuite) TestWebhostingModuleCatalogIsFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/hosting/web/moduleList",
		httpmock.NewStringResponder(200, `[1,2]`))
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/hosting/web/moduleList/1",
		httpmock.NewStringResponder(200, `{"id":1,"name":"wordpress","branch":"6","version":"6.5","active":true,"latest":true}`))
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/hosting/web/moduleList/2",
		httpmock.NewStringResponder(200, `{"id":2,"name":"joomla","branch":"5","version":"5.1","active":true,"latest":true}`))

	out, err := cmd.Execute("webhosting", "module", "catalog", "list", "--filter", `name=="wordpress"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("joomla")), "the excluded module must be gone")
	assert.Cmp(out, td.Contains("wordpress"))
}

// This one does not render a table: it wraps its list in an object and hands it
// to a template, so the filter has to run BEFORE the wrapping — once the list is
// inside an object there are no rows left to select from.
func (ms *MockSuite) TestWebhostingOvhConfigCapabilitiesAreFiltered(assert, require *td.T) {
	httpmock.RegisterResponder("GET",
		"https://eu.api.ovh.com/v1/hosting/web/site.example/ovhConfigCapabilities",
		httpmock.NewStringResponder(200,
			`[{"version":"8.3","container":"jessie.i386"},{"version":"7.4","container":"stretch.i386"}]`))

	out, err := cmd.Execute("webhosting", "ovh-config", "capabilities", "site.example",
		"--filter", `version=="8.3"`)

	require.CmpNoError(err)
	assert.Cmp(out, td.Not(td.Contains("7.4")), "the excluded capability must be gone")
	assert.Cmp(out, td.Contains("8.3"))
}

// The other half of the fix: two commands render one object, so there was never
// anything for --filter to select. Removing the flag is the honest correction —
// applying it there would only have made the lie harder to see.
func (ms *MockSuite) TestCommandsRenderingOneObjectDoNotOfferFilter(assert, require *td.T) {
	// Asserted by running the command, not by looking the flag up on a FlagSet:
	// withFilterFlag registers on PersistentFlags, so Flags().Lookup("filter")
	// answers nil whether the flag is there or not — an assertion that passes
	// either way. Found by sabotage; the FlagSet version stayed green with the
	// flag put back.
	for _, path := range [][]string{
		{"cloud", "loadbalancer", "log", "list-kinds"},
		{"cloud", "storage", "object", "quota", "get"},
	} {
		found, _, err := cmd.GetRootCommand().Find(path)
		require.CmpNoError(err, "%v must resolve", path)
		require.Cmp(found.Name(), path[len(path)-1], "%v must resolve to its leaf", path)

		cmd.PostExecute()
		_, err = cmd.Execute(append(append([]string{}, path...), "GRA11", "--filter", `x=="y"`)...)

		require.CmpError(err, "%v", path)
		assert.Cmp(err.Error(), td.Contains("unknown flag: --filter"),
			"%v renders one object, so --filter must not be accepted at all", path)
	}
}
