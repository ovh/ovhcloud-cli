// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package apicall

import (
	"encoding/json"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/flags"
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

// An identifier above 2^53 must reach the API as it was written.
//
// Decoding into `any` turns every JSON number into a float64, and float64 has
// 53 bits of mantissa: 9007199254740993 comes back as 9007199254740992 and is
// signed and sent as a different object. This is the same defect the edit
// commands carried, and it is worse here — this command exists to send exactly
// what the operator wrote, with the CLI knowing nothing about the payload.
func TestReadBodyDoesNotRoundLargeIntegers(t *testing.T) {
	assert := td.Assert(t)

	origRead, origFile := readFile, flags.ParametersFile
	readFile = func(string) ([]byte, error) {
		return []byte(`{"id": 9007199254740993, "ratio": 0.30000000000000004}`), nil
	}
	flags.ParametersFile = "payload.json"
	defer func() { readFile, flags.ParametersFile = origRead, origFile }()

	body, err := readBody()
	assert.CmpNoError(err)

	// Marshalled the way the client will marshal it before signing.
	sent, err := json.Marshal(body)
	assert.CmpNoError(err)
	assert.Cmp(string(sent), td.Contains("9007199254740993"))
	assert.Cmp(string(sent), td.Not(td.Contains("9007199254740992")))
	assert.Cmp(string(sent), td.Contains("0.30000000000000004"),
		"and a float is not renormalised either")
}
