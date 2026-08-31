// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// Regression test for a bug where logging in with a custom endpoint stored
// the picker's menu label ("Custom endpoint") instead of the URL typed by
// the user, which made serviceconfig.SetEndpoint reject it with
// `given url has an invalid scheme, only "http" and "https" are allowed`.
func TestLegacyEndpointArg_CustomEndpoint(t *testing.T) {
	got := legacyEndpointArg("Custom endpoint", "https://api.eu.ovhcloud.com/1.0", true)
	td.Cmp(t, got, "https://api.eu.ovhcloud.com/1.0")
}

func TestLegacyEndpointArg_Region(t *testing.T) {
	got := legacyEndpointArg("eu", "ovh-eu", false)
	td.Cmp(t, got, "EU")
}

func TestParseHeadersInput(t *testing.T) {
	got := parseHeadersInput("X-Routing-Key: internal-build-eu, X-Debug-Bypass=true")
	td.Cmp(t, got, map[string]string{
		"X-Routing-Key":  "internal-build-eu",
		"X-Debug-Bypass": "true",
	})
}

func TestParseHeadersInput_Blank(t *testing.T) {
	td.Cmp(t, parseHeadersInput(""), map[string]string{})
	td.Cmp(t, parseHeadersInput("   "), map[string]string{})
}

func TestParseHeadersInput_SkipsMalformedPairs(t *testing.T) {
	got := parseHeadersInput("no-separator-here, X-Valid: ok, : no-name")
	td.Cmp(t, got, map[string]string{"X-Valid": "ok"})
}
