// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestCustomHeadersTransport_AppliesHeaders(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newCustomHeadersTransport(recorder, map[string]string{
		"X-Routing-Key": "internal-build-eu",
	})

	req, _ := http.NewRequest(http.MethodGet, "https://api.eu.ovhcloud.com/1.0/me", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	td.Cmp(t, recorder.req.Header.Get("X-Routing-Key"), "internal-build-eu")
}

func TestCustomHeadersTransport_NoopWhenEmpty(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newCustomHeadersTransport(recorder, nil)

	req, _ := http.NewRequest(http.MethodGet, "https://eu.api.ovh.com/v1/me", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	// Passed through untouched: same *http.Request instance, no headers added.
	td.Cmp(t, recorder.req, req)
}

func TestCustomHeadersTransport_DoesNotMutateOriginal(t *testing.T) {
	recorder := &recordingTransport{}
	transport := newCustomHeadersTransport(recorder, map[string]string{"X-Routing-Key": "internal-build-eu"})

	req, _ := http.NewRequest(http.MethodGet, "https://api.eu.ovhcloud.com/1.0/me", nil)
	_, err := transport.RoundTrip(req)
	td.Require(t).CmpNoError(err)
	// The original request must not have been mutated.
	td.Cmp(t, req.Header.Get("X-Routing-Key"), "")
}
