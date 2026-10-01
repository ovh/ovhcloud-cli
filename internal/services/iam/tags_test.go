// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestTagsAreReadAsKeyEqualsValue(t *testing.T) {
	assert := td.Assert(t)

	got, err := parseTagAssignments([]string{"env=prod", "team=infra"})

	assert.CmpNoError(err)
	assert.Cmp(got, []tagAssignment{{Key: "env", Value: "prod"}, {Key: "team", Value: "infra"}})
}

// Only the first "=" separates, so a value may contain one.
func TestAValueMayContainAnEqualsSign(t *testing.T) {
	assert := td.Assert(t)

	got, err := parseTagAssignments([]string{"filter=a=b"})

	assert.CmpNoError(err)
	assert.Cmp(got, []tagAssignment{{Key: "filter", Value: "a=b"}})
}

// An empty value is legal and is not a removal: the API stores {"key": ""}.
// That is exactly the trap `iam resource edit --tag only=` falls into.
func TestAnEmptyValueIsATagNotARemoval(t *testing.T) {
	assert := td.Assert(t)

	got, err := parseTagAssignments([]string{"only="})

	assert.CmpNoError(err)
	assert.Cmp(got, []tagAssignment{{Key: "only", Value: ""}})
}

func TestSomethingThatIsNotAPairIsRefused(t *testing.T) {
	assert := td.Assert(t)

	_, err := parseTagAssignments([]string{"env"})

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("is not a key=value pair"))
}

func TestAPairWithNoKeyIsRefused(t *testing.T) {
	assert := td.Assert(t)

	_, err := parseTagAssignments([]string{"=prod"})

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("has no key"))
}

// One POST per tag, so a key given twice would be two calls and the second
// would win without a word.
func TestTheSameKeyTwiceIsRefusedRatherThanSilentlyOverwritten(t *testing.T) {
	assert := td.Assert(t)

	_, err := parseTagAssignments([]string{"env=prod", "env=staging"})

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("given twice"))
}

func TestNoTagAtAllIsRefused(t *testing.T) {
	assert := td.Assert(t)

	_, err := parseTagAssignments(nil)

	assert.CmpError(err)
	assert.Cmp(err.Error(), td.Contains("no tag given"))
}

// The mistake behind a missing key is a typo, so the answer is what is there.
func TestTheKeysInPlaceAreNamedWhenOneIsMissing(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(knownKeys(map[string]any{"team": "infra", "env": "prod"}), "env, team")
	assert.Cmp(knownKeys(map[string]any{}), "no tag at all")
	assert.Cmp(knownKeys(nil), "no tag at all")
}
