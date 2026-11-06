// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// Completion skips PersistentPreRun, so a <tab> starts with no API client at
// all. That is the path the cache key has to be right on: derived any earlier
// than the client, it carries no account and every account on the machine
// shares one cache entry.
//
// This test therefore never sets httpLib.Client itself — it starts from nil,
// the way a shell does, and switches account through the environment.
func TestServiceList_TwoAccountsDoNotShareTheCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	httpmock.Activate(t)

	origClient, origProfile := httpLib.Client, flags.Profile
	t.Cleanup(func() { httpLib.Client, flags.Profile = origClient, origProfile })

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(200, `["ns3168421.ip-51-77-12.eu"]`))

	t.Setenv("OVH_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_APPLICATION_KEY", "app_key")
	t.Setenv("OVH_APPLICATION_SECRET", "app_secret")

	completeAs := func(consumerKey string) []string {
		t.Setenv("OVH_CONSUMER_KEY", consumerKey)
		httpLib.Client = nil // what the shell hands us on every <tab>
		suggestions, _ := ServiceList("/v1/dedicated/server")(nil, nil, "")
		return suggestions
	}

	completeAs("consumer_a")
	callsAfterA := httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/dedicated/server"]

	completeAs("consumer_b")
	callsAfterB := httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/dedicated/server"]

	td.Cmp(t, callsAfterA, 1, "the first account must be fetched")
	td.Cmp(t, callsAfterB, 2,
		"the second account must be fetched too, not served the first one's cache")
}

// And the cache must still work for one account, or the fix above would have
// bought correctness by disabling it.
func TestServiceList_SameAccountStillHitsTheCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	httpmock.Activate(t)

	origClient, origProfile := httpLib.Client, flags.Profile
	t.Cleanup(func() { httpLib.Client, flags.Profile = origClient, origProfile })

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(200, `["ns3168421.ip-51-77-12.eu"]`))

	t.Setenv("OVH_ENDPOINT", "ovh-eu")
	t.Setenv("OVH_APPLICATION_KEY", "app_key")
	t.Setenv("OVH_APPLICATION_SECRET", "app_secret")
	t.Setenv("OVH_CONSUMER_KEY", "consumer_a")

	complete := ServiceList("/v1/dedicated/server")

	httpLib.Client = nil
	first, _ := complete(nil, nil, "")
	httpLib.Client = nil
	second, _ := complete(nil, nil, "")

	td.Cmp(t, second, first)
	td.Cmp(t, httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/dedicated/server"], 1,
		"the second completion of the same account must not call the API again")
}
