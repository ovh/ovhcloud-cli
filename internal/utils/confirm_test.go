// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package utils

import (
	"strings"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

// withInteractiveInput runs f as if the operator were typing answer at an
// interactive terminal.
func withInteractiveInput(t *testing.T, answer string, f func()) {
	t.Helper()

	origInput, origInteractive := confirmInput, confirmInteractive
	confirmInput = strings.NewReader(answer)
	confirmInteractive = func() bool { return true }
	defer func() { confirmInput, confirmInteractive = origInput, origInteractive }()

	f()
}

// The exact name is the guard between a typo and an erased disk, so the
// accepting path has to be exercised, not only the refusals around it.
func TestConfirmByName_AcceptsTheExactName(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "ns3168421.ip-51-77-12.eu\n", func() {
		confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, confirmed, true)
}

// Surrounding whitespace is the operator's, not the server's.
func TestConfirmByName_AcceptsTheNameWithSurroundingSpaces(t *testing.T) {
	var confirmed bool
	withInteractiveInput(t, "  ns3168421.ip-51-77-12.eu  \n", func() {
		confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
	})

	td.Cmp(t, confirmed, true)
}

// Anything else is a refusal: a near miss must not pass.
func TestConfirmByName_RejectsAnythingElse(t *testing.T) {
	for _, answer := range []string{
		"ns3168421.ip-51-77-12.e\n",  // one character short
		"ns3168422.ip-51-77-12.eu\n", // the neighbouring server
		"yes\n",
		"\n",
		"", // stream closed before any answer
	} {
		var confirmed bool
		withInteractiveInput(t, answer, func() {
			confirmed = ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks")
		})

		td.Cmp(t, confirmed, false, "answer %q must not confirm", answer)
	}
}

// Without a terminal there is no prompt at all: an unattended run has to opt
// in with --yes rather than be asked a question nobody will read.
func TestConfirmByName_RefusesWhenNotInteractive(t *testing.T) {
	origInput, origInteractive := confirmInput, confirmInteractive
	confirmInput = strings.NewReader("ns3168421.ip-51-77-12.eu\n")
	confirmInteractive = func() bool { return false }
	defer func() { confirmInput, confirmInteractive = origInput, origInteractive }()

	td.Cmp(t, ConfirmByName("ns3168421.ip-51-77-12.eu", "this wipes both disks"), false)
}
