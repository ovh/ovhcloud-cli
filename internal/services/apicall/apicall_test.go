// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package apicall

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// The point of accepting several shapes is that a user reading the API
// documentation types what it shows, which carries no version prefix.
func TestNormalizePath(t *testing.T) {
	for _, tc := range []struct{ given, want string }{
		{"dedicated/server", "/v1/dedicated/server"},
		{"/dedicated/server", "/v1/dedicated/server"},
		{"/v1/dedicated/server", "/v1/dedicated/server"},

		// /v2 exists; prefixing it with /v1 would call a different API.
		{"/v2/publicCloud/project", "/v2/publicCloud/project"},

		// A path whose first segment merely starts with "v1" is not versioned.
		{"/v1beta/thing", "/v1/v1beta/thing"},
	} {
		td.Cmp(t, normalizePath(tc.given), tc.want, "normalizePath(%q)", tc.given)
	}
}
