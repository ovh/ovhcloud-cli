// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package baremetal

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// Powering two servers off and one back on must not make the CLI forget where
// the other one boots: that memory is the only thing standing between the
// second server and a boot entry it was never on.
func TestForgettingOneServerKeepsTheOther(t *testing.T) {
	assert := td.Assert(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	rememberBoot("ns1.example.net", 230242)
	rememberBoot("ns2.example.net", 1234)

	forgetBoot("ns1.example.net")

	_, found := recallBoot("ns1.example.net")
	assert.False(found, "the server that was powered back on is forgotten")

	boot, found := recallBoot("ns2.example.net")
	assert.True(found, "the server still powered off is still remembered")
	assert.Cmp(boot, 1234)
}
