// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package http

import "net/http"

// customHeadersTransport is an http.RoundTripper middleware that injects
// user-configured custom HTTP headers on every request. It is used to
// support custom API endpoints that require extra headers (e.g. an internal
// routing header) that the built-in OVHcloud regions don't need.
type customHeadersTransport struct {
	next    http.RoundTripper
	headers map[string]string
}

func (t *customHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.headers) == 0 {
		return t.next.RoundTrip(req)
	}

	// Clone the request so we don't mutate the caller's original.
	req = req.Clone(req.Context())
	for name, value := range t.headers {
		req.Header.Set(name, value)
	}

	return t.next.RoundTrip(req)
}

// newCustomHeadersTransport wraps the given RoundTripper with injection of
// the given custom headers on every request.
func newCustomHeadersTransport(next http.RoundTripper, headers map[string]string) http.RoundTripper {
	return &customHeadersTransport{next: next, headers: headers}
}
