// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"fmt"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// Repeating "urn:v1:eu:resource:" on every row buys nothing.
func TestAUrnIsShownByThePartThatIdentifiesIt(t *testing.T) {
	assert := td.Assert(t)

	assert.Cmp(shortURN("urn:v1:eu:resource:dedicatedServer:ns1.example"),
		"dedicatedServer:ns1.example")
	assert.Cmp(shortURN("not-a-urn"), "not-a-urn", "anything unexpected is left alone")
}

// The API validates the action names before the permissions, and one unknown
// name fails the whole batch — so a typo in the fourth action reads as if none
// of them could be answered.
func TestAnUnknownActionExplainsWhyTheWholeCheckFailed(t *testing.T) {
	assert := td.Assert(t)

	message := explainCheckFailure(
		fmt.Errorf(`Client::BadRequest: "Unknown action \"dedicatedServer:apiovh:nope\""`),
		[]string{"dedicatedServer:apiovh:get", "dedicatedServer:apiovh:nope", "dedicatedServer:apiovh:reboot"})

	assert.Cmp(message, td.Contains("One unknown action fails the whole check"))
	assert.Cmp(message, td.Contains("the other 2 were not answered"))
	assert.Cmp(message, td.Contains("iam reference actions"))
}

// A permission failure is not an unknown action, and must not be dressed up
// as one.
func TestAnyOtherFailureIsReportedAsItself(t *testing.T) {
	assert := td.Assert(t)

	message := explainCheckFailure(fmt.Errorf("connection refused"), []string{"a"})

	assert.Cmp(message, "failed to check authorization: connection refused")
	assert.Cmp(message, td.Not(td.Contains("unknown action")))
}

func TestCategoriesAreMatchedWithoutCaringAboutCase(t *testing.T) {
	assert := td.Assert(t)
	action := referenceAction{Categories: []string{"READ", "OPERATE"}}

	assert.Cmp(hasCategory(action, "read"), true)
	assert.Cmp(hasCategory(action, "READ"), true)
	assert.Cmp(hasCategory(action, "DELETE"), false)
	assert.Cmp(hasCategory(referenceAction{}, "READ"), false)
}

// The description is searched too: an operator looking for "reinstall" should
// find the action whose name spells it differently.
func TestSearchLooksAtTheDescriptionAsWellAsTheName(t *testing.T) {
	assert := td.Assert(t)
	action := referenceAction{
		Action:      "dedicatedServer:apiovh:reinstall",
		Description: "Start the installation of a server",
	}

	assert.Cmp(matchesSearch(action, "reinstall"), true)
	assert.Cmp(matchesSearch(action, "INSTALLATION"), true)
	assert.Cmp(matchesSearch(action, "vrack"), false)
}
