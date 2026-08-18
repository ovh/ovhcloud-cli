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
	Client = client
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
