// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/ovhcloud-cli/internal/flags"
)

// TestFetchObjectsParallel_IgnoredErrorLogging verifies that when ignoreErrors
// is true, an expected per-item error (e.g. a feature not available in a given
// region) does not pollute normal output: it is only logged in debug mode
// (issue #173). The error must never be fatal.
func TestFetchObjectsParallel_IgnoredErrorLogging(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)

	previousClient := Client
	Client = NewAPIClient(client)
	defer func() { Client = previousClient }()

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	// GOOD region answers normally.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/region/GOOD/floatingip",
		httpmock.NewStringResponder(200, "[]"))
	// BAD region rejects the call because the feature is not available there.
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v1/cloud/region/BAD/floatingip",
		httpmock.NewStringResponder(400, `{"message":"floating ip feature is not available on region BAD"}`))

	run := func(debug bool) string {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		previousDebug := flags.Debug
		flags.Debug = debug
		defer func() { flags.Debug = previousDebug }()

		// The ignored error must never make the call fail.
		_, err := FetchObjectsParallel[[]any]("/v1/cloud/region/%s/floatingip", []any{"GOOD", "BAD"}, true)
		td.Require(t).CmpNoError(err)

		return buf.String()
	}

	// Without debug: the expected per-region error stays silent.
	td.Cmp(t, strings.Contains(run(false), "error fetching"), false,
		"ignored errors must not be logged when debug is disabled")

	// With debug: the error is surfaced for troubleshooting.
	td.Cmp(t, strings.Contains(run(true), "error fetching"), true,
		"ignored errors must be logged when debug is enabled")
}

// TestAPIErrorDetailsMessage verifies that an API error shows the reason the API
// gives in details.message, which go-ovh leaves out of the error text, and stays
// an *ovh.APIError for the callers that check its code.
func TestAPIErrorDetailsMessage(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client, err := ovh.NewClient("ovh-eu", "app_key", "app_secret", "consumer_key")
	td.Require(t).CmpNoError(err)

	previousClient := Client
	Client = NewAPIClient(client)
	defer func() { Client = previousClient }()

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	httpmock.RegisterResponder("PUT", "https://eu.api.ovh.com/v2/topic",
		httpmock.NewStringResponder(409, `{"class":"Client::Conflict::BadRequest","message":"Bad Request","details":{"message":"Topic is not yet available"}}`))
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/v2/missing",
		httpmock.NewStringResponder(404, `{"class":"Client::NotFound","message":"Not Found"}`))

	err = Client.Put("/v2/topic", map[string]any{}, nil)
	td.Cmp(t, err, td.Smuggle(func(err error) string { return err.Error() },
		`OVHcloud API error (status code 409): Client::Conflict::BadRequest: "Bad Request: Topic is not yet available"`))

	err = Client.Get("/v2/missing", nil)
	td.Cmp(t, err, td.Smuggle(func(err error) string { return err.Error() },
		`OVHcloud API error (status code 404): Client::NotFound: "Not Found"`))
	apiErr, ok := err.(*ovh.APIError)
	td.Require(t).True(ok, "the error must stay an *ovh.APIError")
	td.Cmp(t, apiErr.Code, 404)
}

// The index alignment of FetchObjectsParallel is a contract that call sites in
// internal/services/browser depend on: they pair objects[i] with a name held in
// a parallel slice, so a result compacted to drop failures would attach one
// region's details to another region's name.
//
// The padding reads like a bug — a slice with nil holes in it — which is exactly
// why it needs a test. This one fails if anyone makes the result dense.
func TestFetchObjectsParallelKeepsFailedItemsInPlace(t *testing.T) {
	httpmock.Activate(t)
	client, err := ovh.NewClient("ovh-eu", "k", "s", "c")
	td.Require(t).CmpNoError(err)
	saved := Client
	Client = NewAPIClient(client)
	t.Cleanup(func() { Client = saved })

	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/auth/time",
		httpmock.NewStringResponder(200, "0"))
	for _, id := range []string{"a", "c"} {
		httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/thing/"+id,
			httpmock.NewStringResponder(200, `{"name":"`+id+`"}`))
	}
	httpmock.RegisterResponder("GET", "https://eu.api.ovh.com/1.0/thing/b",
		httpmock.NewStringResponder(500, `{"message":"nope"}`))

	objects, err := FetchObjectsParallel[map[string]any]("/thing/%s", []any{"a", "b", "c"}, true)

	td.Require(t).CmpNoError(err, "ignoreErrors means the batch still succeeds")
	td.Require(t).Cmp(len(objects), 3, "one slot per id, however many failed")
	td.Cmp(t, objects[0]["name"], "a")
	td.CmpNil(t, objects[1], "the failed id keeps its slot, holding the zero value")
	td.Cmp(t, objects[2]["name"], "c", "and the ids after it are not shifted up")
}
