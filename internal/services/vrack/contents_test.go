// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package vrack

import (
	"slices"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
)

// interfaceType is read by the servers section, which resolves interfaces to
// their machine and its display name. It is the one value of the enum that must
// not appear in contentTypes, because listing it twice would print the same
// attachments once as servers and once as raw UUIDs.
const interfaceType = "dedicatedServerInterface"

// The listing of vRack contents is a list of API paths retyped into Go, and such
// a list stops being true in silence. It already had: `dedicatedServer` — the
// legacy attachment, the only path under which a vRack bound that way shows up
// at all — and `dedicatedCloudDatacenter` were both missing, so a vRack whose
// only content was one of them printed "This vRack is empty."
//
// Zero of the 71 vRacks on the account measured use the legacy attachment, which
// is why the omission was invisible rather than harmless. This test is what makes
// the next one visible: it fails when the schema grows a value the listing does
// not read.
func TestOtherContentsCoversEveryAttachableTypeTheSchemaDeclares(t *testing.T) {
	assert := td.NewT(t)

	declared, err := openapi.GetComponentEnum(assets.VrackOpenapiSchema, "vrack.AllowedServiceEnum")
	td.Require(t).CmpNoError(err, "the enum has to be readable for this test to mean anything")
	assert.Cmp(len(declared) > 1, true, "positive control: the enum is not empty")

	covered := make([]string, 0, len(contentTypes))
	for _, ct := range contentTypes {
		covered = append(covered, ct.path)
	}

	for _, want := range declared {
		if want == interfaceType {
			assert.Cmp(slices.Contains(covered, want), false,
				"%s is read by the servers section and must not be listed twice", want)
			continue
		}
		assert.Cmp(slices.Contains(covered, want), true,
			"the schema declares %q as attachable, so `vrack get` has to list it", want)
	}
}

// And nothing beyond it: a path that is not in the enum is a request that will
// answer with an error on every vRack, and errors are counted as "could not be
// read" — so an invented path would permanently stop the summary from ever
// saying a vRack is empty.
func TestOtherContentsListsNoTypeTheSchemaDoesNotDeclare(t *testing.T) {
	assert := td.NewT(t)

	declared, err := openapi.GetComponentEnum(assets.VrackOpenapiSchema, "vrack.AllowedServiceEnum")
	td.Require(t).CmpNoError(err)

	for _, ct := range contentTypes {
		assert.Cmp(slices.Contains(declared, ct.path), true,
			"%q is listed but the schema does not declare it attachable", ct.path)
	}
}
