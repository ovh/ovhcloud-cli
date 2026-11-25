// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/flags"
)

// roundTripperFunc turns a function into an http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// debugRoundTrip runs one request through the debug transport and returns
// everything it logged.
func debugRoundTrip(t *testing.T, url, responseBody string) string {
	t.Helper()

	origDebug := flags.Debug
	flags.Debug = true
	defer func() { flags.Debug = origDebug }()

	var logged bytes.Buffer
	origWriter := log.Writer()
	log.SetOutput(&logged)
	defer log.SetOutput(origWriter)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(`{"id":"secret-id-1"}`))
	td.Require(t).CmpNoError(err)

	tr := &transport{
		name: "test",
		transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Request:    r,
			}, nil
		}),
	}

	_, err = tr.RoundTrip(req)
	td.Require(t).CmpNoError(err)

	return logged.String()
}

// --debug is a diagnostic switch, not a way around the masking that --reveal
// governs: the body of a credential endpoint must never reach the log.
func TestDebugTransport_DoesNotLogCredentialBodies(t *testing.T) {
	for _, endpoint := range []string{
		"https://eu.api.ovh.com/v1/secret/retrieve",
		"https://eu.api.ovh.com/v1/dedicated/server/ns1234/authenticationSecret",
	} {
		logged := debugRoundTrip(t, endpoint, `{"secret":"Tk9uY2VQYXNzMTIz"}`)

		td.Cmp(t, logged, td.Not(td.Contains("Tk9uY2VQYXNzMTIz")),
			"the response body of %s must not be logged", endpoint)
		td.Cmp(t, logged, td.Not(td.Contains("secret-id-1")),
			"nor the request body of %s", endpoint)
		td.Cmp(t, logged, td.Contains("200"), "the exchange is still traced")
	}
}

// Every other endpoint keeps its full debug dump: the redaction must stay
// narrow, otherwise --debug loses its purpose.
func TestDebugTransport_StillLogsOrdinaryBodies(t *testing.T) {
	logged := debugRoundTrip(t,
		"https://eu.api.ovh.com/v1/dedicated/server/ns1234", `{"state":"ok"}`)

	td.Cmp(t, logged, td.Contains(`"state": "ok"`), "the debug logger pretty-prints, so the body appears spaced")
}
