// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// 139 of the rules on the account measured are on "/*". A rule count would
// have said "4 rules" for a key that can call anything.
func TestAKeyOnEverythingSaysSoRatherThanCountingRules(t *testing.T) {
	assert := td.Assert(t)

	whole := credential{Rules: []credentialRule{
		{Method: "GET", Path: "/*"}, {Method: "POST", Path: "/*"},
		{Method: "PUT", Path: "/*"}, {Method: "DELETE", Path: "/*"},
	}}
	assert.Cmp(whole.scope(), "whole API: DELETE,GET,POST,PUT")

	star := credential{Rules: []credentialRule{{Method: "GET", Path: "*"}}}
	assert.Cmp(star.scope(), "whole API: GET", `"*" reaches as far as "/*"`)
}

func TestANarrowKeyIsCountedByItsPaths(t *testing.T) {
	assert := td.Assert(t)

	narrow := credential{Rules: []credentialRule{
		{Method: "GET", Path: "/me"}, {Method: "GET", Path: "/dedicated/server"},
	}}
	assert.Cmp(narrow.scope(), "2 path(s)")
}

// A key that is broad on one verb and narrow elsewhere is broad. Reporting the
// path count would hide the part that matters.
func TestAMixedKeyLeadsWithWhatItCanReach(t *testing.T) {
	assert := td.Assert(t)

	mixed := credential{Rules: []credentialRule{
		{Method: "GET", Path: "/*"}, {Method: "POST", Path: "/me/api/credential"},
	}}
	assert.Cmp(mixed.scope(), "whole API: GET, and 1 more")
}

func TestAKeyWithNoRuleReachesNothing(t *testing.T) {
	td.Assert(t).Cmp(credential{}.scope(), "nothing")
}

// 64 of the 66 keys measured work from anywhere, which is the answer, not a
// blank cell.
func TestAnUnrestrictedKeySaysAnywhere(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(credential{}.restriction(), "anywhere")
	assert.Cmp(credential{AllowedIPs: []string{"203.0.113.7/32"}}.restriction(), "203.0.113.7/32")
}

// null means two different things on these fields, and neither is "unknown".
func TestANullTimestampIsSaidInItsOwnTerms(t *testing.T) {
	assert := td.Assert(t)
	used := "2024-12-19T17:13:01+01:00"

	assert.Cmp(never(nil, "never used"), "never used")
	assert.Cmp(never(nil, "never expires"), "never expires")
	assert.Cmp(never(&used, "never used"), used)
}
