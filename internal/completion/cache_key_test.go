// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package completion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"gopkg.in/ini.v1"
)

// withMockedAPI points the shared client at httpmock and gives the test its own
// cache directory, so nothing leaks between tests or onto the developer's disk.
func withMockedAPI(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// Same wiring as the command test suite: httpmock patches the default
	// transport, which is the one go-ovh builds its client on.
	httpmock.Activate(t)

	origClient, origProfile := httpLib.Client, flags.Profile
	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)
	httpLib.Client = client

	// go-ovh computes its clock delta before the first signed call; without
	// this the request never leaves and the test measures nothing.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))

	t.Cleanup(func() {
		httpLib.Client, flags.Profile = origClient, origProfile
	})
}

// The acceptance criterion of the requirement: two consecutive <tab> keystrokes
// must cost one API call, not two. Completion runs as a fresh process each
// time, which is why the cache lives on disk.
func TestServiceList_SecondCompletionHitsTheCache(t *testing.T) {
	withMockedAPI(t)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(200, `["ns3168421.ip-51-77-12.eu","ns5012993.ip-141-95-4.eu"]`),
	)

	complete := ServiceList("/v1/dedicated/server")

	first, _ := complete(nil, nil, "")
	second, _ := complete(nil, nil, "")

	td.Cmp(t, first, []string{"ns3168421.ip-51-77-12.eu", "ns5012993.ip-141-95-4.eu"})
	td.Cmp(t, second, first, "the cached answer is the one that was fetched")
	td.Cmp(t, httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/dedicated/server"], 1,
		"the second completion must not call the API again")
}

// A failed call must not be cached, otherwise a transient error would silence
// completion for the whole TTL.
func TestServiceList_FailureIsNotCached(t *testing.T) {
	withMockedAPI(t)
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/dedicated/server",
		httpmock.NewStringResponder(500, `{"message":"boom"}`),
	)

	complete := ServiceList("/v1/dedicated/server")
	complete(nil, nil, "")
	complete(nil, nil, "")

	td.Cmp(t, httpmock.GetCallCountInfo()["GET https://eu.api.ovh.com/v1/dedicated/server"], 2,
		"a failure must be retried rather than remembered")
}

// Profiles point at different accounts: one must never be served the other's
// fleet, however fresh the cache is.
func TestCacheKeyFor_IsScopedToTheProfile(t *testing.T) {
	origProfile := flags.Profile
	defer func() { flags.Profile = origProfile }()

	flags.Profile = ""
	withoutProfile := cacheKeyFor("/v1/dedicated/server")
	flags.Profile = "prod"
	withProfile := cacheKeyFor("/v1/dedicated/server")

	td.Cmp(t, withProfile, td.Not(withoutProfile))
}

// The flag is the one way of picking a profile that this cache used to see. The
// two others select an account just as effectively, and on a real keystroke
// they are the likely ones — a shell completes without anybody typing
// --profile. Both were sharing a key, and with it the previous account's
// identifiers, for the whole TTL.
func TestCacheKeyFor_IsScopedToTheProfileChosenByEnvironment(t *testing.T) {
	origProfile, origEnv := flags.Profile, os.Getenv("OVH_PROFILE")
	defer func() { flags.Profile = origProfile; os.Setenv("OVH_PROFILE", origEnv) }()

	flags.Profile = ""
	os.Unsetenv("OVH_PROFILE")
	plain := cacheKeyFor("/v1/dedicated/server")

	os.Setenv("OVH_PROFILE", "prod")
	fromEnv := cacheKeyFor("/v1/dedicated/server")
	os.Setenv("OVH_PROFILE", "staging")
	otherFromEnv := cacheKeyFor("/v1/dedicated/server")

	td.Cmp(t, fromEnv, td.Not(plain), "OVH_PROFILE must reach the key")
	td.Cmp(t, otherFromEnv, td.Not(fromEnv), "and two of them must not collide")
}

func TestCacheKeyFor_IsScopedToTheProfileChosenByConfig(t *testing.T) {
	origProfile, origEnv := flags.Profile, os.Getenv("OVH_PROFILE")
	origConfig := flags.CliConfig
	defer func() {
		flags.Profile = origProfile
		os.Setenv("OVH_PROFILE", origEnv)
		flags.CliConfig = origConfig
	}()

	flags.Profile = ""
	os.Unsetenv("OVH_PROFILE")
	flags.CliConfig = nil
	plain := cacheKeyFor("/v1/dedicated/server")

	cfg := ini.Empty()
	cfg.Section("default").Key("profile").SetValue("prod")
	flags.CliConfig = cfg

	td.Cmp(t, cacheKeyFor("/v1/dedicated/server"), td.Not(plain),
		"the profile named by the configuration file must reach the key")
}

// Two endpoints must not share a cache entry, including when their readable
// prefixes collide after slugging.
func TestCacheKeyFor_DistinguishesEndpoints(t *testing.T) {
	td.Cmp(t, cacheKeyFor("/v1/dedicated/server"), td.Not(cacheKeyFor("/v1/dedicated/nasha")))
	td.Cmp(t, cacheKeyFor("/v1/a/b"), td.Not(cacheKeyFor("/v1/a-b")),
		"slugging must not merge two different endpoints")
}

// The key becomes a file name, so it must carry nothing a path cannot hold.
func TestCacheKeyFor_IsSafeAsAFileName(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	key := cacheKeyFor("/v1/cloud/project/aa..bb/instance?flavor=b2-7")

	td.Cmp(t, strings.ContainsAny(key, `/\:*?"<>| `), false, "key %q must be path-safe", key)
	td.Cmp(t, strings.Contains(key, ".."), false, "key %q must not walk up a directory", key)

	writeCachedSuggestions(key, []string{"i-1"})
	_, err := os.Stat(filepath.Join(completionCacheDir(), key))
	td.CmpNoError(t, err, "the key must be usable as a file name")
}

// The profile is not the only thing that selects an account: environment
// variables override it, and in legacy mode there is no profile at all. Two
// credentials must therefore never share a cache entry.
func TestCacheKeyFor_IsScopedToTheCredentials(t *testing.T) {
	origClient := httpLib.Client
	defer func() { httpLib.Client = origClient }()

	newKey := func(consumerKey string) string {
		client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", consumerKey)
		td.Require(t).CmpNoError(err)
		httpLib.Client = client
		return cacheKeyFor("/v1/dedicated/server")
	}

	td.Cmp(t, newKey("consumer_a"), td.Not(newKey("consumer_b")),
		"two accounts must not share suggestions")
}

// Regions are separate fleets behind the same profile, and OVH_ENDPOINT is
// enough to move between them.
func TestCacheKeyFor_IsScopedToTheRegion(t *testing.T) {
	origClient := httpLib.Client
	defer func() { httpLib.Client = origClient }()

	newKey := func(endpoint string) string {
		client, err := ovh.NewClient(endpoint, "app_key", "app_secret", "consumer_key")
		td.Require(t).CmpNoError(err)
		httpLib.Client = client
		return cacheKeyFor("/v1/dedicated/server")
	}

	td.Cmp(t, newKey("ovh-eu"), td.Not(newKey("ovh-ca")),
		"two regions must not share suggestions")
}
